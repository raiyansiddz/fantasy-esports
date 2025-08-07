package models

import (
	"time"
	"encoding/json"
	"strconv"
)

// ContentIDValue is a custom type that can handle both string and int64 for content_id
type ContentIDValue struct {
	Value *int64
}

// UnmarshalJSON implements custom JSON unmarshaling for content_id
func (c *ContentIDValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int64 first
	var intVal int64
	if err := json.Unmarshal(data, &intVal); err == nil {
		c.Value = &intVal
		return nil
	}
	
	// Try to unmarshal as string and convert to int64
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		if strVal == "" {
			c.Value = nil
			return nil
		}
		
		if intVal, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			c.Value = &intVal
			return nil
		}
	}
	
	// If both fail, set as nil
	c.Value = nil
	return nil
}

type SocialShare struct {
	ID         int64           `json:"id" db:"id"`
	UserID     int64           `json:"user_id" db:"user_id"`
	ShareType  string          `json:"share_type" db:"share_type"`
	Platform   string          `json:"platform" db:"platform"`
	ContentID  *int64          `json:"content_id" db:"content_id"`
	ShareData  json.RawMessage `json:"share_data" db:"share_data"`
	ShareURL   *string         `json:"share_url" db:"share_url"`
	ClickCount int             `json:"click_count" db:"click_count"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type CreateShareRequest struct {
	ShareType string                 `json:"share_type" validate:"required" binding:"required"`
	Platform  string                 `json:"platform" validate:"required" binding:"required"`
	ContentID ContentIDValue         `json:"content_id"`
	ShareData map[string]interface{} `json:"share_data" validate:"required" binding:"required"`
}

type ShareContent struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	ImageURL    *string                `json:"image_url"`
	URL         string                 `json:"url"`
	Hashtags    []string               `json:"hashtags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PlatformShareURLs struct {
	Twitter   string `json:"twitter"`
	Facebook  string `json:"facebook"`
	WhatsApp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
}