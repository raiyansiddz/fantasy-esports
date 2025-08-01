package models

import (
	"time"
	"encoding/json"
)

type Game struct {
	ID              int             `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Code            string          `json:"code" db:"code"`
	Category        *string         `json:"category" db:"category"`
	Description     *string         `json:"description" db:"description"`
	LogoURL         *string         `json:"logo_url" db:"logo_url"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	ScoringRules    json.RawMessage `json:"scoring_rules" db:"scoring_rules"`
	PlayerRoles     json.RawMessage `json:"player_roles" db:"player_roles"`
	TeamComposition json.RawMessage `json:"team_composition" db:"team_composition"`
	MinPlayersPerTeam int           `json:"min_players_per_team" db:"min_players_per_team"`
	MaxPlayersPerTeam int           `json:"max_players_per_team" db:"max_players_per_team"`
	TotalTeamSize     int           `json:"total_team_size" db:"total_team_size"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type Team struct {
	ID          int64           `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	ShortName   *string         `json:"short_name" db:"short_name"`
	LogoURL     *string         `json:"logo_url" db:"logo_url"`
	Region      *string         `json:"region" db:"region"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	SocialLinks json.RawMessage `json:"social_links" db:"social_links"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type Player struct {
	ID          int64           `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	TeamID      int64           `json:"team_id" db:"team_id"`
	GameID      int             `json:"game_id" db:"game_id"`
	Role        *string         `json:"role" db:"role"`
	CreditValue float64         `json:"credit_value" db:"credit_value"`
	IsPlaying   bool            `json:"is_playing" db:"is_playing"`
	AvatarURL   *string         `json:"avatar_url" db:"avatar_url"`
	Country     *string         `json:"country" db:"country"`
	Stats       json.RawMessage `json:"stats" db:"stats"`
	FormScore   float64         `json:"form_score" db:"form_score"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	TeamName    *string `json:"team_name,omitempty"`
	GameName    *string `json:"game_name,omitempty"`
}

type Tournament struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	GameID      int       `json:"game_id" db:"game_id"`
	Description *string   `json:"description" db:"description"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	PrizePool   *float64  `json:"prize_pool" db:"prize_pool"`
	TotalTeams  *int      `json:"total_teams" db:"total_teams"`
	Status      string    `json:"status" db:"status"`
	IsFeatured  bool      `json:"is_featured" db:"is_featured"`
	LogoURL     *string   `json:"logo_url" db:"logo_url"`
	BannerURL   *string   `json:"banner_url" db:"banner_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TournamentStage struct {
	ID           int64           `json:"id" db:"id"`
	TournamentID int64           `json:"tournament_id" db:"tournament_id"`
	Name         string          `json:"name" db:"name"`
	StageOrder   int             `json:"stage_order" db:"stage_order"`
	StageType    string          `json:"stage_type" db:"stage_type"`
	StartDate    *time.Time      `json:"start_date" db:"start_date"`
	EndDate      *time.Time      `json:"end_date" db:"end_date"`
	MaxTeams     *int            `json:"max_teams" db:"max_teams"`
	Rules        json.RawMessage `json:"rules" db:"rules"`
}

type Match struct {
	ID            int64           `json:"id" db:"id"`
	TournamentID  *int64          `json:"tournament_id" db:"tournament_id"`
	StageID       *int64          `json:"stage_id" db:"stage_id"`
	GameID        int             `json:"game_id" db:"game_id"`
	Name          *string         `json:"name" db:"name"`
	ScheduledAt   time.Time       `json:"scheduled_at" db:"scheduled_at"`
	LockTime      time.Time       `json:"lock_time" db:"lock_time"`
	Status        string          `json:"status" db:"status"`
	MatchType     string          `json:"match_type" db:"match_type"`
	Map           *string         `json:"map" db:"map"`
	BestOf        int             `json:"best_of" db:"best_of"`
	Result        json.RawMessage `json:"result" db:"result"`
	WinnerTeamID  *int64          `json:"winner_team_id" db:"winner_team_id"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	Teams         []Team          `json:"teams,omitempty"`
	TournamentName *string        `json:"tournament_name,omitempty"`
	GameName      *string         `json:"game_name,omitempty"`
}

type MatchParticipant struct {
	ID            int64     `json:"id" db:"id"`
	MatchID       int64     `json:"match_id" db:"match_id"`
	TeamID        int64     `json:"team_id" db:"team_id"`
	Seed          *int      `json:"seed" db:"seed"`
	FinalPosition *int      `json:"final_position" db:"final_position"`
	TeamScore     int       `json:"team_score" db:"team_score"`
	PointsEarned  float64   `json:"points_earned" db:"points_earned"`
	EliminatedAt  *time.Time `json:"eliminated_at" db:"eliminated_at"`
	JoinedAt      time.Time `json:"joined_at" db:"joined_at"`
}

type MatchEvent struct {
	ID             int64           `json:"id" db:"id"`
	MatchID        int64           `json:"match_id" db:"match_id"`
	PlayerID       int64           `json:"player_id" db:"player_id"`
	EventType      string          `json:"event_type" db:"event_type"`
	Points         float64         `json:"points" db:"points"`
	RoundNumber    *int            `json:"round_number" db:"round_number"`
	GameTime       *string         `json:"game_time" db:"game_time"`
	Description    *string         `json:"description" db:"description"`
	AdditionalData json.RawMessage `json:"additional_data" db:"additional_data"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	CreatedBy      int64           `json:"created_by" db:"created_by"`
	
	// Joined fields
	PlayerName     *string         `json:"player_name,omitempty"`
	TeamName       *string         `json:"team_name,omitempty"`
}