package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	db            *sql.DB
	cfg           *config.Config
	friendService *services.FriendService
}

func NewFriendHandler(db *sql.DB, cfg *config.Config) *FriendHandler {
	return &FriendHandler{
		db:            db,
		cfg:           cfg,
		friendService: services.NewFriendService(db),
	}
}

// Friend management
func (h *FriendHandler) AddFriend(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req models.AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	err = h.friendService.AddFriend(userID, req)
	if err != nil {
		// Determine appropriate status code based on error type
		statusCode := http.StatusInternalServerError
		errorCode := "ADD_FRIEND_FAILED"
		
		errMsg := err.Error()
		
		// Check for user input validation errors (should be 400)
		if strings.Contains(errMsg, "user not found") || 
		   strings.Contains(errMsg, "must provide") ||
		   strings.Contains(errMsg, "cannot add yourself") ||
		   strings.Contains(errMsg, "friendship already exists") {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_REQUEST"
		}
		
		c.JSON(statusCode, models.ErrorResponse{
			Success: false,
			Error:   errMsg,
			Code:    errorCode,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Friend request sent",
	})
}

func (h *FriendHandler) AcceptFriend(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	friendID, err := strconv.ParseInt(c.Param("friend_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid friend ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.friendService.AcceptFriend(userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to accept friend request",
			Code:    "ACCEPT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Friend request accepted",
	})
}

func (h *FriendHandler) DeclineFriend(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	friendID, err := strconv.ParseInt(c.Param("friend_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid friend ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.friendService.DeclineFriend(userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to decline friend request",
			Code:    "DECLINE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Friend request declined",
	})
}

func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	friendID, err := strconv.ParseInt(c.Param("friend_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid friend ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.friendService.RemoveFriend(userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to remove friend",
			Code:    "REMOVE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Friend removed successfully",
	})
}

func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	status := c.Query("status") // pending, accepted, all

	friends, err := h.friendService.GetFriends(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch friends",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"friends": friends,
	})
}

// Friend challenges
func (h *FriendHandler) CreateChallenge(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req models.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	challenge, err := h.friendService.CreateChallenge(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CHALLENGE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"challenge": challenge,
		"message":   "Challenge created successfully",
	})
}

func (h *FriendHandler) AcceptChallenge(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("challenge_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid challenge ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req models.AcceptChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	err = h.friendService.AcceptChallenge(challengeID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to accept challenge",
			Code:    "ACCEPT_CHALLENGE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Challenge accepted",
	})
}

func (h *FriendHandler) DeclineChallenge(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("challenge_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid challenge ID",
			Code:    "INVALID_ID",
		})
		return
	}

	err = h.friendService.DeclineChallenge(challengeID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to decline challenge",
			Code:    "DECLINE_CHALLENGE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Challenge declined",
	})
}

func (h *FriendHandler) GetChallenges(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	status := c.Query("status") // pending, accepted, completed, all

	challenges, err := h.friendService.GetChallenges(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch challenges",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"challenges": challenges,
	})
}

// Activity feed
func (h *FriendHandler) GetFriendActivities(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 50 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	activities, err := h.friendService.GetFriendActivities(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch activities",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"activities": activities,
	})
}