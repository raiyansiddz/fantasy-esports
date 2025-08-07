package models

import (
	"time"
	"encoding/json"
)

type TournamentBracket struct {
	ID           int64           `json:"id" db:"id"`
	TournamentID int64           `json:"tournament_id" db:"tournament_id"`
	StageID      int64           `json:"stage_id" db:"stage_id"`
	BracketType  string          `json:"bracket_type" db:"bracket_type"`
	BracketData  json.RawMessage `json:"bracket_data" db:"bracket_data"`
	CurrentRound int             `json:"current_round" db:"current_round"`
	TotalRounds  int             `json:"total_rounds" db:"total_rounds"`
	Status       string          `json:"status" db:"status"`
	AutoAdvance  bool            `json:"auto_advance" db:"auto_advance"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

type PlayerPrediction struct {
	ID              int64           `json:"id" db:"id"`
	PlayerID        int64           `json:"player_id" db:"player_id"`
	MatchID         int64           `json:"match_id" db:"match_id"`
	PredictionDate  time.Time       `json:"prediction_date" db:"prediction_date"`
	PredictedPoints float64         `json:"predicted_points" db:"predicted_points"`
	ConfidenceScore float64         `json:"confidence_score" db:"confidence_score"`
	Factors         json.RawMessage `json:"factors" db:"factors"`
	ActualPoints    *float64        `json:"actual_points" db:"actual_points"`
	AccuracyScore   *float64        `json:"accuracy_score" db:"accuracy_score"`
	ModelVersion    string          `json:"model_version" db:"model_version"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	PlayerName *string `json:"player_name,omitempty"`
	TeamName   *string `json:"team_name,omitempty"`
	MatchName  *string `json:"match_name,omitempty"`
}

type GameAnalyticsAdvanced struct {
	ID          int64           `json:"id" db:"id"`
	GameID      int             `json:"game_id" db:"game_id"`
	Date        time.Time       `json:"date" db:"date"`
	MetricType  string          `json:"metric_type" db:"metric_type"`
	MetricValue float64         `json:"metric_value" db:"metric_value"`
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type FraudAlert struct {
	ID              int64      `json:"id" db:"id"`
	UserID          *int64     `json:"user_id" db:"user_id"`
	AlertType       string     `json:"alert_type" db:"alert_type"`
	Severity        string     `json:"severity" db:"severity"`
	Description     string     `json:"description" db:"description"`
	DetectionData   json.RawMessage `json:"detection_data" db:"detection_data"`
	Status          string     `json:"status" db:"status"`
	AssignedTo      *int64     `json:"assigned_to" db:"assigned_to"`
	ResolvedAt      *time.Time `json:"resolved_at" db:"resolved_at"`
	ResolutionNotes *string    `json:"resolution_notes" db:"resolution_notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	
	// Joined fields
	UserName     *string `json:"user_name,omitempty"`
	AssignedName *string `json:"assigned_name,omitempty"`
}

type UserBehaviorLog struct {
	ID          int64           `json:"id" db:"id"`
	UserID      *int64          `json:"user_id" db:"user_id"`
	SessionID   *string         `json:"session_id" db:"session_id"`
	Action      string          `json:"action" db:"action"`
	ContextData json.RawMessage `json:"context_data" db:"context_data"`
	IPAddress   *string         `json:"ip_address" db:"ip_address"`
	UserAgent   *string         `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type PredictionFactors struct {
	RecentForm        float64                `json:"recent_form"`
	HeadToHeadRecord  float64                `json:"head_to_head_record"`
	TeamStrength      float64                `json:"team_strength"`
	MapPerformance    float64                `json:"map_performance"`
	WeatherConditions *string                `json:"weather_conditions"`
	PlayerHealth      *string                `json:"player_health"`
	TeamMorale        float64                `json:"team_morale"`
	ExtraFactors      map[string]interface{} `json:"extra_factors"`
}

type AdvancedGameMetrics struct {
	PlayerEfficiency     float64 `json:"player_efficiency"`
	TeamSynergy         float64 `json:"team_synergy"`
	StrategicDiversity  float64 `json:"strategic_diversity"`
	ComebackPotential   float64 `json:"comeback_potential"`
	ClutchPerformance   float64 `json:"clutch_performance"`
	ConsistencyIndex    float64 `json:"consistency_index"`
	AdaptabilityScore   float64 `json:"adaptability_score"`
}

type CreateBracketRequest struct {
	TournamentID int64  `json:"tournament_id" validate:"required"`
	StageID      int64  `json:"stage_id" validate:"required"`
	BracketType  string `json:"bracket_type" validate:"required"`
	AutoAdvance  bool   `json:"auto_advance"`
}