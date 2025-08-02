package models

import (
	"time"
	"encoding/json"
)

type AdminUser struct {
	ID          int64           `json:"id" db:"id"`
	Username    string          `json:"username" db:"username"`
	Email       string          `json:"email" db:"email"`
	PasswordHash string         `json:"-" db:"password_hash"`
	FullName    *string         `json:"full_name" db:"full_name"`
	Role        string          `json:"role" db:"role"`
	Permissions json.RawMessage `json:"permissions" db:"permissions"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	LastLoginAt *time.Time      `json:"last_login_at" db:"last_login_at"`
	CreatedBy   *int64          `json:"created_by" db:"created_by"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type SystemConfig struct {
	ID          int             `json:"id" db:"id"`
	ConfigKey   string          `json:"config_key" db:"config_key"`
	ConfigValue json.RawMessage `json:"config_value" db:"config_value"`
	Description *string         `json:"description" db:"description"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	UpdatedBy   *int64          `json:"updated_by" db:"updated_by"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AdminLoginResponse struct {
	Success     bool      `json:"success"`
	AccessToken string    `json:"access_token"`
	AdminUser   AdminUser `json:"admin_user"`
}

type MatchScoringRequest struct {
	ActualStartTime time.Time    `json:"actual_start_time"`
	InitialSetup    InitialSetup `json:"initial_setup"`
}

type InitialSetup struct {
	Team1Side string `json:"team1_side"`
	Team2Side string `json:"team2_side"`
	Map       string `json:"map"`
}

type AddEventRequest struct {
	PlayerID       int64           `json:"player_id" validate:"required"`
	EventType      string          `json:"event_type" validate:"required"`
	Points         float64         `json:"points" validate:"required"`
	RoundNumber    *int            `json:"round_number"`
	Timestamp      time.Time       `json:"timestamp" validate:"required"`
	Description    *string         `json:"description"`
	AdditionalData json.RawMessage `json:"additional_data"`
}

type BulkEventsRequest struct {
	Events                   []AddEventRequest `json:"events" validate:"required"`
	AutoCalculateFantasyPoints bool            `json:"auto_calculate_fantasy_points"`
}

type UpdatePlayerStatsRequest struct {
	Kills      int `json:"kills"`
	Deaths     int `json:"deaths"`
	Assists    int `json:"assists"`
	Headshots  int `json:"headshots"`
	Aces       int `json:"aces"`
	Plants     int `json:"plants"`
	Defuses    int `json:"defuses"`
	FirstKills int `json:"first_kills"`
}

type UpdateMatchScoreRequest struct {
	Team1Score     int     `json:"team1_score"`
	Team2Score     int     `json:"team2_score"`
	CurrentRound   int     `json:"current_round"`
	MatchStatus    string  `json:"match_status"`
	WinnerTeamID   *int64  `json:"winner_team_id"`
	FinalScore     string  `json:"final_score"`
	MatchDuration  string  `json:"match_duration"`
}

type RecalculatePointsRequest struct {
	ForceRecalculate       bool `json:"force_recalculate"`
	NotifyUsers           bool `json:"notify_users"`
	RecalculateLeaderboards bool `json:"recalculate_leaderboards"`
}

type CompleteMatchRequest struct {
	FinalResult        FinalResult `json:"final_result"`
	DistributePrizes   bool        `json:"distribute_prizes"`
	SendNotifications  bool        `json:"send_notifications"`
}

type FinalResult struct {
	WinnerTeamID   int64   `json:"winner_team_id"`
	FinalScore     string  `json:"final_score"`
	MVPPlayerID    int64   `json:"mvp_player_id"`
	MatchDuration  int     `json:"match_duration"` // in seconds
}

type LiveScoringDashboard struct {
	MatchInfo      Match              `json:"match_info"`
	TeamStats      map[string]TeamStats `json:"team_stats"`
	PlayerStats    []PlayerPerformance `json:"player_stats"`
	RecentEvents   []MatchEvent       `json:"recent_events"`
	FantasyImpact  FantasyImpact      `json:"fantasy_impact"`
}

type TeamStats struct {
	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`
}

type PlayerPerformance struct {
	PlayerID    int64            `json:"player_id"`
	Name        string           `json:"name"`
	TeamName    string           `json:"team_name"`
	Stats       PlayerGameStats  `json:"stats"`
	FantasyPoints float64        `json:"fantasy_points"`
	Events      []MatchEvent     `json:"events"`
}

type PlayerGameStats struct {
	Kills     int `json:"kills"`
	Deaths    int `json:"deaths"`
	Assists   int `json:"assists"`
	Headshots int `json:"headshots"`
	Aces      int `json:"aces"`
}

type FantasyImpact struct {
	AffectedTeams      int `json:"affected_teams"`
	LeaderboardChanges int `json:"leaderboard_changes"`
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type WebSocketEventAdded struct {
	EventID              int64  `json:"event_id"`
	PlayerName           string `json:"player_name"`
	EventType            string `json:"event_type"`
	Points               float64 `json:"points"`
	FantasyTeamsAffected int    `json:"fantasy_teams_affected"`
}

// KYC Admin Management Models
type KYCDocumentWithUser struct {
	KYCDocument
	UserMobile    string  `json:"user_mobile"`
	UserName      string  `json:"user_name"`
	UserEmail     *string `json:"user_email"`
}

type KYCApprovalRequest struct {
	Status          string  `json:"status" validate:"required,oneof=verified rejected"`
	RejectionReason *string `json:"rejection_reason"`
	Notes           *string `json:"notes"`
}

type KYCListResponse struct {
	Documents []KYCDocumentWithUser `json:"documents"`
	Total     int                   `json:"total"`
	Page      int                   `json:"page"`
	Pages     int                   `json:"pages"`
	Filters   KYCFilters            `json:"filters"`
}

type KYCFilters struct {
	Status       string `json:"status"`
	DocumentType string `json:"document_type"`
	DateFrom     string `json:"date_from"`
	DateTo       string `json:"date_to"`
}