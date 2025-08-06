package services

import (
	"database/sql"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	""
	"strings"
	"time"
)

type ContentService struct {
	db *sql.DB
}

func NewContentService(db *sql.DB) *ContentService {
	return &ContentService{db: db}
}

// ========================= BANNER MANAGEMENT =========================

func (s *ContentService) CreateBanner(req *models.BannerCreateRequest, createdBy int64) (*models.Banner, error) {
	query := `
		INSERT INTO banners (title, description, image_url, link_url, position, type, priority, 
			start_date, end_date, target_roles, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`
	
	var banner models.Banner
	err := s.db.QueryRow(query, req.Title, req.Description, req.ImageURL, req.LinkURL,
		req.Position, req.Type, req.Priority, req.StartDate, req.EndDate,
		req.TargetRoles, req.Metadata, createdBy).Scan(&banner.ID, &banner.CreatedAt, &banner.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create banner: %w", err)
	}

	// Set other fields
	banner.Title = req.Title
	banner.Description = req.Description
	banner.ImageURL = req.ImageURL
	banner.LinkURL = req.LinkURL
	banner.Position = req.Position
	banner.Type = req.Type
	banner.Priority = req.Priority
	banner.StartDate = req.StartDate
	banner.EndDate = req.EndDate
	banner.TargetRoles = req.TargetRoles
	banner.Metadata = req.Metadata
	banner.IsActive = true
	banner.CreatedBy = createdBy

	return &banner, nil
}

func (s *ContentService) UpdateBanner(id int64, req *models.BannerCreateRequest, updatedBy int64) error {
	query := `
		UPDATE banners 
		SET title = $2, description = $3, image_url = $4, link_url = $5, position = $6, 
			type = $7, priority = $8, start_date = $9, end_date = $10, target_roles = $11, 
			metadata = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $13`
	
	result, err := s.db.Exec(query, id, req.Title, req.Description, req.ImageURL, req.LinkURL,
		req.Position, req.Type, req.Priority, req.StartDate, req.EndDate,
		req.TargetRoles, req.Metadata, updatedBy)
	
	if err != nil {
		return fmt.Errorf("failed to update banner: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("banner not found or unauthorized")
	}

	return nil
}

func (s *ContentService) GetBanner(id int64) (*models.BannerResponse, error) {
	query := `
		SELECT b.id, b.title, b.description, b.image_url, b.link_url, b.position, b.type, 
			b.priority, b.start_date, b.end_date, b.is_active, b.target_roles, b.metadata, 
			b.click_count, b.view_count, b.created_by, b.created_at, b.updated_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM banners b
		LEFT JOIN admin_users au ON b.created_by = au.id
		WHERE b.id = $1`
	
	var banner models.BannerResponse
	err := s.db.QueryRow(query, id).Scan(
		&banner.ID, &banner.Title, &banner.Description, &banner.ImageURL, &banner.LinkURL,
		&banner.Position, &banner.Type, &banner.Priority, &banner.StartDate, &banner.EndDate,
		&banner.IsActive, &banner.TargetRoles, &banner.Metadata, &banner.ClickCount,
		&banner.ViewCount, &banner.CreatedBy, &banner.CreatedAt, &banner.UpdatedAt,
		&banner.CreatorName)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("banner not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get banner: %w", err)
	}

	return &banner, nil
}

func (s *ContentService) ListBanners(page, limit int, position, bannerType, status string) (*models.BannerListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if position != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.position = $%d", argIndex))
		args = append(args, position)
		argIndex++
	}

	if bannerType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.type = $%d", argIndex))
		args = append(args, bannerType)
		argIndex++
	}

	if status == "active" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.is_active = true AND b.start_date <= CURRENT_TIMESTAMP AND b.end_date >= CURRENT_TIMESTAMP"))
	} else if status == "inactive" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.is_active = false OR b.start_date > CURRENT_TIMESTAMP OR b.end_date < CURRENT_TIMESTAMP"))
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM banners b WHERE %s`, whereClause)
	
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count banners: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT b.id, b.title, b.description, b.image_url, b.link_url, b.position, b.type, 
			b.priority, b.start_date, b.end_date, b.is_active, b.target_roles, b.metadata, 
			b.click_count, b.view_count, b.created_by, b.created_at, b.updated_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM banners b
		LEFT JOIN admin_users au ON b.created_by = au.id
		WHERE %s
		ORDER BY b.priority DESC, b.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list banners: %w", err)
	}
	defer rows.Close()

	var banners []models.BannerResponse
	for rows.Next() {
		var banner models.BannerResponse
		err := rows.Scan(
			&banner.ID, &banner.Title, &banner.Description, &banner.ImageURL, &banner.LinkURL,
			&banner.Position, &banner.Type, &banner.Priority, &banner.StartDate, &banner.EndDate,
			&banner.IsActive, &banner.TargetRoles, &banner.Metadata, &banner.ClickCount,
			&banner.ViewCount, &banner.CreatedBy, &banner.CreatedAt, &banner.UpdatedAt,
			&banner.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan banner: %w", err)
		}
		banners = append(banners, banner)
	}

	if banners == nil {
		banners = []models.BannerResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.BannerListResponse{
		Banners: banners,
		Total:   total,
		Page:    page,
		Pages:   pages,
		Success: true,
	}, nil
}

func (s *ContentService) DeleteBanner(id, deletedBy int64) error {
	query := `DELETE FROM banners WHERE id = $1 AND created_by = $2`
	result, err := s.db.Exec(query, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete banner: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("banner not found or unauthorized")
	}

	return nil
}

func (s *ContentService) ToggleBannerStatus(id int64, updatedBy int64) error {
	query := `
		UPDATE banners 
		SET is_active = NOT is_active, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $2`
	
	result, err := s.db.Exec(query, id, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to toggle banner status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("banner not found or unauthorized")
	}

	return nil
}

func (s *ContentService) IncrementBannerView(id int64) error {
	query := `UPDATE banners SET view_count = view_count + 1 WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *ContentService) IncrementBannerClick(id int64) error {
	query := `UPDATE banners SET click_count = click_count + 1 WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// ========================= EMAIL TEMPLATE MANAGEMENT =========================

func (s *ContentService) CreateEmailTemplate(name, description, subject, htmlContent, textContent, category string, variables models.JSONMap, createdBy int64) (*models.EmailTemplate, error) {
	query := `
		INSERT INTO email_templates (name, description, subject, html_content, text_content, category, variables, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	
	var template models.EmailTemplate
	err := s.db.QueryRow(query, name, description, subject, htmlContent, textContent, category, variables, createdBy).
		Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	template.Name = name
	template.Description = description
	template.Subject = subject
	template.HTMLContent = htmlContent
	template.TextContent = textContent
	template.Category = category
	template.Variables = variables
	template.IsActive = true
	template.CreatedBy = createdBy

	return &template, nil
}

func (s *ContentService) ListEmailTemplates(page, limit int, category string, active *bool) ([]models.EmailTemplate, int64, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if active != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM email_templates WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count email templates: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, name, description, subject, html_content, text_content, category, 
			variables, is_active, created_by, created_at, updated_at
		FROM email_templates 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list email templates: %w", err)
	}
	defer rows.Close()

	var templates []models.EmailTemplate
	for rows.Next() {
		var template models.EmailTemplate
		err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.Subject,
			&template.HTMLContent, &template.TextContent, &template.Category,
			&template.Variables, &template.IsActive, &template.CreatedBy,
			&template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan email template: %w", err)
		}
		templates = append(templates, template)
	}

	if templates == nil {
		templates = []models.EmailTemplate{}
	}

	return templates, total, nil
}

// ========================= MARKETING CAMPAIGN MANAGEMENT =========================

func (s *ContentService) CreateMarketingCampaign(req *models.MarketingCampaignCreateRequest, createdBy int64) (*models.MarketingCampaign, error) {
	query := `
		INSERT INTO marketing_campaigns (name, subject, email_template, target_segment, target_criteria, scheduled_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	
	var campaign models.MarketingCampaign
	err := s.db.QueryRow(query, req.Name, req.Subject, req.EmailTemplate, req.TargetSegment, 
		req.TargetCriteria, req.ScheduledAt, createdBy).
		Scan(&campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create marketing campaign: %w", err)
	}

	campaign.Name = req.Name
	campaign.Subject = req.Subject
	campaign.EmailTemplate = req.EmailTemplate
	campaign.TargetSegment = req.TargetSegment
	campaign.TargetCriteria = req.TargetCriteria
	campaign.ScheduledAt = req.ScheduledAt
	campaign.Status = "draft"
	campaign.CreatedBy = createdBy

	return &campaign, nil
}

func (s *ContentService) ListMarketingCampaigns(page, limit int, status, segment string) (*models.CampaignListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("c.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if segment != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("c.target_segment = $%d", argIndex))
		args = append(args, segment)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM marketing_campaigns c WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT c.id, c.name, c.subject, c.email_template, c.target_segment, c.target_criteria, 
			c.scheduled_at, c.status, c.total_recipients, c.sent_count, c.delivered_count, 
			c.open_count, c.click_count, c.unsubscribe_count, c.bounce_count, c.metadata, 
			c.created_by, c.created_at, c.updated_at, c.sent_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM marketing_campaigns c
		LEFT JOIN admin_users au ON c.created_by = au.id
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.CampaignResponse
	for rows.Next() {
		var campaign models.CampaignResponse
		err := rows.Scan(
			&campaign.ID, &campaign.Name, &campaign.Subject, &campaign.EmailTemplate,
			&campaign.TargetSegment, &campaign.TargetCriteria, &campaign.ScheduledAt,
			&campaign.Status, &campaign.TotalRecipients, &campaign.SentCount,
			&campaign.DeliveredCount, &campaign.OpenCount, &campaign.ClickCount,
			&campaign.UnsubscribeCount, &campaign.BounceCount, &campaign.Metadata,
			&campaign.CreatedBy, &campaign.CreatedAt, &campaign.UpdatedAt,
			&campaign.SentAt, &campaign.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns = append(campaigns, campaign)
	}

	if campaigns == nil {
		campaigns = []models.CampaignResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.CampaignListResponse{
		Campaigns: campaigns,
		Total:     total,
		Page:      page,
		Pages:     pages,
		Success:   true,
	}, nil
}

func (s *ContentService) UpdateCampaignStatus(id int64, status string, updatedBy int64) error {
	validStatuses := map[string]bool{
		"draft": true, "scheduled": true, "sending": true, "sent": true, "cancelled": true,
	}
	
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE marketing_campaigns 
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $3`
	
	result, err := s.db.Exec(query, id, status, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("campaign not found or unauthorized")
	}

	return nil
}

func (s *ContentService) CalculateCampaignRecipients(segment string, criteria models.JSONMap) (int, error) {
	// This is a simplified implementation - in real world you'd have complex segmentation logic
	baseQuery := "SELECT COUNT(*) FROM users WHERE is_verified = true"
	
	switch segment {
	case "all":
		// No additional filters
	case "kyc_verified":
		baseQuery += " AND kyc_status = 'verified'"
	case "high_value":
		baseQuery += " AND id IN (SELECT user_id FROM user_wallets WHERE total_balance > 10000)"
	case "new_users":
		baseQuery += " AND created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'"
	default:
		return 0, fmt.Errorf("unsupported segment: %s", segment)
	}

	var count int
	err := s.db.QueryRow(baseQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate recipients: %w", err)
	}

	return count, nil
}