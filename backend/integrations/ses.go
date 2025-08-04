package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fantasy-esports-backend/models"
)

// SESNotifier implements email notifications via Amazon SES
type SESNotifier struct{}

// SESRequest represents the request structure for SES API
type SESRequest struct {
	Action      string `json:"Action"`
	Source      string `json:"Source"`
	Destination string `json:"Destination.ToAddresses.member.1"`
	Subject     string `json:"Message.Subject.Data"`
	Body        string `json:"Message.Body.Text.Data"`
	Version     string `json:"Version"`
}

// NewSESNotifier creates a new SES notifier
func NewSESNotifier() *SESNotifier {
	return &SESNotifier{}
}

// Send sends email via Amazon SES
func (s *SESNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := s.ValidateConfig(config); err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration validation failed",
			Error:   &errMsg,
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

	// Prepare email content
	subject := "Notification"
	if request.Subject != nil {
		subject = *request.Subject
	}

	body := ""
	if request.Body != nil {
		body = *request.Body
	}

	fromEmail := config["from_email"]
	fromName := config["from_name"]
	source := fmt.Sprintf("%s <%s>", fromName, fromEmail)

	// Create form data
	formData := url.Values{}
	formData.Set("Action", "SendEmail")
	formData.Set("Source", source)
	formData.Set("Destination.ToAddresses.member.1", request.Recipient)
	formData.Set("Message.Subject.Data", subject)
	formData.Set("Message.Body.Text.Data", body)
	formData.Set("Version", "2010-12-01")

	// Create HTTP request
	region := config["region"]
	sesURL := fmt.Sprintf("https://ses.%s.amazonaws.com/", region)
	
	httpReq, err := http.NewRequest("POST", sesURL, strings.NewReader(formData.Encode()))
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
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Sign request
	if err := s.signRequest(httpReq, config); err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to sign request",
			Error:   &errMsg,
		}, NewNotificationError(ErrAuthError, "Failed to sign request", err)
	}

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

	// Check response
	if resp.StatusCode == 200 {
		return &models.NotificationResponse{
			Success: true,
			Status:  models.StatusSent,
			Message: "Email sent successfully via SES",
		}, nil
	}

	return &models.NotificationResponse{
		Success: false,
		Status:  models.StatusFailed,
		Message: fmt.Sprintf("SES API error: %d", resp.StatusCode),
		Error:   stringPtr(fmt.Sprintf("HTTP %d", resp.StatusCode)),
	}, NewNotificationError(ErrProviderUnavailable, fmt.Sprintf("SES API error: %d", resp.StatusCode), nil)
}

// ValidateConfig validates SES configuration
func (s *SESNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"access_key_id", "secret_access_key", "region", "from_email"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}
	return nil
}

// GetProviderName returns the provider name
func (s *SESNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderSES
}

// GetChannel returns the notification channel
func (s *SESNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelEmail
}

// signRequest signs the AWS request using AWS Signature Version 4
func (s *SESNotifier) signRequest(req *http.Request, config map[string]string) error {
	// This is a simplified AWS signing implementation
	// In production, use the official AWS SDK
	
	accessKey := config["access_key_id"]
	secretKey := config["secret_access_key"]
	region := config["region"]
	service := "ses"
	
	// Create canonical request
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	date := timestamp[:8]
	
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Amz-Date", timestamp)
	
	// Create authorization header
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", date, region, service)
	
	// Simplified signing (for demo - use AWS SDK in production)
	signature := s.createSignature(secretKey, date, region, service, "demo-request")
	
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=host;x-amz-date, Signature=%s",
		accessKey, credentialScope, signature)
	
	req.Header.Set("Authorization", authHeader)
	
	return nil
}

func (s *SESNotifier) createSignature(secretKey, date, region, service, stringToSign string) string {
	// Simplified signature creation for demo
	h := hmac.New(sha256.New, []byte("AWS4"+secretKey))
	h.Write([]byte(date))
	dateKey := h.Sum(nil)
	
	h = hmac.New(sha256.New, dateKey)
	h.Write([]byte(region))
	regionKey := h.Sum(nil)
	
	h = hmac.New(sha256.New, regionKey)
	h.Write([]byte(service))
	serviceKey := h.Sum(nil)
	
	h = hmac.New(sha256.New, serviceKey)
	h.Write([]byte("aws4_request"))
	signingKey := h.Sum(nil)
	
	h = hmac.New(sha256.New, signingKey)
	h.Write([]byte(stringToSign))
	
	return hex.EncodeToString(h.Sum(nil))
}