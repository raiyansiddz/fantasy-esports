package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONMap represents a JSON object
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, j)
}

// StringArray represents a PostgreSQL text array
type StringArray []string

// Value implements the driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// Scan implements the sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, sa)
}

// Banner represents promotional banners and advertisements
type Banner struct {
	ID           int64     `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	ImageURL     string    `json:"image_url" db:"image_url"`
	LinkURL      string    `json:"link_url" db:"link_url"`
	Position     string    `json:"position" db:"position"` // top, middle, bottom, sidebar
	Type         string    `json:"type" db:"type"`         // promotion, announcement, sponsored
	Priority     int       `json:"priority" db:"priority"`
	StartDate    time.Time `json:"start_date" db:"start_date"`
	EndDate      time.Time `json:"end_date" db:"end_date"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	TargetRoles  JSONMap   `json:"target_roles" db:"target_roles"` // user roles to show banner
	Metadata     JSONMap   `json:"metadata" db:"metadata"`
	ClickCount   int64     `json:"click_count" db:"click_count"`
	ViewCount    int64     `json:"view_count" db:"view_count"`
	CreatedBy    int64     `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// MarketingCampaign represents email marketing campaigns
type MarketingCampaign struct {
	ID               int64       `json:"id" db:"id"`
	Name             string      `json:"name" db:"name"`
	Subject          string      `json:"subject" db:"subject"`
	EmailTemplate    string      `json:"email_template" db:"email_template"`
	TargetSegment    string      `json:"target_segment" db:"target_segment"` // all, kyc_verified, high_value, etc.
	TargetCriteria   JSONMap     `json:"target_criteria" db:"target_criteria"`
	ScheduledAt      *time.Time  `json:"scheduled_at" db:"scheduled_at"`
	Status           string      `json:"status" db:"status"` // draft, scheduled, sending, sent, cancelled
	TotalRecipients  int         `json:"total_recipients" db:"total_recipients"`
	SentCount        int         `json:"sent_count" db:"sent_count"`
	DeliveredCount   int         `json:"delivered_count" db:"delivered_count"`
	OpenCount        int         `json:"open_count" db:"open_count"`
	ClickCount       int         `json:"click_count" db:"click_count"`
	UnsubscribeCount int         `json:"unsubscribe_count" db:"unsubscribe_count"`
	BounceCount      int         `json:"bounce_count" db:"bounce_count"`
	Metadata         JSONMap     `json:"metadata" db:"metadata"`
	CreatedBy        int64       `json:"created_by" db:"created_by"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
	SentAt           *time.Time  `json:"sent_at" db:"sent_at"`
}

// EmailTemplate represents reusable email templates
type EmailTemplate struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Subject     string    `json:"subject" db:"subject"`
	HTMLContent string    `json:"html_content" db:"html_content"`
	TextContent string    `json:"text_content" db:"text_content"`
	Category    string    `json:"category" db:"category"` // welcome, promotional, transactional
	Variables   JSONMap   `json:"variables" db:"variables"` // template variables
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SEOContent represents SEO content for different pages
type SEOContent struct {
	ID          int64     `json:"id" db:"id"`
	PageType    string    `json:"page_type" db:"page_type"` // home, games, tournaments, etc.
	PageSlug    string    `json:"page_slug" db:"page_slug"` // unique identifier
	MetaTitle   string    `json:"meta_title" db:"meta_title"`
	MetaDesc    string    `json:"meta_description" db:"meta_description"`
	Keywords    StringArray `json:"keywords" db:"keywords"`
	OGTitle     string    `json:"og_title" db:"og_title"`
	OGDesc      string    `json:"og_description" db:"og_description"`
	OGImage     string    `json:"og_image" db:"og_image"`
	TwitterCard string    `json:"twitter_card" db:"twitter_card"`
	StructData  JSONMap   `json:"structured_data" db:"structured_data"` // JSON-LD structured data
	Content     string    `json:"content" db:"content"` // Page content
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// FAQSection represents FAQ categories/sections
type FAQSection struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// FAQItem represents individual FAQ questions and answers
type FAQItem struct {
	ID         int64     `json:"id" db:"id"`
	SectionID  int64     `json:"section_id" db:"section_id"`
	Question   string    `json:"question" db:"question"`
	Answer     string    `json:"answer" db:"answer"`
	SortOrder  int       `json:"sort_order" db:"sort_order"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	ViewCount  int64     `json:"view_count" db:"view_count"`
	LikeCount  int64     `json:"like_count" db:"like_count"`
	Tags       StringArray `json:"tags" db:"tags"`
	CreatedBy  int64     `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// LegalDocument represents legal documents like Terms, Privacy Policy, etc.
type LegalDocument struct {
	ID            int64      `json:"id" db:"id"`
	DocumentType  string     `json:"document_type" db:"document_type"` // terms, privacy, refund, etc.
	Title         string     `json:"title" db:"title"`
	Content       string     `json:"content" db:"content"`
	Version       string     `json:"version" db:"version"`
	EffectiveDate time.Time  `json:"effective_date" db:"effective_date"`
	Status        string     `json:"status" db:"status"` // draft, published, archived
	IsActive      bool       `json:"is_active" db:"is_active"`
	Metadata      JSONMap    `json:"metadata" db:"metadata"`
	CreatedBy     int64      `json:"created_by" db:"created_by"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt   *time.Time `json:"published_at" db:"published_at"`
}

// ContentAnalytics represents content performance analytics
type ContentAnalytics struct {
	ID           int64     `json:"id" db:"id"`
	ContentType  string    `json:"content_type" db:"content_type"` // banner, campaign, seo, faq, legal
	ContentID    int64     `json:"content_id" db:"content_id"`
	Date         time.Time `json:"date" db:"date"`
	Views        int64     `json:"views" db:"views"`
	Clicks       int64     `json:"clicks" db:"clicks"`
	Engagements  int64     `json:"engagements" db:"engagements"`
	Conversions  int64     `json:"conversions" db:"conversions"`
	Metrics      JSONMap   `json:"metrics" db:"metrics"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Content Management Request/Response DTOs
type BannerCreateRequest struct {
	Title       string    `json:"title" binding:"required,max=200"`
	Description string    `json:"description" binding:"max=500"`
	ImageURL    string    `json:"image_url" binding:"required,url"`
	LinkURL     string    `json:"link_url" binding:"url"`
	Position    string    `json:"position" binding:"required,oneof=top middle bottom sidebar"`
	Type        string    `json:"type" binding:"required,oneof=promotion announcement sponsored"`
	Priority    int       `json:"priority" binding:"min=0,max=100"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	TargetRoles JSONMap   `json:"target_roles"`
	Metadata    JSONMap   `json:"metadata"`
}

type MarketingCampaignCreateRequest struct {
	Name           string     `json:"name" binding:"required,max=200"`
	Subject        string     `json:"subject" binding:"required,max=200"`
	EmailTemplate  string     `json:"email_template" binding:"required"`
	TargetSegment  string     `json:"target_segment" binding:"required"`
	TargetCriteria JSONMap    `json:"target_criteria"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
}

type SEOContentCreateRequest struct {
	PageType    string      `json:"page_type" binding:"required,max=100"`
	PageSlug    string      `json:"page_slug" binding:"required,max=200"`
	MetaTitle   string      `json:"meta_title" binding:"required,max=60"`
	MetaDesc    string      `json:"meta_description" binding:"required,max=160"`
	Keywords    StringArray `json:"keywords"`
	OGTitle     string      `json:"og_title" binding:"max=60"`
	OGDesc      string      `json:"og_description" binding:"max=160"`
	OGImage     string      `json:"og_image" binding:"url"`
	TwitterCard string      `json:"twitter_card"`
	StructData  JSONMap     `json:"structured_data"`
	Content     string      `json:"content"`
}

type FAQSectionCreateRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description" binding:"max=500"`
	SortOrder   int    `json:"sort_order" binding:"min=0"`
}

type FAQItemCreateRequest struct {
	SectionID int64       `json:"section_id" binding:"required"`
	Question  string      `json:"question" binding:"required,max=500"`
	Answer    string      `json:"answer" binding:"required"`
	SortOrder int         `json:"sort_order" binding:"min=0"`
	Tags      StringArray `json:"tags"`
}

type LegalDocumentCreateRequest struct {
	DocumentType  string    `json:"document_type" binding:"required,oneof=terms privacy refund cookie disclaimer"`
	Title         string    `json:"title" binding:"required,max=200"`
	Content       string    `json:"content" binding:"required"`
	Version       string    `json:"version" binding:"required,max=20"`
	EffectiveDate time.Time `json:"effective_date" binding:"required"`
	Metadata      JSONMap   `json:"metadata"`
}

// Response DTOs with additional fields
type BannerResponse struct {
	Banner
	CreatorName string `json:"creator_name"`
}

type CampaignResponse struct {
	MarketingCampaign
	CreatorName string `json:"creator_name"`
}

type SEOContentResponse struct {
	SEOContent
	CreatorName string `json:"creator_name"`
}

type FAQSectionResponse struct {
	FAQSection
	CreatorName string `json:"creator_name"`
	ItemCount   int64  `json:"item_count"`
}

type FAQItemResponse struct {
	FAQItem
	SectionName string `json:"section_name"`
	CreatorName string `json:"creator_name"`
}

type LegalDocumentResponse struct {
	LegalDocument
	CreatorName string `json:"creator_name"`
}

// List responses with pagination
type BannerListResponse struct {
	Banners []BannerResponse `json:"banners"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Pages   int              `json:"pages"`
	Success bool             `json:"success"`
}

type CampaignListResponse struct {
	Campaigns []CampaignResponse `json:"campaigns"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	Pages     int                `json:"pages"`
	Success   bool               `json:"success"`
}

type SEOContentListResponse struct {
	Contents []SEOContentResponse `json:"contents"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	Pages    int                 `json:"pages"`
	Success  bool                `json:"success"`
}

type FAQSectionListResponse struct {
	Sections []FAQSectionResponse `json:"sections"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	Pages    int                 `json:"pages"`
	Success  bool                `json:"success"`
}

type FAQItemListResponse struct {
	Items   []FAQItemResponse `json:"items"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Pages   int              `json:"pages"`
	Success bool             `json:"success"`
}

type LegalDocumentListResponse struct {
	Documents []LegalDocumentResponse `json:"documents"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	Pages     int                     `json:"pages"`
	Success   bool                    `json:"success"`
}