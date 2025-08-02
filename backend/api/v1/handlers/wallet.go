package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	db              *sql.DB
	config          *config.Config
	referralService *services.ReferralService
}

func NewWalletHandler(db *sql.DB, cfg *config.Config) *WalletHandler {
	return &WalletHandler{
		db:              db,
		config:          cfg,
		referralService: services.NewReferralService(db),
	}
}

// @Summary Get wallet balance
// @Description Get user's wallet balance breakdown
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.WalletBalance
// @Router /wallet/balance [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var wallet models.UserWallet
	err := h.db.QueryRow(`
		SELECT user_id, bonus_balance, deposit_balance, winning_balance, total_balance, updated_at
		FROM user_wallets WHERE user_id = $1`, userID).Scan(
		&wallet.UserID, &wallet.BonusBalance, &wallet.DepositBalance,
		&wallet.WinningBalance, &wallet.TotalBalance, &wallet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create wallet if doesn't exist
		_, err = h.db.Exec("INSERT INTO user_wallets (user_id) VALUES ($1)", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "Failed to create wallet",
				Code:    "WALLET_CREATION_FAILED",
			})
			return
		}
		wallet = models.UserWallet{UserID: userID}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error",
			Code:    "DB_ERROR",
		})
		return
	}

	balance := models.WalletBalance{
		BonusBalance:        wallet.BonusBalance,
		DepositBalance:      wallet.DepositBalance,
		WinningBalance:      wallet.WinningBalance,
		TotalBalance:        wallet.TotalBalance,
		WithdrawableBalance: wallet.DepositBalance + wallet.WinningBalance,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"balance": balance,
	})
}

// @Summary Deposit money
// @Description Add money to wallet using payment gateway
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.DepositRequest true "Deposit request"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /wallet/deposit [post]
func (h *WalletHandler) Deposit(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if req.Amount < 10 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Minimum deposit amount is ₹10",
			Code:    "MIN_AMOUNT_ERROR",
		})
		return
	}

	// Generate payment transaction
	paymentID := uuid.New().String()
	gatewayOrderID := "order_" + paymentID

	// In production, you would integrate with actual payment gateway
	// For now, create a pending transaction
	_, err := h.db.Exec(`
		INSERT INTO payment_transactions (user_id, transaction_id, gateway, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, 'deposit', 'initiated', NOW())`,
		userID, paymentID, req.PaymentMethod, req.Amount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create payment transaction",
			Code:    "PAYMENT_CREATION_FAILED",
		})
		return
	}

	response := models.PaymentResponse{
		PaymentID:      paymentID,
		GatewayOrderID: gatewayOrderID,
		Amount:         req.Amount,
		GatewayURL:     "https://mock-payment-gateway.com/pay/" + paymentID,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"payment": response,
	})
}

// @Summary Withdraw money
// @Description Withdraw money from wallet to bank account
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.WithdrawRequest true "Withdrawal request"
// @Success 200 {object} models.WithdrawResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /wallet/withdraw [post]
func (h *WalletHandler) Withdraw(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if req.Amount < 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Minimum withdrawal amount is ₹100",
			Code:    "MIN_WITHDRAWAL_ERROR",
		})
		return
	}

	// Check withdrawable balance
	var withdrawableBalance float64
	err := h.db.QueryRow(`
		SELECT deposit_balance + winning_balance FROM user_wallets WHERE user_id = $1`, userID).Scan(&withdrawableBalance)

	if err != nil || withdrawableBalance < req.Amount {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Insufficient withdrawable balance",
			Code:    "INSUFFICIENT_BALANCE",
		})
		return
	}

	withdrawalID := uuid.New().String()
	processingFee := 5.0 // Fixed fee for now
	netAmount := req.Amount - processingFee

	// Create withdrawal transaction
	_, err = h.db.Exec(`
		INSERT INTO payment_transactions (user_id, transaction_id, gateway, amount, type, status, created_at)
		VALUES ($1, $2, 'bank_transfer', $3, 'withdrawal', 'pending', NOW())`,
		userID, withdrawalID, req.Amount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create withdrawal request",
			Code:    "WITHDRAWAL_CREATION_FAILED",
		})
		return
	}

	response := models.WithdrawResponse{
		WithdrawalID:  withdrawalID,
		Amount:        req.Amount,
		ProcessingFee: processingFee,
		NetAmount:     netAmount,
		Status:        "pending",
		EstimatedTime: "1-2 business days",
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"withdrawal": response,
	})
}

// @Summary Get transactions
// @Description Get user's wallet transaction history
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Transaction type" Enums(all, deposit, withdrawal, contest_fee, prize, bonus)
// @Param status query string false "Transaction status" Enums(all, pending, completed, failed, cancelled)
// @Param date_from query string false "From date (YYYY-MM-DD)"
// @Param date_to query string false "To date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /wallet/transactions [get]
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// For demo, return empty transactions
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"transactions": []models.WalletTransaction{},
		"total":        0,
		"user_id":      userID,
	})
}

// @Summary Get payment methods
// @Description Get user's saved payment methods
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /wallet/payment-methods [get]
func (h *WalletHandler) GetPaymentMethods(c *gin.Context) {
	userID := c.GetInt64("user_id")

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"payment_methods": []map[string]interface{}{},
		"user_id":         userID,
	})
}

// @Summary Add payment method
// @Description Add a new payment method
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "Payment method data"
// @Success 200 {object} map[string]interface{}
// @Router /wallet/payment-methods [post]
func (h *WalletHandler) AddPaymentMethod(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment method added successfully",
		"user_id": userID,
	})
}

// @Summary Get payment status
// @Description Get status of a payment transaction
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]interface{}
// @Router /payments/{id}/status [get]
func (h *WalletHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("id")
	userID := c.GetInt64("user_id")

	var status string
	var amount float64
	err := h.db.QueryRow(`
		SELECT status, amount FROM payment_transactions 
		WHERE transaction_id = $1 AND user_id = $2`, paymentID, userID).Scan(&status, &amount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Payment not found",
			Code:    "PAYMENT_NOT_FOUND",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error",
			Code:    "DB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"payment_id": paymentID,
		"status":     status,
		"amount":     amount,
	})
}

// @Summary Get referral statistics
// @Description Get detailed referral statistics for the user
// @Tags Referrals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ReferralStats
// @Failure 500 {object} models.ErrorResponse
// @Router /referrals/my-stats [get]
func (h *WalletHandler) GetReferralStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	stats, err := h.referralService.GetUserReferralStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get referral statistics: " + err.Error(),
			Code:    "REFERRAL_STATS_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"referral_stats": stats,
	})
}

// @Summary Get referral history
// @Description Get paginated referral history for the user
// @Tags Referrals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(all, pending, completed, expired)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.ReferralHistoryResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /referrals/history [get]
func (h *WalletHandler) GetReferralHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	status := c.DefaultQuery("status", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate page and limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	referrals, total, err := h.referralService.GetUserReferralHistory(userID, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get referral history: " + err.Error(),
			Code:    "REFERRAL_HISTORY_ERROR",
		})
		return
	}

	response := models.ReferralHistoryResponse{
		Success:   true,
		Referrals: referrals,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Apply referral code
// @Description Apply a referral code (can only be done once per user)
// @Tags Referrals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ApplyReferralCodeRequest true "Referral code to apply"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /referrals/apply [post]
func (h *WalletHandler) ApplyReferralCode(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.ApplyReferralCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if req.ReferralCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Referral code is required",
			Code:    "REFERRAL_CODE_REQUIRED",
		})
		return
	}

	err := h.referralService.ApplyReferralCode(userID, req.ReferralCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "REFERRAL_APPLICATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Referral code applied successfully",
		"referral_code": req.ReferralCode,
	})
}

// @Summary Share referral code
// @Description Share referral code via different methods
// @Tags Referrals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ShareReferralRequest true "Sharing method and details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /referrals/share [post]
func (h *WalletHandler) ShareReferral(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.ShareReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get user's referral code
	var referralCode string
	err := h.db.QueryRow("SELECT referral_code FROM users WHERE id = $1", userID).Scan(&referralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get referral code",
			Code:    "DB_ERROR",
		})
		return
	}

	// Generate sharing URL
	shareURL := "https://fantasy-esports.com/signup?ref=" + referralCode
	
	// Default sharing message
	defaultMessage := "Join Fantasy Esports with my referral code " + referralCode + " and get bonus rewards! " + shareURL

	message := req.Message
	if message == "" {
		message = defaultMessage
	}

	response := gin.H{
		"success":      true,
		"referral_code": referralCode,
		"share_url":     shareURL,
		"message":       message,
		"method":        req.Method,
	}

	// Add method-specific data
	switch req.Method {
	case "whatsapp":
		whatsappURL := "https://wa.me/?text=" + message
		response["whatsapp_url"] = whatsappURL
	case "sms":
		response["sms_message"] = message
		if len(req.Contacts) > 0 {
			response["contacts"] = req.Contacts
		}
	case "email":
		response["email_subject"] = "Join Fantasy Esports!"
		response["email_body"] = message
		if len(req.Contacts) > 0 {
			response["contacts"] = req.Contacts
		}
	case "copy":
		response["copy_text"] = message
	default:
		response["share_text"] = message
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get referral leaderboard
// @Description Get top referrers leaderboard
// @Tags Referrals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of top referrers to return" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /referrals/leaderboard [get]
func (h *WalletHandler) GetReferralLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	leaderboard, err := h.referralService.GetReferralLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get referral leaderboard: " + err.Error(),
			Code:    "REFERRAL_LEADERBOARD_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"leaderboard": leaderboard,
		"limit":       limit,
	})
}

// Internal method to trigger referral completion checks
func (h *WalletHandler) TriggerReferralCheck(userID int64, action string) {
	err := h.referralService.CheckAndCompleteReferral(userID, action)
	if err != nil {
		// Log error but don't fail the main operation
		// In production, you might want to queue this for retry
	}
}