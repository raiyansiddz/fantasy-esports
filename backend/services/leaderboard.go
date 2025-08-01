package services

import (
	"database/sql"
	"fmt"
	"time"
	"sync"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
)

type LeaderboardService struct {
	db             *sql.DB
	cache          map[string]*models.CachedLeaderboard
	cacheMutex     sync.RWMutex
	snapshots      map[int64]*models.RankingSnapshot
	snapshotMutex  sync.RWMutex
	updateChannel  chan models.RealTimeLeaderboardUpdate
}

func NewLeaderboardService(db *sql.DB) *LeaderboardService {
	service := &LeaderboardService{
		db:             db,
		cache:          make(map[string]*models.CachedLeaderboard),
		snapshots:      make(map[int64]*models.RankingSnapshot),
		updateChannel:  make(chan models.RealTimeLeaderboardUpdate, 100),
	}
	
	// Start the real-time update processor
	go service.processRealTimeUpdates()
	
	return service
}

// CalculateContestLeaderboard calculates real-time leaderboard for a contest
func (s *LeaderboardService) CalculateContestLeaderboard(contestID int64) (*models.Leaderboard, error) {
	logger.Info(fmt.Sprintf("Calculating leaderboard for contest %d", contestID))
	
	// Get contest info
	var contest models.Contest
	err := s.db.QueryRow(`
		SELECT id, match_id, name, total_prize_pool, current_participants, status
		FROM contests WHERE id = $1`, contestID).Scan(
		&contest.ID, &contest.MatchID, &contest.Name, &contest.TotalPrizePool,
		&contest.CurrentParticipants, &contest.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest info: %w", err)
	}

	// Calculate team points and rankings
	topPerformers, err := s.getTopPerformers(contestID, 50) // Top 50
	if err != nil {
		return nil, fmt.Errorf("failed to get top performers: %w", err)
	}

	leaderboard := &models.Leaderboard{
		ContestID:         contestID,
		TotalParticipants: contest.CurrentParticipants,
		TopPerformers:     topPerformers,
		LastUpdated:       time.Now(),
	}

	return leaderboard, nil
}

// GetLiveLeaderboard gets real-time leaderboard with live updates
func (s *LeaderboardService) GetLiveLeaderboard(contestID int64, userID int64) (*models.Leaderboard, error) {
	logger.Info(fmt.Sprintf("Getting live leaderboard for contest %d, user %d", contestID, userID))
	
	// Get base leaderboard
	leaderboard, err := s.CalculateContestLeaderboard(contestID)
	if err != nil {
		return nil, err
	}

	// Get user's rank and points
	userRank, userPoints, userTeamID, err := s.GetUserRankInContest(contestID, userID)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to get user rank for user %d in contest %d: %v", userID, contestID, err))
		userRank = 0
		userPoints = 0.0
		userTeamID = 0
	}

	leaderboard.MyRank = userRank
	leaderboard.MyPoints = userPoints
	leaderboard.MyTeamID = userTeamID

	// Get rankings around user
	if userRank > 0 {
		aroundMe, err := s.GetRankingsAroundUser(contestID, userRank, 5)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to get rankings around user: %v", err))
		} else {
			leaderboard.AroundMe = aroundMe
		}
	}

	return leaderboard, nil
}

// RecalculateFantasyPoints recalculates fantasy points for all teams in a match
func (s *LeaderboardService) RecalculateFantasyPoints(matchID int64) error {
	logger.Info(fmt.Sprintf("Recalculating fantasy points for match %d", matchID))
	
	// Get all teams in this match
	teams, err := s.getTeamsInMatch(matchID)
	if err != nil {
		return fmt.Errorf("failed to get teams in match: %w", err)
	}

	// Update points for each team
	for _, team := range teams {
		totalPoints, err := s.calculateTeamPoints(team.ID, matchID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to calculate points for team %d: %v", team.ID, err))
			continue
		}

		// Update team total points
		_, err = s.db.Exec(`
			UPDATE user_teams 
			SET total_points = $1, updated_at = NOW() 
			WHERE id = $2`, totalPoints, team.ID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to update points for team %d: %v", team.ID, err))
		}
	}

	// Update all contest leaderboards for this match
	err = s.updateContestRankings(matchID)
	if err != nil {
		return fmt.Errorf("failed to update contest rankings: %w", err)
	}

	logger.Info(fmt.Sprintf("Successfully recalculated fantasy points for match %d", matchID))
	return nil
}

// getTopPerformers gets top performing teams in a contest
func (s *LeaderboardService) getTopPerformers(contestID int64, limit int) ([]models.LeaderboardEntry, error) {
	query := `
		SELECT 
			ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, ut.created_at ASC) as rank,
			cp.user_id, 
			COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '') as username,
			ut.team_name,
			ut.total_points,
			u.avatar_url,
			cp.prize_won
		FROM contest_participants cp
		JOIN user_teams ut ON cp.team_id = ut.id
		JOIN users u ON cp.user_id = u.id
		WHERE cp.contest_id = $1
		ORDER BY ut.total_points DESC, ut.created_at ASC
		LIMIT $2`

	rows, err := s.db.Query(query, contestID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top performers: %w", err)
	}
	defer rows.Close()

	var performers []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		var username sql.NullString
		
		err := rows.Scan(
			&entry.Rank, &entry.UserID, &username, &entry.TeamName,
			&entry.Points, &entry.AvatarURL, &entry.PrizeWon,
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to scan leaderboard entry: %v", err))
			continue
		}
		
		if username.Valid {
			entry.Username = username.String
		} else {
			entry.Username = "Anonymous"
		}
		
		performers = append(performers, entry)
	}

	return performers, nil
}

// getUserRankInContest gets user's current rank in a contest
func (s *LeaderboardService) GetUserRankInContest(contestID int64, userID int64) (int, float64, int64, error) {
	query := `
		WITH ranked_teams AS (
			SELECT 
				cp.user_id,
				cp.team_id,
				ut.total_points,
				ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, ut.created_at ASC) as rank
			FROM contest_participants cp
			JOIN user_teams ut ON cp.team_id = ut.id
			WHERE cp.contest_id = $1
		)
		SELECT rank, total_points, team_id
		FROM ranked_teams
		WHERE user_id = $2`

	var rank int
	var points float64
	var teamID int64
	
	err := s.db.QueryRow(query, contestID, userID).Scan(&rank, &points, &teamID)
	if err != nil {
		return 0, 0.0, 0, err
	}

	return rank, points, teamID, nil
}

// getRankingsAroundUser gets rankings around a specific user
func (s *LeaderboardService) GetRankingsAroundUser(contestID int64, userRank int, radius int) ([]models.LeaderboardEntry, error) {
	startRank := userRank - radius
	if startRank < 1 {
		startRank = 1
	}
	endRank := userRank + radius

	query := `
		WITH ranked_teams AS (
			SELECT 
				ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, ut.created_at ASC) as rank,
				cp.user_id,
				COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '') as username,
				ut.team_name,
				ut.total_points,
				u.avatar_url,
				cp.prize_won
			FROM contest_participants cp
			JOIN user_teams ut ON cp.team_id = ut.id
			JOIN users u ON cp.user_id = u.id
			WHERE cp.contest_id = $1
		)
		SELECT rank, user_id, username, team_name, total_points, avatar_url, prize_won
		FROM ranked_teams
		WHERE rank BETWEEN $2 AND $3
		ORDER BY rank`

	rows, err := s.db.Query(query, contestID, startRank, endRank)
	if err != nil {
		return nil, fmt.Errorf("failed to query rankings around user: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		var username sql.NullString
		
		err := rows.Scan(
			&entry.Rank, &entry.UserID, &username, &entry.TeamName,
			&entry.Points, &entry.AvatarURL, &entry.PrizeWon,
		)
		if err != nil {
			continue
		}
		
		if username.Valid {
			entry.Username = username.String
		} else {
			entry.Username = "Anonymous"
		}
		
		entries = append(entries, entry)
	}

	return entries, nil
}

// getTeamsInMatch gets all fantasy teams for a specific match
func (s *LeaderboardService) getTeamsInMatch(matchID int64) ([]models.UserTeam, error) {
	query := `
		SELECT id, user_id, match_id, team_name, captain_player_id, vice_captain_player_id, 
		       total_credits_used, total_points
		FROM user_teams 
		WHERE match_id = $1`

	rows, err := s.db.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.UserTeam
	for rows.Next() {
		var team models.UserTeam
		err := rows.Scan(
			&team.ID, &team.UserID, &team.MatchID, &team.TeamName,
			&team.CaptainPlayerID, &team.ViceCaptainPlayerID,
			&team.TotalCreditsUsed, &team.TotalPoints,
		)
		if err != nil {
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// calculateTeamPoints calculates total fantasy points for a team
func (s *LeaderboardService) calculateTeamPoints(teamID int64, matchID int64) (float64, error) {
	// Get team players with their roles
	query := `
		SELECT tp.player_id, tp.is_captain, tp.is_vice_captain
		FROM team_players tp
		WHERE tp.team_id = $1`

	rows, err := s.db.Query(query, teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to get team players: %w", err)
	}
	defer rows.Close()

	var totalPoints float64
	for rows.Next() {
		var playerID int64
		var isCaptain, isViceCaptain bool
		
		err := rows.Scan(&playerID, &isCaptain, &isViceCaptain)
		if err != nil {
			continue
		}

		// Calculate player points from match events
		playerPoints, err := s.calculatePlayerPoints(playerID, matchID)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to calculate points for player %d: %v", playerID, err))
			continue
		}

		// Apply multipliers
		if isCaptain {
			playerPoints *= 2.0
		} else if isViceCaptain {
			playerPoints *= 1.5
		}

		totalPoints += playerPoints

		// Update individual player points
		_, err = s.db.Exec(`
			UPDATE team_players 
			SET points_earned = $1 
			WHERE team_id = $2 AND player_id = $3`,
			playerPoints, teamID, playerID)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to update player points: %v", err))
		}
	}

	return totalPoints, nil
}

// calculatePlayerPoints calculates fantasy points for a player in a match
func (s *LeaderboardService) calculatePlayerPoints(playerID int64, matchID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(points), 0) as total_points
		FROM match_events
		WHERE player_id = $1 AND match_id = $2`

	var points float64
	err := s.db.QueryRow(query, playerID, matchID).Scan(&points)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate player points: %w", err)
	}

	return points, nil
}

// updateContestRankings updates rankings for all contests in a match
func (s *LeaderboardService) updateContestRankings(matchID int64) error {
	// Get all contests for this match
	contestQuery := `
		SELECT id FROM contests WHERE match_id = $1 AND status IN ('upcoming', 'live')`

	rows, err := s.db.Query(contestQuery, matchID)
	if err != nil {
		return fmt.Errorf("failed to get contests for match: %w", err)
	}
	defer rows.Close()

	var contestIDs []int64
	for rows.Next() {
		var contestID int64
		if err := rows.Scan(&contestID); err != nil {
			continue
		}
		contestIDs = append(contestIDs, contestID)
	}

	// Update rankings for each contest
	for _, contestID := range contestIDs {
		err := s.updateSingleContestRankings(contestID)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to update rankings for contest %d: %v", contestID, err))
		}
	}

	return nil
}

// updateSingleContestRankings updates the rank for all participants in a contest
func (s *LeaderboardService) updateSingleContestRankings(contestID int64) error {
	// Update ranks based on points
	updateQuery := `
		WITH ranked_participants AS (
			SELECT 
				cp.id,
				ROW_NUMBER() OVER (ORDER BY ut.total_points DESC, ut.created_at ASC) as new_rank
			FROM contest_participants cp
			JOIN user_teams ut ON cp.team_id = ut.id
			WHERE cp.contest_id = $1
		)
		UPDATE contest_participants 
		SET rank = rp.new_rank
		FROM ranked_participants rp
		WHERE contest_participants.id = rp.id`

	_, err := s.db.Exec(updateQuery, contestID)
	if err != nil {
		return fmt.Errorf("failed to update contest rankings: %w", err)
	}

	logger.Info(fmt.Sprintf("Updated rankings for contest %d", contestID))
	return nil
}

// GetUserTeamPerformance gets detailed performance breakdown for a user's team
func (s *LeaderboardService) GetUserTeamPerformance(teamID int64, userID int64) (*models.TeamPerformance, error) {
	// Verify team ownership
	var teamUserID int64
	err := s.db.QueryRow("SELECT user_id FROM user_teams WHERE id = $1", teamID).Scan(&teamUserID)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}
	if teamUserID != userID {
		return nil, fmt.Errorf("unauthorized access to team")
	}

	// Get team details
	var team models.UserTeam
	err = s.db.QueryRow(`
		SELECT id, team_name, total_points, final_rank, match_id
		FROM user_teams WHERE id = $1`, teamID).Scan(
		&team.ID, &team.TeamName, &team.TotalPoints, &team.FinalRank, &team.MatchID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get team details: %w", err)
	}

	// Get player breakdown
	playerQuery := `
		SELECT 
			tp.player_id, tp.points_earned, tp.is_captain, tp.is_vice_captain,
			p.name as player_name, p.credit_value, p.role,
			t.name as team_name
		FROM team_players tp
		JOIN players p ON tp.player_id = p.id
		JOIN teams t ON p.team_id = t.id
		WHERE tp.team_id = $1
		ORDER BY tp.points_earned DESC`

	rows, err := s.db.Query(playerQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player breakdown: %w", err)
	}
	defer rows.Close()

	var playerBreakdown []models.FantasyPlayerPerformance
	for rows.Next() {
		var player models.FantasyPlayerPerformance
		var playerName, role, teamName string
		var creditValue float64
		
		err := rows.Scan(
			&player.PlayerID, &player.BasePoints, &player.IsCaptain, &player.IsViceCaptain,
			&playerName, &creditValue, &role, &teamName,
		)
		if err != nil {
			continue
		}

		player.PlayerName = playerName
		player.Role = role
		player.TeamName = teamName
		player.CreditValue = creditValue

		// Calculate bonuses
		if player.IsCaptain {
			player.CaptainBonus = player.BasePoints // 2x total - base = base as bonus
			player.TotalPoints = player.BasePoints * 2.0
		} else if player.IsViceCaptain {
			player.ViceCaptainBonus = player.BasePoints * 0.5 // 1.5x total - base = 0.5x base as bonus
			player.TotalPoints = player.BasePoints * 1.5
		} else {
			player.TotalPoints = player.BasePoints
		}

		playerBreakdown = append(playerBreakdown, player)
	}

	performance := &models.TeamPerformance{
		TeamID:           team.ID,
		TeamName:         team.TeamName,
		TotalPoints:      team.TotalPoints,
		FinalRank:        team.FinalRank,
		PlayerBreakdown:  playerBreakdown,
	}

	return performance, nil
}

// ================================
// REAL-TIME LEADERBOARD METHODS ⭐
// ================================

// processRealTimeUpdates processes real-time leaderboard updates in background
func (s *LeaderboardService) processRealTimeUpdates() {
	logger.Info("Started real-time leaderboard update processor")
	
	for update := range s.updateChannel {
		s.handleRealTimeUpdate(update)
	}
}

// TriggerRealTimeUpdate triggers a real-time leaderboard update
func (s *LeaderboardService) TriggerRealTimeUpdate(contestID int64, triggerSource string, matchEventID *int64) error {
	// Create snapshot of current state
	snapshot, err := s.createRankingSnapshot(contestID)
	if err != nil {
		return fmt.Errorf("failed to create ranking snapshot: %w", err)
	}
	
	// Recalculate leaderboard
	newLeaderboard, err := s.CalculateContestLeaderboard(contestID)
	if err != nil {
		return fmt.Errorf("failed to recalculate leaderboard: %w", err)
	}
	
	// Detect rank changes
	rankChanges := s.detectRankChanges(contestID, snapshot, newLeaderboard)
	
	// Create update message
	update := models.RealTimeLeaderboardUpdate{
		ContestID:         contestID,
		UpdateID:          s.generateUpdateID(contestID),
		UpdateType:        s.determineUpdateType(rankChanges),
		UpdateTimestamp:   time.Now(),
		AffectedUserIDs:   s.extractAffectedUserIDs(rankChanges),
		RankChanges:       rankChanges,
		TopPerformers:     newLeaderboard.TopPerformers,
		TotalParticipants: newLeaderboard.TotalParticipants,
		MatchEventID:      matchEventID,
		TriggerSource:     triggerSource,
	}
	
	// Send to update channel
	select {
	case s.updateChannel <- update:
		logger.Info(fmt.Sprintf("Triggered real-time update for contest %d", contestID))
	default:
		logger.Warn(fmt.Sprintf("Update channel full, dropped update for contest %d", contestID))
	}
	
	return nil
}

// GetCachedLeaderboard gets leaderboard from cache or calculates fresh
func (s *LeaderboardService) GetCachedLeaderboard(contestID int64, maxAge time.Duration) (*models.Leaderboard, error) {
	cacheKey := s.generateCacheKey(contestID)
	
	s.cacheMutex.RLock()
	cached, exists := s.cache[cacheKey]
	s.cacheMutex.RUnlock()
	
	if exists && !cached.IsDirty && time.Since(cached.CachedAt) < maxAge {
		logger.Info(fmt.Sprintf("Returning cached leaderboard for contest %d", contestID))
		return cached.Leaderboard, nil
	}
	
	// Calculate fresh leaderboard
	leaderboard, err := s.CalculateContestLeaderboard(contestID)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	s.cacheLeaderboard(contestID, leaderboard)
	
	return leaderboard, nil
}

// InvalidateCache invalidates the cache for a contest
func (s *LeaderboardService) InvalidateCache(contestID int64) {
	cacheKey := s.generateCacheKey(contestID)
	
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	if cached, exists := s.cache[cacheKey]; exists {
		cached.IsDirty = true
		logger.Info(fmt.Sprintf("Invalidated cache for contest %d", contestID))
	}
}

// GetRankingSnapshot gets the current ranking snapshot for a contest
func (s *LeaderboardService) GetRankingSnapshot(contestID int64) (*models.RankingSnapshot, error) {
	s.snapshotMutex.RLock()
	snapshot, exists := s.snapshots[contestID]
	s.snapshotMutex.RUnlock()
	
	if !exists {
		// Create initial snapshot
		return s.createRankingSnapshot(contestID)
	}
	
	return snapshot, nil
}

// handleRealTimeUpdate processes a single real-time update
func (s *LeaderboardService) handleRealTimeUpdate(update models.RealTimeLeaderboardUpdate) {
	logger.Info(fmt.Sprintf("Processing real-time update %s for contest %d", update.UpdateID, update.ContestID))
	
	// Update cache with fresh data
	s.InvalidateCache(update.ContestID)
	
	// Store the snapshot for future comparisons
	s.storeRankingSnapshot(update.ContestID, update.TopPerformers)
	
	// Here you would typically broadcast to WebSocket connections
	// This will be handled by the connection manager
	
	logger.Info(fmt.Sprintf("Processed update affecting %d users", len(update.AffectedUserIDs)))
}

// createRankingSnapshot creates a snapshot of current rankings
func (s *LeaderboardService) createRankingSnapshot(contestID int64) (*models.RankingSnapshot, error) {
	leaderboard, err := s.CalculateContestLeaderboard(contestID)
	if err != nil {
		return nil, err
	}
	
	snapshot := &models.RankingSnapshot{
		ContestID:   contestID,
		SnapshotID:  s.generateSnapshotID(contestID),
		CreatedAt:   time.Now(),
		Rankings:    make(map[int64]models.RankPosition),
		TotalPoints: make(map[int64]float64),
	}
	
	for _, entry := range leaderboard.TopPerformers {
		snapshot.Rankings[entry.UserID] = models.RankPosition{
			Rank:     entry.Rank,
			Points:   entry.Points,
			TeamID:   0, // Will be filled from team query if needed
			Username: entry.Username,
			TeamName: entry.TeamName,
		}
		snapshot.TotalPoints[entry.UserID] = entry.Points
	}
	
	// Store snapshot
	s.snapshotMutex.Lock()
	s.snapshots[contestID] = snapshot
	s.snapshotMutex.Unlock()
	
	return snapshot, nil
}

// detectRankChanges compares current leaderboard with previous snapshot
func (s *LeaderboardService) detectRankChanges(contestID int64, previousSnapshot *models.RankingSnapshot, newLeaderboard *models.Leaderboard) []models.LeaderboardRankChange {
	var changes []models.LeaderboardRankChange
	
	if previousSnapshot == nil {
		return changes // No previous data to compare
	}
	
	for _, entry := range newLeaderboard.TopPerformers {
		if prevPos, existed := previousSnapshot.Rankings[entry.UserID]; existed {
			// User existed in previous snapshot, check for changes
			if prevPos.Rank != entry.Rank || prevPos.Points != entry.Points {
				changes = append(changes, models.LeaderboardRankChange{
					UserID:         entry.UserID,
					TeamID:         0, // Will be filled if needed
					Username:       entry.Username,
					TeamName:       entry.TeamName,
					PreviousRank:   prevPos.Rank,
					NewRank:        entry.Rank,
					RankChange:     prevPos.Rank - entry.Rank, // positive = moved up
					PreviousPoints: prevPos.Points,
					NewPoints:      entry.Points,
					PointsChange:   entry.Points - prevPos.Points,
					AvatarURL:      entry.AvatarURL,
				})
			}
		} else {
			// New user in leaderboard
			changes = append(changes, models.LeaderboardRankChange{
				UserID:         entry.UserID,
				TeamID:         0,
				Username:       entry.Username,
				TeamName:       entry.TeamName,
				PreviousRank:   0, // New entry
				NewRank:        entry.Rank,
				RankChange:     entry.Rank, // negative for new entries
				PreviousPoints: 0,
				NewPoints:      entry.Points,
				PointsChange:   entry.Points,
				AvatarURL:      entry.AvatarURL,
			})
		}
	}
	
	return changes
}

// storeRankingSnapshot stores a ranking snapshot for future comparisons
func (s *LeaderboardService) storeRankingSnapshot(contestID int64, topPerformers []models.LeaderboardEntry) {
	snapshot := &models.RankingSnapshot{
		ContestID:   contestID,
		SnapshotID:  s.generateSnapshotID(contestID),
		CreatedAt:   time.Now(),
		Rankings:    make(map[int64]models.RankPosition),
		TotalPoints: make(map[int64]float64),
	}
	
	for _, entry := range topPerformers {
		snapshot.Rankings[entry.UserID] = models.RankPosition{
			Rank:     entry.Rank,
			Points:   entry.Points,
			Username: entry.Username,
			TeamName: entry.TeamName,
		}
		snapshot.TotalPoints[entry.UserID] = entry.Points
	}
	
	s.snapshotMutex.Lock()
	s.snapshots[contestID] = snapshot
	s.snapshotMutex.Unlock()
}

// cacheLeaderboard caches a leaderboard result
func (s *LeaderboardService) cacheLeaderboard(contestID int64, leaderboard *models.Leaderboard) {
	cacheKey := s.generateCacheKey(contestID)
	
	cached := &models.CachedLeaderboard{
		Leaderboard: leaderboard,
		CacheKey:    cacheKey,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5-minute cache
		IsDirty:     false,
		LastEventID: 0, // Would be set based on latest match event
	}
	
	s.cacheMutex.Lock()
	s.cache[cacheKey] = cached
	s.cacheMutex.Unlock()
	
	logger.Info(fmt.Sprintf("Cached leaderboard for contest %d", contestID))
}

// Helper methods
func (s *LeaderboardService) generateCacheKey(contestID int64) string {
	return fmt.Sprintf("leaderboard:contest:%d", contestID)
}

func (s *LeaderboardService) generateUpdateID(contestID int64) string {
	return fmt.Sprintf("%d_%d", contestID, time.Now().UnixNano())
}

func (s *LeaderboardService) generateSnapshotID(contestID int64) string {
	return fmt.Sprintf("snapshot_%d_%d", contestID, time.Now().UnixNano())
}

func (s *LeaderboardService) determineUpdateType(changes []models.LeaderboardRankChange) string {
	if len(changes) == 0 {
		return "no_change"
	}
	
	hasRankChanges := false
	hasPointsChanges := false
	hasNewEntries := false
	
	for _, change := range changes {
		if change.RankChange != 0 {
			hasRankChanges = true
		}
		if change.PointsChange != 0 {
			hasPointsChanges = true
		}
		if change.PreviousRank == 0 {
			hasNewEntries = true
		}
	}
	
	if hasNewEntries {
		return "new_entry"
	} else if hasRankChanges {
		return "rank_change"
	} else if hasPointsChanges {
		return "points_update"
	}
	
	return "full_refresh"
}

func (s *LeaderboardService) extractAffectedUserIDs(changes []models.LeaderboardRankChange) []int64 {
	userIDs := make([]int64, len(changes))
	for i, change := range changes {
		userIDs[i] = change.UserID
	}
	return userIDs
}