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

type SocialSharingHandler struct {
	db                   *sql.DB
	cfg                  *config.Config
	socialSharingService *services.SocialSharingService
}

func NewSocialSharingHandler(db *sql.DB, cfg *config.Config) *SocialSharingHandler {
	return &SocialSharingHandler{
		db:                   db,
		cfg:                  cfg,
		socialSharingService: services.NewSocialSharingService(db, cfg.BaseURL),
	}
}

func (h *SocialSharingHandler) CreateShare(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req models.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request data: %v", err.Error()),
			Code:    "INVALID_DATA",
		})
		return
	}

	// Validate required fields
	if req.ShareType == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "share_type is required",
			Code:    "MISSING_SHARE_TYPE",
		})
		return
	}

	if req.Platform == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "platform is required",
			Code:    "MISSING_PLATFORM",
		})
		return
	}

	if req.ShareData == nil || len(req.ShareData) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "share_data is required and must not be empty",
			Code:    "MISSING_SHARE_DATA",
		})
		return
	}

	// Validate share type
	validShareTypes := []string{"team_composition", "contest_win", "achievement_unlock", "challenge_result"}
	isValidType := false
	for _, validType := range validShareTypes {
		if req.ShareType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid share_type. Must be one of: team_composition, contest_win, achievement_unlock, challenge_result",
			Code:    "INVALID_SHARE_TYPE",
		})
		return
	}

	// Validate platform
	validPlatforms := []string{"twitter", "facebook", "whatsapp", "instagram"}
	isValidPlatform := false
	for _, validPlatform := range validPlatforms {
		if req.Platform == validPlatform {
			isValidPlatform = true
			break
		}
	}
	if !isValidPlatform {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid platform. Must be one of: twitter, facebook, whatsapp, instagram",
			Code:    "INVALID_PLATFORM",
		})
		return
	}

	share, err := h.socialSharingService.CreateShare(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create share: %v", err.Error()),
			Code:    "CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"share":   share,
	})
}

func (h *SocialSharingHandler) GenerateTeamShareURLs(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	teamID, err := strconv.ParseInt(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid team ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Generate content for team composition
	content, err := h.socialSharingService.GenerateTeamCompositionContent(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate team content",
			Code:    "CONTENT_FAILED",
		})
		return
	}

	// Get platform URLs
	urls, err := h.socialSharingService.GetPlatformURLs(userID, "team_composition", &teamID, *content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate share URLs",
			Code:    "URL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": content,
		"urls":    urls,
	})
}

func (h *SocialSharingHandler) GenerateContestWinShareURLs(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid contest ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Generate content for contest win
	content, err := h.socialSharingService.GenerateContestWinContent(userID, contestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate contest win content",
			Code:    "CONTENT_FAILED",
		})
		return
	}

	// Get platform URLs
	urls, err := h.socialSharingService.GetPlatformURLs(userID, "contest_win", &contestID, *content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate share URLs",
			Code:    "URL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": content,
		"urls":    urls,
	})
}

func (h *SocialSharingHandler) GenerateAchievementShareURLs(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	achievementID, err := strconv.ParseInt(c.Param("achievement_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid achievement ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Generate content for achievement
	content, err := h.socialSharingService.GenerateAchievementContent(userID, achievementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate achievement content",
			Code:    "CONTENT_FAILED",
		})
		return
	}

	// Get platform URLs
	urls, err := h.socialSharingService.GetPlatformURLs(userID, "achievement", &achievementID, *content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate share URLs",
			Code:    "URL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": content,
		"urls":    urls,
	})
}

func (h *SocialSharingHandler) TrackShareClick(c *gin.Context) {
	shareID, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid share ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.socialSharingService.TrackShareClick(shareID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to track click",
			Code:    "TRACK_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Click tracked",
	})
}

func (h *SocialSharingHandler) GetUserShares(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	platform := c.Query("platform")

	shares, err := h.socialSharingService.GetUserShares(userID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch shares",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"shares":  shares,
	})
}

func (h *SocialSharingHandler) GetShareAnalytics(c *gin.Context) {
	// Optional user filter
	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &uid
		}
	}

	platform := c.Query("platform")
	daysStr := c.Query("days")
	days := 30 // Default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	analytics, err := h.socialSharingService.GetShareAnalytics(userID, platform, days)
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
		"period":    gin.H{
			"days":     days,
			"platform": platform,
		},
	})
}