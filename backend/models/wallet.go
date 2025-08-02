package models

import (
	"time"
	"encoding/json"
)

type UserWallet struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	BonusBalance    float64   `json:"bonus_balance" db:"bonus_balance"`
	DepositBalance  float64   `json:"deposit_balance" db:"deposit_balance"`
	WinningBalance  float64   `json:"winning_balance" db:"winning_balance"`
	TotalBalance    float64   `json:"total_balance" db:"total_balance"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type WalletTransaction struct {
	ID              int64           `json:"id" db:"id"`
	UserID          int64           `json:"user_id" db:"user_id"`
	TransactionType string          `json:"transaction_type" db:"transaction_type"`
	Amount          float64         `json:"amount" db:"amount"`
	BalanceType     string          `json:"balance_type" db:"balance_type"`
	Description     *string         `json:"description" db:"description"`
	ReferenceID     *string         `json:"reference_id" db:"reference_id"`
	Status          string          `json:"status" db:"status"`
	GatewayResponse json.RawMessage `json:"gateway_response" db:"gateway_response"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	CompletedAt     *time.Time      `json:"completed_at" db:"completed_at"`
}

type PaymentTransaction struct {
	ID                    int64           `json:"id" db:"id"`
	UserID                int64           `json:"user_id" db:"user_id"`
	TransactionID         string          `json:"transaction_id" db:"transaction_id"`
	Gateway               string          `json:"gateway" db:"gateway"`
	GatewayTransactionID  *string         `json:"gateway_transaction_id" db:"gateway_transaction_id"`
	Amount                float64         `json:"amount" db:"amount"`
	Currency              string          `json:"currency" db:"currency"`
	Type                  string          `json:"type" db:"type"`
	Status                string          `json:"status" db:"status"`
	GatewayResponse       json.RawMessage `json:"gateway_response" db:"gateway_response"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	CompletedAt           *time.Time      `json:"completed_at" db:"completed_at"`
}

type Referral struct {
	ID                 int64      `json:"id" db:"id"`
	ReferrerUserID     int64      `json:"referrer_user_id" db:"referrer_user_id"`
	ReferredUserID     int64      `json:"referred_user_id" db:"referred_user_id"`
	ReferralCode       string     `json:"referral_code" db:"referral_code"`
	Status             string     `json:"status" db:"status"`
	RewardAmount       float64    `json:"reward_amount" db:"reward_amount"`
	CompletionCriteria string     `json:"completion_criteria" db:"completion_criteria"`
	CompletedAt        *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type DepositRequest struct {
	Amount        float64 `json:"amount" validate:"required,min=10"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	ReturnURL     string  `json:"return_url" validate:"required"`
	PromoCode     *string `json:"promo_code"`
}

type WithdrawRequest struct {
	Amount         float64        `json:"amount" validate:"required,min=100"`
	AccountType    string         `json:"account_type" validate:"required"`
	AccountDetails AccountDetails `json:"account_details" validate:"required"`
	OTP            string         `json:"otp" validate:"required"`
}

type AccountDetails struct {
	AccountNumber string `json:"account_number" validate:"required"`
	IFSC          string `json:"ifsc" validate:"required"`
	HolderName    string `json:"holder_name" validate:"required"`
	AccountType   string `json:"account_type" validate:"required"`
}

type WalletBalance struct {
	BonusBalance       float64    `json:"bonus_balance"`
	DepositBalance     float64    `json:"deposit_balance"`
	WinningBalance     float64    `json:"winning_balance"`
	TotalBalance       float64    `json:"total_balance"`
	WithdrawableBalance float64   `json:"withdrawable_balance"`
	BonusExpiry        *BonusExpiry `json:"bonus_expiry,omitempty"`
}

type BonusExpiry struct {
	Amount    float64   `json:"amount"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PaymentResponse struct {
	PaymentID       string    `json:"payment_id"`
	GatewayOrderID  string    `json:"gateway_order_id"`
	Amount          float64   `json:"amount"`
	GatewayURL      string    `json:"gateway_url"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type WithdrawResponse struct {
	WithdrawalID    string  `json:"withdrawal_id"`
	Amount          float64 `json:"amount"`
	ProcessingFee   float64 `json:"processing_fee"`
	NetAmount       float64 `json:"net_amount"`
	Status          string  `json:"status"`
	EstimatedTime   string  `json:"estimated_time"`
}

type ReferralStats struct {
	ReferralCode       string  `json:"referral_code"`
	TotalReferrals     int     `json:"total_referrals"`
	SuccessfulReferrals int    `json:"successful_referrals"`
	TotalEarnings      float64 `json:"total_earnings"`
	PendingEarnings    float64 `json:"pending_earnings"`
	LifetimeEarnings   float64 `json:"lifetime_earnings"`
	CurrentTier        string  `json:"current_tier"`
	NextTierRequirement int    `json:"next_tier_requirement"`
}

type ApplyReferralCodeRequest struct {
	ReferralCode string `json:"referral_code" validate:"required"`
}

type ShareReferralRequest struct {
	Method   string   `json:"method" validate:"required"`
	Contacts []string `json:"contacts,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type ReferralHistoryResponse struct {
	Success    bool            `json:"success"`
	Referrals  []Referral      `json:"referrals"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

type ReferralLeaderboardEntry struct {
	Rank               int     `json:"rank"`
	UserID             int64   `json:"user_id"`
	Name               string  `json:"name"`
	ReferralCode       string  `json:"referral_code"`
	TotalReferrals     int     `json:"total_referrals"`
	SuccessfulReferrals int    `json:"successful_referrals"`
	TotalEarnings      float64 `json:"total_earnings"`
	CurrentTier        string  `json:"current_tier"`
}