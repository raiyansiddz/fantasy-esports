package services

import (
	"database/sql"
	"fmt"
	"time"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
)

type ReferralService struct {
	db *sql.DB
}

// ReferralTier represents a tier in the referral system
type ReferralTier struct {
	Name              string  `json:"name"`
	MinReferrals      int     `json:"min_referrals"`
	RewardPerReferral float64 `json:"reward_per_referral"`
	BonusReward       float64 `json:"bonus_reward"`
}

// ReferralConfig contains all referral system configuration
type ReferralConfig struct {
	BaseReward            float64        `json:"base_reward"`
	CompletionCriteria    string         `json:"completion_criteria"`
	RewardExpiryDays      int            `json:"reward_expiry_days"`
	MaxRewardPerUser      float64        `json:"max_reward_per_user"`
	Tiers                 []ReferralTier `json:"tiers"`
}

// NewReferralService creates a new referral service instance
func NewReferralService(db *sql.DB) *ReferralService {
	return &ReferralService{
		db: db,
	}
}

// GetReferralConfig returns the current referral system configuration
func (s *ReferralService) GetReferralConfig() ReferralConfig {
	return ReferralConfig{
		BaseReward:         50.0, // Base reward for each successful referral
		CompletionCriteria: "first_deposit",
		RewardExpiryDays:   30,
		MaxRewardPerUser:   5000.0,
		Tiers: []ReferralTier{
			{Name: "bronze", MinReferrals: 0, RewardPerReferral: 50.0, BonusReward: 0.0},
			{Name: "silver", MinReferrals: 10, RewardPerReferral: 75.0, BonusReward: 200.0},
			{Name: "gold", MinReferrals: 25, RewardPerReferral: 100.0, BonusReward: 500.0},
			{Name: "platinum", MinReferrals: 50, RewardPerReferral: 150.0, BonusReward: 1000.0},
			{Name: "diamond", MinReferrals: 100, RewardPerReferral: 200.0, BonusReward: 2500.0},
		},
	}
}

// ApplyReferralCode applies a referral code when a user registers
func (s *ReferralService) ApplyReferralCode(referredUserID int64, referralCode string) error {
	logger.Info(fmt.Sprintf("Applying referral code %s for user %d", referralCode, referredUserID))
	
	// Check if referral code exists and get referrer
	var referrerUserID int64
	err := s.db.QueryRow(`
		SELECT id FROM users WHERE referral_code = $1 AND is_active = true`, referralCode).Scan(&referrerUserID)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid referral code")
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	
	// Check if user is trying to refer themselves
	if referrerUserID == referredUserID {
		return fmt.Errorf("cannot refer yourself")
	}
	
	// Check if this user was already referred
	var existingReferral int64
	err = s.db.QueryRow(`
		SELECT id FROM referrals WHERE referred_user_id = $1`, referredUserID).Scan(&existingReferral)
	
	if err == nil {
		return fmt.Errorf("user has already been referred")
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("database error: %w", err)
	}
	
	// Create referral record
	config := s.GetReferralConfig()
	_, err = s.db.Exec(`
		INSERT INTO referrals (referrer_user_id, referred_user_id, referral_code, status, 
							   reward_amount, completion_criteria, created_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, NOW())`,
		referrerUserID, referredUserID, referralCode, config.BaseReward, config.CompletionCriteria)
	
	if err != nil {
		return fmt.Errorf("failed to create referral record: %w", err)
	}
	
	// Update user's referred_by_code
	_, err = s.db.Exec(`
		UPDATE users SET referred_by_code = $1, updated_at = NOW() WHERE id = $2`,
		referralCode, referredUserID)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update user's referred_by_code: %v", err))
	}
	
	logger.Info(fmt.Sprintf("Successfully created referral: user %d referred by %d", referredUserID, referrerUserID))
	return nil
}

// CheckAndCompleteReferral checks if a referral should be completed based on criteria
func (s *ReferralService) CheckAndCompleteReferral(userID int64, action string) error {
	logger.Info(fmt.Sprintf("Checking referral completion for user %d, action: %s", userID, action))
	
	// Get pending referral for this user
	var referralID int64
	var referrerUserID int64
	var rewardAmount float64
	var completionCriteria string
	
	err := s.db.QueryRow(`
		SELECT id, referrer_user_id, reward_amount, completion_criteria
		FROM referrals 
		WHERE referred_user_id = $1 AND status = 'pending'`, userID).Scan(
		&referralID, &referrerUserID, &rewardAmount, &completionCriteria)
	
	if err == sql.ErrNoRows {
		// No pending referral found
		return nil
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	
	// Check if the action matches completion criteria
	shouldComplete := false
	switch completionCriteria {
	case "first_deposit":
		shouldComplete = (action == "deposit")
	case "first_contest":
		shouldComplete = (action == "contest_join")
	case "immediate":
		shouldComplete = true
	}
	
	if !shouldComplete {
		return nil
	}
	
	// Get current tier reward for referrer
	tierReward, err := s.GetUserTierReward(referrerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get tier reward: %v", err))
		tierReward = rewardAmount // Use base reward as fallback
	}
	
	// Start transaction for reward distribution
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Mark referral as completed
	_, err = tx.Exec(`
		UPDATE referrals 
		SET status = 'completed', completed_at = NOW(), reward_amount = $1
		WHERE id = $2`, tierReward, referralID)
	
	if err != nil {
		return fmt.Errorf("failed to update referral status: %w", err)
	}
	
	// Add bonus balance to referrer's wallet
	err = s.addBonusBalance(tx, referrerUserID, tierReward, 
		fmt.Sprintf("Referral reward for user %d", userID))
	if err != nil {
		return fmt.Errorf("failed to add bonus balance: %w", err)
	}
	
	// Check for tier upgrade bonus
	err = s.checkAndApplyTierBonus(tx, referrerUserID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to check tier bonus: %v", err))
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.Info(fmt.Sprintf("Completed referral: user %d earned %.2f from referring user %d", 
		referrerUserID, tierReward, userID))
	
	return nil
}

// GetUserReferralStats returns comprehensive referral statistics for a user
func (s *ReferralService) GetUserReferralStats(userID int64) (*models.ReferralStats, error) {
	// Get user's referral code
	var referralCode string
	err := s.db.QueryRow(`SELECT referral_code FROM users WHERE id = $1`, userID).Scan(&referralCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get referral code: %w", err)
	}
	
	// Get referral counts and earnings
	var totalReferrals int
	var successfulReferrals int
	var totalEarnings float64
	var pendingEarnings float64
	
	// Total referrals
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM referrals WHERE referrer_user_id = $1`, userID).Scan(&totalReferrals)
	if err != nil {
		totalReferrals = 0
	}
	
	// Successful referrals and total earnings
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(reward_amount), 0)
		FROM referrals 
		WHERE referrer_user_id = $1 AND status = 'completed'`, userID).Scan(&successfulReferrals, &totalEarnings)
	if err != nil {
		successfulReferrals = 0
		totalEarnings = 0.0
	}
	
	// Pending earnings
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(reward_amount), 0)
		FROM referrals 
		WHERE referrer_user_id = $1 AND status = 'pending'`, userID).Scan(&pendingEarnings)
	if err != nil {
		pendingEarnings = 0.0
	}
	
	// Calculate lifetime earnings (including tier bonuses from wallet transactions)
	var lifetimeEarnings float64
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM wallet_transactions 
		WHERE user_id = $1 AND transaction_type = 'bonus_credit' 
		AND description LIKE '%referral%'`, userID).Scan(&lifetimeEarnings)
	if err != nil {
		lifetimeEarnings = totalEarnings // Fallback to just completed referrals
	}
	
	// Determine current tier and next tier requirement
	currentTier, nextTierRequirement := s.getUserTier(successfulReferrals)
	
	stats := &models.ReferralStats{
		ReferralCode:        referralCode,
		TotalReferrals:      totalReferrals,
		SuccessfulReferrals: successfulReferrals,
		TotalEarnings:       totalEarnings,
		PendingEarnings:     pendingEarnings,
		LifetimeEarnings:    lifetimeEarnings,
		CurrentTier:         currentTier,
		NextTierRequirement: nextTierRequirement,
	}
	
	return stats, nil
}

// GetUserReferralHistory returns the referral history for a user
func (s *ReferralService) GetUserReferralHistory(userID int64, status string, page, limit int) ([]models.Referral, int, error) {
	offset := (page - 1) * limit
	
	// Build query based on status filter
	var query string
	var countQuery string
	var args []interface{}
	
	if status == "all" || status == "" {
		query = `
			SELECT r.id, r.referrer_user_id, r.referred_user_id, r.referral_code, r.status,
				   r.reward_amount, r.completion_criteria, r.completed_at, r.created_at,
				   u.first_name, u.last_name
			FROM referrals r
			JOIN users u ON r.referred_user_id = u.id
			WHERE r.referrer_user_id = $1
			ORDER BY r.created_at DESC
			LIMIT $2 OFFSET $3`
		countQuery = `SELECT COUNT(*) FROM referrals WHERE referrer_user_id = $1`
		args = []interface{}{userID, limit, offset}
	} else {
		query = `
			SELECT r.id, r.referrer_user_id, r.referred_user_id, r.referral_code, r.status,
				   r.reward_amount, r.completion_criteria, r.completed_at, r.created_at,
				   u.first_name, u.last_name
			FROM referrals r
			JOIN users u ON r.referred_user_id = u.id
			WHERE r.referrer_user_id = $1 AND r.status = $2
			ORDER BY r.created_at DESC
			LIMIT $3 OFFSET $4`
		countQuery = `SELECT COUNT(*) FROM referrals WHERE referrer_user_id = $1 AND status = $2`
		args = []interface{}{userID, status, limit, offset}
	}
	
	// Get referrals
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query referrals: %w", err)
	}
	defer rows.Close()
	
	var referrals []models.Referral
	for rows.Next() {
		var referral models.Referral
		var firstName, lastName sql.NullString
		
		err := rows.Scan(
			&referral.ID, &referral.ReferrerUserID, &referral.ReferredUserID,
			&referral.ReferralCode, &referral.Status, &referral.RewardAmount,
			&referral.CompletionCriteria, &referral.CompletedAt, &referral.CreatedAt,
			&firstName, &lastName,
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to scan referral: %v", err))
			continue
		}
		
		referrals = append(referrals, referral)
	}
	
	// Get total count
	var total int
	if status == "all" || status == "" {
		err = s.db.QueryRow(countQuery, userID).Scan(&total)
	} else {
		err = s.db.QueryRow(countQuery, userID, status).Scan(&total)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get referral count: %v", err))
		total = len(referrals)
	}
	
	return referrals, total, nil
}

// GetUserTierReward calculates the reward amount based on user's current tier
func (s *ReferralService) GetUserTierReward(userID int64) (float64, error) {
	// Get successful referrals count
	var successfulReferrals int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM referrals 
		WHERE referrer_user_id = $1 AND status = 'completed'`, userID).Scan(&successfulReferrals)
	if err != nil {
		successfulReferrals = 0
	}
	
	config := s.GetReferralConfig()
	
	// Find appropriate tier
	for i := len(config.Tiers) - 1; i >= 0; i-- {
		tier := config.Tiers[i]
		if successfulReferrals >= tier.MinReferrals {
			return tier.RewardPerReferral, nil
		}
	}
	
	// Default to bronze tier
	return config.Tiers[0].RewardPerReferral, nil
}

// Private helper methods

// getUserTier determines user's current tier and next tier requirement
func (s *ReferralService) getUserTier(successfulReferrals int) (string, int) {
	config := s.GetReferralConfig()
	
	// Find current tier
	currentTier := config.Tiers[0]
	for i := len(config.Tiers) - 1; i >= 0; i-- {
		tier := config.Tiers[i]
		if successfulReferrals >= tier.MinReferrals {
			currentTier = tier
			break
		}
	}
	
	// Find next tier requirement
	nextTierRequirement := 0
	for _, tier := range config.Tiers {
		if tier.MinReferrals > successfulReferrals {
			nextTierRequirement = tier.MinReferrals - successfulReferrals
			break
		}
	}
	
	return currentTier.Name, nextTierRequirement
}

// addBonusBalance adds bonus balance to user's wallet within a transaction
func (s *ReferralService) addBonusBalance(tx *sql.Tx, userID int64, amount float64, description string) error {
	// Create or update wallet
	_, err := tx.Exec(`
		INSERT INTO user_wallets (user_id, bonus_balance, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			bonus_balance = user_wallets.bonus_balance + $2,
			updated_at = NOW()`, userID, amount)
	
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	
	// Create transaction record
	_, err = tx.Exec(`
		INSERT INTO wallet_transactions (user_id, transaction_type, amount, balance_type, 
										description, status, created_at, completed_at)
		VALUES ($1, 'bonus_credit', $2, 'bonus', $3, 'completed', NOW(), NOW())`,
		userID, amount, description)
	
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}
	
	return nil
}

// checkAndApplyTierBonus checks if user qualifies for tier upgrade bonus
func (s *ReferralService) checkAndApplyTierBonus(tx *sql.Tx, userID int64) error {
	// Get current successful referrals count
	var successfulReferrals int
	err := tx.QueryRow(`
		SELECT COUNT(*) FROM referrals 
		WHERE referrer_user_id = $1 AND status = 'completed'`, userID).Scan(&successfulReferrals)
	if err != nil {
		return fmt.Errorf("failed to get referral count: %w", err)
	}
	
	config := s.GetReferralConfig()
	
	// Check if user just reached a tier milestone
	for _, tier := range config.Tiers[1:] { // Skip bronze tier (no bonus)
		if successfulReferrals == tier.MinReferrals && tier.BonusReward > 0 {
			// Check if bonus was already given
			var bonusCount int
			err = tx.QueryRow(`
				SELECT COUNT(*) FROM wallet_transactions 
				WHERE user_id = $1 AND transaction_type = 'bonus_credit' 
				AND description = $2`, userID, fmt.Sprintf("Tier upgrade bonus - %s", tier.Name)).Scan(&bonusCount)
			
			if err != nil || bonusCount > 0 {
				continue // Bonus already given or error occurred
			}
			
			// Give tier upgrade bonus
			err = s.addBonusBalance(tx, userID, tier.BonusReward, 
				fmt.Sprintf("Tier upgrade bonus - %s", tier.Name))
			if err != nil {
				return fmt.Errorf("failed to give tier bonus: %w", err)
			}
			
			logger.Info(fmt.Sprintf("User %d received tier upgrade bonus: %s (%.2f)", 
				userID, tier.Name, tier.BonusReward))
			break
		}
	}
	
	return nil
}

// ValidateReferralCode checks if a referral code is valid
func (s *ReferralService) ValidateReferralCode(referralCode string) (bool, int64, error) {
	var referrerUserID int64
	err := s.db.QueryRow(`
		SELECT id FROM users WHERE referral_code = $1 AND is_active = true`, referralCode).Scan(&referrerUserID)
	
	if err == sql.ErrNoRows {
		return false, 0, nil
	} else if err != nil {
		return false, 0, fmt.Errorf("database error: %w", err)
	}
	
	return true, referrerUserID, nil
}

// GetReferralLeaderboard returns top referrers for a leaderboard
func (s *ReferralService) GetReferralLeaderboard(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.referral_code,
			   COUNT(r.id) as total_referrals,
			   COUNT(CASE WHEN r.status = 'completed' THEN 1 END) as successful_referrals,
			   COALESCE(SUM(CASE WHEN r.status = 'completed' THEN r.reward_amount END), 0) as total_earnings
		FROM users u
		LEFT JOIN referrals r ON u.id = r.referrer_user_id
		WHERE u.is_active = true 
		GROUP BY u.id, u.first_name, u.last_name, u.referral_code
		HAVING COUNT(r.id) > 0
		ORDER BY successful_referrals DESC, total_earnings DESC
		LIMIT $1`
	
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()
	
	var leaderboard []map[string]interface{}
	rank := 1
	
	for rows.Next() {
		var userID int64
		var firstName, lastName, referralCode sql.NullString
		var totalReferrals, successfulReferrals int
		var totalEarnings float64
		
		err := rows.Scan(&userID, &firstName, &lastName, &referralCode,
			&totalReferrals, &successfulReferrals, &totalEarnings)
		if err != nil {
			continue
		}
		
		// Determine tier
		tier, _ := s.getUserTier(successfulReferrals)
		
		entry := map[string]interface{}{
			"rank":                rank,
			"user_id":             userID,
			"name":                fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"referral_code":       referralCode.String,
			"total_referrals":     totalReferrals,
			"successful_referrals": successfulReferrals,
			"total_earnings":      totalEarnings,
			"current_tier":        tier,
		}
		
		leaderboard = append(leaderboard, entry)
		rank++
	}
	
	return leaderboard, nil
}