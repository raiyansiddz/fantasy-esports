package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID               int64     `json:"id" db:"id"`
	Mobile           string    `json:"mobile" db:"mobile" validate:"required"`
	Email            *string   `json:"email" db:"email"`
	PasswordHash     *string   `json:"-" db:"password_hash"`
	FirstName        *string   `json:"first_name" db:"first_name"`
	LastName         *string   `json:"last_name" db:"last_name"`
	DateOfBirth      *time.Time `json:"date_of_birth" db:"date_of_birth"`
	Gender           *string   `json:"gender" db:"gender"`
	AvatarURL        *string   `json:"avatar_url" db:"avatar_url"`
	IsVerified       bool      `json:"is_verified" db:"is_verified"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	AccountStatus    string    `json:"account_status" db:"account_status"`
	KYCStatus        string    `json:"kyc_status" db:"kyc_status"`
	ReferralCode     string    `json:"referral_code" db:"referral_code"`
	ReferredByCode   *string   `json:"referred_by_code" db:"referred_by_code"`
	State            *string   `json:"state" db:"state"`
	City             *string   `json:"city" db:"city"`
	Pincode          *string   `json:"pincode" db:"pincode"`
	LastLoginAt      *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type KYCDocument struct {
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	DocumentType   string    `json:"document_type" db:"document_type"`
	DocumentFrontURL string  `json:"document_front_url" db:"document_front_url"`
	DocumentBackURL  *string `json:"document_back_url" db:"document_back_url"`
	DocumentNumber   *string `json:"document_number" db:"document_number"`
	AdditionalData   *string `json:"additional_data" db:"additional_data"`
	Status          string    `json:"status" db:"status"`
	VerifiedAt      *time.Time `json:"verified_at" db:"verified_at"`
	VerifiedBy      *int64    `json:"verified_by" db:"verified_by"`
	RejectionReason *string   `json:"rejection_reason" db:"rejection_reason"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type AuthRequest struct {
	Mobile      string  `json:"mobile" validate:"required"`
	CountryCode string  `json:"country_code" validate:"required"`
	ReferralCode *string `json:"referral_code"`
	DeviceID    string  `json:"device_id" validate:"required"`
	AppVersion  string  `json:"app_version" validate:"required"`
	Platform    string  `json:"platform" validate:"required"`
}

type OTPVerifyRequest struct {
	SessionID   string      `json:"session_id" validate:"required"`
	OTP         string      `json:"otp" validate:"required"`
	DeviceInfo  DeviceInfo  `json:"device_info" validate:"required"`
	ProfileData *ProfileData `json:"profile_data"`
	ReferralCode *string     `json:"referral_code"`
}

type DeviceInfo struct {
	Platform   string `json:"platform" validate:"required"`
	DeviceID   string `json:"device_id" validate:"required"`
	AppVersion string `json:"app_version" validate:"required"`
	FCMToken   *string `json:"fcm_token"`
}

type ProfileData struct {
	FirstName   string    `json:"first_name" validate:"required"`
	LastName    string    `json:"last_name" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
	State       string    `json:"state" validate:"required"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
	IsNewUser    bool   `json:"is_new_user"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code"`
}

type OTPSession struct {
	SessionID   string    `json:"session_id"`
	Mobile      string    `json:"mobile"`
	OTP         string    `json:"otp"`
	IsNewUser   bool      `json:"is_new_user"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func GenerateReferralCode() string {
	return uuid.New().String()[:8]
}