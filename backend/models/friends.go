package models

import (
	"time"
	"encoding/json"
)

type Friend struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	FriendID    int64     `json:"friend_id" db:"friend_id"`
	Status      string    `json:"status" db:"status"`
	RequestedBy int64     `json:"requested_by" db:"requested_by"`
	RequestedAt time.Time `json:"requested_at" db:"requested_at"`
	AcceptedAt  *time.Time `json:"accepted_at" db:"accepted_at"`
	
	// Joined fields
	FriendName   *string `json:"friend_name,omitempty"`
	FriendAvatar *string `json:"friend_avatar,omitempty"`
	FriendEmail  *string `json:"friend_email,omitempty"`
}

type FriendChallenge struct {
	ID               int64     `json:"id" db:"id"`
	ChallengerID     int64     `json:"challenger_id" db:"challenger_id"`
	ChallengedID     int64     `json:"challenged_id" db:"challenged_id"`
	MatchID          int64     `json:"match_id" db:"match_id"`
	ChallengeType    string    `json:"challenge_type" db:"challenge_type"`
	EntryFee         float64   `json:"entry_fee" db:"entry_fee"`
	PrizeAmount      *float64  `json:"prize_amount" db:"prize_amount"`
	Status           string    `json:"status" db:"status"`
	WinnerID         *int64    `json:"winner_id" db:"winner_id"`
	ChallengerTeamID *int64    `json:"challenger_team_id" db:"challenger_team_id"`
	ChallengedTeamID *int64    `json:"challenged_team_id" db:"challenged_team_id"`
	Message          *string   `json:"message" db:"message"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	AcceptedAt       *time.Time `json:"accepted_at" db:"accepted_at"`
	CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
	
	// Joined fields
	ChallengerName   *string `json:"challenger_name,omitempty"`
	ChallengedName   *string `json:"challenged_name,omitempty"`
	MatchName        *string `json:"match_name,omitempty"`
	WinnerName       *string `json:"winner_name,omitempty"`
}

type FriendActivity struct {
	ID           int64           `json:"id" db:"id"`
	UserID       int64           `json:"user_id" db:"user_id"`
	ActivityType string          `json:"activity_type" db:"activity_type"`
	ActivityData json.RawMessage `json:"activity_data" db:"activity_data"`
	IsPublic     bool            `json:"is_public" db:"is_public"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	
	// Joined fields
	UserName   *string `json:"user_name,omitempty"`
	UserAvatar *string `json:"user_avatar,omitempty"`
}

type AddFriendRequest struct {
	FriendID int64   `json:"friend_id" validate:"required"`
	Message  *string `json:"message"`
}

type CreateChallengeRequest struct {
	ChallengedID  int64   `json:"challenged_id" validate:"required"`
	MatchID       int64   `json:"match_id" validate:"required"`
	ChallengeType string  `json:"challenge_type" validate:"required"`
	EntryFee      float64 `json:"entry_fee"`
	Message       *string `json:"message"`
}

type AcceptChallengeRequest struct {
	TeamID int64 `json:"team_id" validate:"required"`
}