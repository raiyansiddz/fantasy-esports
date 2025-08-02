package models

import (
	"time"
)

// Real-time leaderboard specific models

type RealTimeLeaderboardUpdate struct {
	ContestID         int64                    `json:"contest_id"`
	UpdateID          string                   `json:"update_id"`
	UpdateType        string                   `json:"update_type"` // "rank_change", "points_update", "new_entry", "full_refresh"
	UpdateTimestamp   time.Time                `json:"update_timestamp"`
	AffectedUserIDs   []int64                  `json:"affected_user_ids"`
	RankChanges       []LeaderboardRankChange  `json:"rank_changes"`
	TopPerformers     []LeaderboardEntry       `json:"top_performers"`
	TotalParticipants int                      `json:"total_participants"`
	MatchEventID      *int64                   `json:"match_event_id,omitempty"`
	TriggerSource     string                   `json:"trigger_source"` // "match_event", "score_update", "manual_recalc"
}

type LeaderboardRankChange struct {
	UserID           int64   `json:"user_id"`
	TeamID           int64   `json:"team_id"`
	Username         string  `json:"username"`
	TeamName         string  `json:"team_name"`
	PreviousRank     int     `json:"previous_rank"`
	NewRank          int     `json:"new_rank"`
	RankChange       int     `json:"rank_change"` // positive = moved up, negative = moved down
	PreviousPoints   float64 `json:"previous_points"`
	NewPoints        float64 `json:"new_points"`
	PointsChange     float64 `json:"points_change"`
	AvatarURL        *string `json:"avatar_url"`
}

type LeaderboardSubscription struct {
	UserID         int64     `json:"user_id"`
	ContestID      int64     `json:"contest_id"`
	SubscribedAt   time.Time `json:"subscribed_at"`
	LastUpdateID   string    `json:"last_update_id"`
	ConnectionID   string    `json:"connection_id"`
	IsActive       bool      `json:"is_active"`
}

type RealTimeWebSocketMessage struct {
	Type      string      `json:"type"`      // "leaderboard_update", "rank_change", "connection_status", "error"
	ContestID int64       `json:"contest_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	MessageID string      `json:"message_id"`
}

type LeaderboardConnectionStatus struct {
	Connected        bool      `json:"connected"`
	ContestID        int64     `json:"contest_id"`
	UserID           int64     `json:"user_id"`
	MyCurrentRank    int       `json:"my_current_rank"`
	MyCurrentPoints  float64   `json:"my_current_points"`
	TotalParticipants int      `json:"total_participants"`
	LastUpdated      time.Time `json:"last_updated"`
}

type RankingSnapshot struct {
	ContestID     int64                  `json:"contest_id"`
	SnapshotID    string                 `json:"snapshot_id"`
	CreatedAt     time.Time              `json:"created_at"`
	Rankings      map[int64]RankPosition `json:"rankings"` // userID -> position
	TotalPoints   map[int64]float64      `json:"total_points"` // userID -> points
}

type RankPosition struct {
	Rank     int     `json:"rank"`
	Points   float64 `json:"points"`
	TeamID   int64   `json:"team_id"`
	Username string  `json:"username"`
	TeamName string  `json:"team_name"`
}

type LiveLeaderboardRequest struct {
	ContestID        int64 `json:"contest_id"`
	IncludeAroundMe  bool  `json:"include_around_me"`
	TopCount         int   `json:"top_count"`         // Default 50
	AroundMeRadius   int   `json:"around_me_radius"`  // Default 5
}

type LiveLeaderboardResponse struct {
	Success           bool                         `json:"success"`
	ContestID         int64                        `json:"contest_id"`
	Leaderboard       *Leaderboard                 `json:"leaderboard"`
	RealTimeEnabled   bool                         `json:"real_time_enabled"`
	UpdateFrequency   int                          `json:"update_frequency"` // seconds
	LastUpdateID      string                       `json:"last_update_id"`
	WebSocketEndpoint string                       `json:"websocket_endpoint"`
}

// Enhanced contest leaderboard with caching metadata
type CachedLeaderboard struct {
	*Leaderboard
	CacheKey      string    `json:"cache_key"`
	CachedAt      time.Time `json:"cached_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsDirty       bool      `json:"is_dirty"`
	LastEventID   int64     `json:"last_event_id"`
}

// WebSocket connection info for leaderboard subscriptions
type LeaderboardConnection struct {
	UserID       int64     `json:"user_id"`
	ContestID    int64     `json:"contest_id"`
	ConnectionID string    `json:"connection_id"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastPing     time.Time `json:"last_ping"`
	IsActive     bool      `json:"is_active"`
}