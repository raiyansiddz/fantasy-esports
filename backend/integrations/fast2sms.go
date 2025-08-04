package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fantasy-esports-backend/models"
)

// Fast2SMSNotifier implements SMS notifications via Fast2SMS
type Fast2SMSNotifier struct{}

// Fast2SMSRequest represents the request structure for Fast2SMS API
type Fast2SMSRequest struct {
	Route     string `json:"route"`
	Message   string `json:"message"`
	Numbers   string `json:"numbers"`
	SenderID  string `json:"sender_id"`
	Language  string `json:"language"`
}

// Fast2SMSResponse represents the response from Fast2SMS API
type Fast2SMSResponse struct {
	Return    bool   `json:"return"`
	RequestID string `json:"request_id"`
	Message   []string `json:"message"`
}

// NewFast2SMSNotifier creates a new Fast2SMS notifier
func NewFast2SMSNotifier() *Fast2SMSNotifier {
	return &Fast2SMSNotifier{}
}

// Send sends SMS via Fast2SMS
func (f *Fast2SMSNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := f.ValidateConfig(config); err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
			Error:   &errMsg,
		}, err
	}

	// Prepare request body
	body := request.Body
	if body == nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Message body is required",
			Error:   stringPtr("Message body is required"),
		}, NewNotificationError(ErrInvalidRecipient, "Message body is required", nil)
	}

	// Clean phone number
	phoneNumber := cleanPhoneNumber(request.Recipient)
	if !isValidIndianMobile(phoneNumber) {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Invalid mobile number format",
			Error:   stringPtr("Invalid mobile number format"),
		}, NewNotificationError(ErrInvalidRecipient, "Invalid mobile number format", nil)
	}

	// Create API request
	apiRequest := Fast2SMSRequest{
		Route:     "v3",
		Message:   *body,
		Numbers:   phoneNumber,
		SenderID:  config["sender_id"],
		Language:  "english",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(apiRequest)
	if err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to encode request",
			Error:   &errMsg,
		}, NewNotificationError(ErrTemplateParsing, "Failed to encode request", err)
	}

	// Create HTTP request
	baseURL := config["base_url"]
	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to create HTTP request",
			Error:   &err.Error(),
		}, NewNotificationError(ErrNetworkError, "Failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", config["api_key"])

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Network error",
			Error:   &err.Error(),
		}, NewNotificationError(ErrNetworkError, "Network error", err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResponse Fast2SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to parse response",
			Error:   &err.Error(),
		}, NewNotificationError(ErrNetworkError, "Failed to parse response", err)
	}

	// Check if successful
	if apiResponse.Return && resp.StatusCode == 200 {
		return &models.NotificationResponse{
			Success:    true,
			ProviderID: &apiResponse.RequestID,
			Status:     models.StatusSent,
			Message:    "SMS sent successfully",
		}, nil
	}

	// Handle error response
	errorMsg := "SMS sending failed"
	if len(apiResponse.Message) > 0 {
		errorMsg = strings.Join(apiResponse.Message, ", ")
	}

	return &models.NotificationResponse{
		Success: false,
		Status:  models.StatusFailed,
		Message: errorMsg,
		Error:   &errorMsg,
	}, NewNotificationError(ErrProviderUnavailable, errorMsg, nil)
}

// ValidateConfig validates Fast2SMS configuration
func (f *Fast2SMSNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"api_key", "sender_id", "base_url"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (f *Fast2SMSNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderFast2SMS
}

// GetChannel returns the notification channel
func (f *Fast2SMSNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelSMS
}

// Helper functions
func cleanPhoneNumber(phone string) string {
	// Remove all non-digit characters
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	
	// Remove +91 prefix if present
	if strings.HasPrefix(cleaned, "91") && len(cleaned) == 12 {
		cleaned = cleaned[2:]
	}
	
	return cleaned
}

func isValidIndianMobile(phone string) bool {
	// Indian mobile numbers: 10 digits starting with 6, 7, 8, or 9
	if len(phone) != 10 {
		return false
	}
	
	firstDigit := phone[0]
	return firstDigit >= '6' && firstDigit <= '9'
}

func stringPtr(s string) *string {
	return &s
}