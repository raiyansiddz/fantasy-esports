package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
)

type FriendService struct {
	db *sql.DB
}

func NewFriendService(db *sql.DB) *FriendService {
	return &FriendService{db: db}
}

// Friend management
func (s *FriendService) AddFriend(userID int64, req models.AddFriendRequest) error {
	// Resolve friend ID from the request
	var friendID int64
	var err error
	
	if req.FriendID != nil {
		friendID = *req.FriendID
	} else if req.Username != nil {
		friendID, err = s.getUserByUsername(*req.Username)
		if err != nil {
			return fmt.Errorf("user not found with username: %s", *req.Username)
		}
	} else if req.Mobile != nil {
		friendID, err = s.getUserByMobile(*req.Mobile)
		if err != nil {
			return fmt.Errorf("user not found with mobile: %s", *req.Mobile)
		}
	} else {
		return fmt.Errorf("must provide friend_id, username, or mobile")
	}

	if userID == friendID {
		return fmt.Errorf("cannot add yourself as a friend")
	}

	// Check if friendship already exists
	var count int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM user_friends 
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`, userID, friendID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("friendship already exists")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert friend request
	_, err = tx.Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, requested_by)
		VALUES ($1, $2, 'pending', $1)
	`, userID, friendID)
	if err != nil {
		return err
	}

	// Insert reverse relationship for easier querying
	_, err = tx.Exec(`
		INSERT INTO user_friends (user_id, friend_id, status, requested_by)
		VALUES ($1, $2, 'pending', $3)
	`, friendID, userID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FriendService) AcceptFriend(userID, friendID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update both relationships to accepted
	_, err = tx.Exec(`
		UPDATE user_friends SET status = 'accepted', accepted_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND friend_id = $2
	`, userID, friendID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE user_friends SET status = 'accepted', accepted_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND friend_id = $2
	`, friendID, userID)
	if err != nil {
		return err
	}

	// Create activity for both users
	activityData := map[string]interface{}{
		"friend_id":   friendID,
		"friend_name": "",
	}
	activityJSON, _ := json.Marshal(activityData)

	_, err = tx.Exec(`
		INSERT INTO friend_activities (user_id, activity_type, activity_data)
		VALUES ($1, 'friend_added', $2)
	`, userID, activityJSON)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FriendService) DeclineFriend(userID, friendID int64) error {
	// Delete both relationships
	_, err := s.db.Exec(`
		DELETE FROM user_friends 
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`, userID, friendID)
	return err
}

func (s *FriendService) RemoveFriend(userID, friendID int64) error {
	// Delete both relationships
	_, err := s.db.Exec(`
		DELETE FROM user_friends 
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`, userID, friendID)
	return err
}

func (s *FriendService) GetFriends(userID int64, status string) ([]models.Friend, error) {
	query := `
		SELECT uf.id, uf.user_id, uf.friend_id, uf.status, uf.requested_by, uf.requested_at, uf.accepted_at,
			u.first_name, u.last_name, u.avatar_url, u.email
		FROM user_friends uf
		JOIN users u ON uf.friend_id = u.id
		WHERE uf.user_id = $1`
	
	args := []interface{}{userID}
	if status != "" {
		query += " AND uf.status = $2"
		args = append(args, status)
	}
	
	query += " ORDER BY uf.requested_at DESC"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []models.Friend
	for rows.Next() {
		var f models.Friend
		var firstName, lastName *string
		
		err := rows.Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.RequestedBy, 
			&f.RequestedAt, &f.AcceptedAt, &firstName, &lastName, &f.FriendAvatar, &f.FriendEmail)
		if err != nil {
			return nil, err
		}

		// Combine first and last name
		if firstName != nil && lastName != nil {
			fullName := *firstName + " " + *lastName
			f.FriendName = &fullName
		}
		
		friends = append(friends, f)
	}

	return friends, nil
}

// Friend challenges
func (s *FriendService) CreateChallenge(challengerID int64, req models.CreateChallengeRequest) (*models.FriendChallenge, error) {
	// Verify they are friends
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM user_friends 
		WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted'
	`, challengerID, req.ChallengedID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("can only challenge friends")
	}

	// Calculate prize amount
	prizeAmount := req.EntryFee * 2 * 0.9 // 90% of total entry fees

	query := `
		INSERT INTO friend_challenges (challenger_id, challenged_id, match_id, challenge_type, 
			entry_fee, prize_amount, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	
	challenge := &models.FriendChallenge{
		ChallengerID:  challengerID,
		ChallengedID:  req.ChallengedID,
		MatchID:       req.MatchID,
		ChallengeType: req.ChallengeType,
		EntryFee:      req.EntryFee,
		PrizeAmount:   &prizeAmount,
		Message:       req.Message,
		Status:        "pending",
	}

	err = s.db.QueryRow(query, challenge.ChallengerID, challenge.ChallengedID, 
		challenge.MatchID, challenge.ChallengeType, challenge.EntryFee, 
		challenge.PrizeAmount, challenge.Message).Scan(&challenge.ID, &challenge.CreatedAt)
	
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

func (s *FriendService) AcceptChallenge(challengeID, challengedID int64, req models.AcceptChallengeRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update challenge status and set team
	_, err = tx.Exec(`
		UPDATE friend_challenges 
		SET status = 'accepted', challenged_team_id = $1, accepted_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND challenged_id = $3 AND status = 'pending'
	`, req.TeamID, challengeID, challengedID)
	if err != nil {
		return err
	}

	// Deduct entry fee from both users' wallets (if applicable)
	var entryFee float64
	var challengerID int64
	err = tx.QueryRow("SELECT entry_fee, challenger_id FROM friend_challenges WHERE id = $1", 
		challengeID).Scan(&entryFee, &challengerID)
	if err != nil {
		return err
	}

	if entryFee > 0 {
		// Deduct from challenged user
		_, err = tx.Exec(`
			UPDATE user_wallets 
			SET deposit_balance = CASE 
				WHEN deposit_balance >= $1 THEN deposit_balance - $1
				ELSE GREATEST(0, deposit_balance + winning_balance - $1)
			END,
			winning_balance = CASE 
				WHEN deposit_balance >= $1 THEN winning_balance
				ELSE GREATEST(0, winning_balance - ($1 - deposit_balance))
			END
			WHERE user_id = $2
		`, entryFee, challengedID)
		if err != nil {
			return err
		}

		// Deduct from challenger user
		_, err = tx.Exec(`
			UPDATE user_wallets 
			SET deposit_balance = CASE 
				WHEN deposit_balance >= $1 THEN deposit_balance - $1
				ELSE GREATEST(0, deposit_balance + winning_balance - $1)
			END,
			winning_balance = CASE 
				WHEN deposit_balance >= $1 THEN winning_balance
				ELSE GREATEST(0, winning_balance - ($1 - deposit_balance))
			END
			WHERE user_id = $2
		`, entryFee, challengerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *FriendService) DeclineChallenge(challengeID, challengedID int64) error {
	_, err := s.db.Exec(`
		UPDATE friend_challenges 
		SET status = 'declined'
		WHERE id = $1 AND challenged_id = $2 AND status = 'pending'
	`, challengeID, challengedID)
	return err
}

func (s *FriendService) GetChallenges(userID int64, challengeType string) ([]models.FriendChallenge, error) {
	query := `
		SELECT fc.id, fc.challenger_id, fc.challenged_id, fc.match_id, fc.challenge_type,
			fc.entry_fee, fc.prize_amount, fc.status, fc.winner_id, fc.challenger_team_id,
			fc.challenged_team_id, fc.message, fc.created_at, fc.accepted_at, fc.completed_at,
			u1.first_name as challenger_first, u1.last_name as challenger_last,
			u2.first_name as challenged_first, u2.last_name as challenged_last,
			m.name as match_name
		FROM friend_challenges fc
		JOIN users u1 ON fc.challenger_id = u1.id
		JOIN users u2 ON fc.challenged_id = u2.id
		LEFT JOIN matches m ON fc.match_id = m.id
		WHERE (fc.challenger_id = $1 OR fc.challenged_id = $1)`
	
	args := []interface{}{userID}
	if challengeType != "" {
		query += " AND fc.status = $2"
		args = append(args, challengeType)
	}
	
	query += " ORDER BY fc.created_at DESC"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []models.FriendChallenge
	for rows.Next() {
		var c models.FriendChallenge
		var challengerFirst, challengerLast, challengedFirst, challengedLast *string
		
		err := rows.Scan(&c.ID, &c.ChallengerID, &c.ChallengedID, &c.MatchID, &c.ChallengeType,
			&c.EntryFee, &c.PrizeAmount, &c.Status, &c.WinnerID, &c.ChallengerTeamID,
			&c.ChallengedTeamID, &c.Message, &c.CreatedAt, &c.AcceptedAt, &c.CompletedAt,
			&challengerFirst, &challengerLast, &challengedFirst, &challengedLast, &c.MatchName)
		if err != nil {
			return nil, err
		}

		// Combine names
		if challengerFirst != nil && challengerLast != nil {
			name := *challengerFirst + " " + *challengerLast
			c.ChallengerName = &name
		}
		if challengedFirst != nil && challengedLast != nil {
			name := *challengedFirst + " " + *challengedLast
			c.ChallengedName = &name
		}
		
		challenges = append(challenges, c)
	}

	return challenges, nil
}

// Activity feed
func (s *FriendService) GetFriendActivities(userID int64, limit int) ([]models.FriendActivity, error) {
	query := `
		SELECT fa.id, fa.user_id, fa.activity_type, fa.activity_data, fa.is_public, fa.created_at,
			u.first_name, u.last_name, u.avatar_url
		FROM friend_activities fa
		JOIN users u ON fa.user_id = u.id
		WHERE fa.user_id IN (
			SELECT friend_id FROM user_friends 
			WHERE user_id = $1 AND status = 'accepted'
		) OR fa.user_id = $1
		ORDER BY fa.created_at DESC
		LIMIT $2
	`
	
	rows, err := s.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.FriendActivity
	for rows.Next() {
		var a models.FriendActivity
		var firstName, lastName *string
		
		err := rows.Scan(&a.ID, &a.UserID, &a.ActivityType, &a.ActivityData, 
			&a.IsPublic, &a.CreatedAt, &firstName, &lastName, &a.UserAvatar)
		if err != nil {
			return nil, err
		}

		// Combine names
		if firstName != nil && lastName != nil {
			name := *firstName + " " + *lastName
			a.UserName = &name
		}
		
		activities = append(activities, a)
	}

	return activities, nil
}

func (s *FriendService) CreateActivity(userID int64, activityType string, activityData map[string]interface{}) error {
	dataJSON, err := json.Marshal(activityData)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO friend_activities (user_id, activity_type, activity_data)
		VALUES ($1, $2, $3)
	`, userID, activityType, dataJSON)
	
	return err
}

// Challenge completion (called after match ends)
func (s *FriendService) CompleteChallenges(matchID int64) error {
	// Get all accepted challenges for this match
	challenges, err := s.getActiveChallenges(matchID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, challenge := range challenges {
		// Determine winner based on team points
		winnerID, err := s.determineWinner(challenge)
		if err != nil {
			continue // Skip this challenge if can't determine winner
		}

		// Update challenge
		_, err = tx.Exec(`
			UPDATE friend_challenges 
			SET status = 'completed', winner_id = $1, completed_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, winnerID, challenge.ID)
		if err != nil {
			continue
		}

		// Award prize
		if challenge.PrizeAmount != nil && *challenge.PrizeAmount > 0 {
			_, err = tx.Exec(`
				UPDATE user_wallets 
				SET winning_balance = winning_balance + $1
				WHERE user_id = $2
			`, *challenge.PrizeAmount, winnerID)
			if err != nil {
				continue
			}

			// Log transaction
			_, err = tx.Exec(`
				INSERT INTO wallet_transactions (user_id, transaction_type, amount, balance_type, 
					description, status)
				VALUES ($1, 'prize_credit', $2, 'winning', 'Challenge win prize', 'completed')
			`, winnerID, *challenge.PrizeAmount)
		}
	}

	return tx.Commit()
}

func (s *FriendService) getActiveChallenges(matchID int64) ([]models.FriendChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, match_id, challenger_team_id, 
			challenged_team_id, prize_amount
		FROM friend_challenges
		WHERE match_id = $1 AND status = 'accepted'
	`
	
	rows, err := s.db.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []models.FriendChallenge
	for rows.Next() {
		var c models.FriendChallenge
		err := rows.Scan(&c.ID, &c.ChallengerID, &c.ChallengedID, &c.MatchID, 
			&c.ChallengerTeamID, &c.ChallengedTeamID, &c.PrizeAmount)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, c)
	}

	return challenges, nil
}

func (s *FriendService) determineWinner(challenge models.FriendChallenge) (int64, error) {
	if challenge.ChallengerTeamID == nil || challenge.ChallengedTeamID == nil {
		return 0, fmt.Errorf("teams not set")
	}

	// Get points for both teams
	var challengerPoints, challengedPoints float64
	
	err := s.db.QueryRow("SELECT total_points FROM user_teams WHERE id = $1", 
		*challenge.ChallengerTeamID).Scan(&challengerPoints)
	if err != nil {
		return 0, err
	}

	err = s.db.QueryRow("SELECT total_points FROM user_teams WHERE id = $1", 
		*challenge.ChallengedTeamID).Scan(&challengedPoints)
	if err != nil {
		return 0, err
	}

	if challengerPoints > challengedPoints {
		return challenge.ChallengerID, nil
	} else if challengedPoints > challengerPoints {
		return challenge.ChallengedID, nil
	}

	// Tie - return challenger as winner by default
	return challenge.ChallengerID, nil
}

// Helper methods for user lookup
func (s *FriendService) getUserByUsername(username string) (int64, error) {
	var userID int64
	err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return userID, nil
}

func (s *FriendService) getUserByMobile(mobile string) (int64, error) {
	var userID int64
	err := s.db.QueryRow("SELECT id FROM users WHERE mobile = $1", mobile).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return userID, nil
}