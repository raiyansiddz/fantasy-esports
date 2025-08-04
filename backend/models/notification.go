package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NotificationChannel represents different notification channels
type NotificationChannel string

const (
	ChannelSMS      NotificationChannel = "sms"
	ChannelEmail    NotificationChannel = "email"
	ChannelPush     NotificationChannel = "push"
	ChannelWhatsApp NotificationChannel = "whatsapp"
)

// NotificationProvider represents different providers for each channel
type NotificationProvider string

const (
	// SMS Providers
	ProviderFast2SMS NotificationProvider = "fast2sms"
	
	// Email Providers
	ProviderSMTP     NotificationProvider = "smtp"
	ProviderSES      NotificationProvider = "amazon_ses"
	ProviderMailchimp NotificationProvider = "mailchimp"
	
	// Push Providers
	ProviderFCM      NotificationProvider = "firebase_fcm"
	ProviderOneSignal NotificationProvider = "onesignal"
	
	// WhatsApp Providers
	ProviderWhatsAppCloud NotificationProvider = "whatsapp_cloud"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
)

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID          int64               `json:"id" db:"id"`
	Name        string              `json:"name" db:"name"`
	Channel     NotificationChannel `json:"channel" db:"channel"`
	Provider    NotificationProvider `json:"provider" db:"provider"`
	Subject     *string             `json:"subject" db:"subject"`
	Body        string              `json:"body" db:"body"`
	Variables   TemplateVariables   `json:"variables" db:"variables"`
	IsDLTApproved bool              `json:"is_dlt_approved" db:"is_dlt_approved"`
	DLTTemplateID *string           `json:"dlt_template_id" db:"dlt_template_id"`
	IsActive    bool                `json:"is_active" db:"is_active"`
	CreatedBy   int64               `json:"created_by" db:"created_by"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" db:"updated_at"`
}

// TemplateVariables represents template variables that can be replaced
type TemplateVariables []string

// Value implements driver.Valuer interface for database storage
func (tv TemplateVariables) Value() (driver.Value, error) {
	return json.Marshal(tv)
}

// Scan implements sql.Scanner interface for database retrieval
func (tv *TemplateVariables) Scan(value interface{}) error {
	if value == nil {
		*tv = TemplateVariables{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TemplateVariables", value)
	}
	
	return json.Unmarshal(bytes, tv)
}

// NotificationLog represents a sent notification record
type NotificationLog struct {
	ID          int64               `json:"id" db:"id"`
	TemplateID  *int64              `json:"template_id" db:"template_id"`
	Channel     NotificationChannel `json:"channel" db:"channel"`
	Provider    NotificationProvider `json:"provider" db:"provider"`
	Recipient   string              `json:"recipient" db:"recipient"`
	Subject     *string             `json:"subject" db:"subject"`
	Body        string              `json:"body" db:"body"`
	Status      NotificationStatus  `json:"status" db:"status"`
	ProviderID  *string             `json:"provider_id" db:"provider_id"`
	Response    *string             `json:"response" db:"response"`
	ErrorMsg    *string             `json:"error_msg" db:"error_msg"`
	RetryCount  int                 `json:"retry_count" db:"retry_count"`
	UserID      *int64              `json:"user_id" db:"user_id"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	SentAt      *time.Time          `json:"sent_at" db:"sent_at"`
	DeliveredAt *time.Time          `json:"delivered_at" db:"delivered_at"`
}

// NotificationConfig represents configuration for notification providers
type NotificationConfig struct {
	ID          int64                `json:"id" db:"id"`
	Provider    NotificationProvider `json:"provider" db:"provider"`
	Channel     NotificationChannel  `json:"channel" db:"channel"`
	ConfigKey   string               `json:"config_key" db:"config_key"`
	ConfigValue string               `json:"config_value" db:"config_value"`
	IsActive    bool                 `json:"is_active" db:"is_active"`
	UpdatedBy   int64                `json:"updated_by" db:"updated_by"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

// SendNotificationRequest represents a request to send notification
type SendNotificationRequest struct {
	Channel    NotificationChannel    `json:"channel" validate:"required"`
	Provider   *NotificationProvider  `json:"provider,omitempty"`
	TemplateID *int64                 `json:"template_id,omitempty"`
	Recipient  string                 `json:"recipient" validate:"required"`
	Subject    *string                `json:"subject,omitempty"`
	Body       *string                `json:"body,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	UserID     *int64                 `json:"user_id,omitempty"`
}

// BulkNotificationRequest represents a bulk notification request
type BulkNotificationRequest struct {
	Channel    NotificationChannel    `json:"channel" validate:"required"`
	Provider   *NotificationProvider  `json:"provider,omitempty"`
	TemplateID *int64                 `json:"template_id,omitempty"`
	Recipients []string               `json:"recipients" validate:"required"`
	Subject    *string                `json:"subject,omitempty"`
	Body       *string                `json:"body,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	UserFilter *UserFilter            `json:"user_filter,omitempty"`
}

// UserFilter represents filters for targeting users
type UserFilter struct {
	KYCStatus       *string   `json:"kyc_status,omitempty"`
	AccountStatus   *string   `json:"account_status,omitempty"`
	InactiveDays    *int      `json:"inactive_days,omitempty"`
	WalletBalance   *float64  `json:"wallet_balance,omitempty"`
	LastContestDays *int      `json:"last_contest_days,omitempty"`
	States          []string  `json:"states,omitempty"`
}

// NotificationResponse represents the response after sending notification
type NotificationResponse struct {
	Success    bool                `json:"success"`
	LogID      int64               `json:"log_id"`
	ProviderID *string             `json:"provider_id,omitempty"`
	Status     NotificationStatus  `json:"status"`
	Message    string              `json:"message"`
	Error      *string             `json:"error,omitempty"`
}

// TemplateCreateRequest represents a request to create a template
type TemplateCreateRequest struct {
	Name          string                 `json:"name" validate:"required"`
	Channel       NotificationChannel    `json:"channel" validate:"required"`
	Provider      NotificationProvider   `json:"provider" validate:"required"`
	Subject       *string                `json:"subject,omitempty"`
	Body          string                 `json:"body" validate:"required"`
	Variables     []string               `json:"variables"`
	IsDLTApproved bool                   `json:"is_dlt_approved"`
	DLTTemplateID *string                `json:"dlt_template_id,omitempty"`
}

// TemplateUpdateRequest represents a request to update a template
type TemplateUpdateRequest struct {
	Name          *string                `json:"name,omitempty"`
	Subject       *string                `json:"subject,omitempty"`
	Body          *string                `json:"body,omitempty"`
	Variables     []string               `json:"variables,omitempty"`
	IsDLTApproved *bool                  `json:"is_dlt_approved,omitempty"`
	DLTTemplateID *string                `json:"dlt_template_id,omitempty"`
	IsActive      *bool                  `json:"is_active,omitempty"`
}

// ConfigUpdateRequest represents a request to update notification config
type ConfigUpdateRequest struct {
	Provider    NotificationProvider `json:"provider" validate:"required"`
	Channel     NotificationChannel  `json:"channel" validate:"required"`
	ConfigKey   string               `json:"config_key" validate:"required"`
	ConfigValue string               `json:"config_value" validate:"required"`
	IsActive    bool                 `json:"is_active"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent      int64 `json:"total_sent"`
	TotalDelivered int64 `json:"total_delivered"`
	TotalFailed    int64 `json:"total_failed"`
	TotalPending   int64 `json:"total_pending"`
	DeliveryRate   float64 `json:"delivery_rate"`
	FailureRate    float64 `json:"failure_rate"`
}

// ChannelStats represents statistics per channel
type ChannelStats struct {
	Channel        NotificationChannel `json:"channel"`
	Provider       NotificationProvider `json:"provider"`
	Stats          NotificationStats   `json:"stats"`
	LastSent       *time.Time          `json:"last_sent"`
	AvgResponseTime float64            `json:"avg_response_time_ms"`
}