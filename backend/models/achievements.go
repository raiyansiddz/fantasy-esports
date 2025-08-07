package models

import (
	"time"
	"encoding/json"
)

type Achievement struct {
	ID              int64           `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	BadgeIcon       *string         `json:"badge_icon" db:"badge_icon"`
	BadgeColor      string          `json:"badge_color" db:"badge_color"`
	Category        string          `json:"category" db:"category"`
	TriggerType     string          `json:"trigger_type" db:"trigger_type"`
	TriggerCriteria json.RawMessage `json:"trigger_criteria" db:"trigger_criteria"`
	RewardType      *string         `json:"reward_type" db:"reward_type"`
	RewardValue     float64         `json:"reward_value" db:"reward_value"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	IsHidden        bool            `json:"is_hidden" db:"is_hidden"`
	SortOrder       int             `json:"sort_order" db:"sort_order"`
	CreatedBy       int64           `json:"created_by" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type UserAchievement struct {
	ID           int64           `json:"id" db:"id"`
	UserID       int64           `json:"user_id" db:"user_id"`
	AchievementID int64          `json:"achievement_id" db:"achievement_id"`
	EarnedAt     time.Time       `json:"earned_at" db:"earned_at"`
	ProgressData json.RawMessage `json:"progress_data" db:"progress_data"`
	IsFeatured   bool            `json:"is_featured" db:"is_featured"`
	
	// Joined fields
	Achievement  *Achievement    `json:"achievement,omitempty"`
}

type CreateAchievementRequest struct {
	Name            string                 `json:"name" validate:"required"`
	Description     string                 `json:"description" validate:"required"`
	BadgeIcon       *string               `json:"badge_icon"`
	BadgeColor      string                `json:"badge_color"`
	Category        string                `json:"category" validate:"required"`
	TriggerType     string                `json:"trigger_type" validate:"required"`
	TriggerCriteria map[string]interface{} `json:"trigger_criteria" validate:"required"`
	RewardType      *string               `json:"reward_type"`
	RewardValue     float64               `json:"reward_value"`
	IsHidden        bool                  `json:"is_hidden"`
	SortOrder       int                   `json:"sort_order"`
}

type AchievementProgress struct {
	AchievementID   int64                  `json:"achievement_id"`
	CurrentProgress map[string]interface{} `json:"current_progress"`
	IsCompleted     bool                   `json:"is_completed"`
	CompletionRate  float64               `json:"completion_rate"`
}