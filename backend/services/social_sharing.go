package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
	"net/url"
	"strings"
)

type SocialSharingService struct {
	db      *sql.DB
	baseURL string
}

func NewSocialSharingService(db *sql.DB, baseURL string) *SocialSharingService {
	return &SocialSharingService{
		db:      db,
		baseURL: baseURL,
	}
}

func (s *SocialSharingService) CreateShare(userID int64, req models.CreateShareRequest) (*models.SocialShare, error) {
	shareData, err := json.Marshal(req.ShareData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal share data: %v", err)
	}

	// Enhanced content_id validation based on share_type
	if err := s.validateContentID(req.ShareType, req.ContentID); err != nil {
		return nil, fmt.Errorf("content validation failed: %v", err)
	}

	// Generate share URL based on content
	shareURL := s.generateShareURL(req.ShareType, req.ContentID, userID)

	query := `
		INSERT INTO social_shares (user_id, share_type, platform, content_id, share_data, share_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	
	share := &models.SocialShare{
		UserID:    userID,
		ShareType: req.ShareType,
		Platform:  req.Platform,
		ContentID: req.ContentID,
		ShareData: shareData,
		ShareURL:  &shareURL,
	}

	err = s.db.QueryRow(query, share.UserID, share.ShareType, share.Platform, 
		share.ContentID, share.ShareData, share.ShareURL).Scan(&share.ID, &share.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create share record: %v", err)
	}

	return share, nil
}

func (s *SocialSharingService) GetPlatformURLs(userID int64, shareType string, contentID *int64, content models.ShareContent) (*models.PlatformShareURLs, error) {
	baseShareURL := s.generateShareURL(shareType, contentID, userID)
	
	urls := &models.PlatformShareURLs{
		Twitter:   s.generateTwitterURL(content, baseShareURL),
		Facebook:  s.generateFacebookURL(content, baseShareURL),
		WhatsApp:  s.generateWhatsAppURL(content, baseShareURL),
		Instagram: s.generateInstagramURL(content, baseShareURL),
	}

	return urls, nil
}

func (s *SocialSharingService) GenerateTeamCompositionContent(teamID int64) (*models.ShareContent, error) {
	// Get team details
	var teamName, userName string
	var totalCredits float64
	
	err := s.db.QueryRow(`
		SELECT ut.team_name, u.first_name || ' ' || u.last_name, ut.total_credits_used
		FROM user_teams ut
		JOIN users u ON ut.user_id = u.id
		WHERE ut.id = $1
	`, teamID).Scan(&teamName, &userName, &totalCredits)
	if err != nil {
		return nil, err
	}

	// Get player details
	rows, err := s.db.Query(`
		SELECT p.name, t.name, tp.is_captain, tp.is_vice_captain
		FROM team_players tp
		JOIN players p ON tp.player_id = p.id
		JOIN teams t ON tp.real_team_id = t.id
		WHERE tp.team_id = $1
		ORDER BY tp.is_captain DESC, tp.is_vice_captain DESC, p.name
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []string
	for rows.Next() {
		var playerName, teamName string
		var isCaptain, isViceCaptain bool
		
		err := rows.Scan(&playerName, &teamName, &isCaptain, &isViceCaptain)
		if err != nil {
			continue
		}

		role := ""
		if isCaptain {
			role = " (C)"
		} else if isViceCaptain {
			role = " (VC)"
		}
		
		players = append(players, fmt.Sprintf("%s%s", playerName, role))
	}

	content := &models.ShareContent{
		Title:       fmt.Sprintf("Check out %s's fantasy team: %s", userName, teamName),
		Description: fmt.Sprintf("Team: %s\nPlayers: %s\nCredits used: %.1f/100", 
			teamName, strings.Join(players, ", "), totalCredits),
		URL:         s.generateShareURL("team_composition", &teamID, 0),
		Hashtags:    []string{"FantasyEsports", "Gaming", "Esports", "FantasyTeam"},
		Metadata: map[string]interface{}{
			"team_id":       teamID,
			"team_name":     teamName,
			"player_count":  len(players),
			"credits_used":  totalCredits,
		},
	}

	return content, nil
}

func (s *SocialSharingService) GenerateContestWinContent(userID int64, contestID int64) (*models.ShareContent, error) {
	// Get contest win details
	var contestName, userName string
	var prizeWon float64
	var rank int
	
	err := s.db.QueryRow(`
		SELECT c.name, u.first_name || ' ' || u.last_name, cp.prize_won, cp.rank
		FROM contest_participants cp
		JOIN contests c ON cp.contest_id = c.id
		JOIN users u ON cp.user_id = u.id
		WHERE cp.user_id = $1 AND cp.contest_id = $2
	`, userID, contestID).Scan(&contestName, &userName, &prizeWon, &rank)
	if err != nil {
		return nil, err
	}

	var title, description string
	if rank == 1 {
		title = fmt.Sprintf("🏆 %s won %s!", userName, contestName)
		description = fmt.Sprintf("Champion! Won ₹%.2f in %s", prizeWon, contestName)
	} else {
		title = fmt.Sprintf("🎉 %s finished #%d in %s", userName, rank, contestName)
		description = fmt.Sprintf("Great performance! Won ₹%.2f by finishing #%d", prizeWon, rank)
	}

	content := &models.ShareContent{
		Title:       title,
		Description: description,
		URL:         s.generateShareURL("contest_win", &contestID, userID),
		Hashtags:    []string{"FantasyEsports", "Winner", "Gaming", "Esports", "Contest"},
		Metadata: map[string]interface{}{
			"contest_id":   contestID,
			"contest_name": contestName,
			"prize_won":    prizeWon,
			"rank":         rank,
		},
	}

	return content, nil
}

func (s *SocialSharingService) GenerateAchievementContent(userID int64, achievementID int64) (*models.ShareContent, error) {
	// Get achievement details
	var achievementName, achievementDesc, userName string
	var rewardValue float64
	
	err := s.db.QueryRow(`
		SELECT a.name, a.description, u.first_name || ' ' || u.last_name, COALESCE(a.reward_value, 0)
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		JOIN users u ON ua.user_id = u.id
		WHERE ua.user_id = $1 AND ua.achievement_id = $2
	`, userID, achievementID).Scan(&achievementName, &achievementDesc, &userName, &rewardValue)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("🏅 %s unlocked: %s", userName, achievementName)
	description := achievementDesc
	if rewardValue > 0 {
		description += fmt.Sprintf("\nReward: ₹%.2f bonus", rewardValue)
	}

	content := &models.ShareContent{
		Title:       title,
		Description: description,
		URL:         s.generateShareURL("achievement", &achievementID, userID),
		Hashtags:    []string{"Achievement", "FantasyEsports", "Gaming", "Badge"},
		Metadata: map[string]interface{}{
			"achievement_id":   achievementID,
			"achievement_name": achievementName,
			"reward_value":     rewardValue,
		},
	}

	return content, nil
}

func (s *SocialSharingService) TrackShareClick(shareID int64) error {
	_, err := s.db.Exec("UPDATE social_shares SET click_count = click_count + 1 WHERE id = $1", shareID)
	return err
}

func (s *SocialSharingService) GetUserShares(userID int64, platform string) ([]models.SocialShare, error) {
	query := `
		SELECT id, user_id, share_type, platform, content_id, share_data, share_url, click_count, created_at
		FROM social_shares
		WHERE user_id = $1`
	
	args := []interface{}{userID}
	if platform != "" {
		query += " AND platform = $2"
		args = append(args, platform)
	}
	
	query += " ORDER BY created_at DESC LIMIT 50"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []models.SocialShare
	for rows.Next() {
		var s models.SocialShare
		err := rows.Scan(&s.ID, &s.UserID, &s.ShareType, &s.Platform, &s.ContentID,
			&s.ShareData, &s.ShareURL, &s.ClickCount, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}

	return shares, nil
}

// Helper methods for generating platform URLs
func (s *SocialSharingService) generateShareURL(shareType string, contentID *int64, userID int64) string {
	baseURL := s.baseURL
	if contentID != nil {
		return fmt.Sprintf("%s/share/%s/%d?ref=%d", baseURL, shareType, *contentID, userID)
	}
	return fmt.Sprintf("%s/share/%s?ref=%d", baseURL, shareType, userID)
}

func (s *SocialSharingService) generateTwitterURL(content models.ShareContent, shareURL string) string {
	text := fmt.Sprintf("%s\n%s", content.Title, content.Description)
	hashtags := strings.Join(content.Hashtags, ",")
	
	params := url.Values{}
	params.Add("text", text)
	params.Add("url", shareURL)
	if hashtags != "" {
		params.Add("hashtags", hashtags)
	}
	
	return fmt.Sprintf("https://twitter.com/intent/tweet?%s", params.Encode())
}

func (s *SocialSharingService) generateFacebookURL(content models.ShareContent, shareURL string) string {
	params := url.Values{}
	params.Add("u", shareURL)
	params.Add("quote", fmt.Sprintf("%s - %s", content.Title, content.Description))
	
	return fmt.Sprintf("https://www.facebook.com/sharer/sharer.php?%s", params.Encode())
}

func (s *SocialSharingService) generateWhatsAppURL(content models.ShareContent, shareURL string) string {
	text := fmt.Sprintf("%s\n%s\n%s", content.Title, content.Description, shareURL)
	
	params := url.Values{}
	params.Add("text", text)
	
	return fmt.Sprintf("https://wa.me/?%s", params.Encode())
}

func (s *SocialSharingService) generateInstagramURL(content models.ShareContent, shareURL string) string {
	// Instagram doesn't support direct URL sharing like other platforms
	// This would typically return instructions or deep link
	return fmt.Sprintf("instagram://story?text=%s %s", 
		url.QueryEscape(content.Title), url.QueryEscape(shareURL))
}

func (s *SocialSharingService) GetShareAnalytics(userID *int64, platform string, days int) (map[string]interface{}, error) {
	query := `
		SELECT 
			share_type,
			COUNT(*) as share_count,
			SUM(click_count) as total_clicks,
			AVG(click_count) as avg_clicks
		FROM social_shares
		WHERE created_at >= CURRENT_DATE - INTERVAL '%d days'`
	
	args := []interface{}{}
	argIndex := 0
	
	if userID != nil {
		query += " AND user_id = $%d"
		argIndex++
		args = append(args, *userID)
	}
	
	if platform != "" {
		query += fmt.Sprintf(" AND platform = $%d", argIndex+1)
		args = append(args, platform)
	}
	
	query += " GROUP BY share_type ORDER BY share_count DESC"
	
	finalQuery := fmt.Sprintf(query, days)
	
	rows, err := s.db.Query(finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analytics := map[string]interface{}{
		"by_share_type": []map[string]interface{}{},
		"total_shares":  0,
		"total_clicks":  0,
	}

	totalShares := 0
	totalClicks := 0
	
	var shareTypes []map[string]interface{}
	
	for rows.Next() {
		var shareType string
		var shareCount, totalShareClicks int
		var avgClicks float64
		
		err := rows.Scan(&shareType, &shareCount, &totalShareClicks, &avgClicks)
		if err != nil {
			continue
		}
		
		shareTypes = append(shareTypes, map[string]interface{}{
			"share_type":    shareType,
			"share_count":   shareCount,
			"total_clicks":  totalShareClicks,
			"avg_clicks":    avgClicks,
		})
		
		totalShares += shareCount
		totalClicks += totalShareClicks
	}
	
	analytics["by_share_type"] = shareTypes
	analytics["total_shares"] = totalShares
	analytics["total_clicks"] = totalClicks
	
	if totalShares > 0 {
		analytics["avg_clicks_per_share"] = float64(totalClicks) / float64(totalShares)
	} else {
		analytics["avg_clicks_per_share"] = 0.0
	}

	return analytics, nil
}