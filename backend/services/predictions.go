package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	"time"
)

type PlayerPredictionService struct {
	db *sql.DB
}

func NewPlayerPredictionService(db *sql.DB) *PlayerPredictionService {
	return &PlayerPredictionService{db: db}
}

// Generate predictions for all players in a match
func (s *PlayerPredictionService) GenerateMatchPredictions(matchID int64) error {
	// Get all players in the match
	players, err := s.getMatchPlayers(matchID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, playerID := range players {
		prediction, err := s.generatePlayerPrediction(playerID, matchID)
		if err != nil {
			continue // Skip this player if prediction fails
		}

		// Store prediction
		err = s.storePrediction(tx, prediction)
		if err != nil {
			continue
		}
	}

	return tx.Commit()
}

func (s *PlayerPredictionService) generatePlayerPrediction(playerID, matchID int64) (*models.PlayerPrediction, error) {
	factors := &models.PredictionFactors{}

	// Recent form (last 5 matches)
	recentForm, err := s.calculateRecentForm(playerID)
	if err != nil {
		recentForm = 5.0 // Default
	}
	factors.RecentForm = recentForm

	// Head-to-head record against opponents
	headToHead, err := s.calculateHeadToHeadRecord(playerID, matchID)
	if err != nil {
		headToHead = 5.0 // Default
	}
	factors.HeadToHeadRecord = headToHead

	// Team strength
	teamStrength, err := s.calculateTeamStrength(playerID, matchID)
	if err != nil {
		teamStrength = 5.0 // Default
	}
	factors.TeamStrength = teamStrength

	// Map performance
	mapPerformance, err := s.calculateMapPerformance(playerID, matchID)
	if err != nil {
		mapPerformance = 5.0 // Default
	}
	factors.MapPerformance = mapPerformance

	// Team morale (based on recent results)
	teamMorale, err := s.calculateTeamMorale(playerID)
	if err != nil {
		teamMorale = 5.0 // Default
	}
	factors.TeamMorale = teamMorale

	// Calculate predicted points using weighted formula
	predictedPoints := s.calculatePredictedPoints(factors)

	// Calculate confidence score
	confidenceScore := s.calculateConfidenceScore(factors)

	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, err
	}

	prediction := &models.PlayerPrediction{
		PlayerID:        playerID,
		MatchID:         matchID,
		PredictionDate:  time.Now(),
		PredictedPoints: predictedPoints,
		ConfidenceScore: confidenceScore,
		Factors:         factorsJSON,
		ModelVersion:    "1.0",
	}

	return prediction, nil
}

func (s *PlayerPredictionService) calculateRecentForm(playerID int64) (float64, error) {
	// Get last 5 match performances
	query := `
		SELECT tp.points_earned
		FROM team_players tp
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		WHERE tp.player_id = $1 AND m.status = 'completed'
		ORDER BY m.scheduled_at DESC
		LIMIT 5
	`

	rows, err := s.db.Query(query, playerID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var points []float64
	for rows.Next() {
		var point float64
		if err := rows.Scan(&point); err != nil {
			continue
		}
		points = append(points, point)
	}

	if len(points) == 0 {
		return 5.0, nil // Default neutral form
	}

	// Calculate weighted average (more recent matches have higher weight)
	totalWeight := 0.0
	weightedSum := 0.0
	for i, point := range points {
		weight := float64(len(points) - i) // Higher weight for recent matches
		weightedSum += point * weight
		totalWeight += weight
	}

	return weightedSum / totalWeight, nil
}

func (s *PlayerPredictionService) calculateHeadToHeadRecord(playerID, matchID int64) (float64, error) {
	// Get opponent teams for this match
	opponentTeams, err := s.getOpponentTeams(playerID, matchID)
	if err != nil {
		return 5.0, nil
	}

	if len(opponentTeams) == 0 {
		return 5.0, nil
	}

	// Calculate performance against these opponents historically
	query := `
		SELECT AVG(tp.points_earned) as avg_points
		FROM team_players tp
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		JOIN match_participants mp ON m.id = mp.match_id
		WHERE tp.player_id = $1 AND mp.team_id = ANY($2) AND m.status = 'completed'
	`

	var avgPoints sql.NullFloat64
	err = s.db.QueryRow(query, playerID, pq.Array(opponentTeams)).Scan(&avgPoints)
	if err != nil || !avgPoints.Valid {
		return 5.0, nil
	}

	// Normalize to 0-10 scale
	return math.Min(10.0, math.Max(0.0, avgPoints.Float64/2.0)), nil
}

func (s *PlayerPredictionService) calculateTeamStrength(playerID, matchID int64) (float64, error) {
	// Get player's team average rating
	query := `
		SELECT AVG(p.form_score) as team_strength
		FROM players p
		JOIN teams t ON p.team_id = t.id
		WHERE t.id = (SELECT team_id FROM players WHERE id = $1)
		AND p.is_playing = true
	`

	var teamStrength sql.NullFloat64
	err := s.db.QueryRow(query, playerID).Scan(&teamStrength)
	if err != nil || !teamStrength.Valid {
		return 5.0, nil
	}

	return teamStrength.Float64, nil
}

func (s *PlayerPredictionService) calculateMapPerformance(playerID, matchID int64) (float64, error) {
	// Get the map for this match
	var mapName sql.NullString
	err := s.db.QueryRow("SELECT map FROM matches WHERE id = $1", matchID).Scan(&mapName)
	if err != nil || !mapName.Valid {
		return 5.0, nil // Default if no map specified
	}

	// Get player's historical performance on this map
	query := `
		SELECT AVG(tp.points_earned) as map_performance
		FROM team_players tp
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		WHERE tp.player_id = $1 AND m.map = $2 AND m.status = 'completed'
	`

	var mapPerformance sql.NullFloat64
	err = s.db.QueryRow(query, playerID, mapName.String).Scan(&mapPerformance)
	if err != nil || !mapPerformance.Valid {
		return 5.0, nil
	}

	// Normalize to 0-10 scale
	return math.Min(10.0, math.Max(0.0, mapPerformance.Float64/2.0)), nil
}

func (s *PlayerPredictionService) calculateTeamMorale(playerID int64) (float64, error) {
	// Get team's win rate in last 10 matches
	query := `
		SELECT 
			COUNT(*) as total_matches,
			SUM(CASE WHEN m.winner_team_id = t.id THEN 1 ELSE 0 END) as wins
		FROM matches m
		JOIN match_participants mp ON m.id = mp.match_id
		JOIN teams t ON mp.team_id = t.id
		WHERE t.id = (SELECT team_id FROM players WHERE id = $1)
		AND m.status = 'completed'
		ORDER BY m.scheduled_at DESC
		LIMIT 10
	`

	var totalMatches, wins int
	err := s.db.QueryRow(query, playerID).Scan(&totalMatches, &wins)
	if err != nil || totalMatches == 0 {
		return 5.0, nil
	}

	winRate := float64(wins) / float64(totalMatches)
	return winRate * 10.0, nil // Scale to 0-10
}

func (s *PlayerPredictionService) calculatePredictedPoints(factors *models.PredictionFactors) float64 {
	// Weighted formula for prediction
	weights := map[string]float64{
		"recent_form":      0.3,
		"head_to_head":     0.2,
		"team_strength":    0.2,
		"map_performance":  0.15,
		"team_morale":      0.15,
	}

	weightedSum := factors.RecentForm*weights["recent_form"] +
		factors.HeadToHeadRecord*weights["head_to_head"] +
		factors.TeamStrength*weights["team_strength"] +
		factors.MapPerformance*weights["map_performance"] +
		factors.TeamMorale*weights["team_morale"]

	// Scale to typical fantasy points range (0-50)
	return math.Max(0.0, math.Min(50.0, weightedSum*2.5))
}

func (s *PlayerPredictionService) calculateConfidenceScore(factors *models.PredictionFactors) float64 {
	// Confidence based on data availability and consistency
	scores := []float64{
		factors.RecentForm,
		factors.HeadToHeadRecord,
		factors.TeamStrength,
		factors.MapPerformance,
		factors.TeamMorale,
	}

	// Calculate standard deviation
	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(scores))
	stdDev := math.Sqrt(variance)

	// Lower standard deviation = higher confidence
	confidence := math.Max(0.1, math.Min(1.0, 1.0-(stdDev/5.0)))
	return confidence
}

func (s *PlayerPredictionService) UpdatePredictionAccuracy(matchID int64) error {
	// Get all predictions for this match
	predictions, err := s.getMatchPredictions(matchID)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, prediction := range predictions {
		// Get actual points scored
		actualPoints, err := s.getActualPoints(prediction.PlayerID, matchID)
		if err != nil {
			continue
		}

		// Calculate accuracy
		accuracy := s.calculateAccuracy(prediction.PredictedPoints, actualPoints)

		// Update prediction
		_, err = tx.Exec(`
			UPDATE player_predictions 
			SET actual_points = $1, accuracy_score = $2, updated_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, actualPoints, accuracy, prediction.ID)
		if err != nil {
			continue
		}
	}

	return tx.Commit()
}

func (s *PlayerPredictionService) calculateAccuracy(predicted, actual float64) float64 {
	if predicted == 0 {
		return 0.0
	}
	
	// Calculate percentage error
	error := math.Abs(predicted-actual) / predicted
	
	// Convert to accuracy (1 - error), capped at 0
	return math.Max(0.0, 1.0-error)
}

func (s *PlayerPredictionService) GetPlayerPredictions(matchID int64) ([]models.PlayerPrediction, error) {
	query := `
		SELECT pp.id, pp.player_id, pp.match_id, pp.prediction_date, pp.predicted_points,
			pp.confidence_score, pp.factors, pp.actual_points, pp.accuracy_score, pp.model_version,
			pp.created_at, pp.updated_at, p.name as player_name, t.name as team_name,
			m.name as match_name
		FROM player_predictions pp
		JOIN players p ON pp.player_id = p.id
		JOIN teams t ON p.team_id = t.id
		LEFT JOIN matches m ON pp.match_id = m.id
		WHERE pp.match_id = $1
		ORDER BY pp.confidence_score DESC, pp.predicted_points DESC
	`

	rows, err := s.db.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []models.PlayerPrediction
	for rows.Next() {
		var p models.PlayerPrediction
		err := rows.Scan(&p.ID, &p.PlayerID, &p.MatchID, &p.PredictionDate, &p.PredictedPoints,
			&p.ConfidenceScore, &p.Factors, &p.ActualPoints, &p.AccuracyScore, &p.ModelVersion,
			&p.CreatedAt, &p.UpdatedAt, &p.PlayerName, &p.TeamName, &p.MatchName)
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, p)
	}

	return predictions, nil
}

func (s *PlayerPredictionService) GetPredictionAnalytics(days int) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_predictions,
			AVG(accuracy_score) as avg_accuracy,
			AVG(confidence_score) as avg_confidence,
			COUNT(CASE WHEN accuracy_score >= 0.8 THEN 1 END) as high_accuracy_count,
			model_version
		FROM player_predictions
		WHERE created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND actual_points IS NOT NULL
		GROUP BY model_version
		ORDER BY model_version
	`

	rows, err := s.db.Query(fmt.Sprintf(query, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	analytics := map[string]interface{}{
		"models": []map[string]interface{}{},
		"overall": map[string]interface{}{},
	}

	var models []map[string]interface{}
	totalPreds, totalAcc, totalConf, totalHighAcc := 0, 0.0, 0.0, 0

	for rows.Next() {
		var count, highAccCount int
		var avgAcc, avgConf float64
		var version string

		err := rows.Scan(&count, &avgAcc, &avgConf, &highAccCount, &version)
		if err != nil {
			continue
		}

		models = append(models, map[string]interface{}{
			"version":           version,
			"total_predictions": count,
			"avg_accuracy":      avgAcc,
			"avg_confidence":    avgConf,
			"high_accuracy_rate": float64(highAccCount) / float64(count),
		})

		totalPreds += count
		totalAcc += avgAcc * float64(count)
		totalConf += avgConf * float64(count)
		totalHighAcc += highAccCount
	}

	analytics["models"] = models

	if totalPreds > 0 {
		analytics["overall"] = map[string]interface{}{
			"total_predictions":  totalPreds,
			"avg_accuracy":       totalAcc / float64(totalPreds),
			"avg_confidence":     totalConf / float64(totalPreds),
			"high_accuracy_rate": float64(totalHighAcc) / float64(totalPreds),
		}
	}

	return analytics, nil
}

// Helper methods
func (s *PlayerPredictionService) getMatchPlayers(matchID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT p.id
		FROM players p
		JOIN teams t ON p.team_id = t.id
		JOIN match_participants mp ON t.id = mp.team_id
		WHERE mp.match_id = $1 AND p.is_playing = true
	`

	rows, err := s.db.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playerIDs []int64
	for rows.Next() {
		var playerID int64
		if err := rows.Scan(&playerID); err != nil {
			continue
		}
		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs, nil
}

func (s *PlayerPredictionService) getOpponentTeams(playerID, matchID int64) ([]int64, error) {
	query := `
		SELECT mp.team_id
		FROM match_participants mp
		WHERE mp.match_id = $1
		AND mp.team_id != (SELECT team_id FROM players WHERE id = $2)
	`

	rows, err := s.db.Query(query, matchID, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamIDs []int64
	for rows.Next() {
		var teamID int64
		if err := rows.Scan(&teamID); err != nil {
			continue
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

func (s *PlayerPredictionService) storePrediction(tx *sql.Tx, prediction *models.PlayerPrediction) error {
	_, err := tx.Exec(`
		INSERT INTO player_predictions (player_id, match_id, prediction_date, predicted_points,
			confidence_score, factors, model_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, prediction.PlayerID, prediction.MatchID, prediction.PredictionDate, prediction.PredictedPoints,
		prediction.ConfidenceScore, prediction.Factors, prediction.ModelVersion)
	return err
}

func (s *PlayerPredictionService) getMatchPredictions(matchID int64) ([]models.PlayerPrediction, error) {
	query := `
		SELECT id, player_id, match_id, predicted_points, confidence_score
		FROM player_predictions
		WHERE match_id = $1
	`

	rows, err := s.db.Query(query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []models.PlayerPrediction
	for rows.Next() {
		var p models.PlayerPrediction
		err := rows.Scan(&p.ID, &p.PlayerID, &p.MatchID, &p.PredictedPoints, &p.ConfidenceScore)
		if err != nil {
			continue
		}
		predictions = append(predictions, p)
	}

	return predictions, nil
}

func (s *PlayerPredictionService) getActualPoints(playerID, matchID int64) (float64, error) {
	query := `
		SELECT COALESCE(AVG(tp.points_earned), 0) as actual_points
		FROM team_players tp
		JOIN user_teams ut ON tp.team_id = ut.id
		WHERE tp.player_id = $1 AND ut.match_id = $2
	`

	var actualPoints float64
	err := s.db.QueryRow(query, playerID, matchID).Scan(&actualPoints)
	return actualPoints, err
}

