package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
)

type TournamentService struct {
	db *sql.DB
}

// TournamentBracket represents a complete bracket structure
type TournamentBracket struct {
	TournamentID   int64                     `json:"tournament_id"`
	TournamentName string                    `json:"tournament_name"`
	Stages         []TournamentStageWithData `json:"stages"`
	Teams          []models.Team             `json:"teams"`
	Status         string                    `json:"status"`
	CurrentStage   *int64                    `json:"current_stage_id"`
}

// TournamentStageWithData includes matches and bracket data
type TournamentStageWithData struct {
	models.TournamentStage
	Matches          []MatchWithParticipants `json:"matches"`
	BracketStructure *BracketStructure       `json:"bracket_structure,omitempty"`
}

// BracketStructure represents the bracket tree
type BracketStructure struct {
	Type      string         `json:"type"` // "single_elimination", "double_elimination", "round_robin"
	Rounds    []BracketRound `json:"rounds"`
	Finals    *BracketMatch  `json:"finals,omitempty"`
	ThirdPlace *BracketMatch `json:"third_place,omitempty"`
}

// BracketRound represents a round in the bracket
type BracketRound struct {
	Round   int             `json:"round"`
	Name    string          `json:"name"`
	Matches []BracketMatch  `json:"matches"`
}

// BracketMatch represents a match in the bracket
type BracketMatch struct {
	MatchID     int64  `json:"match_id"`
	Position    int    `json:"position"`
	Team1       *BracketTeam `json:"team1"`
	Team2       *BracketTeam `json:"team2"`
	Winner      *BracketTeam `json:"winner,omitempty"`
	Status      string `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// BracketTeam represents a team in bracket context
type BracketTeam struct {
	TeamID   int64  `json:"team_id"`
	Name     string `json:"name"`
	Seed     int    `json:"seed"`
	LogoURL  *string `json:"logo_url,omitempty"`
}

// MatchWithParticipants includes participant details
type MatchWithParticipants struct {
	models.Match
	Participants     []models.MatchParticipant `json:"participants"`
	LiveStreamURL    *string                   `json:"live_stream_url,omitempty"`
	LiveStreamActive bool                      `json:"live_stream_active"`
}

// NewTournamentService creates a new tournament service instance
func NewTournamentService(db *sql.DB) *TournamentService {
	return &TournamentService{
		db: db,
	}
}

// GetTournamentBracket generates and returns the complete bracket structure
func (s *TournamentService) GetTournamentBracket(tournamentID int64) (*TournamentBracket, error) {
	logger.Info(fmt.Sprintf("Generating tournament bracket for tournament %d", tournamentID))

	// Get tournament details
	var tournament models.Tournament
	err := s.db.QueryRow(`
		SELECT id, name, game_id, description, start_date, end_date, 
			   prize_pool, total_teams, status, is_featured, logo_url, banner_url, created_at
		FROM tournaments WHERE id = $1`, tournamentID).Scan(
		&tournament.ID, &tournament.Name, &tournament.GameID, &tournament.Description,
		&tournament.StartDate, &tournament.EndDate, &tournament.PrizePool,
		&tournament.TotalTeams, &tournament.Status, &tournament.IsFeatured,
		&tournament.LogoURL, &tournament.BannerURL, &tournament.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	// Get tournament stages with matches
	stages, err := s.getTournamentStages(tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament stages: %w", err)
	}

	// Get participating teams
	teams, err := s.getTournamentTeams(tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament teams: %w", err)
	}

	// Determine current stage
	currentStageID := s.getCurrentStageID(stages)

	bracket := &TournamentBracket{
		TournamentID:   tournament.ID,
		TournamentName: tournament.Name,
		Stages:         stages,
		Teams:          teams,
		Status:         tournament.Status,
		CurrentStage:   currentStageID,
	}

	return bracket, nil
}

// getTournamentStages retrieves all stages for a tournament with their matches
func (s *TournamentService) getTournamentStages(tournamentID int64) ([]TournamentStageWithData, error) {
	// Get stages
	stageRows, err := s.db.Query(`
		SELECT id, tournament_id, name, stage_order, stage_type, start_date, end_date, max_teams, rules
		FROM tournament_stages 
		WHERE tournament_id = $1 
		ORDER BY stage_order`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer stageRows.Close()

	var stages []TournamentStageWithData
	for stageRows.Next() {
		var stage TournamentStageWithData
		err := stageRows.Scan(
			&stage.ID, &stage.TournamentID, &stage.Name, &stage.StageOrder,
			&stage.StageType, &stage.StartDate, &stage.EndDate,
			&stage.MaxTeams, &stage.Rules,
		)
		if err != nil {
			continue
		}

		// Get matches for this stage
		matches, err := s.getStageMatches(stage.ID)
		if err != nil {
			log.Printf("Failed to get matches for stage %d: %v", stage.ID, err)
			matches = []MatchWithParticipants{}
		}
		stage.Matches = matches

		// Generate bracket structure for elimination stages
		if stage.StageType == "single_elimination" || stage.StageType == "double_elimination" {
			bracketStructure, err := s.generateBracketStructure(stage.ID, stage.StageType, matches)
			if err != nil {
				log.Printf("Failed to generate bracket structure for stage %d: %v", stage.ID, err)
			} else {
				stage.BracketStructure = bracketStructure
			}
		}

		stages = append(stages, stage)
	}

	return stages, nil
}

// getStageMatches retrieves all matches for a tournament stage
func (s *TournamentService) getStageMatches(stageID int64) ([]MatchWithParticipants, error) {
	matchRows, err := s.db.Query(`
		SELECT m.id, m.tournament_id, m.stage_id, m.game_id, m.name, m.scheduled_at,
			   m.lock_time, m.status, m.match_type, m.map, m.best_of, m.result,
			   m.winner_team_id, m.created_at, m.updated_at
		FROM matches m
		WHERE m.stage_id = $1
		ORDER BY m.scheduled_at`, stageID)
	if err != nil {
		return nil, err
	}
	defer matchRows.Close()

	var matches []MatchWithParticipants
	for matchRows.Next() {
		var match MatchWithParticipants
		err := matchRows.Scan(
			&match.ID, &match.TournamentID, &match.StageID, &match.GameID,
			&match.Name, &match.ScheduledAt, &match.LockTime, &match.Status,
			&match.MatchType, &match.Map, &match.BestOf, &match.Result,
			&match.WinnerTeamID, &match.CreatedAt, &match.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Get match participants
		participants, err := s.getMatchParticipants(match.ID)
		if err != nil {
			log.Printf("Failed to get participants for match %d: %v", match.ID, err)
			participants = []models.MatchParticipant{}
		}
		match.Participants = participants

		// Check for live stream
		liveStreamURL, isActive := s.getMatchLiveStream(match.ID)
		match.LiveStreamURL = liveStreamURL
		match.LiveStreamActive = isActive

		matches = append(matches, match)
	}

	return matches, nil
}

// getMatchParticipants retrieves all participants for a match
func (s *TournamentService) getMatchParticipants(matchID int64) ([]models.MatchParticipant, error) {
	rows, err := s.db.Query(`
		SELECT id, match_id, team_id, seed, final_position, team_score, 
			   points_earned, eliminated_at, joined_at
		FROM match_participants 
		WHERE match_id = $1 
		ORDER BY seed NULLS LAST, joined_at`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.MatchParticipant
	for rows.Next() {
		var participant models.MatchParticipant
		err := rows.Scan(
			&participant.ID, &participant.MatchID, &participant.TeamID,
			&participant.Seed, &participant.FinalPosition, &participant.TeamScore,
			&participant.PointsEarned, &participant.EliminatedAt, &participant.JoinedAt,
		)
		if err != nil {
			continue
		}
		participants = append(participants, participant)
	}

	return participants, nil
}

// getTournamentTeams retrieves all teams participating in a tournament
func (s *TournamentService) getTournamentTeams(tournamentID int64) ([]models.Team, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT t.id, t.name, t.short_name, t.logo_url, t.region, t.is_active, t.social_links, t.created_at
		FROM teams t
		JOIN match_participants mp ON t.id = mp.team_id
		JOIN matches m ON mp.match_id = m.id
		WHERE m.tournament_id = $1
		ORDER BY t.name`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID, &team.Name, &team.ShortName, &team.LogoURL,
			&team.Region, &team.IsActive, &team.SocialLinks, &team.CreatedAt,
		)
		if err != nil {
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// getCurrentStageID determines the current active stage
func (s *TournamentService) getCurrentStageID(stages []TournamentStageWithData) *int64 {
	now := time.Now()
	
	for _, stage := range stages {
		// Check if stage is currently active
		if stage.StartDate != nil && stage.EndDate != nil {
			if now.After(*stage.StartDate) && now.Before(*stage.EndDate) {
				return &stage.ID
			}
		}
		
		// Check if any matches in this stage are live or upcoming
		for _, match := range stage.Matches {
			if match.Status == "live" || match.Status == "upcoming" {
				return &stage.ID
			}
		}
	}
	
	return nil
}

// generateBracketStructure creates bracket structure for elimination tournaments
func (s *TournamentService) generateBracketStructure(stageID int64, stageType string, matches []MatchWithParticipants) (*BracketStructure, error) {
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for stage")
	}

	bracket := &BracketStructure{
		Type:   stageType,
		Rounds: []BracketRound{},
	}

	// Group matches by round (determine by match count and structure)
	rounds := s.groupMatchesIntoRounds(matches)
	
	// Create bracket rounds
	for roundNum, roundMatches := range rounds {
		bracketRound := BracketRound{
			Round:   roundNum + 1,
			Name:    s.getRoundName(roundNum+1, len(rounds), stageType),
			Matches: []BracketMatch{},
		}

		// Convert matches to bracket format
		for i, match := range roundMatches {
			bracketMatch := s.convertToBracketMatch(match, i)
			bracketRound.Matches = append(bracketRound.Matches, bracketMatch)
		}

		bracket.Rounds = append(bracket.Rounds, bracketRound)
	}

	// Set finals reference
	if len(bracket.Rounds) > 0 {
		finalRound := &bracket.Rounds[len(bracket.Rounds)-1]
		if len(finalRound.Matches) > 0 {
			bracket.Finals = &finalRound.Matches[0]
		}
	}

	return bracket, nil
}

// groupMatchesIntoRounds groups matches into tournament rounds
func (s *TournamentService) groupMatchesIntoRounds(matches []MatchWithParticipants) [][]MatchWithParticipants {
	// Sort matches by scheduled time
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ScheduledAt.Before(matches[j].ScheduledAt)
	})

	// Simple grouping logic - can be enhanced based on specific tournament structure
	totalMatches := len(matches)
	if totalMatches <= 1 {
		return [][]MatchWithParticipants{matches}
	}

	// Calculate expected rounds for single elimination
	expectedRounds := int(math.Log2(float64(totalMatches*2))) + 1
	
	rounds := make([][]MatchWithParticipants, expectedRounds)
	matchesPerRound := totalMatches / expectedRounds
	
	for i, match := range matches {
		roundIndex := i / (matchesPerRound + 1)
		if roundIndex >= expectedRounds {
			roundIndex = expectedRounds - 1
		}
		rounds[roundIndex] = append(rounds[roundIndex], match)
	}

	// Remove empty rounds
	var filteredRounds [][]MatchWithParticipants
	for _, round := range rounds {
		if len(round) > 0 {
			filteredRounds = append(filteredRounds, round)
		}
	}

	return filteredRounds
}

// getRoundName generates appropriate round names
func (s *TournamentService) getRoundName(roundNum, totalRounds int, stageType string) string {
	if stageType == "single_elimination" {
		switch totalRounds - roundNum {
		case 0:
			return "Finals"
		case 1:
			return "Semi-Finals"
		case 2:
			return "Quarter-Finals"
		case 3:
			return "Round of 16"
		case 4:
			return "Round of 32"
		default:
			return fmt.Sprintf("Round %d", roundNum)
		}
	}
	return fmt.Sprintf("Round %d", roundNum)
}

// convertToBracketMatch converts a match to bracket format
func (s *TournamentService) convertToBracketMatch(match MatchWithParticipants, position int) BracketMatch {
	bracketMatch := BracketMatch{
		MatchID:     match.ID,
		Position:    position,
		Status:      match.Status,
		ScheduledAt: &match.ScheduledAt,
	}

	// Set teams from participants
	if len(match.Participants) >= 2 {
		bracketMatch.Team1 = &BracketTeam{
			TeamID: match.Participants[0].TeamID,
			Seed:   s.getSeedValue(match.Participants[0].Seed),
		}
		bracketMatch.Team2 = &BracketTeam{
			TeamID: match.Participants[1].TeamID,
			Seed:   s.getSeedValue(match.Participants[1].Seed),
		}

		// Set team names (would need additional query in real implementation)
		s.populateTeamNames(bracketMatch.Team1, bracketMatch.Team2)
	}

	// Set winner if match is completed
	if match.Status == "completed" && match.WinnerTeamID != nil {
		if bracketMatch.Team1 != nil && bracketMatch.Team1.TeamID == *match.WinnerTeamID {
			bracketMatch.Winner = bracketMatch.Team1
		} else if bracketMatch.Team2 != nil && bracketMatch.Team2.TeamID == *match.WinnerTeamID {
			bracketMatch.Winner = bracketMatch.Team2
		}
	}

	return bracketMatch
}

// getSeedValue safely gets seed value
func (s *TournamentService) getSeedValue(seed *int) int {
	if seed != nil {
		return *seed
	}
	return 0
}

// populateTeamNames fills team names from database
func (s *TournamentService) populateTeamNames(team1, team2 *BracketTeam) {
	if team1 != nil {
		s.db.QueryRow("SELECT name, logo_url FROM teams WHERE id = $1", team1.TeamID).Scan(&team1.Name, &team1.LogoURL)
	}
	if team2 != nil {
		s.db.QueryRow("SELECT name, logo_url FROM teams WHERE id = $1", team2.TeamID).Scan(&team2.Name, &team2.LogoURL)
	}
}

// getMatchLiveStream checks if match has live stream
func (s *TournamentService) getMatchLiveStream(matchID int64) (*string, bool) {
	var streamURL sql.NullString
	var isActive bool
	
	err := s.db.QueryRow(`
		SELECT stream_url, is_stream_active 
		FROM match_streams 
		WHERE match_id = $1`, matchID).Scan(&streamURL, &isActive)
	
	if err != nil || !streamURL.Valid {
		return nil, false
	}
	
	return &streamURL.String, isActive
}

// CreateTournamentStage creates a new tournament stage
func (s *TournamentService) CreateTournamentStage(tournamentID int64, stage models.TournamentStage) (*models.TournamentStage, error) {
	var newStage models.TournamentStage
	
	err := s.db.QueryRow(`
		INSERT INTO tournament_stages (tournament_id, name, stage_order, stage_type, start_date, end_date, max_teams, rules)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tournament_id, name, stage_order, stage_type, start_date, end_date, max_teams, rules`,
		tournamentID, stage.Name, stage.StageOrder, stage.StageType, 
		stage.StartDate, stage.EndDate, stage.MaxTeams, stage.Rules).Scan(
		&newStage.ID, &newStage.TournamentID, &newStage.Name, &newStage.StageOrder,
		&newStage.StageType, &newStage.StartDate, &newStage.EndDate, &newStage.MaxTeams, &newStage.Rules,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create tournament stage: %w", err)
	}
	
	return &newStage, nil
}

// AdvanceToNextStage advances teams from one stage to the next
func (s *TournamentService) AdvanceToNextStage(currentStageID int64) error {
	// Get current stage details
	var currentStage models.TournamentStage
	err := s.db.QueryRow(`
		SELECT tournament_id, stage_order, stage_type 
		FROM tournament_stages 
		WHERE id = $1`, currentStageID).Scan(&currentStage.TournamentID, &currentStage.StageOrder, &currentStage.StageType)
	if err != nil {
		return fmt.Errorf("failed to get current stage: %w", err)
	}

	// Get next stage
	var nextStageID int64
	err = s.db.QueryRow(`
		SELECT id FROM tournament_stages 
		WHERE tournament_id = $1 AND stage_order = $2`,
		currentStage.TournamentID, currentStage.StageOrder+1).Scan(&nextStageID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no next stage found")
	} else if err != nil {
		return fmt.Errorf("failed to find next stage: %w", err)
	}

	// Get winners from current stage
	winners, err := s.getStageWinners(currentStageID)
	if err != nil {
		return fmt.Errorf("failed to get stage winners: %w", err)
	}

	// Create matches for next stage
	err = s.createNextStageMatches(nextStageID, winners)
	if err != nil {
		return fmt.Errorf("failed to create next stage matches: %w", err)
	}

	logger.Info(fmt.Sprintf("Successfully advanced %d teams to stage %d", len(winners), nextStageID))
	return nil
}

// getStageWinners gets all winning teams from a completed stage
func (s *TournamentService) getStageWinners(stageID int64) ([]int64, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT winner_team_id 
		FROM matches 
		WHERE stage_id = $1 AND status = 'completed' AND winner_team_id IS NOT NULL
		ORDER BY winner_team_id`, stageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var winners []int64
	for rows.Next() {
		var teamID int64
		if err := rows.Scan(&teamID); err == nil {
			winners = append(winners, teamID)
		}
	}

	return winners, nil
}

// createNextStageMatches creates matches for the next tournament stage
func (s *TournamentService) createNextStageMatches(stageID int64, teams []int64) error {
	if len(teams) < 2 {
		return fmt.Errorf("insufficient teams for next stage matches")
	}

	// Simple pairing logic - pair teams sequentially
	matchCount := 0
	for i := 0; i < len(teams); i += 2 {
		if i+1 >= len(teams) {
			break // Odd number of teams, last team gets bye
		}

		// Create match
		matchID, err := s.createMatch(stageID, teams[i], teams[i+1], matchCount+1)
		if err != nil {
			return fmt.Errorf("failed to create match %d: %w", matchCount+1, err)
		}

		logger.Info(fmt.Sprintf("Created match %d for stage %d", matchID, stageID))
		matchCount++
	}

	return nil
}

// createMatch creates a new match between two teams
func (s *TournamentService) createMatch(stageID int64, team1ID, team2ID int64, matchNumber int) (int64, error) {
	// Get tournament and game info
	var tournamentID int64
	var gameID int
	err := s.db.QueryRow(`
		SELECT tournament_id, 
			   (SELECT game_id FROM tournaments WHERE id = ts.tournament_id) as game_id
		FROM tournament_stages ts 
		WHERE id = $1`, stageID).Scan(&tournamentID, &gameID)
	if err != nil {
		return 0, err
	}

	// Create match
	scheduledAt := time.Now().Add(24 * time.Hour) // Schedule for tomorrow
	lockTime := scheduledAt.Add(-10 * time.Minute)

	var matchID int64
	err = s.db.QueryRow(`
		INSERT INTO matches (tournament_id, stage_id, game_id, name, scheduled_at, lock_time, status, match_type, best_of)
		VALUES ($1, $2, $3, $4, $5, $6, 'upcoming', 'elimination', 1)
		RETURNING id`,
		tournamentID, stageID, gameID, fmt.Sprintf("Match %d", matchNumber),
		scheduledAt, lockTime).Scan(&matchID)
	if err != nil {
		return 0, err
	}

	// Add participants
	_, err = s.db.Exec(`
		INSERT INTO match_participants (match_id, team_id, seed, joined_at)
		VALUES ($1, $2, 1, NOW()), ($1, $3, 2, NOW())`,
		matchID, team1ID, team2ID)
	if err != nil {
		return 0, fmt.Errorf("failed to add participants: %w", err)
	}

	return matchID, nil
}