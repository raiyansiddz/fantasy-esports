package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/pkg/cdn"
	"fantasy-esports-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	db              *sql.DB
	config          *config.Config
	cdn             *cdn.CloudinaryClient
	referralService *services.ReferralService
	otpSessions     map[string]models.OTPSession // In-memory storage for demo
}

func NewAuthHandler(db *sql.DB, cfg *config.Config, cdn *cdn.CloudinaryClient) *AuthHandler {
	return &AuthHandler{
		db:              db,
		config:          cfg,
		cdn:             cdn,
		referralService: services.NewReferralService(db),
		otpSessions:     make(map[string]models.OTPSession),
	}
}

// @Summary Verify mobile number (one-step registration/login)
// @Description Send OTP to mobile number. Creates new user if not exists.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.AuthRequest true "Mobile verification request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/verify-mobile [post]
func (h *AuthHandler) VerifyMobile(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate mobile number
	if !utils.ValidateMobile(req.Mobile) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid mobile number format",
			Code:    "INVALID_MOBILE",
		})
		return
	}

	// Check if user exists
	var userID int64
	var isNewUser bool
	err := h.db.QueryRow("SELECT id FROM users WHERE mobile = $1", req.Mobile).Scan(&userID)
	if err == sql.ErrNoRows {
		isNewUser = true
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error",
			Code:    "DB_ERROR",
		})
		return
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	sessionID := uuid.New().String()

	// Store OTP session (in-memory for demo)
	h.otpSessions[sessionID] = models.OTPSession{
		SessionID: sessionID,
		Mobile:    req.Mobile,
		OTP:       otp,
		IsNewUser: isNewUser,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	// Print OTP to console (as per requirements)
	utils.PrintOTPToConsole(req.Mobile, otp)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"otp_sent":    true,
		"session_id":  sessionID,
		"is_new_user": isNewUser,
		"message":     "OTP sent successfully",
	})
}

// @Summary Verify OTP and complete login/registration
// @Description Verify OTP and return JWT tokens. Creates user if new.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.OTPVerifyRequest true "OTP verification request"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get OTP session
	session, exists := h.otpSessions[req.SessionID]
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid session",
			Code:    "INVALID_SESSION",
		})
		return
	}

	// Validate OTP
	if !utils.ValidateOTP(req.OTP, session.OTP, session.CreatedAt) && !utils.IsDevelopmentOTP(req.OTP) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid or expired OTP",
			Code:    "INVALID_OTP",
		})
		return
	}

	var user models.User
	var err error

	if session.IsNewUser {
		// Create new user
		if req.ProfileData == nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "Profile data required for new user",
				Code:    "PROFILE_DATA_REQUIRED",
			})
			return
		}

		referralCode := models.GenerateReferralCode()
		
		err = h.db.QueryRow(`
			INSERT INTO users (mobile, first_name, last_name, email, date_of_birth, state, 
							  is_verified, referral_code, referred_by_code, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, true, $7, $8, NOW(), NOW())
			RETURNING id, mobile, email, first_name, last_name, date_of_birth, 
					  is_verified, account_status, kyc_status, referral_code, 
					  state, created_at, updated_at`,
			session.Mobile, req.ProfileData.FirstName, req.ProfileData.LastName,
			req.ProfileData.Email, req.ProfileData.DateOfBirth, req.ProfileData.State,
			referralCode, req.ReferralCode,
		).Scan(&user.ID, &user.Mobile, &user.Email, &user.FirstName, &user.LastName,
			&user.DateOfBirth, &user.IsVerified, &user.AccountStatus, &user.KYCStatus,
			&user.ReferralCode, &user.State, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "Failed to create user",
				Code:    "USER_CREATION_FAILED",
			})
			return
		}

		// Create wallet for new user
		_, err = h.db.Exec("INSERT INTO user_wallets (user_id) VALUES ($1)", user.ID)
		if err != nil {
			// Log error but continue - wallet can be created later
		}

	} else {
		// Get existing user
		err = h.db.QueryRow(`
			SELECT id, mobile, email, first_name, last_name, date_of_birth, 
				   is_verified, account_status, kyc_status, referral_code, 
				   state, created_at, updated_at
			FROM users WHERE mobile = $1`,
			session.Mobile,
		).Scan(&user.ID, &user.Mobile, &user.Email, &user.FirstName, &user.LastName,
			&user.DateOfBirth, &user.IsVerified, &user.AccountStatus, &user.KYCStatus,
			&user.ReferralCode, &user.State, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "Failed to fetch user",
				Code:    "USER_FETCH_FAILED",
			})
			return
		}

		// Update last login
		_, err = h.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := utils.GenerateTokens(user.ID, user.Mobile, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate tokens",
			Code:    "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// Clean up OTP session
	delete(h.otpSessions, req.SessionID)

	c.JSON(http.StatusOK, models.AuthResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		IsNewUser:    session.IsNewUser,
	})
}

// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Authorization header required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	claims, err := utils.ValidateToken(tokenString, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Invalid token",
			Code:    "INVALID_TOKEN",
		})
		return
	}

	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Invalid token type",
			Code:    "INVALID_TOKEN_TYPE",
		})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := utils.GenerateTokens(claims.UserID, claims.Mobile, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate tokens",
			Code:    "TOKEN_GENERATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// @Summary Logout user
// @Description Logout user (invalidate tokens)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a production system, you would maintain a blacklist of tokens
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// @Summary Social login
// @Description Login using social providers (Google, Facebook)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Social login request"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/social-login [post]
func (h *AuthHandler) SocialLogin(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// This is a placeholder implementation
	// In production, you would validate the social provider token
	
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Success: false,
		Error:   "Social login not implemented yet",
		Code:    "NOT_IMPLEMENTED",
	})
}