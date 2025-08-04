package integrations

import "fmt"

// NotificationError represents a notification-specific error
type NotificationError struct {
	Code    string
	Message string
	Cause   error
}

// Error implements error interface
func (e *NotificationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewNotificationError creates a new notification error
func NewNotificationError(code, message string, cause error) *NotificationError {
	return &NotificationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Common error codes
const (
	ErrInvalidConfig     = "INVALID_CONFIG"
	ErrProviderUnavailable = "PROVIDER_UNAVAILABLE"
	ErrRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrInvalidRecipient  = "INVALID_RECIPIENT"
	ErrTemplateParsing   = "TEMPLATE_PARSING_ERROR"
	ErrNetworkError      = "NETWORK_ERROR"
	ErrAuthError         = "AUTHENTICATION_ERROR"
)