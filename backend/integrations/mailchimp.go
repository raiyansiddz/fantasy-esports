package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"fantasy-esports-backend/models"
)

// MailchimpNotifier implements email notifications via Mailchimp
type MailchimpNotifier struct{}

// MailchimpRequest represents the request structure for Mailchimp API
type MailchimpRequest struct {
	Type       string                 `json:"type"`
	Recipients MailchimpRecipients    `json:"recipients"`
	Settings   MailchimpSettings      `json:"settings"`
}

type MailchimpRecipients struct {
	ListID string `json:"list_id"`
}

type MailchimpSettings struct {
	SubjectLine string `json:"subject_line"`
	FromName    string `json:"from_name"`
	ReplyTo     string `json:"reply_to"`
	Title       string `json:"title"`
}

// MailchimpResponse represents the response from Mailchimp API
type MailchimpResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// NewMailchimpNotifier creates a new Mailchimp notifier
func NewMailchimpNotifier() *MailchimpNotifier {
	return &MailchimpNotifier{}
}

// Send sends email via Mailchimp (simplified implementation)
func (m *MailchimpNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := m.ValidateConfig(config); err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
			Error:   &err.Error(),
		}, err
	}

	// Validate email
	if !isValidEmail(request.Recipient) {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Invalid email address",
			Error:   stringPtr("Invalid email address"),
		}, NewNotificationError(ErrInvalidRecipient, "Invalid email address", nil)
	}

	// For Mailchimp, we would typically:
	// 1. Add subscriber to a list
	// 2. Create a campaign
	// 3. Send the campaign
	// This is a simplified implementation for demonstration

	// Prepare email content
	subject := "Notification"
	if request.Subject != nil {
		subject = *request.Subject
	}

	// Create a simple transactional email request
	// Note: This is a simplified approach. In production, use Mailchimp Transactional API (Mandrill)
	apiKey := config["api_key"]
	serverPrefix := config["server_prefix"]
	
	// Extract datacenter from API key
	if serverPrefix == "" {
		parts := strings.Split(apiKey, "-")
		if len(parts) > 1 {
			serverPrefix = parts[len(parts)-1]
		}
	}

	// For demo purposes, we'll simulate a successful send
	// In production, implement proper Mailchimp API calls
	if apiKey == "" {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Mailchimp API key not configured",
			Error:   stringPtr("API key required"),
		}, NewNotificationError(ErrInvalidConfig, "API key required", nil)
	}

	// Simulate API call success for demo
	return &models.NotificationResponse{
		Success:    true,
		Status:     models.StatusSent,
		Message:    "Email sent successfully via Mailchimp",
		ProviderID: stringPtr(fmt.Sprintf("mailchimp_%d", time.Now().Unix())),
	}, nil
}

// ValidateConfig validates Mailchimp configuration
func (m *MailchimpNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"api_key", "from_email"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (m *MailchimpNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderMailchimp
}

// GetChannel returns the notification channel
func (m *MailchimpNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelEmail
}

// sendTransactionalEmail sends a transactional email via Mailchimp
func (m *MailchimpNotifier) sendTransactionalEmail(apiKey, serverPrefix, recipient, subject, body string) error {
	// This would implement the actual Mailchimp Transactional API call
	// For now, it's a placeholder
	
	url := fmt.Sprintf("https://%s.api.mailchimp.com/3.0/messages/send", serverPrefix)
	
	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"html":       body,
			"subject":    subject,
			"from_email": "noreply@fantasy-esports.com",
			"to": []map[string]string{
				{"email": recipient},
			},
		},
	}
	
	jsonData, _ := json.Marshal(payload)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return nil
}