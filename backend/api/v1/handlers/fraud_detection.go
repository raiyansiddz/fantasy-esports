package handlers

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FraudDetectionHandler struct {
	db                    *sql.DB
	cfg                   *config.Config
	fraudDetectionService *services.FraudDetectionService
}

func NewFraudDetectionHandler(db *sql.DB, cfg *config.Config) *FraudDetectionHandler {
	return &FraudDetectionHandler{
		db:                    db,
		cfg:                   cfg,
		fraudDetectionService: services.NewFraudDetectionService(db),
	}
}

// Admin endpoints
func (h *FraudDetectionHandler) GetAlerts(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	status := c.Query("status")
	severity := c.Query("severity")
	limitStr := c.Query("limit")
	
	limit := 50 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	alerts, err := h.fraudDetectionService.GetAlerts(status, severity, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch alerts",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"alerts":  alerts,
		"filters": gin.H{
			"status":   status,
			"severity": severity,
			"limit":    limit,
		},
	})
}

func (h *FraudDetectionHandler) UpdateAlertStatus(c *gin.Context) {
	// Verify admin access
	adminID, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	alertID, err := strconv.ParseInt(c.Param("alert_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid alert ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req struct {
		Status          string  `json:"status" binding:"required"`
		ResolutionNotes *string `json:"resolution_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"open":           true,
		"investigating":  true,
		"resolved":       true,
		"false_positive": true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid status",
			Code:    "INVALID_STATUS",
		})
		return
	}

	adminIDInt64 := adminID.(int64)
	err = h.fraudDetectionService.UpdateAlertStatus(alertID, req.Status, &adminIDInt64, req.ResolutionNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update alert",
			Code:    "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alert updated successfully",
	})
}

func (h *FraudDetectionHandler) GetFraudStatistics(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
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

	stats, err := h.fraudDetectionService.GetFraudStatistics(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch statistics",
			Code:    "STATS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"statistics": stats,
		"period":     days,
	})
}

// Middleware for fraud detection on user actions
func (h *FraudDetectionHandler) FraudDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID if available
		userID := int64(0)
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(int64)
		}

		// Get IP address and User-Agent
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// Determine action based on route
		action := h.getActionFromRoute(c.FullPath(), c.Request.Method)

		if action != "" && userID != 0 {
			// Prepare context data (can be enhanced based on specific endpoints)
			contextData := map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.FullPath(),
				"user_agent": userAgent,
				"ip_address": ipAddress,
			}

			// Run fraud detection asynchronously to not block the request
			go func() {
				err := h.fraudDetectionService.CheckUserAction(userID, action, contextData, ipAddress, userAgent)
				if err != nil {
					// Log error but don't affect user experience
				}
			}()
		}

		c.Next()
	}
}

func (h *FraudDetectionHandler) getActionFromRoute(path, method string) string {
	// Map routes to actions for fraud detection
	routeActions := map[string]string{
		"POST:/api/v1/teams":                    "team_created",
		"POST:/api/v1/contests/:id/join":       "contest_joined",
		"POST:/api/v1/wallet/deposit":          "wallet_deposit",
		"POST:/api/v1/wallet/withdraw":         "wallet_withdrawal",
		"PUT:/api/v1/profile":                  "profile_updated",
		"POST:/api/v1/friends":                 "friend_added",
		"POST:/api/v1/auth/verify-otp":         "login_attempt",
		"POST:/api/v1/contests/private":        "private_contest_created",
	}

	key := method + ":" + path
	if action, exists := routeActions[key]; exists {
		return action
	}

	// Generic API call tracking
	return "api_call"
}

// Public endpoint for reporting suspicious activity
func (h *FraudDetectionHandler) ReportSuspiciousActivity(c *gin.Context) {
	var req struct {
		UserID      *int64                 `json:"user_id"`
		ActivityType string                 `json:"activity_type" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		Evidence    map[string]interface{} `json:"evidence"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	// Create fraud alert from user report
	evidenceJSON, _ := json.Marshal(req.Evidence)
	
	alert := models.FraudAlert{
		UserID:        req.UserID,
		AlertType:     "user_reported_" + req.ActivityType,
		Severity:      "medium",
		Description:   "User reported: " + req.Description,
		DetectionData: evidenceJSON,
		Status:        "open",
	}

	// This would typically go through a service method
	_, err := h.db.Exec(`
		INSERT INTO fraud_alerts (user_id, alert_type, severity, description, detection_data, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, alert.UserID, alert.AlertType, alert.Severity, alert.Description, alert.DetectionData, alert.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to report activity",
			Code:    "REPORT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Suspicious activity reported successfully",
	})
}

// Webhook endpoint for external fraud detection services
func (h *FraudDetectionHandler) FraudWebhook(c *gin.Context) {
	// Verify webhook authenticity (implement signature verification)
	
	var webhook struct {
		Source      string                 `json:"source"`
		EventType   string                 `json:"event_type"`
		UserID      *int64                 `json:"user_id"`
		Severity    string                 `json:"severity"`
		Description string                 `json:"description"`
		Data        map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid webhook data",
			Code:    "INVALID_WEBHOOK",
		})
		return
	}

	// Create alert from external source
	dataJSON, _ := json.Marshal(webhook.Data)
	
	alert := models.FraudAlert{
		UserID:        webhook.UserID,
		AlertType:     "external_" + webhook.EventType,
		Severity:      webhook.Severity,
		Description:   fmt.Sprintf("[%s] %s", webhook.Source, webhook.Description),
		DetectionData: dataJSON,
		Status:        "open",
	}

	_, err := h.db.Exec(`
		INSERT INTO fraud_alerts (user_id, alert_type, severity, description, detection_data, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, alert.UserID, alert.AlertType, alert.Severity, alert.Description, alert.DetectionData, alert.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to process webhook",
			Code:    "WEBHOOK_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook processed successfully",
	})
}