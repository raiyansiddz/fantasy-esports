package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fantasy-esports-backend/models"
)

// WhatsAppNotifier implements WhatsApp notifications via WhatsApp Cloud API
type WhatsAppNotifier struct{}

// WhatsAppRequest represents the request structure for WhatsApp Cloud API
type WhatsAppRequest struct {
	MessagingProduct string          `json:"messaging_product"`
	To               string          `json:"to"`
	Type             string          `json:"type"`
	Text             WhatsAppText    `json:"text"`
	Template         *WhatsAppTemplate `json:"template,omitempty"`
}

type WhatsAppText struct {
	Body string `json:"body"`
}

type WhatsAppTemplate struct {
	Name       string                    `json:"name"`
	Language   WhatsAppLanguage          `json:"language"`
	Components []WhatsAppComponent       `json:"components,omitempty"`
}

type WhatsAppLanguage struct {
	Code string `json:"code"`
}

type WhatsAppComponent struct {
	Type       string                  `json:"type"`
	Parameters []WhatsAppParameter     `json:"parameters,omitempty"`
}

type WhatsAppParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// WhatsAppResponse represents the response from WhatsApp Cloud API
type WhatsAppResponse struct {
	MessagingProduct string            `json:"messaging_product"`
	Contacts         []WhatsAppContact `json:"contacts"`
	Messages         []WhatsAppMessage `json:"messages"`
	Error            *WhatsAppError    `json:"error,omitempty"`
}

type WhatsAppContact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type WhatsAppMessage struct {
	ID string `json:"id"`
}

type WhatsAppError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
}

// NewWhatsAppNotifier creates a new WhatsApp notifier
func NewWhatsAppNotifier() *WhatsAppNotifier {
	return &WhatsAppNotifier{}
}

// Send sends WhatsApp message via WhatsApp Cloud API
func (w *WhatsAppNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := w.ValidateConfig(config); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
			errMsg := err.Error()
			Error:   &errMsg,
		}, err
	}

	// Clean phone number for WhatsApp
	phoneNumber := cleanWhatsAppNumber(request.Recipient)
	if !isValidWhatsAppNumber(phoneNumber) {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Invalid WhatsApp number format",
			Error:   stringPtr("Invalid WhatsApp number format"),
		}, NewNotificationError(ErrInvalidRecipient, "Invalid WhatsApp number format", nil)
	}

	// Prepare message content
	body := ""
	if request.Body != nil {
		body = *request.Body
	}

	// Create WhatsApp request
	waRequest := WhatsAppRequest{
		MessagingProduct: "whatsapp",
		To:               phoneNumber,
		Type:             "text",
		Text: WhatsAppText{
			Body: body,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(waRequest)
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to encode request",
			errMsg := err.Error()
			Error:   &errMsg,
		}, NewNotificationError(ErrTemplateParsing, "Failed to encode request", err)
	}

	// Create HTTP request
	baseURL := config["base_url"]
	phoneNumberID := config["phone_number_id"]
	url := fmt.Sprintf("%s/%s/messages", baseURL, phoneNumberID)
	
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to create HTTP request",
			errMsg := err.Error()
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config["access_token"]))

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Network error",
			errMsg := err.Error()
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Network error", err)
	}
	defer resp.Body.Close()

	// Parse response
	var waResponse WhatsAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&waResponse); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to parse response",
			errMsg := err.Error()
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Failed to parse response", err)
	}

	// Check if successful
	if resp.StatusCode == 200 && len(waResponse.Messages) > 0 {
		return &models.NotificationResponse{
			Success:    true,
			ProviderID: &waResponse.Messages[0].ID,
			Status:     models.StatusSent,
			Message:    "WhatsApp message sent successfully",
		}, nil
	}

	// Handle error response
	errorMsg := "WhatsApp message sending failed"
	if waResponse.Error != nil {
		errorMsg = waResponse.Error.Message
	}

	return &models.NotificationResponse{
		Success: false,
		Status:  models.StatusFailed,
		Message: errorMsg,
		Error:   &errorMsg,
	}, NewNotificationError(ErrProviderUnavailable, errorMsg, nil)
}

// ValidateConfig validates WhatsApp configuration
func (w *WhatsAppNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"access_token", "phone_number_id", "base_url"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (w *WhatsAppNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderWhatsAppCloud
}

// GetChannel returns the notification channel
func (w *WhatsAppNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelWhatsApp
}

// Helper functions for WhatsApp
func cleanWhatsAppNumber(phone string) string {
	// Remove all non-digit characters
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	
	// Add country code if not present
	if len(cleaned) == 10 {
		cleaned = "91" + cleaned // Add India country code
	}
	
	return cleaned
}

func isValidWhatsAppNumber(phone string) bool {
	// WhatsApp numbers should be in international format without +
	// For India: 91XXXXXXXXXX (12 digits)
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	
	// Check if all characters are digits
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}