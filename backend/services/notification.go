package services

import (
	"database/sql"
	"fmt"
	"strings"

	"fantasy-esports-backend/integrations"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
)

// NotificationService handles all notification operations
type NotificationService struct {
	db             *sql.DB
	notifierFactory *integrations.NotifierFactory
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{
		db:              db,
		notifierFactory: integrations.NewNotifierFactory(),
	}
}

// SendNotification sends a single notification
func (s *NotificationService) SendNotification(request *models.SendNotificationRequest) (*models.NotificationResponse, error) {
	// Determine provider if not specified
	provider := request.Provider
	if provider == nil {
		defaultProvider := s.getDefaultProvider(request.Channel)
		provider = &defaultProvider
	}

	// Get template if template ID is provided
	var template *models.NotificationTemplate
	if request.TemplateID != nil {
		var err error
		template, err = s.GetTemplate(*request.TemplateID)
		if err != nil {
			errMsg := err.Error()
			return &models.NotificationResponse{
				Success: false,
				Status:  models.StatusFailed,
				Message: "Template not found",
				Error:   &errMsg,
			}, err
		}
		
		// Override provider with template's provider
		provider = &template.Provider
	}

	// Get configuration for the provider
	config, err := s.getProviderConfig(*provider, request.Channel)
	if err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Configuration error",
			Error:   &errMsg,
		}, err
	}

	// Process template variables if template is used
	finalRequest := *request
	if template != nil {
		processedBody, err := s.processTemplate(template, request.Variables)
		if err != nil {
			errMsg := err.Error()
			return &models.NotificationResponse{
				Success: false,
				Status:  models.StatusFailed,
				Message: "Template processing failed",
				Error:   &errMsg,
			}, err
		}
		
		finalRequest.Body = &processedBody
		if template.Subject != nil {
			processedSubject, err := s.processTemplateString(*template.Subject, request.Variables)
			if err == nil {
				finalRequest.Subject = &processedSubject
			}
		}
	}

	// Create notifier
	notifier, err := s.notifierFactory.CreateNotifier(*provider, request.Channel)
	if err != nil {
		errMsg := err.Error()
		return &models.NotificationResponse{
			Success: false,
			Status:  models.StatusFailed,
			Message: "Notifier creation failed",
			Error:   &errMsg,
		}, err
	}

	// Create log entry
	logID, err := s.createNotificationLog(&finalRequest, template, *provider)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create notification log: %v", err))
	}

	// Send notification
	response, err := notifier.Send(&finalRequest, config)
	if err != nil {
		// Update log with error
		if logID > 0 {
			errMsg := err.Error()
			s.updateNotificationLog(logID, models.StatusFailed, nil, &errMsg)
		}
		return response, err
	}

	// Update log with success
	if logID > 0 {
		response.LogID = logID
		s.updateNotificationLog(logID, response.Status, response.ProviderID, nil)
	} else {
		// Create log entry if not created earlier
		logID, _ = s.createNotificationLog(&finalRequest, template, *provider)
		response.LogID = logID
	}

	return response, nil
}

// SendBulkNotification sends notifications to multiple recipients
func (s *NotificationService) SendBulkNotification(request *models.BulkNotificationRequest) ([]*models.NotificationResponse, error) {
	var responses []*models.NotificationResponse
	
	recipients := request.Recipients
	
	// If user filter is provided, get filtered recipients
	if request.UserFilter != nil {
		filteredRecipients, err := s.getFilteredRecipients(request.Channel, request.UserFilter)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, filteredRecipients...)
	}

	// Send to each recipient
	for _, recipient := range recipients {
		singleRequest := &models.SendNotificationRequest{
			Channel:    request.Channel,
			Provider:   request.Provider,
			TemplateID: request.TemplateID,
			Recipient:  recipient,
			Subject:    request.Subject,
			Body:       request.Body,
			Variables:  request.Variables,
		}

		response, err := s.SendNotification(singleRequest)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to send notification to %s: %v", recipient, err))
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// Template Management Methods

// CreateTemplate creates a new notification template
func (s *NotificationService) CreateTemplate(request *models.TemplateCreateRequest, createdBy int64) (*models.NotificationTemplate, error) {
	query := `
		INSERT INTO notification_templates (name, channel, provider, subject, body, variables, 
			is_dlt_approved, dlt_template_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	var template models.NotificationTemplate
	variables := models.TemplateVariables(request.Variables)
	
	err := s.db.QueryRow(query, request.Name, request.Channel, request.Provider,
		request.Subject, request.Body, variables, request.IsDLTApproved,
		request.DLTTemplateID, createdBy).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	template.Name = request.Name
	template.Channel = request.Channel
	template.Provider = request.Provider
	template.Subject = request.Subject
	template.Body = request.Body
	template.Variables = variables
	template.IsDLTApproved = request.IsDLTApproved
	template.DLTTemplateID = request.DLTTemplateID
	template.CreatedBy = createdBy
	template.IsActive = true

	return &template, nil
}

// GetTemplate retrieves a template by ID
func (s *NotificationService) GetTemplate(templateID int64) (*models.NotificationTemplate, error) {
	query := `
		SELECT id, name, channel, provider, subject, body, variables, is_dlt_approved, 
			dlt_template_id, is_active, created_by, created_at, updated_at
		FROM notification_templates WHERE id = $1`

	var template models.NotificationTemplate
	err := s.db.QueryRow(query, templateID).Scan(
		&template.ID, &template.Name, &template.Channel, &template.Provider,
		&template.Subject, &template.Body, &template.Variables, &template.IsDLTApproved,
		&template.DLTTemplateID, &template.IsActive, &template.CreatedBy,
		&template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// GetTemplates retrieves templates with pagination
func (s *NotificationService) GetTemplates(channel *models.NotificationChannel, provider *models.NotificationProvider, page, limit int) ([]*models.NotificationTemplate, int, error) {
	offset := (page - 1) * limit
	
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argIndex := 1

	if channel != nil {
		whereClause += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, *channel)
		argIndex++
	}

	if provider != nil {
		whereClause += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, *provider)
		argIndex++
	}

	// Get templates
	query := fmt.Sprintf(`
		SELECT id, name, channel, provider, subject, body, variables, is_dlt_approved, 
			dlt_template_id, is_active, created_by, created_at, updated_at
		FROM notification_templates %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.NotificationTemplate
	for rows.Next() {
		template := &models.NotificationTemplate{}
		err := rows.Scan(&template.ID, &template.Name, &template.Channel, &template.Provider,
			&template.Subject, &template.Body, &template.Variables, &template.IsDLTApproved,
			&template.DLTTemplateID, &template.IsActive, &template.CreatedBy,
			&template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			continue
		}
		templates = append(templates, template)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notification_templates %s", whereClause)
	var total int
	err = s.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		total = len(templates)
	}

	return templates, total, nil
}

// UpdateTemplate updates an existing template
func (s *NotificationService) UpdateTemplate(templateID int64, request *models.TemplateUpdateRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if request.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *request.Name)
		argIndex++
	}

	if request.Subject != nil {
		setParts = append(setParts, fmt.Sprintf("subject = $%d", argIndex))
		args = append(args, *request.Subject)
		argIndex++
	}

	if request.Body != nil {
		setParts = append(setParts, fmt.Sprintf("body = $%d", argIndex))
		args = append(args, *request.Body)
		argIndex++
	}

	if request.Variables != nil {
		setParts = append(setParts, fmt.Sprintf("variables = $%d", argIndex))
		variables := models.TemplateVariables(request.Variables)
		args = append(args, variables)
		argIndex++
	}

	if request.IsDLTApproved != nil {
		setParts = append(setParts, fmt.Sprintf("is_dlt_approved = $%d", argIndex))
		args = append(args, *request.IsDLTApproved)
		argIndex++
	}

	if request.DLTTemplateID != nil {
		setParts = append(setParts, fmt.Sprintf("dlt_template_id = $%d", argIndex))
		args = append(args, *request.DLTTemplateID)
		argIndex++
	}

	if request.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *request.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))
	
	query := fmt.Sprintf("UPDATE notification_templates SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, templateID)

	_, err := s.db.Exec(query, args...)
	return err
}

// Configuration Management Methods

// UpdateConfig updates notification configuration
func (s *NotificationService) UpdateConfig(request *models.ConfigUpdateRequest, updatedBy int64) error {
	query := `
		INSERT INTO notification_config (provider, channel, config_key, config_value, is_active, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (provider, channel, config_key) 
		DO UPDATE SET config_value = $4, is_active = $5, updated_by = $6, updated_at = NOW()`

	_, err := s.db.Exec(query, request.Provider, request.Channel, request.ConfigKey,
		request.ConfigValue, request.IsActive, updatedBy)
	
	return err
}

// GetConfig retrieves configuration for a provider and channel
func (s *NotificationService) GetConfig(provider models.NotificationProvider, channel models.NotificationChannel) (map[string]*models.NotificationConfig, error) {
	query := `
		SELECT id, provider, channel, config_key, config_value, is_active, updated_by, updated_at
		FROM notification_config 
		WHERE provider = $1 AND channel = $2`

	rows, err := s.db.Query(query, provider, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}
	defer rows.Close()

	config := make(map[string]*models.NotificationConfig)
	for rows.Next() {
		var cfg models.NotificationConfig
		err := rows.Scan(&cfg.ID, &cfg.Provider, &cfg.Channel, &cfg.ConfigKey,
			&cfg.ConfigValue, &cfg.IsActive, &cfg.UpdatedBy, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		config[cfg.ConfigKey] = &cfg
	}

	return config, nil
}

// Statistics and Monitoring Methods

// GetNotificationStats retrieves notification statistics
func (s *NotificationService) GetNotificationStats(channel *models.NotificationChannel, provider *models.NotificationProvider, days int) (*models.NotificationStats, error) {
	whereClause := fmt.Sprintf("WHERE created_at >= NOW() - INTERVAL '%d days'", days)
	args := []interface{}{}
	argIndex := 1

	if channel != nil {
		whereClause += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, *channel)
		argIndex++
	}

	if provider != nil {
		whereClause += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, *provider)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_sent,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as total_delivered,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as total_failed,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as total_pending
		FROM notification_logs %s`, whereClause)

	var stats models.NotificationStats
	err := s.db.QueryRow(query, args...).Scan(
		&stats.TotalSent, &stats.TotalDelivered, &stats.TotalFailed, &stats.TotalPending)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Calculate rates
	if stats.TotalSent > 0 {
		stats.DeliveryRate = float64(stats.TotalDelivered) / float64(stats.TotalSent) * 100
		stats.FailureRate = float64(stats.TotalFailed) / float64(stats.TotalSent) * 100
	}

	return &stats, nil
}

// GetChannelStats retrieves statistics per channel/provider
func (s *NotificationService) GetChannelStats(days int) ([]*models.ChannelStats, error) {
	query := `
		SELECT 
			channel, provider,
			COUNT(*) as total_sent,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as total_delivered,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as total_failed,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as total_pending,
			MAX(sent_at) as last_sent
		FROM notification_logs 
		WHERE created_at >= NOW() - INTERVAL '%d days'
		GROUP BY channel, provider
		ORDER BY total_sent DESC`

	rows, err := s.db.Query(fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("failed to query channel stats: %w", err)
	}
	defer rows.Close()

	var channelStats []*models.ChannelStats
	for rows.Next() {
		var cs models.ChannelStats
		err := rows.Scan(&cs.Channel, &cs.Provider, &cs.Stats.TotalSent,
			&cs.Stats.TotalDelivered, &cs.Stats.TotalFailed, &cs.Stats.TotalPending,
			&cs.LastSent)
		if err != nil {
			continue
		}

		// Calculate rates
		if cs.Stats.TotalSent > 0 {
			cs.Stats.DeliveryRate = float64(cs.Stats.TotalDelivered) / float64(cs.Stats.TotalSent) * 100
			cs.Stats.FailureRate = float64(cs.Stats.TotalFailed) / float64(cs.Stats.TotalSent) * 100
		}

		channelStats = append(channelStats, &cs)
	}

	return channelStats, nil
}

// Private helper methods

// getDefaultProvider returns the default provider for a channel
func (s *NotificationService) getDefaultProvider(channel models.NotificationChannel) models.NotificationProvider {
	switch channel {
	case models.ChannelSMS:
		return models.ProviderFast2SMS
	case models.ChannelEmail:
		return models.ProviderSMTP
	case models.ChannelPush:
		return models.ProviderFCM
	case models.ChannelWhatsApp:
		return models.ProviderWhatsAppCloud
	}
	return models.ProviderFast2SMS
}

// getProviderConfig retrieves configuration for a provider
func (s *NotificationService) getProviderConfig(provider models.NotificationProvider, channel models.NotificationChannel) (map[string]string, error) {
	query := `SELECT config_key, config_value FROM notification_config 
		WHERE provider = $1 AND channel = $2 AND is_active = true`

	rows, err := s.db.Query(query, provider, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		config[key] = value
	}

	return config, nil
}

// processTemplate processes template with variables
func (s *NotificationService) processTemplate(template *models.NotificationTemplate, variables map[string]interface{}) (string, error) {
	return s.processTemplateString(template.Body, variables)
}

// processTemplateString processes a template string with variables
func (s *NotificationService) processTemplateString(templateStr string, variables map[string]interface{}) (string, error) {
	result := templateStr
	
	if variables != nil {
		for key, value := range variables {
			placeholder := fmt.Sprintf("{%s}", key)
			replacement := fmt.Sprintf("%v", value)
			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}
	
	return result, nil
}

// createNotificationLog creates a log entry
func (s *NotificationService) createNotificationLog(request *models.SendNotificationRequest, template *models.NotificationTemplate, provider models.NotificationProvider) (int64, error) {
	query := `
		INSERT INTO notification_logs (template_id, channel, provider, recipient, subject, body, 
			status, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id`

	var logID int64
	var templateID *int64
	if template != nil {
		templateID = &template.ID
	}

	subject := ""
	if request.Subject != nil {
		subject = *request.Subject
	}

	body := ""
	if request.Body != nil {
		body = *request.Body
	}

	err := s.db.QueryRow(query, templateID, request.Channel, provider, request.Recipient,
		subject, body, models.StatusPending, request.UserID).Scan(&logID)

	return logID, err
}

// updateNotificationLog updates a log entry
func (s *NotificationService) updateNotificationLog(logID int64, status models.NotificationStatus, providerID *string, errorMsg *string) {
	query := `UPDATE notification_logs SET status = $2, provider_id = $3, error_msg = $4, sent_at = NOW() WHERE id = $1`
	s.db.Exec(query, logID, status, providerID, errorMsg)
}

// getFilteredRecipients gets recipients based on user filter
func (s *NotificationService) getFilteredRecipients(channel models.NotificationChannel, filter *models.UserFilter) ([]string, error) {
	var whereClause []string
	var args []interface{}
	argIndex := 1

	if filter.KYCStatus != nil {
		whereClause = append(whereClause, fmt.Sprintf("kyc_status = $%d", argIndex))
		args = append(args, *filter.KYCStatus)
		argIndex++
	}

	if filter.AccountStatus != nil {
		whereClause = append(whereClause, fmt.Sprintf("account_status = $%d", argIndex))
		args = append(args, *filter.AccountStatus)
		argIndex++
	}

	if filter.InactiveDays != nil {
		whereClause = append(whereClause, fmt.Sprintf("last_login_at < NOW() - INTERVAL '%d days'", *filter.InactiveDays))
	}

	if len(filter.States) > 0 {
		placeholders := make([]string, len(filter.States))
		for i, state := range filter.States {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, state)
			argIndex++
		}
		whereClause = append(whereClause, fmt.Sprintf("state IN (%s)", strings.Join(placeholders, ",")))
	}

	var selectField string
	switch channel {
	case models.ChannelSMS, models.ChannelWhatsApp:
		selectField = "mobile"
	case models.ChannelEmail:
		selectField = "email"
	default:
		selectField = "mobile"
	}

	query := fmt.Sprintf("SELECT %s FROM users WHERE is_active = true", selectField)
	if len(whereClause) > 0 {
		query += " AND " + strings.Join(whereClause, " AND ")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered recipients: %w", err)
	}
	defer rows.Close()

	var recipients []string
	for rows.Next() {
		var recipient sql.NullString
		if err := rows.Scan(&recipient); err != nil {
			continue
		}
		if recipient.Valid && recipient.String != "" {
			recipients = append(recipients, recipient.String)
		}
	}

	return recipients, nil
}