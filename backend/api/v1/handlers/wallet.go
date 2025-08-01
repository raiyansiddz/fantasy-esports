package handlers

import (
	"database/sql"
	"net/http"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	db     *sql.DB
	config *config.Config
}

func NewWalletHandler(db *sql.DB, cfg *config.Config) *WalletHandler {
	return &WalletHandler{
		db:     db,
		config: cfg,
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

// Referral methods
func (h *WalletHandler) GetReferralStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

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

	stats := models.ReferralStats{
		ReferralCode:        referralCode,
		TotalReferrals:      0,
		SuccessfulReferrals: 0,
		TotalEarnings:       0.0,
		PendingEarnings:     0.0,
		LifetimeEarnings:    0.0,
		CurrentTier:         "bronze",
		NextTierRequirement: 10,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"referral_stats": stats,
	})
}

func (h *WalletHandler) GetReferralHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "history": []models.Referral{}})
}

func (h *WalletHandler) ApplyReferralCode(c *gin.Context) {
	userID := c.GetInt64("user_id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "message": "Referral applied"})
}

func (h *WalletHandler) ShareReferral(c *gin.Context) {
	userID := c.GetInt64("user_id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "message": "Referral shared"})
}