package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdvancedAnalyticsHandler struct {
	db                       *sql.DB
	cfg                      *config.Config
	advancedAnalyticsService *services.AdvancedAnalyticsService
}

func NewAdvancedAnalyticsHandler(db *sql.DB, cfg *config.Config) *AdvancedAnalyticsHandler {
	return &AdvancedAnalyticsHandler{
		db:                       db,
		cfg:                      cfg,
		advancedAnalyticsService: services.NewAdvancedAnalyticsService(db),
	}
}

func (h *AdvancedAnalyticsHandler) GetAdvancedGameMetrics(c *gin.Context) {
	gameIDStr := c.Param("game_id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid game ID '%s': must be a positive integer", gameIDStr),
			Code:    "INVALID_GAME_ID",
		})
		return
	}

	if gameID <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Game ID must be a positive integer",
			Code:    "INVALID_GAME_ID",
		})
		return
	}

	// Check if game exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)", gameID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error while checking game",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Game with ID %d not found", gameID),
			Code:    "GAME_NOT_FOUND",
		})
		return
	}

	daysStr := c.Query("days")
	days := 30 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		} else if daysStr != "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "Days parameter must be a positive integer between 1 and 365",
				Code:    "INVALID_DAYS",
			})
			return
		}
	}

	metrics, err := h.advancedAnalyticsService.CalculateAdvancedGameMetrics(gameID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to calculate metrics: %v", err.Error()),
			Code:    "METRICS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
		"period":  days,
	})
}

func (h *AdvancedAnalyticsHandler) GetAdvancedMetricsHistory(c *gin.Context) {
	gameID, err := strconv.Atoi(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid game ID",
			Code:    "INVALID_ID",
		})
		return
	}

	daysStr := c.Query("days")
	days := 30 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	history, err := h.advancedAnalyticsService.GetAdvancedMetricsHistory(gameID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch history",
			Code:    "HISTORY_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
		"period":  days,
	})
}

func (h *AdvancedAnalyticsHandler) CompareGames(c *gin.Context) {
	gameIDsStr := c.Query("game_ids")
	if gameIDsStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Game IDs required",
			Code:    "MISSING_GAME_IDS",
		})
		return
	}

	gameIDStrings := strings.Split(gameIDsStr, ",")
	var gameIDs []int
	for _, idStr := range gameIDStrings {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			gameIDs = append(gameIDs, id)
		}
	}

	if len(gameIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Valid game IDs required",
			Code:    "INVALID_GAME_IDS",
		})
		return
	}

	metricType := c.Query("metric_type")
	if metricType == "" {
		metricType = "player_efficiency" // Default
	}

	daysStr := c.Query("days")
	days := 30 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	comparison, err := h.advancedAnalyticsService.GetGameComparison(gameIDs, metricType, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to compare games",
			Code:    "COMPARISON_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"comparison": comparison,
		"parameters": gin.H{
			"game_ids":    gameIDs,
			"metric_type": metricType,
			"days":        days,
		},
	})
}