package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fantasy-esports-backend/models"
)

// OneSignalNotifier implements push notifications via OneSignal
type OneSignalNotifier struct{}

// OneSignalRequest represents the request structure for OneSignal API
type OneSignalRequest struct {
	AppID            string                 `json:"app_id"`
	Contents         map[string]string      `json:"contents"`
	Headings         map[string]string      `json:"headings,omitempty"`
	IncludePlayerIDs []string              `json:"include_player_ids,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
	Priority         int                    `json:"priority"`
}

// OneSignalResponse represents the response from OneSignal API
type OneSignalResponse struct {
	ID         string            `json:"id"`
	Recipients int               `json:"recipients"`
	Errors     map[string]string `json:"errors,omitempty"`
}

// NewOneSignalNotifier creates a new OneSignal notifier
func NewOneSignalNotifier() *OneSignalNotifier {
	return &OneSignalNotifier{}
}

// Send sends push notification via OneSignal
func (o *OneSignalNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := o.ValidateConfig(config); err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
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

	// Create OneSignal request
	osRequest := OneSignalRequest{
		AppID: config["app_id"],
		Contents: map[string]string{
			"en": body,
		},
		Headings: map[string]string{
			"en": title,
		},
		IncludePlayerIDs: []string{request.Recipient}, // OneSignal player ID
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Priority: 10,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(osRequest)
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
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to create HTTP request",
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Failed to create HTTP request", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", config["api_key"]))

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Network error",
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Network error", err)
	}
	defer resp.Body.Close()

	// Parse response
	var osResponse OneSignalResponse
	if err := json.NewDecoder(resp.Body).Decode(&osResponse); err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to parse response",
			Error:   &errMsg,
		}, NewNotificationError(ErrNetworkError, "Failed to parse response", err)
	}

	// Check if successful
	if resp.StatusCode == 200 && osResponse.ID != "" {
		return &models.NotificationResponse{
			Success:    true,
			ProviderID: &osResponse.ID,
			Status:     models.StatusSent,
			Message:    "Push notification sent successfully",
		}, nil
	}

	// Handle error response
	errorMsg := "Push notification sending failed"
	if len(osResponse.Errors) > 0 {
		for _, errMsg := range osResponse.Errors {
			errorMsg = errMsg
			break
		}
	}

	return &models.NotificationResponse{
		Success: false,
		Status:  models.StatusFailed,
		Message: errorMsg,
		Error:   &errorMsg,
	}, NewNotificationError(ErrProviderUnavailable, errorMsg, nil)
}

// ValidateConfig validates OneSignal configuration
func (o *OneSignalNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"app_id", "api_key", "base_url"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (o *OneSignalNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderOneSignal
}

// GetChannel returns the notification channel
func (o *OneSignalNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelPush
}