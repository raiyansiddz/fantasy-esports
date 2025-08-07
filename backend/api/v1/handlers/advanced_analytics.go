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

	metrics, err := h.advancedAnalyticsService.CalculateAdvancedGameMetrics(gameID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to calculate metrics",
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