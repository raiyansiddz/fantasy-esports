package errors

import (
	"fmt"
	"net/http"
	"time"
	
	"fantasy-esports-backend/pkg/logger"
)

// ErrorCode represents standardized error codes
type ErrorCode string

// Standardized Error Codes
const (
	// Authentication & Authorization Errors (A001-A999)
	ErrInvalidCredentials     ErrorCode = "A001"
	ErrInvalidOTP            ErrorCode = "A002"
	ErrOTPExpired           ErrorCode = "A003"
	ErrSessionExpired       ErrorCode = "A004"
	ErrUnauthorized         ErrorCode = "A005"
	ErrForbidden           ErrorCode = "A006"
	ErrInvalidToken        ErrorCode = "A007"
	ErrTokenExpired        ErrorCode = "A008"
	ErrAdminAccess         ErrorCode = "A009"
	ErrInvalidSocialToken  ErrorCode = "A010"
	
	// User & Profile Errors (U001-U999)
	ErrUserNotFound        ErrorCode = "U001"
	ErrUserAlreadyExists   ErrorCode = "U002"
	ErrInvalidMobile       ErrorCode = "U003"
	ErrInvalidEmail        ErrorCode = "U004"
	ErrKYCRequired         ErrorCode = "U005"
	ErrAccountSuspended    ErrorCode = "U006"
	ErrAccountBanned       ErrorCode = "U007"
	ErrInvalidKYCDocument  ErrorCode = "U008"
	ErrKYCAlreadyVerified  ErrorCode = "U009"
	ErrProfileIncomplete   ErrorCode = "U010"
	
	// Contest & Fantasy Errors (C001-C999)
	ErrContestNotFound        ErrorCode = "C001"
	ErrContestFull           ErrorCode = "C002"
	ErrContestClosed         ErrorCode = "C003"
	ErrAlreadyJoined         ErrorCode = "C004"
	ErrInvalidTeam           ErrorCode = "C005"
	ErrInsufficientCredits   ErrorCode = "C006"
	ErrTeamLocked           ErrorCode = "C007"
	ErrInvalidPlayer        ErrorCode = "C008"
	ErrMaxPlayersPerTeam    ErrorCode = "C009"
	ErrInvalidCaptain       ErrorCode = "C010"
	ErrDuplicateEntry       ErrorCode = "C011"
	ErrContestExpired       ErrorCode = "C012"
	
	// Wallet & Payment Errors (W001-W999)
	ErrInsufficientBalance   ErrorCode = "W001"
	ErrInvalidAmount        ErrorCode = "W002"
	ErrPaymentFailed        ErrorCode = "W003"
	ErrWithdrawalFailed     ErrorCode = "W004"
	ErrInvalidPaymentMethod ErrorCode = "W005"
	ErrTransactionNotFound  ErrorCode = "W006"
	ErrMinDepositAmount     ErrorCode = "W007"
	ErrMaxDepositAmount     ErrorCode = "W008"
	ErrWithdrawalLimit      ErrorCode = "W009"
	ErrPendingTransactions  ErrorCode = "W010"
	
	// Game & Match Errors (G001-G999)
	ErrGameNotFound         ErrorCode = "G001"
	ErrMatchNotFound        ErrorCode = "G002"
	ErrTournamentNotFound   ErrorCode = "G003"
	ErrPlayerNotFound       ErrorCode = "G004"
	ErrTeamNotFound         ErrorCode = "G005"
	ErrMatchAlreadyStarted  ErrorCode = "G006"
	ErrMatchNotStarted      ErrorCode = "G007"
	ErrInvalidScore         ErrorCode = "G008"
	ErrScoreAlreadySubmitted ErrorCode = "G009"
	ErrMatchCompleted       ErrorCode = "G010"
	
	// Referral Errors (R001-R999)
	ErrInvalidReferralCode  ErrorCode = "R001"
	ErrSelfReferral         ErrorCode = "R002"
	ErrReferralAlreadyUsed  ErrorCode = "R003"
	ErrReferralExpired      ErrorCode = "R004"
	ErrReferralNotFound     ErrorCode = "R005"
	
	// System & Server Errors (S001-S999)
	ErrInternalServer       ErrorCode = "S001"
	ErrDatabaseConnection   ErrorCode = "S002"
	ErrInvalidRequest       ErrorCode = "S003"
	ErrValidationFailed     ErrorCode = "S004"
	ErrResourceNotFound     ErrorCode = "S005"
	ErrServiceUnavailable   ErrorCode = "S006"
	ErrRateLimitExceeded    ErrorCode = "S007"
	ErrMaintenanceMode      ErrorCode = "S008"
	ErrInvalidJSON          ErrorCode = "S009"
	ErrFileTooLarge         ErrorCode = "S010"
	
	// Business Logic Errors (B001-B999)
	ErrBusinessRule         ErrorCode = "B001"
	ErrDataIntegrity        ErrorCode = "B002"
	ErrConcurrencyConflict  ErrorCode = "B003"
	ErrQuotaExceeded        ErrorCode = "B004"
	ErrFeatureDisabled      ErrorCode = "B005"
	ErrInvalidOperation     ErrorCode = "B006"
	ErrResourceLimit        ErrorCode = "B007"
	ErrSchedulingConflict   ErrorCode = "B008"
	
	// External Service Errors (E001-E999)
	ErrPaymentGateway       ErrorCode = "E001"
	ErrSMSService          ErrorCode = "E002"
	ErrEmailService        ErrorCode = "E003"
	ErrFileUpload          ErrorCode = "E004"
	ErrNotificationService ErrorCode = "E005"
	ErrThirdPartyAPI       ErrorCode = "E006"
)

// AppError represents a standardized application error
type AppError struct {
	Code        ErrorCode   `json:"code"`
	Message     string      `json:"message"`
	Details     interface{} `json:"details,omitempty"`
	HTTPStatus  int         `json:"-"`
	Timestamp   time.Time   `json:"timestamp"`
	RequestID   string      `json:"request_id,omitempty"`
	UserMessage string      `json:"user_message,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorResponse represents the standardized API error response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error"`
	Code      string      `json:"code"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// Error definitions with messages
var errorMessages = map[ErrorCode]struct {
	message     string
	userMessage string
	httpStatus  int
}{
	// Authentication & Authorization
	ErrInvalidCredentials:     {"Invalid username or password", "Please check your login credentials", http.StatusUnauthorized},
	ErrInvalidOTP:            {"Invalid OTP provided", "Please enter the correct OTP", http.StatusBadRequest},
	ErrOTPExpired:           {"OTP has expired", "OTP has expired, please request a new one", http.StatusBadRequest},
	ErrSessionExpired:       {"Session has expired", "Your session has expired, please login again", http.StatusUnauthorized},
	ErrUnauthorized:         {"Unauthorized access", "Access denied", http.StatusUnauthorized},
	ErrForbidden:           {"Forbidden access", "You don't have permission to perform this action", http.StatusForbidden},
	ErrInvalidToken:        {"Invalid authentication token", "Invalid authentication token", http.StatusUnauthorized},
	ErrTokenExpired:        {"Authentication token expired", "Your session has expired, please login again", http.StatusUnauthorized},
	ErrAdminAccess:         {"Admin access required", "Admin privileges required", http.StatusForbidden},
	ErrInvalidSocialToken:  {"Invalid social login token", "Social login failed, please try again", http.StatusBadRequest},

	// User & Profile
	ErrUserNotFound:        {"User not found", "User not found", http.StatusNotFound},
	ErrUserAlreadyExists:   {"User already exists", "An account with this mobile number already exists", http.StatusConflict},
	ErrInvalidMobile:       {"Invalid mobile number format", "Please enter a valid mobile number", http.StatusBadRequest},
	ErrInvalidEmail:        {"Invalid email format", "Please enter a valid email address", http.StatusBadRequest},
	ErrKYCRequired:         {"KYC verification required", "Please complete your KYC verification", http.StatusForbidden},
	ErrAccountSuspended:    {"Account is suspended", "Your account has been suspended", http.StatusForbidden},
	ErrAccountBanned:       {"Account is banned", "Your account has been banned", http.StatusForbidden},
	ErrInvalidKYCDocument:  {"Invalid KYC document", "Please upload valid KYC documents", http.StatusBadRequest},
	ErrKYCAlreadyVerified:  {"KYC already verified", "Your KYC is already verified", http.StatusConflict},
	ErrProfileIncomplete:   {"Profile is incomplete", "Please complete your profile", http.StatusBadRequest},

	// Contest & Fantasy
	ErrContestNotFound:        {"Contest not found", "Contest not found", http.StatusNotFound},
	ErrContestFull:           {"Contest is full", "This contest is already full", http.StatusConflict},
	ErrContestClosed:         {"Contest is closed", "Registration for this contest is closed", http.StatusConflict},
	ErrAlreadyJoined:         {"Already joined contest", "You have already joined this contest", http.StatusConflict},
	ErrInvalidTeam:           {"Invalid team composition", "Invalid team composition", http.StatusBadRequest},
	ErrInsufficientCredits:   {"Insufficient credits for team", "You don't have enough credits for this team", http.StatusBadRequest},
	ErrTeamLocked:           {"Team is locked", "Team cannot be modified after match starts", http.StatusConflict},
	ErrInvalidPlayer:        {"Invalid player selection", "One or more players are not valid", http.StatusBadRequest},
	ErrMaxPlayersPerTeam:    {"Too many players from same team", "Maximum 2 players allowed from same team", http.StatusBadRequest},
	ErrInvalidCaptain:       {"Invalid captain selection", "Please select valid captain and vice-captain", http.StatusBadRequest},
	ErrDuplicateEntry:       {"Duplicate entry not allowed", "Multiple entries not allowed in this contest", http.StatusConflict},
	ErrContestExpired:       {"Contest has expired", "This contest has expired", http.StatusGone},

	// Wallet & Payment
	ErrInsufficientBalance:   {"Insufficient wallet balance", "Insufficient balance in your wallet", http.StatusBadRequest},
	ErrInvalidAmount:        {"Invalid amount", "Please enter a valid amount", http.StatusBadRequest},
	ErrPaymentFailed:        {"Payment processing failed", "Payment failed, please try again", http.StatusBadRequest},
	ErrWithdrawalFailed:     {"Withdrawal processing failed", "Withdrawal failed, please try again", http.StatusBadRequest},
	ErrInvalidPaymentMethod: {"Invalid payment method", "Please select a valid payment method", http.StatusBadRequest},
	ErrTransactionNotFound:  {"Transaction not found", "Transaction not found", http.StatusNotFound},
	ErrMinDepositAmount:     {"Amount below minimum deposit", "Minimum deposit amount is ₹10", http.StatusBadRequest},
	ErrMaxDepositAmount:     {"Amount exceeds maximum deposit", "Maximum deposit amount is ₹10,000", http.StatusBadRequest},
	ErrWithdrawalLimit:      {"Withdrawal limit exceeded", "Daily withdrawal limit exceeded", http.StatusBadRequest},
	ErrPendingTransactions:  {"Pending transactions exist", "Please wait for pending transactions to complete", http.StatusConflict},

	// Game & Match
	ErrGameNotFound:         {"Game not found", "Game not found", http.StatusNotFound},
	ErrMatchNotFound:        {"Match not found", "Match not found", http.StatusNotFound},
	ErrTournamentNotFound:   {"Tournament not found", "Tournament not found", http.StatusNotFound},
	ErrPlayerNotFound:       {"Player not found", "Player not found", http.StatusNotFound},
	ErrTeamNotFound:         {"Team not found", "Team not found", http.StatusNotFound},
	ErrMatchAlreadyStarted:  {"Match already started", "Match has already started", http.StatusConflict},
	ErrMatchNotStarted:      {"Match not started", "Match has not started yet", http.StatusConflict},
	ErrInvalidScore:         {"Invalid score data", "Invalid score data provided", http.StatusBadRequest},
	ErrScoreAlreadySubmitted: {"Score already submitted", "Score has already been submitted for this match", http.StatusConflict},
	ErrMatchCompleted:       {"Match already completed", "This match has already been completed", http.StatusConflict},

	// Referral
	ErrInvalidReferralCode:  {"Invalid referral code", "Invalid referral code", http.StatusBadRequest},
	ErrSelfReferral:         {"Cannot refer yourself", "You cannot use your own referral code", http.StatusBadRequest},
	ErrReferralAlreadyUsed:  {"Referral code already used", "This referral code has already been used", http.StatusConflict},
	ErrReferralExpired:      {"Referral code expired", "This referral code has expired", http.StatusBadRequest},
	ErrReferralNotFound:     {"Referral not found", "Referral not found", http.StatusNotFound},

	// System & Server
	ErrInternalServer:       {"Internal server error", "Something went wrong, please try again", http.StatusInternalServerError},
	ErrDatabaseConnection:   {"Database connection error", "Service temporarily unavailable", http.StatusServiceUnavailable},
	ErrInvalidRequest:       {"Invalid request format", "Invalid request format", http.StatusBadRequest},
	ErrValidationFailed:     {"Request validation failed", "Please check your input data", http.StatusBadRequest},
	ErrResourceNotFound:     {"Resource not found", "Resource not found", http.StatusNotFound},
	ErrServiceUnavailable:   {"Service unavailable", "Service temporarily unavailable", http.StatusServiceUnavailable},
	ErrRateLimitExceeded:    {"Rate limit exceeded", "Too many requests, please try again later", http.StatusTooManyRequests},
	ErrMaintenanceMode:      {"System under maintenance", "System is under maintenance", http.StatusServiceUnavailable},
	ErrInvalidJSON:          {"Invalid JSON format", "Invalid request format", http.StatusBadRequest},
	ErrFileTooLarge:         {"File too large", "File size exceeds maximum limit", http.StatusRequestEntityTooLarge},

	// Business Logic
	ErrBusinessRule:         {"Business rule violation", "Operation violates business rules", http.StatusBadRequest},
	ErrDataIntegrity:        {"Data integrity violation", "Data integrity error", http.StatusConflict},
	ErrConcurrencyConflict:  {"Concurrency conflict", "Resource was modified by another request", http.StatusConflict},
	ErrQuotaExceeded:        {"Quota exceeded", "Usage quota exceeded", http.StatusTooManyRequests},
	ErrFeatureDisabled:      {"Feature disabled", "This feature is currently disabled", http.StatusServiceUnavailable},
	ErrInvalidOperation:     {"Invalid operation", "Operation not allowed in current state", http.StatusBadRequest},
	ErrResourceLimit:        {"Resource limit exceeded", "Resource limit exceeded", http.StatusBadRequest},
	ErrSchedulingConflict:   {"Scheduling conflict", "Scheduling conflict detected", http.StatusConflict},

	// External Service
	ErrPaymentGateway:       {"Payment gateway error", "Payment service error, please try again", http.StatusBadGateway},
	ErrSMSService:          {"SMS service error", "SMS service error, please try again", http.StatusBadGateway},
	ErrEmailService:        {"Email service error", "Email service error, please try again", http.StatusBadGateway},
	ErrFileUpload:          {"File upload error", "File upload failed, please try again", http.StatusBadGateway},
	ErrNotificationService: {"Notification service error", "Notification service error", http.StatusBadGateway},
	ErrThirdPartyAPI:       {"Third-party API error", "External service error, please try again", http.StatusBadGateway},
}

// NewError creates a new AppError
func NewError(code ErrorCode, details interface{}) *AppError {
	errInfo, exists := errorMessages[code]
	if !exists {
		errInfo = errorMessages[ErrInternalServer]
	}
	
	return &AppError{
		Code:        code,
		Message:     errInfo.message,
		UserMessage: errInfo.userMessage,
		HTTPStatus:  errInfo.httpStatus,
		Details:     details,
		Timestamp:   time.Now(),
	}
}

// NewErrorWithMessage creates a new AppError with custom message
func NewErrorWithMessage(code ErrorCode, message string, details interface{}) *AppError {
	errInfo, exists := errorMessages[code]
	if !exists {
		errInfo = errorMessages[ErrInternalServer]
	}
	
	return &AppError{
		Code:        code,
		Message:     message,
		UserMessage: errInfo.userMessage,
		HTTPStatus:  errInfo.httpStatus,
		Details:     details,
		Timestamp:   time.Now(),
	}
}

// SetRequestID adds request ID to the error
func (e *AppError) SetRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// ToResponse converts AppError to ErrorResponse
func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Error:     e.UserMessage,
		Code:      string(e.Code),
		Details:   e.Details,
		Timestamp: e.Timestamp,
		RequestID: e.RequestID,
	}
}

// LogError logs the error with context
func (e *AppError) LogError(context map[string]interface{}) {
	logData := map[string]interface{}{
		"error_code":    e.Code,
		"message":      e.Message,
		"user_message": e.UserMessage,
		"http_status":  e.HTTPStatus,
		"timestamp":    e.Timestamp,
		"request_id":   e.RequestID,
		"details":      e.Details,
	}
	
	// Merge context
	for k, v := range context {
		logData[k] = v
	}
	
	// Log based on severity
	switch e.HTTPStatus {
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		logger.Error("Application Error", logData)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		logger.Warn("Client Error", logData)
	default:
		logger.Info("Application Error", logData)
	}
}

// HandlePanic recovers from panic and creates appropriate error
func HandlePanic() *AppError {
	if r := recover(); r != nil {
		logger.Error("Panic recovered", map[string]interface{}{
			"panic": r,
		})
		return NewError(ErrInternalServer, "Unexpected error occurred")
	}
	return nil
}

// GetErrorInfo returns error information for a given error code
func GetErrorInfo(code ErrorCode) (message, userMessage string, httpStatus int) {
	if info, exists := errorMessages[code]; exists {
		return info.message, info.userMessage, info.httpStatus
	}
	info := errorMessages[ErrInternalServer]
	return info.message, info.userMessage, info.httpStatus
}

// IsClientError checks if error is client-side (4xx)
func IsClientError(code ErrorCode) bool {
	if info, exists := errorMessages[code]; exists {
		return info.httpStatus >= 400 && info.httpStatus < 500
	}
	return false
}

// IsServerError checks if error is server-side (5xx)
func IsServerError(code ErrorCode) bool {
	if info, exists := errorMessages[code]; exists {
		return info.httpStatus >= 500
	}
	return true // Default to server error for unknown codes
}

// ValidationError creates a validation error with field details
func ValidationError(fieldErrors map[string]string) *AppError {
	return NewError(ErrValidationFailed, fieldErrors)
}

// BusinessRuleError creates a business rule violation error
func BusinessRuleError(rule string, details interface{}) *AppError {
	return NewError(ErrBusinessRule, map[string]interface{}{
		"rule":    rule,
		"details": details,
	})
}