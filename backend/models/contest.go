package models

import (
	"time"
	"encoding/json"
)

type Contest struct {
	ID                  int64           `json:"id" db:"id"`
	MatchID             int64           `json:"match_id" db:"match_id"`
	Name                string          `json:"name" db:"name"`
	ContestType         string          `json:"contest_type" db:"contest_type"`
	EntryFee            float64         `json:"entry_fee" db:"entry_fee"`
	MaxParticipants     int             `json:"max_participants" db:"max_participants"`
	CurrentParticipants int             `json:"current_participants" db:"current_participants"`
	TotalPrizePool      float64         `json:"total_prize_pool" db:"total_prize_pool"`
	IsGuaranteed        bool            `json:"is_guaranteed" db:"is_guaranteed"`
	PrizeDistribution   json.RawMessage `json:"prize_distribution" db:"prize_distribution"`
	ContestRules        json.RawMessage `json:"contest_rules" db:"contest_rules"`
	Status              string          `json:"status" db:"status"`
	InviteCode          *string         `json:"invite_code" db:"invite_code"`
	IsMultiEntry        bool            `json:"is_multi_entry" db:"is_multi_entry"`
	MaxEntriesPerUser   int             `json:"max_entries_per_user" db:"max_entries_per_user"`
	CreatedBy           int64           `json:"created_by" db:"created_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	
	// Joined fields
	MatchName           *string         `json:"match_name,omitempty"`
	TournamentName      *string         `json:"tournament_name,omitempty"`
	ScheduledAt         *time.Time      `json:"scheduled_at,omitempty"`
	LockTime            *time.Time      `json:"lock_time,omitempty"`
}

type UserTeam struct {
	ID                   int64     `json:"id" db:"id"`
	UserID               int64     `json:"user_id" db:"user_id"`
	MatchID              int64     `json:"match_id" db:"match_id"`
	TeamName             string    `json:"team_name" db:"team_name"`
	CaptainPlayerID      int64     `json:"captain_player_id" db:"captain_player_id"`
	ViceCaptainPlayerID  int64     `json:"vice_captain_player_id" db:"vice_captain_player_id"`
	TotalCreditsUsed     float64   `json:"total_credits_used" db:"total_credits_used"`
	TotalPoints          float64   `json:"total_points" db:"total_points"`
	FinalRank            int       `json:"final_rank" db:"final_rank"`
	IsLocked             bool      `json:"is_locked" db:"is_locked"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
	
	// Joined fields
	Players              []TeamPlayer `json:"players,omitempty"`
	CaptainName          *string      `json:"captain_name,omitempty"`
	ViceCaptainName      *string      `json:"vice_captain_name,omitempty"`
}

type TeamPlayer struct {
	ID             int64   `json:"id" db:"id"`
	TeamID         int64   `json:"team_id" db:"team_id"`
	PlayerID       int64   `json:"player_id" db:"player_id"`
	RealTeamID     int64   `json:"real_team_id" db:"real_team_id"`
	IsCaptain      bool    `json:"is_captain" db:"is_captain"`
	IsViceCaptain  bool    `json:"is_vice_captain" db:"is_vice_captain"`
	PointsEarned   float64 `json:"points_earned" db:"points_earned"`
	
	// Joined fields
	PlayerName     *string `json:"player_name,omitempty"`
	RealTeamName   *string `json:"real_team_name,omitempty"`
	Role           *string `json:"role,omitempty"`
	CreditValue    *float64 `json:"credit_value,omitempty"`
}

type ContestParticipant struct {
	ID           int64     `json:"id" db:"id"`
	ContestID    int64     `json:"contest_id" db:"contest_id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	TeamID       int64     `json:"team_id" db:"team_id"`
	EntryFeePaid float64   `json:"entry_fee_paid" db:"entry_fee_paid"`
	Rank         int       `json:"rank" db:"rank"`
	PrizeWon     float64   `json:"prize_won" db:"prize_won"`
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	
	// Joined fields
	UserName     *string   `json:"user_name,omitempty"`
	TeamName     *string   `json:"team_name,omitempty"`
	TotalPoints  *float64  `json:"total_points,omitempty"`
}

type CreateTeamRequest struct {
	MatchID   int64                `json:"match_id" validate:"required"`
	TeamName  string               `json:"team_name" validate:"required"`
	Players   []PlayerSelection    `json:"players" validate:"required"`
}

type PlayerSelection struct {
	PlayerID        int64 `json:"player_id" validate:"required"`
	IsCaptain       bool  `json:"is_captain"`
	IsViceCaptain   bool  `json:"is_vice_captain"`
}

type CreateContestRequest struct {
	MatchID                int64           `json:"match_id" validate:"required"`
	ContestName            string          `json:"contest_name" validate:"required"`
	EntryFee               float64         `json:"entry_fee" validate:"required"`
	MaxParticipants        int             `json:"max_participants" validate:"required"`
	PrizeDistributionType  string          `json:"prize_distribution_type" validate:"required"`
	InviteCode             *string         `json:"invite_code"`
	IsMultiEntry           bool            `json:"is_multi_entry"`
}

type JoinContestRequest struct {
	UserTeamID    int64   `json:"user_team_id" validate:"required"`
	JoinMultiple  bool    `json:"join_multiple"`
	Teams         []int64 `json:"teams"`
}

type Leaderboard struct {
	ContestID          int64              `json:"contest_id"`
	TotalParticipants  int                `json:"total_participants"`
	MyRank             int                `json:"my_rank"`
	MyPoints           float64            `json:"my_points"`
	MyTeamID           int64              `json:"my_team_id"`
	TopPerformers      []LeaderboardEntry `json:"top_performers"`
	AroundMe           []LeaderboardEntry `json:"around_me"`
	LastUpdated        time.Time          `json:"last_updated"`
}

type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	UserID       int64   `json:"user_id"`
	Username     string  `json:"username"`
	TeamName     string  `json:"team_name"`
	Points       float64 `json:"points"`
	AvatarURL    *string `json:"avatar_url"`
	PrizeWon     float64 `json:"prize_won,omitempty"`
}

type PlayerPerformance struct {
	PlayerID     int64           `json:"player_id"`
	Name         string          `json:"name"`
	TeamName     string          `json:"team_name"`
	Stats        json.RawMessage `json:"stats"`
	FantasyPoints float64        `json:"fantasy_points"`
	Events       []MatchEvent    `json:"events"`
}