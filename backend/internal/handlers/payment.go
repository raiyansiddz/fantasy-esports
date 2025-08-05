package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	internal_services "fantasy-esports-backend/internal/services"
	"fantasy-esports-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	db             *sql.DB
	config         *config.Config
	paymentService *internal_services.PaymentService
}

func NewPaymentHandler(db *sql.DB, config *config.Config, paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		db:             db,
		config:         config,
		paymentService: paymentService,
	}
}

// CreatePaymentOrder creates a payment order for adding money
// @Summary Create payment order
// @Description Create a payment order using selected gateway
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order details"
// @Success 200 {object} CreateOrderResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/payment/create-order [post]
func (h *PaymentHandler) CreatePaymentOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than 0"})
		return
	}

	if req.Gateway != "razorpay" && req.Gateway != "phonepe" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gateway. Supported: razorpay, phonepe"})
		return
	}

	// Create payment order
	orderReq := &internal_services.CreateOrderRequest{
		Amount:   req.Amount,
		Gateway:  req.Gateway,
		Currency: req.Currency,
	}
	response, err := h.paymentService.CreatePaymentOrder(userID.(int64), orderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment order", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// VerifyPayment verifies payment status after completion
// @Summary Verify payment
// @Description Verify payment status from gateway response
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body VerifyPaymentRequest true "Payment verification details"
// @Success 200 {object} VerifyPaymentResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/payment/verify [post]
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Verify payment
	verifyReq := &internal_services.VerifyPaymentRequest{
		TransactionID: req.TransactionID,
		Gateway:       req.Gateway,
		GatewayData:   req.GatewayData,
	}
	response, err := h.paymentService.VerifyPayment(userID.(int64), verifyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPaymentStatus gets payment status by transaction ID
// @Summary Get payment status  
// @Description Get payment status by transaction ID
// @Tags Payment
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} PaymentStatusResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/payment/status/{transaction_id} [get]
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	status, err := h.paymentService.GetPaymentStatus(userID.(int64), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// Admin APIs for gateway configuration management

// GetGatewayConfigs gets all gateway configurations
// @Summary Get gateway configurations
// @Description Get all payment gateway configurations
// @Tags Admin - Payment
// @Success 200 {object} []PaymentGatewayConfig
// @Router /api/v1/admin/payment/gateways [get]
func (h *PaymentHandler) GetGatewayConfigs(c *gin.Context) {
	configs, err := h.paymentService.GetGatewayConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get gateway configs", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// UpdateGatewayConfig updates gateway configuration
// @Summary Update gateway configuration
// @Description Update payment gateway configuration
// @Tags Admin - Payment
// @Accept json
// @Produce json
// @Param gateway path string true "Gateway name (razorpay/phonepe)"
// @Param request body UpdateGatewayConfigRequest true "Configuration details"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/admin/payment/gateways/{gateway} [put]
func (h *PaymentHandler) UpdateGatewayConfig(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway != "razorpay" && gateway != "phonepe" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gateway. Supported: razorpay, phonepe"})
		return
	}

	var req UpdateGatewayConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	configReq := &internal_services.UpdateGatewayConfigRequest{
		Key1:     req.Key1,
		Key2:     req.Key2,
		IsLive:   req.IsLive,
		Enabled:  req.Enabled,
		Currency: req.Currency,
	}
	err := h.paymentService.UpdateGatewayConfig(gateway, configReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update gateway config", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Gateway configuration updated successfully",
	})
}

// ToggleGatewayStatus enables/disables a payment gateway
// @Summary Toggle gateway status
// @Description Enable or disable a payment gateway
// @Tags Admin - Payment
// @Param gateway path string true "Gateway name (razorpay/phonepe)"
// @Param enabled query bool true "Enable/disable gateway"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/admin/payment/gateways/{gateway}/toggle [put]
func (h *PaymentHandler) ToggleGatewayStatus(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway != "razorpay" && gateway != "phonepe" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gateway. Supported: razorpay, phonepe"})
		return
	}

	enabledStr := c.Query("enabled")
	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enabled parameter. Use true/false"})
		return
	}

	err = h.paymentService.ToggleGatewayStatus(gateway, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle gateway status", "details": err.Error()})
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Gateway " + status + " successfully",
	})
}

// GetTransactionLogs gets all payment transaction logs
// @Summary Get transaction logs
// @Description Get all payment transaction logs with pagination
// @Tags Admin - Payment
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param gateway query string false "Filter by gateway"
// @Param status query string false "Filter by status"
// @Success 200 {object} TransactionLogsResponse
// @Router /api/v1/admin/payment/transactions [get]
func (h *PaymentHandler) GetTransactionLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	gateway := c.Query("gateway")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logs, total, err := h.paymentService.GetTransactionLogs(page, limit, gateway, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction logs", "details": err.Error()})
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         logs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// Request/Response models
type CreateOrderRequest struct {
	Amount   float64 `json:"amount" binding:"required"`
	Gateway  string  `json:"gateway" binding:"required"`
	Currency string  `json:"currency,omitempty"`
}

type CreateOrderResponse struct {
	TransactionID string                 `json:"transaction_id"`
	Gateway       string                 `json:"gateway"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	PaymentData   map[string]interface{} `json:"payment_data"`
}

type VerifyPaymentRequest struct {
	TransactionID   string                 `json:"transaction_id" binding:"required"`
	Gateway         string                 `json:"gateway" binding:"required"`
	GatewayData     map[string]interface{} `json:"gateway_data" binding:"required"`
}

type VerifyPaymentResponse struct {
	Success       bool    `json:"success"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Message       string  `json:"message"`
}

type PaymentStatusResponse struct {
	TransactionID        string  `json:"transaction_id"`
	Gateway              string  `json:"gateway"`
	GatewayTransactionID *string `json:"gateway_transaction_id"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Status               string  `json:"status"`
	CreatedAt            string  `json:"created_at"`
	CompletedAt          *string `json:"completed_at"`
}

type UpdateGatewayConfigRequest struct {
	Key1     string `json:"key1" binding:"required"`
	Key2     string `json:"key2" binding:"required"`
	IsLive   bool   `json:"is_live"`
	Enabled  bool   `json:"enabled"`
	Currency string `json:"currency,omitempty"`
}

type TransactionLogsResponse struct {
	Transactions []models.PaymentTransaction `json:"transactions"`
	Total        int64                       `json:"total"`
	Page         int                         `json:"page"`
	TotalPages   int                         `json:"total_pages"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}