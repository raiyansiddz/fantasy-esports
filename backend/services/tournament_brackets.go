package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	"time"
)

type TournamentBracketService struct {
	db *sql.DB
}

func NewTournamentBracketService(db *sql.DB) *TournamentBracketService {
	return &TournamentBracketService{db: db}
}

// Bracket creation and management
func (s *TournamentBracketService) CreateBracket(req models.CreateBracketRequest) (*models.TournamentBracket, error) {
	// Get tournament and stage info
	var tournamentName, stageName string
	var maxTeams int
	
	err := s.db.QueryRow(`
		SELECT t.name, ts.name, COALESCE(ts.max_teams, 16)
		FROM tournaments t
		JOIN tournament_stages ts ON t.id = ts.tournament_id
		WHERE t.id = $1 AND ts.id = $2
	`, req.TournamentID, req.StageID).Scan(&tournamentName, &stageName, &maxTeams)
	
	if err != nil {
		return nil, fmt.Errorf("tournament or stage not found: %w", err)
	}

	// Get participating teams
	teams, err := s.getStageTeams(req.StageID)
	if err != nil {
		return nil, err
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("no teams found for stage")
	}

	// Generate bracket data based on type
	var bracketData interface{}
	var totalRounds int

	switch req.BracketType {
	case "single_elimination":
		bracketData, totalRounds = s.generateSingleEliminationBracket(teams)
	case "double_elimination":
		bracketData, totalRounds = s.generateDoubleEliminationBracket(teams)
	case "round_robin":
		bracketData, totalRounds = s.generateRoundRobinBracket(teams)
	case "swiss":
		bracketData, totalRounds = s.generateSwissBracket(teams)
	default:
		return nil, fmt.Errorf("unsupported bracket type: %s", req.BracketType)
	}

	bracketDataJSON, err := json.Marshal(bracketData)
	if err != nil {
		return nil, err
	}

	// Create bracket record
	query := `
		INSERT INTO tournament_brackets (tournament_id, stage_id, bracket_type, bracket_data, 
			current_round, total_rounds, auto_advance)
		VALUES ($1, $2, $3, $4, 1, $5, $6)
		RETURNING id, created_at, updated_at
	`

	bracket := &models.TournamentBracket{
		TournamentID: req.TournamentID,
		StageID:      req.StageID,
		BracketType:  req.BracketType,
		BracketData:  bracketDataJSON,
		CurrentRound: 1,
		TotalRounds:  totalRounds,
		Status:       "setup",
		AutoAdvance:  req.AutoAdvance,
	}

	err = s.db.QueryRow(query, bracket.TournamentID, bracket.StageID, bracket.BracketType,
		bracket.BracketData, bracket.TotalRounds, bracket.AutoAdvance).Scan(
		&bracket.ID, &bracket.CreatedAt, &bracket.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return bracket, nil
}

func (s *TournamentBracketService) generateSingleEliminationBracket(teams []BracketTeam) (interface{}, int) {
	numTeams := len(teams)
	
	// Find next power of 2
	bracketSize := 1
	for bracketSize < numTeams {
		bracketSize *= 2
	}
	
	rounds := int(math.Log2(float64(bracketSize)))
	
	bracket := map[string]interface{}{
		"type":        "single_elimination",
		"bracket_size": bracketSize,
		"num_teams":   numTeams,
		"rounds":      []interface{}{},
	}

	// Generate first round matches
	firstRound := []interface{}{}
	
	// Add byes for teams that advance automatically
	_ = bracketSize - numTeams // byeCount unused for now
	
	for i := 0; i < bracketSize/2; i++ {
		match := map[string]interface{}{
			"match_id":     fmt.Sprintf("R1-M%d", i+1),
			"round":        1,
			"position":     i + 1,
			"team1":        nil,
			"team2":        nil,
			"winner":       nil,
			"status":       "pending",
			"next_match":   fmt.Sprintf("R2-M%d", (i/2)+1),
		}

		// Assign teams, accounting for byes
		if i*2 < numTeams {
			match["team1"] = teams[i*2]
		}
		if i*2+1 < numTeams {
			match["team2"] = teams[i*2+1]
		}

		// Handle byes (when only one team is assigned)
		if match["team1"] != nil && match["team2"] == nil {
			match["winner"] = match["team1"]
			match["status"] = "bye"
		}

		firstRound = append(firstRound, match)
	}

	rounds_data := []interface{}{firstRound}

	// Generate subsequent rounds
	for round := 2; round <= rounds; round++ {
		roundMatches := []interface{}{}
		matchesInRound := bracketSize / int(math.Pow(2, float64(round)))
		
		for i := 0; i < matchesInRound; i++ {
			match := map[string]interface{}{
				"match_id":     fmt.Sprintf("R%d-M%d", round, i+1),
				"round":        round,
				"position":     i + 1,
				"team1":        nil,
				"team2":        nil,
				"winner":       nil,
				"status":       "pending",
				"next_match":   nil,
			}

			if round < rounds {
				match["next_match"] = fmt.Sprintf("R%d-M%d", round+1, (i/2)+1)
			}

			roundMatches = append(roundMatches, match)
		}
		
		rounds_data = append(rounds_data, roundMatches)
	}

	bracket["rounds"] = rounds_data
	return bracket, rounds
}

func (s *TournamentBracketService) generateDoubleEliminationBracket(teams []BracketTeam) (interface{}, int) {
	numTeams := len(teams)
	
	// Double elimination needs winners and losers brackets
	bracket := map[string]interface{}{
		"type":           "double_elimination",
		"num_teams":      numTeams,
		"winners_bracket": nil,
		"losers_bracket":  nil,
		"grand_final":    nil,
	}

	// Generate winners bracket (single elimination)
	winnersData, winnersRounds := s.generateSingleEliminationBracket(teams)
	bracket["winners_bracket"] = winnersData

	// Generate losers bracket (more complex structure)
	losersData := s.generateLosersBracket(numTeams)
	bracket["losers_bracket"] = losersData

	// Grand final
	bracket["grand_final"] = map[string]interface{}{
		"match_id": "GF",
		"team1":    nil, // Winner of winners bracket
		"team2":    nil, // Winner of losers bracket
		"winner":   nil,
		"status":   "pending",
	}

	// Total rounds is roughly 2 * log2(teams)
	totalRounds := winnersRounds + winnersRounds - 1 + 1 // winners + losers + grand final

	return bracket, totalRounds
}

func (s *TournamentBracketService) generateRoundRobinBracket(teams []BracketTeam) (interface{}, int) {
	numTeams := len(teams)
	
	// Round robin: every team plays every other team once
	rounds := numTeams - 1
	if numTeams%2 == 1 {
		rounds = numTeams // Add dummy team for odd number
	}

	bracket := map[string]interface{}{
		"type":        "round_robin",
		"num_teams":   numTeams,
		"rounds":      []interface{}{},
		"standings":   []interface{}{},
	}

	// Initialize standings
	standings := []interface{}{}
	for _, team := range teams {
		standings = append(standings, map[string]interface{}{
			"team":    team,
			"played":  0,
			"won":     0,
			"lost":    0,
			"points":  0,
		})
	}
	bracket["standings"] = standings

	// Generate round-robin schedule using circle method
	schedule := s.generateRoundRobinSchedule(teams)
	bracket["rounds"] = schedule

	return bracket, rounds
}

func (s *TournamentBracketService) generateSwissBracket(teams []BracketTeam) (interface{}, int) {
	numTeams := len(teams)
	rounds := int(math.Ceil(math.Log2(float64(numTeams)))) + 1

	bracket := map[string]interface{}{
		"type":        "swiss",
		"num_teams":   numTeams,
		"rounds":      rounds,
		"pairings":    []interface{}{},
		"standings":   []interface{}{},
	}

	// Initialize standings
	standings := []interface{}{}
	for _, team := range teams {
		standings = append(standings, map[string]interface{}{
			"team":           team,
			"points":         0,
			"buchholz":       0,
			"opponents":      []interface{}{},
		})
	}
	bracket["standings"] = standings

	// First round pairings (random or seeded)
	firstRoundPairings := []interface{}{}
	for i := 0; i < len(teams); i += 2 {
		if i+1 < len(teams) {
			firstRoundPairings = append(firstRoundPairings, map[string]interface{}{
				"match_id": fmt.Sprintf("R1-M%d", (i/2)+1),
				"round":    1,
				"team1":    teams[i],
				"team2":    teams[i+1],
				"winner":   nil,
				"status":   "pending",
			})
		}
	}

	pairings := []interface{}{firstRoundPairings}
	bracket["pairings"] = pairings

	return bracket, rounds
}

func (s *TournamentBracketService) AdvanceBracket(bracketID int64, matchResults map[string]interface{}) error {
	// Get current bracket
	bracket, err := s.GetBracket(bracketID)
	if err != nil {
		return err
	}

	// Parse bracket data
	var bracketData map[string]interface{}
	err = json.Unmarshal(bracket.BracketData, &bracketData)
	if err != nil {
		return err
	}

	// Update bracket based on type
	switch bracket.BracketType {
	case "single_elimination":
		err = s.advanceSingleElimination(&bracketData, matchResults)
	case "double_elimination":
		err = s.advanceDoubleElimination(&bracketData, matchResults)
	case "round_robin":
		err = s.advanceRoundRobin(&bracketData, matchResults)
	case "swiss":
		err = s.advanceSwiss(&bracketData, matchResults)
	default:
		return fmt.Errorf("unsupported bracket type for advancement")
	}

	if err != nil {
		return err
	}

	// Update bracket in database
	updatedData, err := json.Marshal(bracketData)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE tournament_brackets 
		SET bracket_data = $1, current_round = current_round + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, updatedData, bracketID)

	return err
}

func (s *TournamentBracketService) GetBracket(bracketID int64) (*models.TournamentBracket, error) {
	query := `
		SELECT id, tournament_id, stage_id, bracket_type, bracket_data, current_round, 
			total_rounds, status, auto_advance, created_at, updated_at
		FROM tournament_brackets
		WHERE id = $1
	`

	bracket := &models.TournamentBracket{}
	err := s.db.QueryRow(query, bracketID).Scan(&bracket.ID, &bracket.TournamentID,
		&bracket.StageID, &bracket.BracketType, &bracket.BracketData, &bracket.CurrentRound,
		&bracket.TotalRounds, &bracket.Status, &bracket.AutoAdvance, &bracket.CreatedAt,
		&bracket.UpdatedAt)

	return bracket, err
}

func (s *TournamentBracketService) GetTournamentBrackets(tournamentID int64) ([]models.TournamentBracket, error) {
	query := `
		SELECT id, tournament_id, stage_id, bracket_type, bracket_data, current_round, 
			total_rounds, status, auto_advance, created_at, updated_at
		FROM tournament_brackets
		WHERE tournament_id = $1
		ORDER BY created_at
	`

	rows, err := s.db.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brackets []models.TournamentBracket
	for rows.Next() {
		var b models.TournamentBracket
		err := rows.Scan(&b.ID, &b.TournamentID, &b.StageID, &b.BracketType,
			&b.BracketData, &b.CurrentRound, &b.TotalRounds, &b.Status,
			&b.AutoAdvance, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			continue
		}
		brackets = append(brackets, b)
	}

	return brackets, nil
}

// Helper methods

func (s *TournamentBracketService) getStageTeams(stageID int64) ([]BracketTeam, error) {
	// This would typically get teams qualified for this stage
	// For now, we'll get all teams from the tournament
	query := `
		SELECT DISTINCT t.id, t.name, COALESCE(mp.seed, 0) as seed
		FROM teams t
		JOIN match_participants mp ON t.id = mp.team_id
		JOIN matches m ON mp.match_id = m.id
		JOIN tournament_stages ts ON m.stage_id = ts.id
		WHERE ts.id = $1 AND t.is_active = true
		ORDER BY seed, t.name
	`

	rows, err := s.db.Query(query, stageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []BracketTeam
	for rows.Next() {
		var team BracketTeam
		err := rows.Scan(&team.ID, &team.Name, &team.Seed)
		if err != nil {
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (s *TournamentBracketService) generateLosersBracket(numTeams int) interface{} {
	// Simplified losers bracket structure
	// In a real implementation, this would be more complex
	return map[string]interface{}{
		"rounds": []interface{}{},
		"type":   "losers",
	}
}

func (s *TournamentBracketService) generateRoundRobinSchedule(teams []BracketTeam) []interface{} {
	numTeams := len(teams)
	schedule := []interface{}{}
	
	// Add dummy team for odd number of teams
	if numTeams%2 == 1 {
		teams = append(teams, BracketTeam{ID: -1, Name: "BYE", Seed: 999})
		numTeams++
	}

	rounds := numTeams - 1
	
	for round := 0; round < rounds; round++ {
		roundMatches := []interface{}{}
		
		for i := 0; i < numTeams/2; i++ {
			team1Idx := i
			team2Idx := numTeams - 1 - i
			
			if team1Idx != team2Idx {
				// Skip BYE matches
				if teams[team1Idx].ID != -1 && teams[team2Idx].ID != -1 {
					roundMatches = append(roundMatches, map[string]interface{}{
						"match_id": fmt.Sprintf("R%d-M%d", round+1, len(roundMatches)+1),
						"round":    round + 1,
						"team1":    teams[team1Idx],
						"team2":    teams[team2Idx],
						"winner":   nil,
						"status":   "pending",
					})
				}
			}
		}
		
		schedule = append(schedule, roundMatches)
		
		// Rotate teams (except first team which stays fixed)
		if numTeams > 2 {
			// Rotate all teams except the first one
			temp := teams[1]
			for i := 1; i < numTeams-1; i++ {
				teams[i] = teams[i+1]
			}
			teams[numTeams-1] = temp
		}
	}

	return schedule
}

func (s *TournamentBracketService) advanceSingleElimination(bracketData *map[string]interface{}, matchResults map[string]interface{}) error {
	// Update match results and advance winners to next round
	rounds, ok := (*bracketData)["rounds"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid bracket data structure")
	}

	// Process each match result
	for matchID, result := range matchResults {
		if resultMap, ok := result.(map[string]interface{}); ok {
			err := s.updateMatchInRounds(rounds, matchID, resultMap)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (s *TournamentBracketService) advanceDoubleElimination(bracketData *map[string]interface{}, matchResults map[string]interface{}) error {
	// More complex logic for double elimination
	return s.advanceSingleElimination(bracketData, matchResults)
}

func (s *TournamentBracketService) advanceRoundRobin(bracketData *map[string]interface{}, matchResults map[string]interface{}) error {
	// Update standings based on results
	standings, ok := (*bracketData)["standings"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid bracket data structure")
	}

	for matchID, result := range matchResults {
		if resultMap, ok := result.(map[string]interface{}); ok {
			err := s.updateRoundRobinStandings(standings, matchID, resultMap)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

func (s *TournamentBracketService) advanceSwiss(bracketData *map[string]interface{}, matchResults map[string]interface{}) error {
	// Update standings and generate next round pairings
	return nil // Simplified implementation
}

func (s *TournamentBracketService) updateMatchInRounds(rounds []interface{}, matchID string, result map[string]interface{}) error {
	// Find and update the match in the rounds structure
	for _, round := range rounds {
		if roundSlice, ok := round.([]interface{}); ok {
			for _, match := range roundSlice {
				if matchMap, ok := match.(map[string]interface{}); ok {
					if matchMap["match_id"] == matchID {
						// Update match result
						matchMap["winner"] = result["winner"]
						matchMap["status"] = "completed"
						
						// Advance winner to next match if specified
						if nextMatch, exists := matchMap["next_match"]; exists && nextMatch != nil {
							s.advanceWinnerToNextMatch(rounds, nextMatch.(string), result["winner"])
						}
						
						return nil
					}
				}
			}
		}
	}
	
	return fmt.Errorf("match not found: %s", matchID)
}

func (s *TournamentBracketService) advanceWinnerToNextMatch(rounds []interface{}, nextMatchID string, winner interface{}) error {
	// Find the next match and add the winner
	for _, round := range rounds {
		if roundSlice, ok := round.([]interface{}); ok {
			for _, match := range roundSlice {
				if matchMap, ok := match.(map[string]interface{}); ok {
					if matchMap["match_id"] == nextMatchID {
						// Add winner to the next match
						if matchMap["team1"] == nil {
							matchMap["team1"] = winner
						} else if matchMap["team2"] == nil {
							matchMap["team2"] = winner
						}
						return nil
					}
				}
			}
		}
	}
	
	return fmt.Errorf("next match not found: %s", nextMatchID)
}

func (s *TournamentBracketService) updateRoundRobinStandings(standings []interface{}, matchID string, result map[string]interface{}) error {
	// Update team standings based on match result
	winner := result["winner"]
	loser := result["loser"]

	for _, standing := range standings {
		if standingMap, ok := standing.(map[string]interface{}); ok {
			team := standingMap["team"]
			
			if team == winner {
				standingMap["won"] = standingMap["won"].(int) + 1
				standingMap["points"] = standingMap["points"].(int) + 3
			} else if team == loser {
				standingMap["lost"] = standingMap["lost"].(int) + 1
			}
			
			standingMap["played"] = standingMap["played"].(int) + 1
		}
	}

	return nil
}