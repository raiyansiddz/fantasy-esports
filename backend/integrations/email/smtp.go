package integrations

import (
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"fantasy-esports-backend/models"
)

// SMTPNotifier implements email notifications via SMTP
type SMTPNotifier struct{}

// NewSMTPNotifier creates a new SMTP notifier
func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

// Send sends email via SMTP
func (s *SMTPNotifier) Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error) {
	// Validate config
	if err := s.ValidateConfig(config); err != nil {
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

	// Get config values
	host := config["host"]
	port := config["port"]
	username := config["username"]
	password := config["password"]
	fromEmail := config["from_email"]
	fromName := config["from_name"]

	// Set up authentication
	auth := smtp.PlainAuth("", username, password, host)

	// Prepare email content
	subject := "Notification"
	if request.Subject != nil {
		subject = *request.Subject
	}

	body := ""
	if request.Body != nil {
		body = *request.Body
	}

	// Create email message
	message := fmt.Sprintf("From: %s <%s>\r\n", fromName, fromEmail)
	message += fmt.Sprintf("To: %s\r\n", request.Recipient)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Send email
	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, fromEmail, []string{request.Recipient}, []byte(message))
	if err != nil {
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Failed to send email",
			Error:   &err.Error(),
		}, NewNotificationError(ErrNetworkError, "Failed to send email", err)
	}

	return &models.NotificationResponse{
		Success: true,
		Status:  models.StatusSent,
		Message: "Email sent successfully",
	}, nil
}

// ValidateConfig validates SMTP configuration
func (s *SMTPNotifier) ValidateConfig(config map[string]string) error {
	required := []string{"host", "port", "username", "password", "from_email"}
	for _, key := range required {
		if config[key] == "" {
			return NewNotificationError(ErrInvalidConfig, fmt.Sprintf("Missing required config: %s", key), nil)
		}
	}

	// Validate port
	if _, err := strconv.Atoi(config["port"]); err != nil {
		return NewNotificationError(ErrInvalidConfig, "Invalid port number", err)
	}

	return nil
}

// GetProviderName returns the provider name
func (s *SMTPNotifier) GetProviderName() models.NotificationProvider {
	return models.ProviderSMTP
}

// GetChannel returns the notification channel
func (s *SMTPNotifier) GetChannel() models.NotificationChannel {
	return models.ChannelEmail
}

// Helper function to validate email
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}