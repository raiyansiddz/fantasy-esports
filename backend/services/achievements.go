package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
	"time"
)

type AchievementService struct {
	db *sql.DB
}

func NewAchievementService(db *sql.DB) *AchievementService {
	return &AchievementService{db: db}
}

// Admin methods
func (s *AchievementService) CreateAchievement(req models.CreateAchievementRequest, createdBy int64) (*models.Achievement, error) {
	criteriaJSON, err := json.Marshal(req.TriggerCriteria)
	if err != nil {
		return nil, fmt.Errorf("error marshaling trigger criteria: %w", err)
	}

	query := `
		INSERT INTO achievements (name, description, badge_icon, badge_color, category, trigger_type, trigger_criteria, 
			reward_type, reward_value, is_hidden, sort_order, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	achievement := &models.Achievement{
		Name:            req.Name,
		Description:     req.Description,
		BadgeIcon:       req.BadgeIcon,
		BadgeColor:      req.BadgeColor,
		Category:        req.Category,
		TriggerType:     req.TriggerType,
		TriggerCriteria: criteriaJSON,
		RewardType:      req.RewardType,
		RewardValue:     req.RewardValue,
		IsActive:        true,
		IsHidden:        req.IsHidden,
		SortOrder:       req.SortOrder,
		CreatedBy:       createdBy,
	}

	err = s.db.QueryRow(query, achievement.Name, achievement.Description, achievement.BadgeIcon,
		achievement.BadgeColor, achievement.Category, achievement.TriggerType, achievement.TriggerCriteria,
		achievement.RewardType, achievement.RewardValue, achievement.IsHidden, achievement.SortOrder,
		achievement.CreatedBy).Scan(&achievement.ID, &achievement.CreatedAt, &achievement.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("error creating achievement: %w", err)
	}

	return achievement, nil
}

func (s *AchievementService) UpdateAchievement(id int64, req models.CreateAchievementRequest) error {
	criteriaJSON, err := json.Marshal(req.TriggerCriteria)
	if err != nil {
		return fmt.Errorf("error marshaling trigger criteria: %w", err)
	}

	query := `
		UPDATE achievements SET 
			name = $1, description = $2, badge_icon = $3, badge_color = $4, 
			category = $5, trigger_type = $6, trigger_criteria = $7, 
			reward_type = $8, reward_value = $9, is_hidden = $10, 
			sort_order = $11, updated_at = CURRENT_TIMESTAMP
		WHERE id = $12
	`
	
	_, err = s.db.Exec(query, req.Name, req.Description, req.BadgeIcon, req.BadgeColor,
		req.Category, req.TriggerType, criteriaJSON, req.RewardType, req.RewardValue,
		req.IsHidden, req.SortOrder, id)
	
	return err
}

func (s *AchievementService) DeleteAchievement(id int64) error {
	_, err := s.db.Exec("DELETE FROM achievements WHERE id = $1", id)
	return err
}

func (s *AchievementService) GetAchievements(isActive *bool) ([]models.Achievement, error) {
	query := "SELECT id, name, description, badge_icon, badge_color, category, trigger_type, trigger_criteria, reward_type, reward_value, is_active, is_hidden, sort_order, created_by, created_at, updated_at FROM achievements"
	args := []interface{}{}
	
	if isActive != nil {
		query += " WHERE is_active = $1"
		args = append(args, *isActive)
	}
	
	query += " ORDER BY sort_order, name"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var a models.Achievement
		err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.BadgeIcon, &a.BadgeColor,
			&a.Category, &a.TriggerType, &a.TriggerCriteria, &a.RewardType, &a.RewardValue,
			&a.IsActive, &a.IsHidden, &a.SortOrder, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}

	return achievements, nil
}

// User achievement methods
func (s *AchievementService) CheckAndAwardAchievements(userID int64, triggerType string, contextData map[string]interface{}) error {
	// Get all active achievements of the trigger type
	achievements, err := s.getAchievementsByTrigger(triggerType)
	if err != nil {
		return err
	}

	for _, achievement := range achievements {
		// Check if user already has this achievement
		hasAchievement, err := s.userHasAchievement(userID, achievement.ID)
		if err != nil {
			return err
		}
		if hasAchievement {
			continue
		}

		// Check if criteria is met
		if s.checkCriteria(achievement.TriggerCriteria, contextData) {
			err = s.awardAchievement(userID, achievement)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AchievementService) GetUserAchievements(userID int64) ([]models.UserAchievement, error) {
	query := `
		SELECT ua.id, ua.user_id, ua.achievement_id, ua.earned_at, ua.progress_data, ua.is_featured,
			a.name, a.description, a.badge_icon, a.badge_color, a.category, a.reward_type, a.reward_value
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
		ORDER BY ua.earned_at DESC
	`
	
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userAchievements []models.UserAchievement
	for rows.Next() {
		var ua models.UserAchievement
		var a models.Achievement
		
		err := rows.Scan(&ua.ID, &ua.UserID, &ua.AchievementID, &ua.EarnedAt, &ua.ProgressData, &ua.IsFeatured,
			&a.Name, &a.Description, &a.BadgeIcon, &a.BadgeColor, &a.Category, &a.RewardType, &a.RewardValue)
		if err != nil {
			return nil, err
		}
		
		ua.Achievement = &a
		userAchievements = append(userAchievements, ua)
	}

	return userAchievements, nil
}

func (s *AchievementService) GetAchievementProgress(userID int64, achievementID int64) (*models.AchievementProgress, error) {
	// Get achievement criteria
	var triggerCriteria json.RawMessage
	err := s.db.QueryRow("SELECT trigger_criteria FROM achievements WHERE id = $1", achievementID).Scan(&triggerCriteria)
	if err != nil {
		return nil, err
	}

	// Get user's current stats (this would need to be implemented based on specific criteria)
	currentStats := s.calculateUserStats(userID)
	
	progress := &models.AchievementProgress{
		AchievementID:   achievementID,
		CurrentProgress: currentStats,
		IsCompleted:     false,
		CompletionRate:  0.0,
	}

	// Check if user has achievement
	hasAchievement, err := s.userHasAchievement(userID, achievementID)
	if err != nil {
		return nil, err
	}
	
	progress.IsCompleted = hasAchievement
	if hasAchievement {
		progress.CompletionRate = 1.0
	}

	return progress, nil
}

// Helper methods
func (s *AchievementService) getAchievementsByTrigger(triggerType string) ([]models.Achievement, error) {
	query := `
		SELECT id, name, description, badge_icon, badge_color, category, trigger_type, 
			trigger_criteria, reward_type, reward_value, is_active, is_hidden, sort_order, 
			created_by, created_at, updated_at 
		FROM achievements 
		WHERE trigger_type = $1 AND is_active = true
	`
	
	rows, err := s.db.Query(query, triggerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var a models.Achievement
		err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.BadgeIcon, &a.BadgeColor,
			&a.Category, &a.TriggerType, &a.TriggerCriteria, &a.RewardType, &a.RewardValue,
			&a.IsActive, &a.IsHidden, &a.SortOrder, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}

	return achievements, nil
}

func (s *AchievementService) userHasAchievement(userID, achievementID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM user_achievements WHERE user_id = $1 AND achievement_id = $2", 
		userID, achievementID).Scan(&count)
	return count > 0, err
}

func (s *AchievementService) checkCriteria(criteriaJSON json.RawMessage, contextData map[string]interface{}) bool {
	var criteria map[string]interface{}
	err := json.Unmarshal(criteriaJSON, &criteria)
	if err != nil {
		return false
	}

	for key, requiredValue := range criteria {
		if contextValue, exists := contextData[key]; !exists {
			return false
		} else {
			// Convert to float64 for comparison
			if reqVal, ok := requiredValue.(float64); ok {
				if ctxVal, ok := contextValue.(float64); ok {
					if ctxVal < reqVal {
						return false
					}
				} else {
					return false
				}
			}
		}
	}

	return true
}

func (s *AchievementService) awardAchievement(userID int64, achievement models.Achievement) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert user achievement
	_, err = tx.Exec(`
		INSERT INTO user_achievements (user_id, achievement_id, earned_at)
		VALUES ($1, $2, $3)
	`, userID, achievement.ID, time.Now())
	if err != nil {
		return err
	}

	// Award bonus if applicable
	if achievement.RewardType != nil && *achievement.RewardType == "bonus" && achievement.RewardValue > 0 {
		_, err = tx.Exec(`
			UPDATE user_wallets SET bonus_balance = bonus_balance + $1 WHERE user_id = $2
		`, achievement.RewardValue, userID)
		if err != nil {
			return err
		}

		// Log wallet transaction
		_, err = tx.Exec(`
			INSERT INTO wallet_transactions (user_id, transaction_type, amount, balance_type, description, status)
			VALUES ($1, 'bonus_credit', $2, 'bonus', $3, 'completed')
		`, userID, achievement.RewardValue, fmt.Sprintf("Achievement bonus: %s", achievement.Name))
		if err != nil {
			return err
		}
	}

	// Create friend activity
	activityData := map[string]interface{}{
		"achievement_id":   achievement.ID,
		"achievement_name": achievement.Name,
		"reward_value":     achievement.RewardValue,
	}
	activityJSON, _ := json.Marshal(activityData)
	
	_, err = tx.Exec(`
		INSERT INTO friend_activities (user_id, activity_type, activity_data)
		VALUES ($1, 'achievement_earned', $2)
	`, userID, activityJSON)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *AchievementService) calculateUserStats(userID int64) map[string]interface{} {
	stats := make(map[string]interface{})

	// Teams created count
	var teamsCreated int
	s.db.QueryRow("SELECT COUNT(*) FROM user_teams WHERE user_id = $1", userID).Scan(&teamsCreated)
	stats["teams_created"] = float64(teamsCreated)

	// Contests won count
	var contestsWon int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM contest_participants cp
		JOIN contests c ON cp.contest_id = c.id
		WHERE cp.user_id = $1 AND cp.rank = 1 AND c.status = 'completed'
	`, userID).Scan(&contestsWon)
	stats["contests_won"] = float64(contestsWon)

	// Successful referrals count
	var successfulReferrals int
	s.db.QueryRow("SELECT COUNT(*) FROM referrals WHERE referrer_user_id = $1 AND status = 'completed'", userID).Scan(&successfulReferrals)
	stats["successful_referrals"] = float64(successfulReferrals)

	// Friends added count
	var friendsAdded int
	s.db.QueryRow("SELECT COUNT(*) FROM user_friends WHERE user_id = $1 AND status = 'accepted'", userID).Scan(&friendsAdded)
	stats["friends_added"] = float64(friendsAdded)

	return stats
}