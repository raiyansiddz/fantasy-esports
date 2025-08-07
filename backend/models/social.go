package models

import (
	"time"
	"encoding/json"
)

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
	ShareType string                 `json:"share_type" validate:"required"`
	Platform  string                 `json:"platform" validate:"required"`
	ContentID *int64                 `json:"content_id"`
	ShareData map[string]interface{} `json:"share_data" validate:"required"`
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