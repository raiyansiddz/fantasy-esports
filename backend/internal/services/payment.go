package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/internal/integrations"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	db           *sql.DB
	razorpay     *integrations.RazorpayClient
	phonepe      *integrations.PhonePeClient
	configService *ConfigService
}

func NewPaymentService(db *sql.DB) *PaymentService {
	configService := NewConfigService(db)
	return &PaymentService{
		db:            db,
		razorpay:      integrations.NewRazorpayClient(),
		phonepe:       integrations.NewPhonePeClient(),
		configService: configService,
	}
}

// CreatePaymentOrder creates a new payment order
func (s *PaymentService) CreatePaymentOrder(userID int64, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Get gateway configuration
	config, err := s.configService.GetGatewayConfig(req.Gateway)
	if err != nil {
		logger.Error("Failed to get gateway config", "gateway", req.Gateway, "error", err)
		return nil, fmt.Errorf("gateway not configured: %v", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("gateway %s is currently disabled", req.Gateway)
	}

	// Generate unique transaction ID
	transactionID := uuid.New().String()

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	// Create payment transaction record
	paymentTx := &models.PaymentTransaction{
		UserID:        userID,
		TransactionID: transactionID,
		Gateway:       req.Gateway,
		Amount:        req.Amount,
		Currency:      currency,
		Type:          "add_money",
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	// Insert into database
	query := `
		INSERT INTO payment_transactions (user_id, transaction_id, gateway, amount, currency, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	
	err = s.db.QueryRow(query, paymentTx.UserID, paymentTx.TransactionID, paymentTx.Gateway,
		paymentTx.Amount, paymentTx.Currency, paymentTx.Type, paymentTx.Status, paymentTx.CreatedAt).Scan(&paymentTx.ID)
	if err != nil {
		logger.Error("Failed to create payment transaction", "error", err)
		return nil, fmt.Errorf("failed to create payment transaction: %v", err)
	}

	// Create order with selected gateway
	var paymentData map[string]interface{}
	
	switch req.Gateway {
	case "razorpay":
		paymentData, err = s.razorpay.CreateOrder(config, req.Amount, currency, transactionID)
	case "phonepe":
		paymentData, err = s.phonepe.InitiatePayment(config, req.Amount, currency, transactionID, userID)
	default:
		return nil, fmt.Errorf("unsupported gateway: %s", req.Gateway)
	}

	if err != nil {
		// Update transaction status to failed
		s.updateTransactionStatus(transactionID, "failed", nil)
		logger.Error("Failed to create payment order", "gateway", req.Gateway, "error", err)
		return nil, fmt.Errorf("failed to create payment order: %v", err)
	}

	// Update transaction with gateway response
	gatewayResponse, _ := json.Marshal(paymentData)
	s.updateTransactionGatewayResponse(transactionID, gatewayResponse)

	logger.Info("Payment order created successfully", "transaction_id", transactionID, "gateway", req.Gateway, "amount", req.Amount)

	return &CreateOrderResponse{
		TransactionID: transactionID,
		Gateway:       req.Gateway,
		Amount:        req.Amount,
		Currency:      currency,
		PaymentData:   paymentData,
	}, nil
}

// VerifyPayment verifies payment with gateway
func (s *PaymentService) VerifyPayment(userID int64, req *VerifyPaymentRequest) (*VerifyPaymentResponse, error) {
	// Get transaction from database
	tx, err := s.getPaymentTransaction(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %v", err)
	}

	if tx.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to transaction")
	}

	// Get gateway configuration
	config, err := s.configService.GetGatewayConfig(req.Gateway)
	if err != nil {
		return nil, fmt.Errorf("gateway not configured: %v", err)
	}

	// Verify payment with gateway
	var verified bool
	var gatewayTxID string
	var gatewayResponse map[string]interface{}

	switch req.Gateway {
	case "razorpay":
		verified, gatewayTxID, gatewayResponse, err = s.razorpay.VerifyPayment(config, req.GatewayData)
	case "phonepe":
		verified, gatewayTxID, gatewayResponse, err = s.phonepe.VerifyPayment(config, req.GatewayData)
	default:
		return nil, fmt.Errorf("unsupported gateway: %s", req.Gateway)
	}

	if err != nil {
		logger.Error("Payment verification failed", "transaction_id", req.TransactionID, "error", err)
		s.updateTransactionStatus(req.TransactionID, "failed", gatewayResponse)
		return &VerifyPaymentResponse{
			Success:       false,
			TransactionID: req.TransactionID,
			Status:        "failed",
			Amount:        tx.Amount,
			Message:       "Payment verification failed",
		}, nil
	}

	status := "failed"
	message := "Payment verification failed"
	
	if verified {
		status = "completed"
		message = "Payment completed successfully"
		
		// Update wallet balance
		err = s.updateWalletBalance(userID, tx.Amount, "deposit", req.TransactionID)
		if err != nil {
			logger.Error("Failed to update wallet balance", "user_id", userID, "error", err)
			// Don't return error here as payment is verified, just log it
		}
		
		// Trigger referral completion check
		s.triggerReferralCheck(userID, tx.Amount)
	}

	// Update transaction status
	s.updateTransactionStatus(req.TransactionID, status, gatewayResponse)
	s.updateTransactionGatewayID(req.TransactionID, gatewayTxID)

	logger.Info("Payment verification completed", "transaction_id", req.TransactionID, "status", status, "verified", verified)

	return &VerifyPaymentResponse{
		Success:       verified,
		TransactionID: req.TransactionID,
		Status:        status,
		Amount:        tx.Amount,
		Message:       message,
	}, nil
}

// GetPaymentStatus gets payment status by transaction ID
func (s *PaymentService) GetPaymentStatus(userID int64, transactionID string) (*PaymentStatusResponse, error) {
	tx, err := s.getPaymentTransaction(transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %v", err)
	}

	if tx.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to transaction")
	}

	var completedAt *string
	if tx.CompletedAt != nil {
		completedAtStr := tx.CompletedAt.Format(time.RFC3339)
		completedAt = &completedAtStr
	}

	return &PaymentStatusResponse{
		TransactionID:        tx.TransactionID,
		Gateway:              tx.Gateway,
		GatewayTransactionID: tx.GatewayTransactionID,
		Amount:               tx.Amount,
		Currency:             tx.Currency,
		Status:               tx.Status,
		CreatedAt:            tx.CreatedAt.Format(time.RFC3339),
		CompletedAt:          completedAt,
	}, nil
}

// Admin methods for gateway configuration

// GetGatewayConfigs gets all gateway configurations
func (s *PaymentService) GetGatewayConfigs() ([]models.PaymentGatewayConfig, error) {
	return s.configService.GetAllGatewayConfigs()
}

// UpdateGatewayConfig updates gateway configuration
func (s *PaymentService) UpdateGatewayConfig(gateway string, req *UpdateGatewayConfigRequest) error {
	config := &models.PaymentGatewayConfig{
		Gateway:  gateway,
		Key1:     req.Key1,
		Key2:     req.Key2,
		IsLive:   req.IsLive,
		Enabled:  req.Enabled,
		Currency: req.Currency,
	}

	return s.configService.UpdateGatewayConfig(config)
}

// ToggleGatewayStatus enables/disables gateway
func (s *PaymentService) ToggleGatewayStatus(gateway string, enabled bool) error {
	return s.configService.ToggleGatewayStatus(gateway, enabled)
}

// GetTransactionLogs gets payment transaction logs
func (s *PaymentService) GetTransactionLogs(page, limit int, gateway, status string) ([]models.PaymentTransaction, int64, error) {
	offset := (page - 1) * limit
	
	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if gateway != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND gateway = $%d", argCount)
		args = append(args, gateway)
	}

	if status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM payment_transactions " + whereClause
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transaction count: %v", err)
	}

	// Get transactions
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount
	
	query := fmt.Sprintf(`
		SELECT id, user_id, transaction_id, gateway, gateway_transaction_id, amount, currency, type, status, created_at, completed_at
		FROM payment_transactions %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, limitArg, offsetArg)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %v", err)
	}
	defer rows.Close()

	var transactions []models.PaymentTransaction
	for rows.Next() {
		var tx models.PaymentTransaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.TransactionID, &tx.Gateway,
			&tx.GatewayTransactionID, &tx.Amount, &tx.Currency, &tx.Type,
			&tx.Status, &tx.CreatedAt, &tx.CompletedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// Helper methods

func (s *PaymentService) getPaymentTransaction(transactionID string) (*models.PaymentTransaction, error) {
	query := `
		SELECT id, user_id, transaction_id, gateway, gateway_transaction_id, amount, currency, type, status, created_at, completed_at
		FROM payment_transactions
		WHERE transaction_id = $1`
	
	var tx models.PaymentTransaction
	err := s.db.QueryRow(query, transactionID).Scan(
		&tx.ID, &tx.UserID, &tx.TransactionID, &tx.Gateway, &tx.GatewayTransactionID,
		&tx.Amount, &tx.Currency, &tx.Type, &tx.Status, &tx.CreatedAt, &tx.CompletedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &tx, nil
}

func (s *PaymentService) updateTransactionStatus(transactionID, status string, gatewayResponse map[string]interface{}) error {
	var gatewayResponseBytes []byte
	if gatewayResponse != nil {
		gatewayResponseBytes, _ = json.Marshal(gatewayResponse)
	}

	completedAt := sql.NullTime{}
	if status == "completed" || status == "failed" {
		completedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	query := `
		UPDATE payment_transactions 
		SET status = $1, gateway_response = $2, completed_at = $3
		WHERE transaction_id = $4`
	
	_, err := s.db.Exec(query, status, gatewayResponseBytes, completedAt, transactionID)
	return err
}

func (s *PaymentService) updateTransactionGatewayResponse(transactionID string, gatewayResponse []byte) error {
	query := `UPDATE payment_transactions SET gateway_response = $1 WHERE transaction_id = $2`
	_, err := s.db.Exec(query, gatewayResponse, transactionID)
	return err
}

func (s *PaymentService) updateTransactionGatewayID(transactionID, gatewayTxID string) error {
	query := `UPDATE payment_transactions SET gateway_transaction_id = $1 WHERE transaction_id = $2`
	_, err := s.db.Exec(query, gatewayTxID, transactionID)
	return err
}

func (s *PaymentService) updateWalletBalance(userID int64, amount float64, transactionType, referenceID string) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Update user wallet - add to deposit balance
	query := `
		UPDATE user_wallets 
		SET deposit_balance = deposit_balance + $1,
			total_balance = total_balance + $1,
			updated_at = NOW()
		WHERE user_id = $2`
	
	_, err = tx.Exec(query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %v", err)
	}

	// Create wallet transaction record
	walletTxQuery := `
		INSERT INTO wallet_transactions (user_id, transaction_type, amount, balance_type, description, reference_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	
	_, err = tx.Exec(walletTxQuery, userID, transactionType, amount, "deposit", 
		"Money added via "+transactionType, referenceID, "completed")
	if err != nil {
		return fmt.Errorf("failed to create wallet transaction: %v", err)
	}

	return tx.Commit()
}

func (s *PaymentService) triggerReferralCheck(userID int64, amount float64) {
	// This would integrate with the existing referral service
	// For now, we'll just log it
	logger.Info("Triggering referral check", "user_id", userID, "amount", amount)
	
	// TODO: Call existing referral service CheckAndCompleteReferral method
	// referralService.CheckAndCompleteReferral(userID, amount)
}

// Request/Response types
type CreateOrderRequest struct {
	Amount   float64 `json:"amount"`
	Gateway  string  `json:"gateway"`
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
	TransactionID string                 `json:"transaction_id"`
	Gateway       string                 `json:"gateway"`
	GatewayData   map[string]interface{} `json:"gateway_data"`
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
	Key1     string `json:"key1"`
	Key2     string `json:"key2"`
	IsLive   bool   `json:"is_live"`
	Enabled  bool   `json:"enabled"`
	Currency string `json:"currency,omitempty"`
}