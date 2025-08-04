package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fantasy-esports-backend/models"
)

// FCMNotifier implements push notifications via Firebase Cloud Messaging
type FCMNotifier struct{}

// FCMRequest represents the request structure for FCM API
type FCMRequest struct {
	To           string      `json:"to,omitempty"`
	RegistrationIDs []string `json:"registration_ids,omitempty"`
	Notification FCMNotification `json:"notification"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Priority     string      `json:"priority"`
}

type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Sound string `json:"sound,omitempty"`
}

// FCMResponse represents the response from FCM API
type FCMResponse struct {
	MulticastID  int64 `json:"multicast_id"`
	Success      int   `json:"success"`
	Failure      int   `json:"failure"`
	CanonicalIDs int   `json:"canonical_ids"`
	Results      []FCMResult `json:"results"`
}

type FCMResult struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// NewFCMNotifier creates a new FCM notifier
func NewFCMNotifier() *FCMNotifier {
	return &FCMNotifier{}
}

// Send sends push notification via FCM
func (f *FCMNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := f.ValidateConfig(config); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
			errMsg := err.Error()
			Error:   &errMsg,
		}, err
	}

	// Prepare notification content
	title := "Notification"
	if request.Subject != nil {
		title = *request.Subject
	}

	body := ""
	if request.Body != nil {
		body = *request.Body
	}

	// Create FCM request
	fcmRequest := FCMRequest{
		To: request.Recipient, // FCM token
		Notification: FCMNotification{
			Title: title,
			Body:  body,
			Icon:  "ic_notification",
			Sound: "default",
		},
		Data: map[string]interface{}{
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
			"timestamp":    time.Now().Unix(),
		},
		Priority: "high",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(fcmRequest)
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
	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("key=%s", config["server_key"]))

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
	var fcmResponse FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResponse); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to parse response",
			errMsg := err.Error()
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Failed to parse response", err)
	}

	// Check if successful
	if resp.StatusCode == 200 && fcmResponse.Success > 0 {
		providerID := ""
		if len(fcmResponse.Results) > 0 && fcmResponse.Results[0].MessageID != "" {
			providerID = fcmResponse.Results[0].MessageID
		}

		return &models.NotificationResponse{
			Success:    true,
			ProviderID: &providerID,
			Status:     models.StatusSent,
			Message:    "Push notification sent successfully",
		}, nil
	}

	// Handle error response
	errorMsg := "Push notification sending failed"
	if len(fcmResponse.Results) > 0 && fcmResponse.Results[0].Error != "" {
		errorMsg = fcmResponse.Results[0].Error
	}

	return &models.NotificationResponse{
		Success: false,
		Status:  models.StatusFailed,
		Message: errorMsg,
		Error:   &errorMsg,
	}, NewNotificationError(ErrProviderUnavailable, errorMsg, nil)
}

// ValidateConfig validates FCM configuration
func (f *FCMNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"server_key", "base_url"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (f *FCMNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderFCM
}

// GetChannel returns the notification channel
func (f *FCMNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelPush
}