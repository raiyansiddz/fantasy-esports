package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/pkg/websocket"
	"fantasy-esports-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	wsocket "github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type RealTimeLeaderboardHandler struct {
	db                *sql.DB
	config            *config.Config
	leaderboardService *services.LeaderboardService
	connectionManager  *websocket.ConnectionManager
	upgrader          wsocket.Upgrader
}

func NewRealTimeLeaderboardHandler(db *sql.DB, cfg *config.Config, leaderboardService *services.LeaderboardService) *RealTimeLeaderboardHandler {
	connectionManager := websocket.NewConnectionManager()
	connectionManager.Start()

	return &RealTimeLeaderboardHandler{
		db:                db,
		config:            cfg,
		leaderboardService: leaderboardService,
		connectionManager:  connectionManager,
		upgrader: wsocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// @Summary Get real-time leaderboard with WebSocket info
// @Description Get leaderboard with real-time update capabilities
// @Tags Real-time Leaderboards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param top_count query int false "Number of top performers to return" default(50)
// @Param include_around_me query bool false "Include rankings around current user" default(true)
// @Success 200 {object} models.LiveLeaderboardResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /leaderboards/real-time/{id} [get]
func (h *RealTimeLeaderboardHandler) GetRealTimeLeaderboard(c *gin.Context) {
	contestID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("user_id")
	
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid contest ID",
			Code:    "INVALID_CONTEST_ID",
		})
		return
	}

	// Parse query parameters
	topCount := 50
	if tc := c.Query("top_count"); tc != "" {
		if parsed, err := strconv.Atoi(tc); err == nil && parsed > 0 {
			topCount = parsed
		}
	}

	includeAroundMe := c.DefaultQuery("include_around_me", "true") == "true"

	// Get cached leaderboard (5-minute cache)
	leaderboard, err := h.leaderboardService.GetCachedLeaderboard(contestID, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get leaderboard: " + err.Error(),
			Code:    "LEADERBOARD_ERROR",
		})
		return
	}

	// Get user's rank if requested
	if includeAroundMe && userID > 0 {
		rank, points, teamID, err := h.leaderboardService.GetUserRankInContest(contestID, userID)
		if err == nil {
			leaderboard.MyRank = rank
			leaderboard.MyPoints = points
			leaderboard.MyTeamID = teamID

			// Get rankings around user
			aroundMe, err := h.leaderboardService.GetRankingsAroundUser(contestID, rank, 5)
			if err == nil {
				leaderboard.AroundMe = aroundMe
			}
		}
	}

	// Limit top performers
	if len(leaderboard.TopPerformers) > topCount {
		leaderboard.TopPerformers = leaderboard.TopPerformers[:topCount]
	}

	// Generate WebSocket endpoint
	wsEndpoint := fmt.Sprintf("/api/v1/leaderboards/ws/contest/%d", contestID)

	response := models.LiveLeaderboardResponse{
		Success:           true,
		ContestID:         contestID,
		Leaderboard:       leaderboard,
		RealTimeEnabled:   true,
		UpdateFrequency:   30, // 30 seconds
		LastUpdateID:      fmt.Sprintf("%d_%d", contestID, time.Now().Unix()),
		WebSocketEndpoint: wsEndpoint,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary WebSocket connection for real-time leaderboard updates
// @Description Connect to real-time leaderboard updates via WebSocket
// @Tags Real-time Leaderboards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contest_id path int true "Contest ID"
// @Router /leaderboards/ws/contest/{contest_id} [get]
func (h *RealTimeLeaderboardHandler) HandleLeaderboardWebSocket(c *gin.Context) {
	contestID, _ := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	userID := c.GetInt64("user_id")

	if contestID == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid contest ID",
			Code:    "INVALID_CONTEST_ID",
		})
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v", err))
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "WebSocket upgrade failed",
			Code:    "WEBSOCKET_ERROR",
		})
		return
	}

	// Create connection wrapper
	connectionID := uuid.New().String()
	leaderboardConn := &websocket.LeaderboardConnection{
		UserID:       userID,
		ContestID:    contestID,
		ConnectionID: connectionID,
		Conn:         conn,
		Send:         make(chan models.RealTimeWebSocketMessage, 256),
		ConnectedAt:  time.Now(),
		LastPing:     time.Now(),
		IsActive:     true,
	}

	// Register connection
	h.connectionManager.RegisterConnection(leaderboardConn)

	// Start goroutines for reading and writing
	go h.handleWebSocketWrite(leaderboardConn)
	go h.handleWebSocketRead(leaderboardConn)
}

func (h *RealTimeLeaderboardHandler) handleWebSocketWrite(conn *websocket.LeaderboardConnection) {
	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer func() {
		ticker.Stop()
		conn.Conn.Close()
		h.connectionManager.UnregisterConnection(conn)
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.Conn.WriteMessage(wsocket.CloseMessage, []byte{})
				return
			}

			if err := conn.Conn.WriteJSON(message); err != nil {
				logger.Error(fmt.Sprintf("WebSocket write error: %v", err))
				return
			}

		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(wsocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *RealTimeLeaderboardHandler) handleWebSocketRead(conn *websocket.LeaderboardConnection) {
	defer func() {
		h.connectionManager.UnregisterConnection(conn)
		conn.Conn.Close()
	}()

	conn.Conn.SetReadLimit(512)
	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		h.connectionManager.UpdateConnectionPing(conn.ConnectionID)
		return nil
	})

	for {
		var message models.RealTimeWebSocketMessage
		err := conn.Conn.ReadJSON(&message)
		if err != nil {
			if wsocket.IsUnexpectedCloseError(err, wsocket.CloseGoingAway, wsocket.CloseAbnormalClosure) {
				logger.Error(fmt.Sprintf("WebSocket error: %v", err))
			}
			break
		}

		// Handle incoming messages
		h.handleIncomingMessage(conn, message)
	}
}

func (h *RealTimeLeaderboardHandler) handleIncomingMessage(conn *websocket.LeaderboardConnection, message models.RealTimeWebSocketMessage) {
	switch message.Type {
	case "ping":
		// Respond with pong
		pongMsg := models.RealTimeWebSocketMessage{
			Type:      "pong",
			ContestID: conn.ContestID,
			Data:      gin.H{"status": "alive"},
			Timestamp: time.Now(),
			MessageID: uuid.New().String(),
		}
		select {
		case conn.Send <- pongMsg:
		default:
		}

	case "subscribe":
		// Send current leaderboard status
		h.sendCurrentLeaderboardStatus(conn)

	case "request_update":
		// Force a leaderboard update
		err := h.leaderboardService.TriggerRealTimeUpdate(conn.ContestID, "user_request", nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to trigger update: %v", err))
		}

	default:
		// Unknown message type
		errorMsg := models.RealTimeWebSocketMessage{
			Type:      "error",
			ContestID: conn.ContestID,
			Data:      gin.H{"error": "Unknown message type", "received_type": message.Type},
			Timestamp: time.Now(),
			MessageID: uuid.New().String(),
		}
		select {
		case conn.Send <- errorMsg:
		default:
		}
	}
}

func (h *RealTimeLeaderboardHandler) sendCurrentLeaderboardStatus(conn *websocket.LeaderboardConnection) {
	leaderboard, err := h.leaderboardService.GetLiveLeaderboard(conn.ContestID, conn.UserID)
	if err != nil {
		errorMsg := models.RealTimeWebSocketMessage{
			Type:      "error",
			ContestID: conn.ContestID,
			Data:      gin.H{"error": "Failed to get leaderboard", "details": err.Error()},
			Timestamp: time.Now(),
			MessageID: uuid.New().String(),
		}
		select {
		case conn.Send <- errorMsg:
		default:
		}
		return
	}

	statusMsg := models.RealTimeWebSocketMessage{
		Type:      "leaderboard_status",
		ContestID: conn.ContestID,
		Data: models.LeaderboardConnectionStatus{
			Connected:         true,
			ContestID:         conn.ContestID,
			UserID:            conn.UserID,
			MyCurrentRank:     leaderboard.MyRank,
			MyCurrentPoints:   leaderboard.MyPoints,
			TotalParticipants: leaderboard.TotalParticipants,
			LastUpdated:       time.Now(),
		},
		Timestamp: time.Now(),
		MessageID: uuid.New().String(),
	}

	select {
	case conn.Send <- statusMsg:
	default:
	}
}

// @Summary Get active connections for contest
// @Description Get number of active WebSocket connections for a contest
// @Tags Real-time Leaderboards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contest_id path int true "Contest ID"
// @Success 200 {object} map[string]interface{}
// @Router /leaderboards/connections/{contest_id} [get]
func (h *RealTimeLeaderboardHandler) GetActiveConnections(c *gin.Context) {
	contestID, _ := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid contest ID",
			Code:    "INVALID_CONTEST_ID",
		})
		return
	}

	connectionCount := h.connectionManager.GetContestConnectionCount(contestID)

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"contest_id":         contestID,
		"active_connections": connectionCount,
		"real_time_enabled":  true,
		"timestamp":          time.Now(),
	})
}

// @Summary Trigger manual leaderboard update
// @Description Manually trigger a real-time leaderboard update for testing
// @Tags Real-time Leaderboards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contest_id path int true "Contest ID"
// @Success 200 {object} map[string]interface{}
// @Router /leaderboards/trigger-update/{contest_id} [post]
func (h *RealTimeLeaderboardHandler) TriggerManualUpdate(c *gin.Context) {
	contestID := parseIntToInt64(c.Param("contest_id"))
	
	if contestID == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid contest ID",
			Code:    "INVALID_CONTEST_ID",
		})
		return
	}

	err := h.leaderboardService.TriggerRealTimeUpdate(contestID, "manual_trigger", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to trigger update: " + err.Error(),
			Code:    "UPDATE_TRIGGER_ERROR",
		})
		return
	}

	connectionCount := h.connectionManager.GetContestConnectionCount(contestID)

	c.JSON(http.StatusOK, gin.H{
		"success":               true,
		"contest_id":            contestID,
		"update_triggered":      true,
		"active_connections":    connectionCount,
		"trigger_source":        "manual_trigger",
		"timestamp":             time.Now(),
	})
}

// GetConnectionManager returns the connection manager for integration with other services
func (h *RealTimeLeaderboardHandler) GetConnectionManager() *websocket.ConnectionManager {
	return h.connectionManager
}

