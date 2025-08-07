package services

import (
	"database/sql"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	"strings"
)

// ========================= SEO CONTENT MANAGEMENT =========================

func (s *ContentService) CreateSEOContent(req *models.SEOContentCreateRequest, createdBy int64) (*models.SEOContent, error) {
	query := `
		INSERT INTO seo_content (page_type, page_slug, meta_title, meta_description, keywords, 
			og_title, og_description, og_image, twitter_card, structured_data, content, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`
	
	var seoContent models.SEOContent
	err := s.db.QueryRow(query, req.PageType, req.PageSlug, req.MetaTitle, req.MetaDesc,
		req.Keywords, req.OGTitle, req.OGDesc, req.OGImage, req.TwitterCard,
		req.StructData, req.Content, createdBy).
		Scan(&seoContent.ID, &seoContent.CreatedAt, &seoContent.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create SEO content: %w", err)
	}

	seoContent.PageType = req.PageType
	seoContent.PageSlug = req.PageSlug
	seoContent.MetaTitle = req.MetaTitle
	seoContent.MetaDesc = req.MetaDesc
	seoContent.Keywords = req.Keywords
	seoContent.OGTitle = req.OGTitle
	seoContent.OGDesc = req.OGDesc
	seoContent.OGImage = req.OGImage
	seoContent.TwitterCard = req.TwitterCard
	seoContent.StructData = req.StructData
	seoContent.Content = req.Content
	seoContent.IsActive = true
	seoContent.CreatedBy = createdBy

	return &seoContent, nil
}

func (s *ContentService) UpdateSEOContent(id int64, req *models.SEOContentCreateRequest, updatedBy int64) error {
	query := `
		UPDATE seo_content 
		SET page_type = $2, page_slug = $3, meta_title = $4, meta_description = $5, 
			keywords = $6, og_title = $7, og_description = $8, og_image = $9, 
			twitter_card = $10, structured_data = $11, content = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $13`
	
	result, err := s.db.Exec(query, id, req.PageType, req.PageSlug, req.MetaTitle, req.MetaDesc,
		req.Keywords, req.OGTitle, req.OGDesc, req.OGImage, req.TwitterCard,
		req.StructData, req.Content, updatedBy)
	
	if err != nil {
		return fmt.Errorf("failed to update SEO content: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("SEO content not found or unauthorized")
	}

	return nil
}

func (s *ContentService) GetSEOContent(id int64) (*models.SEOContentResponse, error) {
	query := `
		SELECT s.id, s.page_type, s.page_slug, s.meta_title, s.meta_description, s.keywords,
			COALESCE(s.og_title, '') as og_title, COALESCE(s.og_description, '') as og_description, 
			COALESCE(s.og_image, '') as og_image, COALESCE(s.twitter_card, '') as twitter_card, 
			s.structured_data, COALESCE(s.content, '') as content, s.is_active, s.created_by, 
			s.created_at, s.updated_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM seo_content s
		LEFT JOIN admin_users au ON s.created_by = au.id
		WHERE s.id = $1`
	
	var seoContent models.SEOContentResponse
	err := s.db.QueryRow(query, id).Scan(
		&seoContent.ID, &seoContent.PageType, &seoContent.PageSlug, &seoContent.MetaTitle,
		&seoContent.MetaDesc, &seoContent.Keywords, &seoContent.OGTitle, &seoContent.OGDesc,
		&seoContent.OGImage, &seoContent.TwitterCard, &seoContent.StructData, &seoContent.Content,
		&seoContent.IsActive, &seoContent.CreatedBy, &seoContent.CreatedAt, &seoContent.UpdatedAt,
		&seoContent.CreatorName)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SEO content not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SEO content: %w", err)
	}

	return &seoContent, nil
}

func (s *ContentService) GetSEOContentBySlug(slug string) (*models.SEOContent, error) {
	query := `
		SELECT id, page_type, page_slug, meta_title, meta_description, keywords,
			og_title, og_description, og_image, twitter_card, structured_data,
			content, is_active, created_by, created_at, updated_at
		FROM seo_content
		WHERE page_slug = $1 AND is_active = true`
	
	var seoContent models.SEOContent
	err := s.db.QueryRow(query, slug).Scan(
		&seoContent.ID, &seoContent.PageType, &seoContent.PageSlug, &seoContent.MetaTitle,
		&seoContent.MetaDesc, &seoContent.Keywords, &seoContent.OGTitle, &seoContent.OGDesc,
		&seoContent.OGImage, &seoContent.TwitterCard, &seoContent.StructData, &seoContent.Content,
		&seoContent.IsActive, &seoContent.CreatedBy, &seoContent.CreatedAt, &seoContent.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SEO content not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SEO content by slug: %w", err)
	}

	return &seoContent, nil
}

func (s *ContentService) ListSEOContent(page, limit int, pageType string, active *bool) (*models.SEOContentListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if pageType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("s.page_type = $%d", argIndex))
		args = append(args, pageType)
		argIndex++
	}

	if active != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("s.is_active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM seo_content s WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count SEO content: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT s.id, s.page_type, s.page_slug, s.meta_title, s.meta_description, s.keywords,
			s.og_title, s.og_description, s.og_image, s.twitter_card, s.structured_data,
			s.content, s.is_active, s.created_by, s.created_at, s.updated_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM seo_content s
		LEFT JOIN admin_users au ON s.created_by = au.id
		WHERE %s
		ORDER BY s.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list SEO content: %w", err)
	}
	defer rows.Close()

	var contents []models.SEOContentResponse
	for rows.Next() {
		var content models.SEOContentResponse
		err := rows.Scan(
			&content.ID, &content.PageType, &content.PageSlug, &content.MetaTitle,
			&content.MetaDesc, &content.Keywords, &content.OGTitle, &content.OGDesc,
			&content.OGImage, &content.TwitterCard, &content.StructData, &content.Content,
			&content.IsActive, &content.CreatedBy, &content.CreatedAt, &content.UpdatedAt,
			&content.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SEO content: %w", err)
		}
		contents = append(contents, content)
	}

	if contents == nil {
		contents = []models.SEOContentResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.SEOContentListResponse{
		Contents: contents,
		Total:    total,
		Page:     page,
		Pages:    pages,
		Success:  true,
	}, nil
}

func (s *ContentService) DeleteSEOContent(id, deletedBy int64) error {
	query := `DELETE FROM seo_content WHERE id = $1 AND created_by = $2`
	result, err := s.db.Exec(query, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete SEO content: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("SEO content not found or unauthorized")
	}

	return nil
}

// ========================= FAQ MANAGEMENT =========================

func (s *ContentService) CreateFAQSection(req *models.FAQSectionCreateRequest, createdBy int64) (*models.FAQSection, error) {
	query := `
		INSERT INTO faq_sections (name, description, sort_order, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	
	var section models.FAQSection
	err := s.db.QueryRow(query, req.Name, req.Description, req.SortOrder, createdBy).
		Scan(&section.ID, &section.CreatedAt, &section.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create FAQ section: %w", err)
	}

	section.Name = req.Name
	section.Description = req.Description
	section.SortOrder = req.SortOrder
	section.IsActive = true
	section.CreatedBy = createdBy

	return &section, nil
}

func (s *ContentService) UpdateFAQSection(id int64, req *models.FAQSectionCreateRequest, updatedBy int64) error {
	query := `
		UPDATE faq_sections 
		SET name = $2, description = $3, sort_order = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $5`
	
	result, err := s.db.Exec(query, id, req.Name, req.Description, req.SortOrder, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to update FAQ section: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("FAQ section not found or unauthorized")
	}

	return nil
}

func (s *ContentService) ListFAQSections(page, limit int, active *bool) (*models.FAQSectionListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if active != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("fs.is_active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM faq_sections fs WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count FAQ sections: %w", err)
	}

	// Data query with item count
	dataQuery := fmt.Sprintf(`
		SELECT fs.id, fs.name, fs.description, fs.sort_order, fs.is_active, fs.created_by, 
			fs.created_at, fs.updated_at,
			COALESCE(au.full_name, au.username) as creator_name,
			COALESCE(ic.item_count, 0) as item_count
		FROM faq_sections fs
		LEFT JOIN admin_users au ON fs.created_by = au.id
		LEFT JOIN (
			SELECT section_id, COUNT(*) as item_count 
			FROM faq_items 
			WHERE is_active = true 
			GROUP BY section_id
		) ic ON fs.id = ic.section_id
		WHERE %s
		ORDER BY fs.sort_order, fs.created_at
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list FAQ sections: %w", err)
	}
	defer rows.Close()

	var sections []models.FAQSectionResponse
	for rows.Next() {
		var section models.FAQSectionResponse
		err := rows.Scan(
			&section.ID, &section.Name, &section.Description, &section.SortOrder,
			&section.IsActive, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt,
			&section.CreatorName, &section.ItemCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan FAQ section: %w", err)
		}
		sections = append(sections, section)
	}

	if sections == nil {
		sections = []models.FAQSectionResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.FAQSectionListResponse{
		Sections: sections,
		Total:    total,
		Page:     page,
		Pages:    pages,
		Success:  true,
	}, nil
}

func (s *ContentService) CreateFAQItem(req *models.FAQItemCreateRequest, createdBy int64) (*models.FAQItem, error) {
	// First verify the section exists
	var sectionExists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM faq_sections WHERE id = $1 AND is_active = true)", 
		req.SectionID).Scan(&sectionExists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify FAQ section: %w", err)
	}
	if !sectionExists {
		return nil, fmt.Errorf("FAQ section not found")
	}

	query := `
		INSERT INTO faq_items (section_id, question, answer, sort_order, tags, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	
	var item models.FAQItem
	err = s.db.QueryRow(query, req.SectionID, req.Question, req.Answer, req.SortOrder, req.Tags, createdBy).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create FAQ item: %w", err)
	}

	item.SectionID = req.SectionID
	item.Question = req.Question
	item.Answer = req.Answer
	item.SortOrder = req.SortOrder
	item.Tags = req.Tags
	item.IsActive = true
	item.CreatedBy = createdBy

	return &item, nil
}

func (s *ContentService) UpdateFAQItem(id int64, req *models.FAQItemCreateRequest, updatedBy int64) error {
	query := `
		UPDATE faq_items 
		SET section_id = $2, question = $3, answer = $4, sort_order = $5, tags = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $7`
	
	result, err := s.db.Exec(query, id, req.SectionID, req.Question, req.Answer, req.SortOrder, req.Tags, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to update FAQ item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("FAQ item not found or unauthorized")
	}

	return nil
}

func (s *ContentService) ListFAQItems(page, limit int, sectionID *int64, active *bool) (*models.FAQItemListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if sectionID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("fi.section_id = $%d", argIndex))
		args = append(args, *sectionID)
		argIndex++
	}

	if active != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("fi.is_active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM faq_items fi WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count FAQ items: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT fi.id, fi.section_id, fi.question, fi.answer, fi.sort_order, fi.is_active, 
			fi.view_count, fi.like_count, fi.tags, fi.created_by, fi.created_at, fi.updated_at,
			fs.name as section_name,
			COALESCE(au.full_name, au.username) as creator_name
		FROM faq_items fi
		LEFT JOIN faq_sections fs ON fi.section_id = fs.id
		LEFT JOIN admin_users au ON fi.created_by = au.id
		WHERE %s
		ORDER BY fi.section_id, fi.sort_order, fi.created_at
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list FAQ items: %w", err)
	}
	defer rows.Close()

	var items []models.FAQItemResponse
	for rows.Next() {
		var item models.FAQItemResponse
		err := rows.Scan(
			&item.ID, &item.SectionID, &item.Question, &item.Answer, &item.SortOrder,
			&item.IsActive, &item.ViewCount, &item.LikeCount, &item.Tags,
			&item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
			&item.SectionName, &item.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan FAQ item: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.FAQItemResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.FAQItemListResponse{
		Items:   items,
		Total:   total,
		Page:    page,
		Pages:   pages,
		Success: true,
	}, nil
}

func (s *ContentService) IncrementFAQView(id int64) error {
	query := `UPDATE faq_items SET view_count = view_count + 1 WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *ContentService) IncrementFAQLike(id int64) error {
	query := `UPDATE faq_items SET like_count = like_count + 1 WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// ========================= LEGAL DOCUMENT MANAGEMENT =========================

func (s *ContentService) CreateLegalDocument(req *models.LegalDocumentCreateRequest, createdBy int64) (*models.LegalDocument, error) {
	query := `
		INSERT INTO legal_documents (document_type, title, content, version, effective_date, metadata, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	
	var document models.LegalDocument
	err := s.db.QueryRow(query, req.DocumentType, req.Title, req.Content, req.Version, 
		req.EffectiveDate, req.Metadata, createdBy).
		Scan(&document.ID, &document.CreatedAt, &document.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create legal document: %w", err)
	}

	document.DocumentType = req.DocumentType
	document.Title = req.Title
	document.Content = req.Content
	document.Version = req.Version
	document.EffectiveDate = req.EffectiveDate
	document.Metadata = req.Metadata
	document.Status = "draft"
	document.IsActive = false
	document.CreatedBy = createdBy

	return &document, nil
}

func (s *ContentService) UpdateLegalDocument(id int64, req *models.LegalDocumentCreateRequest, updatedBy int64) error {
	query := `
		UPDATE legal_documents 
		SET title = $2, content = $3, version = $4, effective_date = $5, metadata = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $7`
	
	result, err := s.db.Exec(query, id, req.Title, req.Content, req.Version, 
		req.EffectiveDate, req.Metadata, updatedBy)
	
	if err != nil {
		return fmt.Errorf("failed to update legal document: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("legal document not found or unauthorized")
	}

	return nil
}

func (s *ContentService) PublishLegalDocument(id int64, publishedBy int64) error {
	// First deactivate any existing active document of the same type
	updateQuery := `
		UPDATE legal_documents 
		SET is_active = false 
		WHERE document_type = (SELECT document_type FROM legal_documents WHERE id = $1)
		AND is_active = true`
	
	_, err := s.db.Exec(updateQuery, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate existing documents: %w", err)
	}

	// Now publish the new version
	publishQuery := `
		UPDATE legal_documents 
		SET status = 'published', is_active = true, published_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND created_by = $2`
	
	result, err := s.db.Exec(publishQuery, id, publishedBy)
	if err != nil {
		return fmt.Errorf("failed to publish legal document: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("legal document not found or unauthorized")
	}

	return nil
}

func (s *ContentService) GetActiveLegalDocument(documentType string) (*models.LegalDocument, error) {
	query := `
		SELECT id, document_type, title, content, version, effective_date, status, is_active,
			metadata, created_by, created_at, updated_at, published_at
		FROM legal_documents
		WHERE document_type = $1 AND is_active = true AND status = 'published'
		ORDER BY effective_date DESC
		LIMIT 1`
	
	var document models.LegalDocument
	err := s.db.QueryRow(query, documentType).Scan(
		&document.ID, &document.DocumentType, &document.Title, &document.Content,
		&document.Version, &document.EffectiveDate, &document.Status, &document.IsActive,
		&document.Metadata, &document.CreatedBy, &document.CreatedAt, &document.UpdatedAt,
		&document.PublishedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active legal document found for type: %s", documentType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active legal document: %w", err)
	}

	return &document, nil
}

func (s *ContentService) ListLegalDocuments(page, limit int, docType, status string) (*models.LegalDocumentListResponse, error) {
	offset := (page - 1) * limit
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if docType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("ld.document_type = $%d", argIndex))
		args = append(args, docType)
		argIndex++
	}

	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("ld.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM legal_documents ld WHERE %s`, whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count legal documents: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT ld.id, ld.document_type, ld.title, ld.content, ld.version, ld.effective_date, 
			ld.status, ld.is_active, ld.metadata, ld.created_by, ld.created_at, ld.updated_at, 
			ld.published_at,
			COALESCE(au.full_name, au.username) as creator_name
		FROM legal_documents ld
		LEFT JOIN admin_users au ON ld.created_by = au.id
		WHERE %s
		ORDER BY ld.document_type, ld.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)
	
	args = append(args, limit, offset)
	
	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list legal documents: %w", err)
	}
	defer rows.Close()

	var documents []models.LegalDocumentResponse
	for rows.Next() {
		var document models.LegalDocumentResponse
		err := rows.Scan(
			&document.ID, &document.DocumentType, &document.Title, &document.Content,
			&document.Version, &document.EffectiveDate, &document.Status, &document.IsActive,
			&document.Metadata, &document.CreatedBy, &document.CreatedAt, &document.UpdatedAt,
			&document.PublishedAt, &document.CreatorName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan legal document: %w", err)
		}
		documents = append(documents, document)
	}

	if documents == nil {
		documents = []models.LegalDocumentResponse{}
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	
	return &models.LegalDocumentListResponse{
		Documents: documents,
		Total:     total,
		Page:      page,
		Pages:     pages,
		Success:   true,
	}, nil
}

func (s *ContentService) DeleteLegalDocument(id, deletedBy int64) error {
	// Don't allow deleting active/published documents
	query := `DELETE FROM legal_documents WHERE id = $1 AND created_by = $2 AND status = 'draft'`
	result, err := s.db.Exec(query, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete legal document: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("legal document not found, unauthorized, or cannot delete published documents")
	}

	return nil
}

// ========================= CONTENT ANALYTICS =========================

func (s *ContentService) RecordContentView(contentType string, contentID int64) error {
	query := `
		INSERT INTO content_analytics (content_type, content_id, date, views)
		VALUES ($1, $2, CURRENT_DATE, 1)
		ON CONFLICT (content_type, content_id, date)
		DO UPDATE SET views = content_analytics.views + 1`
	
	_, err := s.db.Exec(query, contentType, contentID)
	return err
}

func (s *ContentService) RecordContentClick(contentType string, contentID int64) error {
	query := `
		INSERT INTO content_analytics (content_type, content_id, date, clicks)
		VALUES ($1, $2, CURRENT_DATE, 1)
		ON CONFLICT (content_type, content_id, date)
		DO UPDATE SET clicks = content_analytics.clicks + 1`
	
	_, err := s.db.Exec(query, contentType, contentID)
	return err
}

func (s *ContentService) GetContentAnalytics(contentType string, contentID int64, days int) ([]models.ContentAnalytics, error) {
	query := `
		SELECT id, content_type, content_id, date, views, clicks, engagements, conversions, metrics, created_at
		FROM content_analytics
		WHERE content_type = $1 AND content_id = $2 AND date >= CURRENT_DATE - INTERVAL '%d days'
		ORDER BY date DESC`
	
	formattedQuery := fmt.Sprintf(query, days)
	
	rows, err := s.db.Query(formattedQuery, contentType, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content analytics: %w", err)
	}
	defer rows.Close()

	var analytics []models.ContentAnalytics
	for rows.Next() {
		var analytic models.ContentAnalytics
		err := rows.Scan(
			&analytic.ID, &analytic.ContentType, &analytic.ContentID, &analytic.Date,
			&analytic.Views, &analytic.Clicks, &analytic.Engagements, &analytic.Conversions,
			&analytic.Metrics, &analytic.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content analytics: %w", err)
		}
		analytics = append(analytics, analytic)
	}

	if analytics == nil {
		analytics = []models.ContentAnalytics{}
	}

	return analytics, nil
}