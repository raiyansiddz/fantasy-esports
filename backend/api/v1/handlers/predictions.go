package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlayerPredictionHandler struct {
	db                      *sql.DB
	cfg                     *config.Config
	playerPredictionService *services.PlayerPredictionService
}

func NewPlayerPredictionHandler(db *sql.DB, cfg *config.Config) *PlayerPredictionHandler {
	return &PlayerPredictionHandler{
		db:                      db,
		cfg:                     cfg,
		playerPredictionService: services.NewPlayerPredictionService(db),
	}
}

func (h *PlayerPredictionHandler) GenerateMatchPredictions(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid match ID '%s': must be a positive integer", matchIDStr),
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	if matchID <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Match ID must be a positive integer",
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	// Check if match exists
	var exists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1)", matchID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error while checking match",
			Code:    "DATABASE_ERROR",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Match with ID %d not found", matchID),
			Code:    "MATCH_NOT_FOUND",
		})
		return
	}

	err = h.playerPredictionService.GenerateMatchPredictions(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate predictions: %v", err.Error()),
			Code:    "GENERATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Predictions generated successfully",
	})
}

func (h *PlayerPredictionHandler) GetMatchPredictions(c *gin.Context) {
	matchID, err := strconv.ParseInt(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_ID",
		})
		return
	}

	predictions, err := h.playerPredictionService.GetPlayerPredictions(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch predictions",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"predictions": predictions,
	})
}

func (h *PlayerPredictionHandler) UpdatePredictionAccuracy(c *gin.Context) {
	matchID, err := strconv.ParseInt(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.playerPredictionService.UpdatePredictionAccuracy(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update accuracy",
			Code:    "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prediction accuracy updated",
	})
}

func (h *PlayerPredictionHandler) GetPredictionAnalytics(c *gin.Context) {
	daysStr := c.Query("days")
	days := 30 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	analytics, err := h.playerPredictionService.GetPredictionAnalytics(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch analytics",
			Code:    "ANALYTICS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"analytics": analytics,
		"period":    days,
	})
}