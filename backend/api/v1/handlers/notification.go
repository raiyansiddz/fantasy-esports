package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	db                  *sql.DB
	config              *config.Config
	notificationService *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *sql.DB, cfg *config.Config) *NotificationHandler {
	return &NotificationHandler{
		db:                  db,
		config:              cfg,
		notificationService: services.NewNotificationService(db),
	}
}

// @Summary Send single notification
// @Description Send a notification via specified channel and provider
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendNotificationRequest true "Send notification request"
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notify/send [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var request models.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if request.Recipient == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Recipient is required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate channel
	validChannels := []models.NotificationChannel{
		models.ChannelSMS, models.ChannelEmail, models.ChannelPush, models.ChannelWhatsApp,
	}
	validChannel := false
	for _, ch := range validChannels {
		if request.Channel == ch {
			validChannel = true
			break
		}
	}
	if !validChannel {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid channel. Must be one of: sms, email, push, whatsapp",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate recipient format based on channel (MOVED UP)
	if err := h.validateRecipient(request.Channel, request.Recipient); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate that either template_id or body is provided (MOVED DOWN)
	if request.TemplateID == nil && (request.Body == nil || *request.Body == "") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Either template_id or body must be provided",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	response, err := h.notificationService.SendNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "SEND_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Send bulk notifications
// @Description Send notifications to multiple recipients
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.BulkNotificationRequest true "Bulk notification request"
// @Success 200 {object} []models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notify/bulk [post]
func (h *NotificationHandler) SendBulkNotification(c *gin.Context) {
	var request models.BulkNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if len(request.Recipients) == 0 && request.UserFilter == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Recipients list or user filter is required for bulk notifications",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate channel
	validChannels := []models.NotificationChannel{
		models.ChannelSMS, models.ChannelEmail, models.ChannelPush, models.ChannelWhatsApp,
	}
	validChannel := false
	for _, ch := range validChannels {
		if request.Channel == ch {
			validChannel = true
			break
		}
	}
	if !validChannel {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid channel. Must be one of: sms, email, push, whatsapp",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Limit bulk recipients to prevent abuse (MOVED UP)
	if len(request.Recipients) > 1000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Maximum 1000 recipients allowed per bulk request",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate recipients format (MOVED UP)
	for _, recipient := range request.Recipients {
		if err := h.validateRecipient(request.Channel, recipient); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid recipient '%s': %s", recipient, err.Error()),
				Code:    "VALIDATION_ERROR",
			})
			return
		}
	}

	// Validate that either template_id or body is provided (MOVED DOWN)
	if request.TemplateID == nil && (request.Body == nil || *request.Body == "") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Either template_id or body must be provided",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	responses, err := h.notificationService.SendBulkNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "BULK_SEND_FAILED",
		})
		return
	}

	// Calculate summary
	var successCount, failedCount int
	for _, resp := range responses {
		if resp.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total":         len(responses),
		"success_count": successCount,
		"failed_count":  failedCount,
		"responses":     responses,
	})
}

// @Summary Send SMS notification
// @Description Send SMS notification via Fast2SMS
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendNotificationRequest true "SMS notification request"
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /notify/sms [post]
func (h *NotificationHandler) SendSMS(c *gin.Context) {
	var request models.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set channel to SMS
	request.Channel = models.ChannelSMS
	if request.Provider == nil {
		provider := models.ProviderFast2SMS
		request.Provider = &provider
	}

	response, err := h.notificationService.SendNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "SMS_SEND_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Send email notification
// @Description Send email notification via SMTP, SES, or Mailchimp
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendNotificationRequest true "Email notification request"
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /notify/email [post]
func (h *NotificationHandler) SendEmail(c *gin.Context) {
	var request models.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set channel to Email
	request.Channel = models.ChannelEmail
	if request.Provider == nil {
		provider := models.ProviderSMTP
		request.Provider = &provider
	}

	response, err := h.notificationService.SendNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "EMAIL_SEND_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Send push notification
// @Description Send push notification via FCM or OneSignal
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendNotificationRequest true "Push notification request"
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /notify/push [post]
func (h *NotificationHandler) SendPush(c *gin.Context) {
	var request models.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set channel to Push
	request.Channel = models.ChannelPush
	if request.Provider == nil {
		provider := models.ProviderFCM
		request.Provider = &provider
	}

	response, err := h.notificationService.SendNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "PUSH_SEND_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Send WhatsApp notification
// @Description Send WhatsApp notification via WhatsApp Cloud API
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SendNotificationRequest true "WhatsApp notification request"
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /notify/whatsapp [post]
func (h *NotificationHandler) SendWhatsApp(c *gin.Context) {
	var request models.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set channel to WhatsApp
	request.Channel = models.ChannelWhatsApp
	if request.Provider == nil {
		provider := models.ProviderWhatsAppCloud
		request.Provider = &provider
	}

	response, err := h.notificationService.SendNotification(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "WHATSAPP_SEND_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Template Management Endpoints

// @Summary Create notification template
// @Description Create a new notification template
// @Tags Admin/Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TemplateCreateRequest true "Template creation request"
// @Success 201 {object} models.NotificationTemplate
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/templates [post]
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var request models.TemplateCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate required fields
	if request.Name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Template name is required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	if request.Body == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Template body is required",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate channel
	validChannels := []models.NotificationChannel{
		models.ChannelSMS, models.ChannelEmail, models.ChannelPush, models.ChannelWhatsApp,
	}
	validChannel := false
	for _, ch := range validChannels {
		if request.Channel == ch {
			validChannel = true
			break
		}
	}
	if !validChannel {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid channel. Must be one of: sms, email, push, whatsapp",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate provider
	validProviders := []models.NotificationProvider{
		models.ProviderFast2SMS, models.ProviderSMTP, models.ProviderSES, models.ProviderMailchimp,
		models.ProviderFCM, models.ProviderOneSignal, models.ProviderWhatsAppCloud,
	}
	validProvider := false
	for _, pr := range validProviders {
		if request.Provider == pr {
			validProvider = true
			break
		}
	}
	if !validProvider {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid provider",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate name length
	if len(request.Name) > 200 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Template name must not exceed 200 characters",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Validate subject length if provided
	if request.Subject != nil && len(*request.Subject) > 500 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Subject must not exceed 500 characters",
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	// Get admin user ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	template, err := h.notificationService.CreateTemplate(&request, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "TEMPLATE_CREATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// @Summary Get notification templates
// @Description Get notification templates with optional filtering
// @Tags Admin/Templates
// @Produce json
// @Security BearerAuth
// @Param channel query string false "Filter by channel"
// @Param provider query string false "Filter by provider"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} object
// @Router /admin/templates [get]
func (h *NotificationHandler) GetTemplates(c *gin.Context) {
	// Parse query parameters
	channelStr := c.Query("channel")
	providerStr := c.Query("provider")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	var channel *models.NotificationChannel
	if channelStr != "" {
		ch := models.NotificationChannel(channelStr)
		channel = &ch
	}

	var provider *models.NotificationProvider
	if providerStr != "" {
		pr := models.NotificationProvider(providerStr)
		provider = &pr
	}

	templates, total, err := h.notificationService.GetTemplates(channel, provider, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "TEMPLATES_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"templates": templates,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// @Summary Get notification template by ID
// @Description Get a specific notification template
// @Tags Admin/Templates
// @Produce json
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Success 200 {object} models.NotificationTemplate
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/templates/{id} [get]
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	templateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid template ID",
			Code:    "INVALID_ID",
		})
		return
	}

	template, err := h.notificationService.GetTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found",
			Code:    "TEMPLATE_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, template)
}

// @Summary Update notification template
// @Description Update an existing notification template
// @Tags Admin/Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Param request body models.TemplateUpdateRequest true "Template update request"
// @Success 200 {object} object
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/templates/{id} [put]
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	templateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid template ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var request models.TemplateUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	err = h.notificationService.UpdateTemplate(templateID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "TEMPLATE_UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Template updated successfully",
	})
}

// Configuration Management Endpoints

// @Summary Update notification configuration
// @Description Update configuration for notification providers
// @Tags Admin/Config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ConfigUpdateRequest true "Configuration update request"
// @Success 200 {object} object
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/config/notifications [put]
func (h *NotificationHandler) UpdateConfig(c *gin.Context) {
	var request models.ConfigUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get admin user ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	err := h.notificationService.UpdateConfig(&request, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CONFIG_UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// @Summary Get notification configuration
// @Description Get configuration for a specific provider and channel
// @Tags Admin/Config
// @Produce json
// @Security BearerAuth
// @Param provider query string true "Provider name"
// @Param channel query string true "Channel name"
// @Success 200 {object} object
// @Router /admin/config/notifications [get]
func (h *NotificationHandler) GetConfig(c *gin.Context) {
	providerStr := c.Query("provider")
	channelStr := c.Query("channel")

	if providerStr == "" || channelStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Provider and channel are required",
			Code:    "MISSING_PARAMETERS",
		})
		return
	}

	provider := models.NotificationProvider(providerStr)
	channel := models.NotificationChannel(channelStr)

	config, err := h.notificationService.GetConfig(provider, channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CONFIG_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
	})
}

// Statistics Endpoints

// @Summary Get notification statistics
// @Description Get notification statistics for specified period
// @Tags Admin/Stats
// @Produce json
// @Security BearerAuth
// @Param channel query string false "Filter by channel"
// @Param provider query string false "Filter by provider"
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} models.NotificationStats
// @Router /admin/stats/notifications [get]
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	channelStr := c.Query("channel")
	providerStr := c.Query("provider")
	daysStr := c.DefaultQuery("days", "7")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}

	var channel *models.NotificationChannel
	if channelStr != "" {
		ch := models.NotificationChannel(channelStr)
		channel = &ch
	}

	var provider *models.NotificationProvider
	if providerStr != "" {
		pr := models.NotificationProvider(providerStr)
		provider = &pr
	}

	stats, err := h.notificationService.GetNotificationStats(channel, provider, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "STATS_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
		"period":  fmt.Sprintf("%d days", days),
	})
}

// @Summary Get channel statistics
// @Description Get statistics per channel and provider
// @Tags Admin/Stats
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days" default(7)
// @Success 200 {object} []models.ChannelStats
// @Router /admin/stats/channels [get]
func (h *NotificationHandler) GetChannelStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}

	stats, err := h.notificationService.GetChannelStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CHANNEL_STATS_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
		"period":  fmt.Sprintf("%d days", days),
	})
}

// validateRecipient validates recipient format based on channel
func (h *NotificationHandler) validateRecipient(channel models.NotificationChannel, recipient string) error {
	switch channel {
	case models.ChannelSMS, models.ChannelWhatsApp:
		// First check if it starts with proper format (letters/invalid chars first)
		if len(recipient) > 0 && recipient[0] != '+' && (recipient[0] < '0' || recipient[0] > '9') {
			return fmt.Errorf("phone number should start with + or digit")
		}
		// Then validate length
		if len(recipient) < 10 {
			return fmt.Errorf("invalid phone number format")
		}
	case models.ChannelEmail:
		// Basic email validation
		if !strings.Contains(recipient, "@") || !strings.Contains(recipient, ".") {
			return fmt.Errorf("invalid email format")
		}
	case models.ChannelPush:
		// Push notification token should not be empty
		if len(recipient) < 10 {
			return fmt.Errorf("invalid push token format")
		}
	}
	return nil
}