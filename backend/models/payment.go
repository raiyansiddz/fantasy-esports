package models

import (
	"encoding/json"
	"time"
)

// PaymentGatewayConfig stores configuration for payment gateways
type PaymentGatewayConfig struct {
	ID            int64     `json:"id" db:"id"`
	Gateway       string    `json:"gateway" db:"gateway"`                   // razorpay, phonepe
	Key1          string    `json:"key1" db:"key1"`                         // key_id for razorpay, client_id for phonepe
	Key2          string    `json:"key2" db:"key2"`                         // key_secret for razorpay, client_secret for phonepe
	ClientVersion string    `json:"client_version,omitempty" db:"client_version"` // For phonepe
	IsLive        bool      `json:"is_live" db:"is_live"`                   // Test or production environment
	Enabled       bool      `json:"enabled" db:"enabled"`                   // Gateway availability
	Currency      string    `json:"currency" db:"currency"`                 // Default currency
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Enhanced PaymentTransaction for both gateways
type PaymentTransactionEnhanced struct {
	ID                   int64           `json:"id" db:"id"`
	UserID               int64           `json:"user_id" db:"user_id"`
	TransactionID        string          `json:"transaction_id" db:"transaction_id"`                // Our internal transaction ID
	Gateway              string          `json:"gateway" db:"gateway"`                              // razorpay, phonepe
	GatewayTransactionID *string         `json:"gateway_transaction_id" db:"gateway_transaction_id"` // Gateway's transaction ID
	Amount               float64         `json:"amount" db:"amount"`
	Currency             string          `json:"currency" db:"currency"`
	Type                 string          `json:"type" db:"type"`                                    // add_money, withdraw
	Status               string          `json:"status" db:"status"`                               // pending, completed, failed
	RetryCount           int             `json:"retry_count" db:"retry_count"`                     // Number of retry attempts
	GatewayResponse      json.RawMessage `json:"gateway_response" db:"gateway_response"`           // Full gateway response
	FailureReason        *string         `json:"failure_reason" db:"failure_reason"`               // Reason for failure
	Notes                *string         `json:"notes" db:"notes"`                                 // Additional notes
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	CompletedAt          *time.Time      `json:"completed_at" db:"completed_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// WebhookLog stores webhook events from payment gateways
type WebhookLog struct {
	ID                int64           `json:"id" db:"id"`
	Gateway           string          `json:"gateway" db:"gateway"`
	EventType         string          `json:"event_type" db:"event_type"`
	TransactionID     *string         `json:"transaction_id" db:"transaction_id"`
	Payload           json.RawMessage `json:"payload" db:"payload"`
	ProcessingStatus  string          `json:"processing_status" db:"processing_status"` // pending, processed, failed
	ProcessingMessage *string         `json:"processing_message" db:"processing_message"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	ProcessedAt       *time.Time      `json:"processed_at" db:"processed_at"`
}

// PaymentMethod stores saved payment methods for users
type PaymentMethod struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Gateway    string    `json:"gateway" db:"gateway"`
	MethodType string    `json:"method_type" db:"method_type"` // card, upi, netbanking, wallet
	MethodData json.RawMessage `json:"method_data" db:"method_data"` // Encrypted method details
	IsDefault  bool      `json:"is_default" db:"is_default"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// RefundTransaction stores refund information
type RefundTransaction struct {
	ID                    int64           `json:"id" db:"id"`
	PaymentTransactionID  int64           `json:"payment_transaction_id" db:"payment_transaction_id"`
	RefundTransactionID   string          `json:"refund_transaction_id" db:"refund_transaction_id"`
	GatewayRefundID       *string         `json:"gateway_refund_id" db:"gateway_refund_id"`
	Amount                float64         `json:"amount" db:"amount"`
	Reason                string          `json:"reason" db:"reason"`
	Status                string          `json:"status" db:"status"` // pending, completed, failed
	GatewayResponse       json.RawMessage `json:"gateway_response" db:"gateway_response"`
	ProcessedBy           *int64          `json:"processed_by" db:"processed_by"` // Admin user ID
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	CompletedAt           *time.Time      `json:"completed_at" db:"completed_at"`
}

// PaymentAnalytics stores aggregated payment statistics
type PaymentAnalytics struct {
	ID                   int64     `json:"id" db:"id"`
	Date                 time.Time `json:"date" db:"date"`
	Gateway              string    `json:"gateway" db:"gateway"`
	TotalTransactions    int64     `json:"total_transactions" db:"total_transactions"`
	SuccessfulTransactions int64   `json:"successful_transactions" db:"successful_transactions"`
	FailedTransactions   int64     `json:"failed_transactions" db:"failed_transactions"`
	TotalAmount          float64   `json:"total_amount" db:"total_amount"`
	SuccessfulAmount     float64   `json:"successful_amount" db:"successful_amount"`
	AverageAmount        float64   `json:"average_amount" db:"average_amount"`
	SuccessRate          float64   `json:"success_rate" db:"success_rate"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// AdminPaymentConfig stores admin-configurable payment settings
type AdminPaymentConfig struct {
	ID                    int64     `json:"id" db:"id"`
	MinDepositAmount      float64   `json:"min_deposit_amount" db:"min_deposit_amount"`
	MaxDepositAmount      float64   `json:"max_deposit_amount" db:"max_deposit_amount"`
	MinWithdrawAmount     float64   `json:"min_withdraw_amount" db:"min_withdraw_amount"`
	MaxWithdrawAmount     float64   `json:"max_withdraw_amount" db:"max_withdraw_amount"`
	DailyDepositLimit     float64   `json:"daily_deposit_limit" db:"daily_deposit_limit"`
	DailyWithdrawLimit    float64   `json:"daily_withdraw_limit" db:"daily_withdraw_limit"`
	TransactionFeePercent float64   `json:"transaction_fee_percent" db:"transaction_fee_percent"`
	MaintenanceMode       bool      `json:"maintenance_mode" db:"maintenance_mode"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy             int64     `json:"updated_by" db:"updated_by"` // Admin user ID
}

// DTO models for API responses

// PaymentGatewayConfigResponse masks sensitive keys in API responses
type PaymentGatewayConfigResponse struct {
	ID            int64     `json:"id"`
	Gateway       string    `json:"gateway"`
	Key1Masked    string    `json:"key1_masked"`    // Shows only first 4 and last 4 chars
	Key2Masked    string    `json:"key2_masked"`    // Shows only first 4 and last 4 chars
	ClientVersion string    `json:"client_version,omitempty"`
	IsLive        bool      `json:"is_live"`
	Enabled       bool      `json:"enabled"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MaskSensitiveData creates a response-safe version of PaymentGatewayConfig
func (p *PaymentGatewayConfig) MaskSensitiveData() *PaymentGatewayConfigResponse {
	return &PaymentGatewayConfigResponse{
		ID:            p.ID,
		Gateway:       p.Gateway,
		Key1Masked:    maskKey(p.Key1),
		Key2Masked:    maskKey(p.Key2),
		ClientVersion: p.ClientVersion,
		IsLive:        p.IsLive,
		Enabled:       p.Enabled,
		Currency:      p.Currency,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// Helper function to mask sensitive keys
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// PaymentTransactionSummary provides summary view for payment transactions
type PaymentTransactionSummary struct {
	TransactionID        string    `json:"transaction_id"`
	Gateway              string    `json:"gateway"`
	Amount               float64   `json:"amount"`
	Currency             string    `json:"currency"`
	Status               string    `json:"status"`
	Type                 string    `json:"type"`
	CreatedAt            time.Time `json:"created_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
}

// PaymentDashboardStats provides dashboard statistics for admin
type PaymentDashboardStats struct {
	TotalTransactions     int64   `json:"total_transactions"`
	SuccessfulTransactions int64   `json:"successful_transactions"`
	FailedTransactions    int64   `json:"failed_transactions"`
	PendingTransactions   int64   `json:"pending_transactions"`
	TotalAmount           float64 `json:"total_amount"`
	SuccessfulAmount      float64 `json:"successful_amount"`
	SuccessRate           float64 `json:"success_rate"`
	AverageTransactionAmount float64 `json:"average_transaction_amount"`
	GatewayBreakdown      map[string]interface{} `json:"gateway_breakdown"`
}

// Constants for payment statuses
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusFailed    = "failed"
	PaymentStatusRefunded  = "refunded"
	PaymentStatusCancelled = "cancelled"
)

// Constants for payment types
const (
	PaymentTypeAddMoney  = "add_money"
	PaymentTypeWithdraw  = "withdraw"
	PaymentTypeRefund    = "refund"
)

// Constants for gateways
const (
	GatewayRazorpay = "razorpay"
	GatewayPhonePe  = "phonepe"
)