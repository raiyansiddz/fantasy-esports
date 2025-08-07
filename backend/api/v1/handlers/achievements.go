package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AchievementHandler struct {
	db                *sql.DB
	cfg               *config.Config
	achievementService *services.AchievementService
}

func NewAchievementHandler(db *sql.DB, cfg *config.Config) *AchievementHandler {
	return &AchievementHandler{
		db:                db,
		cfg:               cfg,
		achievementService: services.NewAchievementService(db),
	}
}

// Admin endpoints
func (h *AchievementHandler) CreateAchievement(c *gin.Context) {
	var req models.CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	// Get admin user from JWT
	adminID, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	achievement, err := h.achievementService.CreateAchievement(req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create achievement",
			Code:    "CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"achievement": achievement,
	})
}

func (h *AchievementHandler) UpdateAchievement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid achievement ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req models.CreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	err = h.achievementService.UpdateAchievement(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update achievement",
			Code:    "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Achievement updated successfully",
	})
}

func (h *AchievementHandler) DeleteAchievement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid achievement ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.achievementService.DeleteAchievement(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to delete achievement",
			Code:    "DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Achievement deleted successfully",
	})
}

func (h *AchievementHandler) GetAchievements(c *gin.Context) {
	isActiveStr := c.Query("is_active")
	var isActive *bool
	if isActiveStr != "" {
		val := isActiveStr == "true"
		isActive = &val
	}

	achievements, err := h.achievementService.GetAchievements(isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch achievements",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"achievements": achievements,
	})
}

// User endpoints
func (h *AchievementHandler) GetUserAchievements(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	achievements, err := h.achievementService.GetUserAchievements(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch user achievements",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"achievements": achievements,
	})
}

func (h *AchievementHandler) GetAchievementProgress(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	achievementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid achievement ID",
			Code:    "INVALID_ID",
		})
		return
	}

	progress, err := h.achievementService.GetAchievementProgress(userID, achievementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch progress",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"progress": progress,
	})
}

// Trigger achievement check (internal use)
func (h *AchievementHandler) TriggerAchievementCheck(userID int64, triggerType string, contextData map[string]interface{}) {
	go func() {
		err := h.achievementService.CheckAndAwardAchievements(userID, triggerType, contextData)
		if err != nil {
			// Log error but don't block the main flow
		}
	}()
}