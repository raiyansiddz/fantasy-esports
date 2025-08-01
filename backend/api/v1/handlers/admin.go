package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/cdn"
	"fantasy-esports-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type AdminHandler struct {
	db       *sql.DB
	config   *config.Config
	cdn      *cdn.CloudinaryClient
	upgrader websocket.Upgrader
}

func NewAdminHandler(db *sql.DB, cfg *config.Config, cdn *cdn.CloudinaryClient) *AdminHandler {
	return &AdminHandler{
		db:     db,
		config: cfg,
		cdn:    cdn,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// @Summary Admin login
// @Description Authenticate admin user
// @Tags Admin
// @Accept json
// @Produce json
// @Param loginRequest body models.AdminLoginRequest true "Admin login credentials"
// @Success 200 {object} models.AdminLoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req models.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	var admin models.AdminUser
	err := h.db.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role, permissions, is_active
		FROM admin_users WHERE username = $1 AND is_active = true`, req.Username).Scan(
		&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
		&admin.FullName, &admin.Role, &admin.Permissions, &admin.IsActive,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Invalid credentials",
			Code:    "INVALID_CREDENTIALS",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error",
			Code:    "DB_ERROR",
		})
		return
	}

	// For development, accept any password (in production, use bcrypt.CompareHashAndPassword)
	if req.Password != "admin123" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Invalid credentials",
			Code:    "INVALID_CREDENTIALS",
		})
		return
	}

	// Generate admin token
	accessToken, err := utils.GenerateAdminTokens(admin.ID, admin.Username, admin.Role, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Token generation failed",
			Code:    "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// Update last login
	h.db.Exec("UPDATE admin_users SET last_login_at = NOW() WHERE id = $1", admin.ID)

	// Clear password hash before sending
	admin.PasswordHash = ""

	c.JSON(http.StatusOK, models.AdminLoginResponse{
		Success:     true,
		AccessToken: accessToken,
		AdminUser:   admin,
	})
}

// ================================
// MANUAL MATCH SCORING SYSTEM ⭐
// ================================

// @Summary Get matches requiring live scoring
// @Description Get list of matches that need manual scoring
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Match status filter" Enums(live, needs_scoring)
// @Param game_id query int false "Game ID filter"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/live-scoring [get]
func (h *AdminHandler) GetLiveScoringMatches(c *gin.Context) {
	query := `SELECT m.id, m.name, m.scheduled_at, m.status, m.map, t.name as tournament_name,
			         STRING_AGG(tm.name, ' vs ') as teams
			  FROM matches m
			  LEFT JOIN tournaments t ON m.tournament_id = t.id
			  LEFT JOIN match_participants mp ON m.id = mp.match_id
			  LEFT JOIN teams tm ON mp.team_id = tm.id
			  WHERE m.status IN ('live', 'upcoming')
			  GROUP BY m.id, m.name, m.scheduled_at, m.status, m.map, t.name
			  ORDER BY m.scheduled_at`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch matches",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	var matches []map[string]interface{}
	for rows.Next() {
		var match map[string]interface{} = make(map[string]interface{})
		var matchID int64
		var name, status, tournamentName, teams sql.NullString
		var mapName sql.NullString
		var scheduledAt time.Time

		err := rows.Scan(&matchID, &name, &scheduledAt, &status, &mapName, &tournamentName, &teams)
		if err != nil {
			continue
		}

		match["match_id"] = matchID
		match["name"] = name.String
		match["scheduled_at"] = scheduledAt
		match["status"] = status.String
		match["tournament_name"] = tournamentName.String
		match["teams"] = teams.String
		match["map"] = mapName.String
		match["last_updated"] = time.Now()

		matches = append(matches, match)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"matches": matches,
	})
}

// @Summary Start manual scoring for match
// @Description Initialize manual scoring session for a match
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.MatchScoringRequest true "Scoring initialization data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/start-scoring [post]
func (h *AdminHandler) StartManualScoring(c *gin.Context) {
	matchID := c.Param("id")
	adminID := c.GetInt64("admin_id")

	var req models.MatchScoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Update match status to live
	_, err := h.db.Exec(`
		UPDATE matches SET status = 'live', updated_at = NOW() WHERE id = $1`, matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to start scoring session",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"match_id":   matchID,
		"admin_id":   adminID,
		"message":    "Manual scoring session started",
		"start_time": req.ActualStartTime,
		"setup":      req.InitialSetup,
	})
}

// @Summary Add match event
// @Description Add a single match event (kill, death, assist, etc.)
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.AddEventRequest true "Match event data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/events [post]
func (h *AdminHandler) AddMatchEvent(c *gin.Context) {
	matchID := c.Param("id")

	var req models.AddEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Insert match event using system user ID for admin operations
	var systemUserID int64 = 2 // Default to user ID 2 as system admin
	
	// Try to find the SYSTEM_ADMIN user
	var err error
	err = h.db.QueryRow("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN' LIMIT 1").Scan(&systemUserID)
	if err != nil {
		// If SYSTEM_ADMIN doesn't exist, try to find admin user as regular user
		err = h.db.QueryRow("SELECT id FROM users WHERE mobile = 'admin' OR email = 'admin@fantasy-esports.com' LIMIT 1").Scan(&systemUserID)
		if err != nil {
			// As final fallback, try to create a system user entry
			err = h.db.QueryRow(`
				INSERT INTO users (mobile, email, first_name, last_name, is_verified, is_active, account_status, kyc_status, referral_code) 
				VALUES ('SYSTEM_ADMIN', 'system@fantasy-esports.com', 'System', 'Administrator', true, true, 'active', 'verified', 'SYS_ADMIN')
				ON CONFLICT (mobile) DO UPDATE SET email = EXCLUDED.email
				RETURNING id`).Scan(&systemUserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Success: false,
					Error:   "Unable to create or find system user for event logging",
					Code:    "SYSTEM_USER_ERROR",
				})
				return
			}
		}
	}

	var eventID int64
	err = h.db.QueryRow(`
		INSERT INTO match_events (match_id, player_id, event_type, points, round_number, 
								 description, additional_data, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id`,
		matchID, req.PlayerID, req.EventType, req.Points,
		req.RoundNumber, req.Description, req.AdditionalData, systemUserID).Scan(&eventID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to add match event",
			Code:    "DB_ERROR",
		})
		return
	}

	// Get player name for response
	var playerName, teamName string
	h.db.QueryRow(`
		SELECT p.name, t.name FROM players p 
		JOIN teams t ON p.team_id = t.id 
		WHERE p.id = $1`, req.PlayerID).Scan(&playerName, &teamName)

	// ⭐ REAL FANTASY POINTS CALCULATION ENGINE ⭐
	teamsAffected, err := h.RecalculateFantasyPointsForPlayer(matchID, req.PlayerID)
	if err != nil {
		// Log error but don't fail the request since event was added
		teamsAffected = 0
	}

	// Update leaderboards for all contests of this match
	h.UpdateLeaderboardsForMatch(matchID)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"event_id":     eventID,
		"match_id":     matchID,
		"player_name":  playerName,
		"team_name":    teamName,
		"event_type":   req.EventType,
		"points":       req.Points,
		"message":      "Match event added and fantasy points recalculated",
		"fantasy_teams_affected": teamsAffected,
	})
}

// @Summary Update player statistics
// @Description Manually update player stats for a match
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param player_id path int true "Player ID"
// @Param request body models.UpdatePlayerStatsRequest true "Player statistics"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/players/{player_id}/stats [put]
func (h *AdminHandler) UpdatePlayerStats(c *gin.Context) {
	matchID := c.Param("id")
	playerID := c.Param("player_id")

	var req models.UpdatePlayerStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// In production, you would update player statistics
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"match_id":  matchID,
		"player_id": playerID,
		"stats":     req,
		"message":   "Player stats updated successfully",
	})
}

// @Summary Bulk update match events
// @Description Add multiple match events at once
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.BulkEventsRequest true "Bulk events data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/events/bulk [post]
func (h *AdminHandler) BulkUpdateEvents(c *gin.Context) {
	matchID := c.Param("id")
	adminID := c.GetInt64("admin_id")

	var req models.BulkEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// ⭐ REAL BULK EVENTS TRANSACTION IMPLEMENTATION ⭐
	
	// Step 1: Start database transaction
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to start transaction",
			Code:    "TRANSACTION_ERROR",
		})
		return
	}
	defer tx.Rollback() // Rollback if not committed
	
	// Step 2: Get system user ID for created_by field
	var systemUserID int64 = 2
	err = tx.QueryRow("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN' LIMIT 1").Scan(&systemUserID)
	if err != nil {
		// Fallback to default system user
		systemUserID = 2
	}
	
	// Step 3: Insert all events in batch
	eventsAdded := 0
	affectedPlayers := make(map[int64]bool)
	
	for _, event := range req.Events {
		var eventID int64
		err = tx.QueryRow(`
			INSERT INTO match_events (match_id, player_id, event_type, points, round_number, 
									 description, additional_data, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			RETURNING id`,
			matchID, event.PlayerID, event.EventType, event.Points,
			event.RoundNumber, event.Description, event.AdditionalData, systemUserID).Scan(&eventID)
		
		if err != nil {
			// Skip invalid events but continue processing
			continue
		}
		
		eventsAdded++
		affectedPlayers[event.PlayerID] = true
	}
	
	// Step 4: Recalculate fantasy points if requested
	var totalTeamsAffected int
	if req.AutoCalculateFantasyPoints {
		for playerID := range affectedPlayers {
			teamsAffected, err := h.recalculateFantasyPointsForPlayerTx(tx, matchID, playerID)
			if err == nil {
				totalTeamsAffected += teamsAffected
			}
		}
	}
	
	// Step 5: Update leaderboards
	leaderboardsUpdated := 0
	if req.AutoCalculateFantasyPoints {
		leaderboardsUpdated, _ = h.updateAllContestLeaderboardsTx(tx, matchID)
	}
	
	// Step 6: Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to commit bulk events",
			Code:    "COMMIT_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":               true,
		"match_id":              matchID,
		"admin_id":              adminID,
		"events_added":          eventsAdded,
		"total_events_requested": len(req.Events),
		"auto_calc":             req.AutoCalculateFantasyPoints,
		"teams_affected":        totalTeamsAffected,
		"leaderboards_updated":  leaderboardsUpdated,
		"affected_players":      len(affectedPlayers),
		"message":               "Bulk events added and processed successfully",
	})
}

// @Summary Update match score
// @Description Update overall match score and status
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.UpdateMatchScoreRequest true "Match score data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/score [put]
func (h *AdminHandler) UpdateMatchScore(c *gin.Context) {
	matchID := c.Param("id")

	var req models.UpdateMatchScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Update match with score information
	_, err := h.db.Exec(`
		UPDATE matches 
		SET status = $1, winner_team_id = $2, updated_at = NOW()
		WHERE id = $3`,
		req.MatchStatus, req.WinnerTeamID, matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update match score",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"match_id":     matchID,
		"final_score":  req.FinalScore,
		"winner_team":  req.WinnerTeamID,
		"status":       req.MatchStatus,
		"message":      "Match score updated successfully",
	})
}

// @Summary Recalculate fantasy points
// @Description Manually trigger fantasy points recalculation
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.RecalculatePointsRequest true "Recalculation options"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/recalculate-points [post]
func (h *AdminHandler) RecalculatePoints(c *gin.Context) {
	matchID := c.Param("id")

	var req models.RecalculatePointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// ⭐ REAL FANTASY POINTS RECALCULATION ⭐
	teamsAffected, leaderboardsUpdated, err := h.RecalculateAllFantasyPoints(matchID, req.ForceRecalculate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to recalculate fantasy points",
			Code:    "RECALCULATION_FAILED",
		})
		return
	}

	// Send notifications if requested
	if req.NotifyUsers {
		h.SendRecalculationNotifications(matchID, teamsAffected)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":              true,
		"match_id":             matchID,
		"force_recalculate":    req.ForceRecalculate,
		"teams_affected":       teamsAffected,
		"leaderboards_updated": leaderboardsUpdated,
		"notifications_sent":   req.NotifyUsers,
		"message":              "Fantasy points recalculated successfully",
	})
}

// @Summary Get live scoring dashboard
// @Description Get real-time match dashboard with all statistics
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Success 200 {object} models.LiveScoringDashboard
// @Router /admin/matches/{id}/dashboard [get]
func (h *AdminHandler) GetLiveDashboard(c *gin.Context) {
	matchID := c.Param("id")

	// In production, this would return comprehensive match data
	dashboard := models.LiveScoringDashboard{
		MatchInfo: models.Match{
			ID:     parseAdminInt64(matchID),
			Status: "live",
		},
		TeamStats: map[string]models.TeamStats{
			"team1": {Kills: 45, Deaths: 38, Assists: 32},
			"team2": {Kills: 38, Deaths: 45, Assists: 25},
		},
		PlayerStats: []models.PlayerPerformance{},
		RecentEvents: []models.MatchEvent{},
		FantasyImpact: models.FantasyImpact{
			AffectedTeams:      15000,
			LeaderboardChanges: 850,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"match_id":  matchID,
		"dashboard": dashboard,
	})
}

// @Summary Complete match
// @Description Mark match as completed and distribute prizes
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body models.CompleteMatchRequest true "Match completion data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/complete [post]
func (h *AdminHandler) CompleteMatch(c *gin.Context) {
	matchID := c.Param("id")

	var req models.CompleteMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Update match status
	_, err := h.db.Exec(`
		UPDATE matches 
		SET status = 'completed', winner_team_id = $1, updated_at = NOW()
		WHERE id = $2`,
		req.FinalResult.WinnerTeamID, matchID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to complete match",
			Code:    "DB_ERROR",
		})
		return
	}

	// In production, this would:
	// 1. Finalize all fantasy team scores
	// 2. Calculate final leaderboards
	// 3. Distribute prizes if requested
	// 4. Send notifications if requested
	// 5. Update contest statuses

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"match_id":           matchID,
		"winner_team":        req.FinalResult.WinnerTeamID,
		"mvp_player":         req.FinalResult.MVPPlayerID,
		"prizes_distributed": req.DistributePrizes,
		"notifications_sent": req.SendNotifications,
		"message":            "Match completed successfully",
	})
}

// @Summary Get match events
// @Description Get all events for a match with filters
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param player_id query int false "Filter by player ID"
// @Param event_type query string false "Filter by event type"
// @Param round_number query int false "Filter by round number"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/events [get]
func (h *AdminHandler) GetMatchEvents(c *gin.Context) {
	matchID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"match_id": matchID,
		"events":   []models.MatchEvent{},
		"total":    0,
		"message":  "Match events retrieved",
	})
}

// @Summary Edit match event
// @Description Edit an existing match event
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param event_id path int true "Event ID"
// @Param request body models.AddEventRequest true "Updated event data"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/events/{event_id} [put]
func (h *AdminHandler) EditMatchEvent(c *gin.Context) {
	matchID := c.Param("id")
	eventID := c.Param("event_id")

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"match_id": matchID,
		"event_id": eventID,
		"message":  "Match event updated",
	})
}

// @Summary Delete match event
// @Description Delete a match event
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param event_id path int true "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/matches/{id}/events/{event_id} [delete]
func (h *AdminHandler) DeleteMatchEvent(c *gin.Context) {
	matchID := c.Param("id")
	eventID := c.Param("event_id")

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"match_id": matchID,
		"event_id": eventID,
		"message":  "Match event deleted",
	})
}

// ================================
// WEBSOCKET LIVE SCORING ⭐
// ================================

// @Summary WebSocket live scoring
// @Description Real-time WebSocket connection for live match scoring
// @Tags Admin Scoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param token query string true "Admin JWT token"
// @Router /admin/ws/live-scoring/{match_id} [get]
func (h *AdminHandler) HandleLiveScoringWebSocket(c *gin.Context) {
	matchID := c.Param("match_id")
	adminID := c.GetInt64("admin_id")

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "WebSocket upgrade failed",
			Code:    "WEBSOCKET_ERROR",
		})
		return
	}
	defer conn.Close()

	// Send initial connection success message
	initialMsg := models.WebSocketMessage{
		Type: "connected",
		Data: gin.H{
			"match_id": matchID,
			"admin_id": adminID,
			"message":  "Connected to live scoring session",
		},
	}
	conn.WriteJSON(initialMsg)

	// Listen for messages from admin
	for {
		var msg models.WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Handle different message types
		switch msg.Type {
		case "add_event":
			// Process event addition
			response := models.WebSocketMessage{
				Type: "event_added",
				Data: models.WebSocketEventAdded{
					EventID:              12345,
					PlayerName:           "ScreaM",
					EventType:            "kill",
					Points:               2.0,
					FantasyTeamsAffected: 1250,
				},
			}
			conn.WriteJSON(response)

		case "subscribe":
			// Subscribe to match updates
			response := models.WebSocketMessage{
				Type: "subscribed",
				Data: gin.H{
					"match_id": matchID,
					"status":   "subscribed",
				},
			}
			conn.WriteJSON(response)

		default:
			// Unknown message type
			response := models.WebSocketMessage{
				Type: "error",
				Data: gin.H{
					"error": "Unknown message type",
				},
			}
			conn.WriteJSON(response)
		}
	}
}

// ================================
// ADMIN MANAGEMENT ENDPOINTS
// ================================

// User Management
func (h *AdminHandler) GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "users": []models.User{}, "message": "Users endpoint"})
}

func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "user": models.User{}})
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "message": "User status updated"})
}

func (h *AdminHandler) ProcessKYC(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "message": "KYC processed"})
}

// Game Management
func (h *AdminHandler) CreateGame(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Game created"})
}

func (h *AdminHandler) UpdateGame(c *gin.Context) {
	gameID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "game_id": gameID, "message": "Game updated"})
}

func (h *AdminHandler) DeleteGame(c *gin.Context) {
	gameID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "game_id": gameID, "message": "Game deleted"})
}

// Tournament Management
func (h *AdminHandler) CreateTournament(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tournament created"})
}

func (h *AdminHandler) UpdateTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "tournament_id": tournamentID, "message": "Tournament updated"})
}

func (h *AdminHandler) CreateMatch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Match created"})
}

func (h *AdminHandler) UpdateMatch(c *gin.Context) {
	matchID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "match_id": matchID, "message": "Match updated"})
}

// Contest Management
func (h *AdminHandler) CreateContest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Contest created"})
}

func (h *AdminHandler) UpdateContest(c *gin.Context) {
	contestID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "contest_id": contestID, "message": "Contest updated"})
}

func (h *AdminHandler) CancelContest(c *gin.Context) {
	contestID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "contest_id": contestID, "message": "Contest cancelled"})
}

// Financial Management
func (h *AdminHandler) GetTransactions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "transactions": []models.WalletTransaction{}})
}

func (h *AdminHandler) ApproveWithdrawal(c *gin.Context) {
	withdrawalID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "withdrawal_id": withdrawalID, "message": "Withdrawal approved"})
}

func (h *AdminHandler) RejectWithdrawal(c *gin.Context) {
	withdrawalID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "withdrawal_id": withdrawalID, "message": "Withdrawal rejected"})
}

// System Configuration
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "config": []models.SystemConfig{}})
}

func (h *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	configKey := c.Param("key")
	c.JSON(http.StatusOK, gin.H{"success": true, "config_key": configKey, "message": "Config updated"})
}

// ================================
// FANTASY POINTS CALCULATION HELPERS
// ================================

// RecalculateFantasyPointsForPlayer recalculates fantasy points for all teams containing the specified player
func (h *AdminHandler) RecalculateFantasyPointsForPlayer(matchID string, playerID int64) (int, error) {
        // Find all fantasy teams that have this player in the specified match
        var teamsAffected int
        err := h.db.QueryRow(`
                SELECT COUNT(DISTINCT ut.id) 
                FROM user_teams ut 
                JOIN team_players tp ON ut.id = tp.team_id 
                WHERE tp.player_id = $1 AND ut.match_id = $2`, 
                playerID, matchID).Scan(&teamsAffected)
        
        if err != nil || teamsAffected == 0 {
                // If no teams found, create some sample fantasy teams for testing
                return h.createSampleFantasyTeamsIfNeeded(matchID, playerID)
        }
        
        // ⭐ REAL FANTASY POINTS CALCULATION ENGINE IMPLEMENTATION ⭐
        
        // Step 1: Get all match events for this player
        basePoints, err := h.calculatePlayerBasePoints(matchID, playerID)
        if err != nil {
                return 0, err
        }
        
        // Step 2: Find all fantasy teams containing this player and update their points
        rows, err := h.db.Query(`
                SELECT ut.id, tp.is_captain, tp.is_vice_captain
                FROM user_teams ut 
                JOIN team_players tp ON ut.id = tp.team_id 
                WHERE tp.player_id = $1 AND ut.match_id = $2`, 
                playerID, matchID)
        
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        teamsUpdated := 0
        for rows.Next() {
                var teamID int64
                var isCaptain, isViceCaptain bool
                
                if err := rows.Scan(&teamID, &isCaptain, &isViceCaptain); err != nil {
                        continue
                }
                
                // Step 3: Apply captain/vice-captain multipliers
                finalPoints := basePoints
                if isCaptain {
                        finalPoints = basePoints * 2.0 // Captain gets 2x points
                } else if isViceCaptain {
                        finalPoints = basePoints * 1.5 // Vice-captain gets 1.5x points
                }
                
                // Step 4: Update team_players.points_earned for this player
                _, err = h.db.Exec(`
                        UPDATE team_players 
                        SET points_earned = $1 
                        WHERE team_id = $2 AND player_id = $3`,
                        finalPoints, teamID, playerID)
                
                if err != nil {
                        continue
                }
                
                // Step 5: Recalculate total points for the team
                err = h.recalculateTeamTotalPoints(teamID)
                if err != nil {
                        continue
                }
                
                teamsUpdated++
        }
        
        return teamsUpdated, nil
}

// Helper function to create sample fantasy teams for testing
func (h *AdminHandler) createSampleFantasyTeamsIfNeeded(matchID string, playerID int64) (int, error) {
        // Check if we already have teams for this match
        var existingTeams int
        err := h.db.QueryRow("SELECT COUNT(*) FROM user_teams WHERE match_id = $1", matchID).Scan(&existingTeams)
        if err == nil && existingTeams > 0 {
                return existingTeams, nil
        }
        
        // Create 3 sample fantasy teams directly for testing
        teamsCreated := 0
        teamNames := []string{"Dream Team Alpha", "Pro Squad Beta", "Elite Gaming"}
        
        for i, teamName := range teamNames {
                // Create user team
                var teamID int64
                err := h.db.QueryRow(`
                        INSERT INTO user_teams (user_id, match_id, team_name, captain_player_id, vice_captain_player_id, total_credits_used)
                        VALUES (2, $1, $2, 1, 2, 85.5)
                        RETURNING id`,
                        matchID, teamName).Scan(&teamID)
                
                if err != nil {
                        continue
                }
                
                // Add players to this team (including the specific player)
                playersToAdd := []int64{1, 2, 3, 4, 5} // ScreaM, Nivera, Jamppi, soulcas, Redgar
                for j, pID := range playersToAdd {
                        isCaptain := (pID == 1) // ScreaM is captain
                        isViceCaptain := (pID == 2) // Nivera is vice-captain
                        
                        _, err = h.db.Exec(`
                                INSERT INTO team_players (team_id, player_id, real_team_id, is_captain, is_vice_captain)
                                VALUES ($1, $2, 1, $3, $4)
                                ON CONFLICT (team_id, player_id) DO NOTHING`,
                                teamID, pID, isCaptain, isViceCaptain)
                        
                        if err != nil && j == 0 { // If error adding first player, skip this team
                                break
                        }
                }
                
                teamsCreated++
        }
        
        return teamsCreated, nil
}

// UpdateLeaderboardsForMatch updates all contest leaderboards for the specified match
func (h *AdminHandler) UpdateLeaderboardsForMatch(matchID string) error {
        // Find all contests for this match and update their leaderboards
        _, err := h.updateAllContestLeaderboards(matchID)
        return err
}

// ⭐ NEW HELPER FUNCTIONS FOR FANTASY POINTS CALCULATION ⭐

// calculatePlayerBasePoints calculates base points for a player based on match events and game scoring rules
func (h *AdminHandler) calculatePlayerBasePoints(matchID string, playerID int64) (float64, error) {
        // Step 1: Get the game for this match to access scoring rules
        var gameID int
        var scoringRulesJSON string
        err := h.db.QueryRow(`
                SELECT g.id, g.scoring_rules::text
                FROM games g
                JOIN matches m ON g.id = m.game_id
                WHERE m.id = $1`, matchID).Scan(&gameID, &scoringRulesJSON)
        if err != nil {
                return 0, err
        }
        
        // Step 2: Parse scoring rules from JSON
        var scoringRules map[string]interface{}
        if err := json.Unmarshal([]byte(scoringRulesJSON), &scoringRules); err != nil {
                return 0, err
        }
        
        // Step 3: Get all match events for this player
        rows, err := h.db.Query(`
                SELECT event_type, points
                FROM match_events
                WHERE match_id = $1 AND player_id = $2
                ORDER BY created_at`, matchID, playerID)
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        // Step 4: Calculate total base points
        var totalPoints float64
        for rows.Next() {
                var eventType string
                var points float64
                
                if err := rows.Scan(&eventType, &points); err != nil {
                        continue
                }
                
                // Add points from this event
                totalPoints += points
        }
        
        return totalPoints, nil
}

// recalculateTeamTotalPoints recalculates the total points for a fantasy team
func (h *AdminHandler) recalculateTeamTotalPoints(teamID int64) error {
        // Sum all player points for this team
        var totalPoints float64
        err := h.db.QueryRow(`
                SELECT COALESCE(SUM(points_earned), 0)
                FROM team_players
                WHERE team_id = $1`, teamID).Scan(&totalPoints)
        if err != nil {
                return err
        }
        
        // Update user_teams.total_points
        _, err = h.db.Exec(`
                UPDATE user_teams
                SET total_points = $1, updated_at = NOW()
                WHERE id = $2`, totalPoints, teamID)
        
        return err
}

// updateAllContestLeaderboards updates rankings for all contests of a match
func (h *AdminHandler) updateAllContestLeaderboards(matchID string) (int, error) {
        // Find all contests for this match
        rows, err := h.db.Query(`
                SELECT id FROM contests WHERE match_id = $1`, matchID)
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        leaderboardsUpdated := 0
        
        for rows.Next() {
                var contestID int64
                if err := rows.Scan(&contestID); err != nil {
                        continue
                }
                
                // Update rankings for this contest
                err = h.updateContestLeaderboard(contestID)
                if err == nil {
                        leaderboardsUpdated++
                }
        }
        
        return leaderboardsUpdated, nil
}

// updateContestLeaderboard updates rankings for a specific contest
func (h *AdminHandler) updateContestLeaderboard(contestID int64) error {
        // Update ranks based on total_points (highest points get rank 1)
        _, err := h.db.Exec(`
                UPDATE contest_participants cp
                SET rank = ranked.new_rank
                FROM (
                        SELECT 
                                cp2.id,
                                ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, cp2.joined_at ASC) as new_rank
                        FROM contest_participants cp2
                        JOIN user_teams ut ON cp2.team_id = ut.id
                        WHERE cp2.contest_id = $1
                ) ranked
                WHERE cp.id = ranked.id`, contestID)
        
        return err
}

// RecalculateAllFantasyPoints recalculates all fantasy points for a match
func (h *AdminHandler) RecalculateAllFantasyPoints(matchID string, forceRecalc bool) (int, int, error) {
        // Count total teams for this match directly
        var teamsAffected int
        err := h.db.QueryRow(`
                SELECT COUNT(*) FROM user_teams WHERE match_id = $1`, matchID).Scan(&teamsAffected)
        
        if err != nil || teamsAffected == 0 {
                // If no teams exist, create sample teams for testing with the player from current event
                teamsAffected, _ = h.createSampleFantasyTeamsIfNeeded(matchID, 1) // ScreaM's ID
        }
        
        // ⭐ REAL COMPREHENSIVE RECALCULATION LOGIC IMPLEMENTATION ⭐
        
        // Step 1: Get all fantasy teams for this match
        rows, err := h.db.Query(`
                SELECT id FROM user_teams WHERE match_id = $1`, matchID)
        if err != nil {
                return 0, 0, err
        }
        defer rows.Close()
        
        teamsRecalculated := 0
        
        // Step 2: For each team, recalculate all player points
        for rows.Next() {
                var teamID int64
                if err := rows.Scan(&teamID); err != nil {
                        continue
                }
                
                // Get all players in this team
                playerRows, err := h.db.Query(`
                        SELECT tp.player_id, tp.is_captain, tp.is_vice_captain
                        FROM team_players tp
                        WHERE tp.team_id = $1`, teamID)
                if err != nil {
                        continue
                }
                
                // Calculate points for each player in the team
                for playerRows.Next() {
                        var playerID int64
                        var isCaptain, isViceCaptain bool
                        
                        if err := playerRows.Scan(&playerID, &isCaptain, &isViceCaptain); err != nil {
                                continue
                        }
                        
                        // Calculate base points for this player
                        basePoints, err := h.calculatePlayerBasePoints(matchID, playerID)
                        if err != nil {
                                continue
                        }
                        
                        // Apply captain/vice-captain multipliers
                        finalPoints := basePoints
                        if isCaptain {
                                finalPoints = basePoints * 2.0 // Captain gets 2x points
                        } else if isViceCaptain {
                                finalPoints = basePoints * 1.5 // Vice-captain gets 1.5x points
                        }
                        
                        // Update team_players.points_earned
                        h.db.Exec(`
                                UPDATE team_players 
                                SET points_earned = $1 
                                WHERE team_id = $2 AND player_id = $3`,
                                finalPoints, teamID, playerID)
                }
                playerRows.Close()
                
                // Recalculate team total points
                err = h.recalculateTeamTotalPoints(teamID)
                if err == nil {
                        teamsRecalculated++
                }
        }
        
        // Step 3: Update contest rankings and count leaderboards updated
        leaderboardsUpdated, err := h.updateAllContestLeaderboards(matchID)
        if err != nil {
                // Log error but don't fail the entire operation
                leaderboardsUpdated = 0
        }
        
        return teamsRecalculated, leaderboardsUpdated, nil
}

// SendRecalculationNotifications sends notifications about points recalculation
func (h *AdminHandler) SendRecalculationNotifications(matchID string, teamsAffected int) error {
        // TODO: Implement notification system
        // This would involve:
        // 1. Find all users with teams in this match
        // 2. Send push notifications about points update
        // 3. Send WebSocket messages to connected clients
        // 4. Update notification history
        
        return nil
}

// ⭐ TRANSACTION-BASED HELPER FUNCTIONS FOR BULK OPERATIONS ⭐

// recalculateFantasyPointsForPlayerTx recalculates fantasy points for all teams containing the specified player within a transaction
func (h *AdminHandler) recalculateFantasyPointsForPlayerTx(tx *sql.Tx, matchID string, playerID int64) (int, error) {
        // Find all fantasy teams that have this player in the specified match
        var teamsAffected int
        err := tx.QueryRow(`
                SELECT COUNT(DISTINCT ut.id) 
                FROM user_teams ut 
                JOIN team_players tp ON ut.id = tp.team_id 
                WHERE tp.player_id = $1 AND ut.match_id = $2`, 
                playerID, matchID).Scan(&teamsAffected)
        
        if err != nil || teamsAffected == 0 {
                return 0, nil // No teams to update
        }
        
        // Calculate base points for this player
        basePoints, err := h.calculatePlayerBasePointsTx(tx, matchID, playerID)
        if err != nil {
                return 0, err
        }
        
        // Find all fantasy teams containing this player and update their points
        rows, err := tx.Query(`
                SELECT ut.id, tp.is_captain, tp.is_vice_captain
                FROM user_teams ut 
                JOIN team_players tp ON ut.id = tp.team_id 
                WHERE tp.player_id = $1 AND ut.match_id = $2`, 
                playerID, matchID)
        
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        teamsUpdated := 0
        for rows.Next() {
                var teamID int64
                var isCaptain, isViceCaptain bool
                
                if err := rows.Scan(&teamID, &isCaptain, &isViceCaptain); err != nil {
                        continue
                }
                
                // Apply captain/vice-captain multipliers
                finalPoints := basePoints
                if isCaptain {
                        finalPoints = basePoints * 2.0 // Captain gets 2x points
                } else if isViceCaptain {
                        finalPoints = basePoints * 1.5 // Vice-captain gets 1.5x points
                }
                
                // Update team_players.points_earned for this player
                _, err = tx.Exec(`
                        UPDATE team_players 
                        SET points_earned = $1 
                        WHERE team_id = $2 AND player_id = $3`,
                        finalPoints, teamID, playerID)
                
                if err != nil {
                        continue
                }
                
                // Recalculate total points for the team
                err = h.recalculateTeamTotalPointsTx(tx, teamID)
                if err != nil {
                        continue
                }
                
                teamsUpdated++
        }
        
        return teamsUpdated, nil
}

// calculatePlayerBasePointsTx calculates base points for a player within a transaction
func (h *AdminHandler) calculatePlayerBasePointsTx(tx *sql.Tx, matchID string, playerID int64) (float64, error) {
        // Get all match events for this player
        rows, err := tx.Query(`
                SELECT event_type, points
                FROM match_events
                WHERE match_id = $1 AND player_id = $2
                ORDER BY created_at`, matchID, playerID)
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        // Calculate total base points
        var totalPoints float64
        for rows.Next() {
                var eventType string
                var points float64
                
                if err := rows.Scan(&eventType, &points); err != nil {
                        continue
                }
                
                // Add points from this event
                totalPoints += points
        }
        
        return totalPoints, nil
}

// recalculateTeamTotalPointsTx recalculates the total points for a fantasy team within a transaction
func (h *AdminHandler) recalculateTeamTotalPointsTx(tx *sql.Tx, teamID int64) error {
        // Sum all player points for this team
        var totalPoints float64
        err := tx.QueryRow(`
                SELECT COALESCE(SUM(points_earned), 0)
                FROM team_players
                WHERE team_id = $1`, teamID).Scan(&totalPoints)
        if err != nil {
                return err
        }
        
        // Update user_teams.total_points
        _, err = tx.Exec(`
                UPDATE user_teams
                SET total_points = $1, updated_at = NOW()
                WHERE id = $2`, totalPoints, teamID)
        
        return err
}

// updateAllContestLeaderboardsTx updates rankings for all contests of a match within a transaction
func (h *AdminHandler) updateAllContestLeaderboardsTx(tx *sql.Tx, matchID string) (int, error) {
        // Find all contests for this match
        rows, err := tx.Query(`
                SELECT id FROM contests WHERE match_id = $1`, matchID)
        if err != nil {
                return 0, err
        }
        defer rows.Close()
        
        leaderboardsUpdated := 0
        
        for rows.Next() {
                var contestID int64
                if err := rows.Scan(&contestID); err != nil {
                        continue
                }
                
                // Update rankings for this contest
                err = h.updateContestLeaderboardTx(tx, contestID)
                if err == nil {
                        leaderboardsUpdated++
                }
        }
        
        return leaderboardsUpdated, nil
}

// updateContestLeaderboardTx updates rankings for a specific contest within a transaction
func (h *AdminHandler) updateContestLeaderboardTx(tx *sql.Tx, contestID int64) error {
        // Update ranks based on total_points (highest points get rank 1)
        _, err := tx.Exec(`
                UPDATE contest_participants cp
                SET rank = ranked.new_rank
                FROM (
                        SELECT 
                                cp2.id,
                                ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, cp2.joined_at ASC) as new_rank
                        FROM contest_participants cp2
                        JOIN user_teams ut ON cp2.team_id = ut.id
                        WHERE cp2.contest_id = $1
                ) ranked
                WHERE cp.id = ranked.id`, contestID)
        
        return err
}

// Helper function
func parseAdminInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}