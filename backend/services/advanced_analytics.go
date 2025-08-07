package services

import (
	"database/sql"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	"strings"
	"time"
)

type AdvancedAnalyticsService struct {
	db *sql.DB
}

func NewAdvancedAnalyticsService(db *sql.DB) *AdvancedAnalyticsService {
	return &AdvancedAnalyticsService{db: db}
}

// Advanced Game Analytics
func (s *AdvancedAnalyticsService) CalculateAdvancedGameMetrics(gameID int, days int) (*models.AdvancedGameMetrics, error) {
	// Player Efficiency: Average points per credit spent
	playerEfficiency, err := s.calculatePlayerEfficiency(gameID, days)
	if err != nil {
		return nil, err
	}

	// Team Synergy: Correlation between team composition and performance
	teamSynergy, err := s.calculateTeamSynergy(gameID, days)
	if err != nil {
		return nil, err
	}

	// Strategic Diversity: Variety in team compositions
	strategicDiversity, err := s.calculateStrategicDiversity(gameID, days)
	if err != nil {
		return nil, err
	}

	// Comeback Potential: Teams that perform well despite poor starts
	comebackPotential, err := s.calculateComebackPotential(gameID, days)
	if err != nil {
		return nil, err
	}

	// Clutch Performance: Performance in high-stakes situations
	clutchPerformance, err := s.calculateClutchPerformance(gameID, days)
	if err != nil {
		return nil, err
	}

	// Consistency Index: Standard deviation of performance
	consistencyIndex, err := s.calculateConsistencyIndex(gameID, days)
	if err != nil {
		return nil, err
	}

	// Adaptability Score: How well teams adapt to different opponents
	adaptabilityScore, err := s.calculateAdaptabilityScore(gameID, days)
	if err != nil {
		return nil, err
	}

	metrics := &models.AdvancedGameMetrics{
		PlayerEfficiency:   playerEfficiency,
		TeamSynergy:       teamSynergy,
		StrategicDiversity: strategicDiversity,
		ComebackPotential:  comebackPotential,
		ClutchPerformance:  clutchPerformance,
		ConsistencyIndex:   consistencyIndex,
		AdaptabilityScore:  adaptabilityScore,
	}

	// Store in database
	err = s.storeAdvancedMetrics(gameID, metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *AdvancedAnalyticsService) calculatePlayerEfficiency(gameID, days int) (float64, error) {
	query := `
		SELECT AVG(tp.points_earned / p.credit_value) as efficiency
		FROM team_players tp
		JOIN players p ON tp.player_id = p.id
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND tp.points_earned > 0 AND p.credit_value > 0
	`
	
	var efficiency sql.NullFloat64
	err := s.db.QueryRow(fmt.Sprintf(query, days), gameID).Scan(&efficiency)
	if err != nil || !efficiency.Valid {
		return 0.0, err
	}
	
	return efficiency.Float64, nil
}

func (s *AdvancedAnalyticsService) calculateTeamSynergy(gameID, days int) (float64, error) {
	// Calculate correlation between team diversity and performance
	query := `
		SELECT 
			COUNT(DISTINCT tp.real_team_id) as team_diversity,
			AVG(ut.total_points) as avg_points
		FROM user_teams ut
		JOIN team_players tp ON ut.id = tp.team_id
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY ut.id
		HAVING COUNT(DISTINCT tp.real_team_id) >= 2
	`
	
	rows, err := s.db.Query(fmt.Sprintf(query, days), gameID)
	if err != nil {
		return 0.0, err
	}
	defer rows.Close()

	var diversities, points []float64
	for rows.Next() {
		var diversity int
		var avgPoints float64
		
		err := rows.Scan(&diversity, &avgPoints)
		if err != nil {
			continue
		}
		
		diversities = append(diversities, float64(diversity))
		points = append(points, avgPoints)
	}

	if len(diversities) < 2 {
		return 0.5, nil // Default neutral synergy
	}

	return calculateCorrelation(diversities, points), nil
}

func (s *AdvancedAnalyticsService) calculateStrategicDiversity(gameID, days int) (float64, error) {
	// Calculate Shannon diversity index for team compositions
	query := `
		SELECT p.role, COUNT(*) as role_count
		FROM team_players tp
		JOIN players p ON tp.player_id = p.id
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND p.role IS NOT NULL
		GROUP BY p.role
	`
	
	rows, err := s.db.Query(fmt.Sprintf(query, days), gameID)
	if err != nil {
		return 0.0, err
	}
	defer rows.Close()

	var total int
	var roleCounts []int
	
	for rows.Next() {
		var role string
		var count int
		
		err := rows.Scan(&role, &count)
		if err != nil {
			continue
		}
		
		roleCounts = append(roleCounts, count)
		total += count
	}

	if total == 0 {
		return 0.0, nil
	}

	// Calculate Shannon diversity index
	diversity := 0.0
	for _, count := range roleCounts {
		if count > 0 {
			p := float64(count) / float64(total)
			diversity -= p * math.Log2(p)
		}
	}

	return diversity / math.Log2(float64(len(roleCounts))), nil // Normalize
}

func (s *AdvancedAnalyticsService) calculateComebackPotential(gameID, days int) (float64, error) {
	// Teams that score higher in later rounds vs earlier rounds
	query := `
		SELECT 
			AVG(CASE WHEN me.round_number <= 10 THEN me.points ELSE 0 END) as early_points,
			AVG(CASE WHEN me.round_number > 10 THEN me.points ELSE 0 END) as late_points
		FROM match_events me
		JOIN matches m ON me.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND me.round_number IS NOT NULL
	`
	
	var earlyPoints, latePoints sql.NullFloat64
	err := s.db.QueryRow(fmt.Sprintf(query, days), gameID).Scan(&earlyPoints, &latePoints)
	if err != nil || !earlyPoints.Valid || !latePoints.Valid {
		return 0.5, err
	}
	
	if earlyPoints.Float64 == 0 {
		return 0.5, nil
	}
	
	return math.Min(1.0, latePoints.Float64/earlyPoints.Float64), nil
}

func (s *AdvancedAnalyticsService) calculateClutchPerformance(gameID, days int) (float64, error) {
	// Performance in high-stakes contests (high entry fee, many participants)
	query := `
		SELECT AVG(ut.total_points) as clutch_points
		FROM user_teams ut
		JOIN contest_participants cp ON ut.id = cp.team_id
		JOIN contests c ON cp.contest_id = c.id
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND c.entry_fee >= 100 AND c.current_participants >= 1000
	`
	
	var clutchPoints sql.NullFloat64
	err := s.db.QueryRow(fmt.Sprintf(query, days), gameID).Scan(&clutchPoints)
	if err != nil || !clutchPoints.Valid {
		return 0.0, err
	}
	
	// Normalize against average points
	var avgPoints sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG(ut.total_points)
		FROM user_teams ut
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
	`, fmt.Sprintf(query, days), gameID).Scan(&avgPoints)
	
	if err != nil || !avgPoints.Valid || avgPoints.Float64 == 0 {
		return 0.5, nil
	}
	
	return math.Min(1.0, clutchPoints.Float64/avgPoints.Float64), nil
}

func (s *AdvancedAnalyticsService) calculateConsistencyIndex(gameID, days int) (float64, error) {
	// Calculate coefficient of variation for player performances
	query := `
		SELECT 
			p.id,
			STDDEV(tp.points_earned) as std_dev,
			AVG(tp.points_earned) as avg_points
		FROM team_players tp
		JOIN players p ON tp.player_id = p.id
		JOIN user_teams ut ON tp.team_id = ut.id
		JOIN matches m ON ut.match_id = m.id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		AND tp.points_earned > 0
		GROUP BY p.id
		HAVING COUNT(*) >= 5
	`
	
	rows, err := s.db.Query(fmt.Sprintf(query, days), gameID)
	if err != nil {
		return 0.0, err
	}
	defer rows.Close()

	var cvs []float64
	for rows.Next() {
		var playerID int64
		var stdDev, avgPoints sql.NullFloat64
		
		err := rows.Scan(&playerID, &stdDev, &avgPoints)
		if err != nil || !stdDev.Valid || !avgPoints.Valid || avgPoints.Float64 == 0 {
			continue
		}
		
		cv := stdDev.Float64 / avgPoints.Float64
		cvs = append(cvs, cv)
	}

	if len(cvs) == 0 {
		return 0.5, nil
	}

	// Calculate average coefficient of variation
	sum := 0.0
	for _, cv := range cvs {
		sum += cv
	}
	avgCV := sum / float64(len(cvs))
	
	// Lower CV means higher consistency (invert for index)
	return math.Max(0.0, 1.0-math.Min(1.0, avgCV)), nil
}

func (s *AdvancedAnalyticsService) calculateAdaptabilityScore(gameID, days int) (float64, error) {
	// How well teams perform against different opponents
	query := `
		SELECT 
			ut.user_id,
			COUNT(DISTINCT mp.team_id) as unique_opponents,
			AVG(ut.total_points) as avg_performance
		FROM user_teams ut
		JOIN matches m ON ut.match_id = m.id
		JOIN match_participants mp ON m.id = mp.match_id
		WHERE m.game_id = $1 AND m.created_at >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY ut.user_id
		HAVING COUNT(DISTINCT mp.team_id) >= 3
	`
	
	rows, err := s.db.Query(fmt.Sprintf(query, days), gameID)
	if err != nil {
		return 0.0, err
	}
	defer rows.Close()

	var adaptabilityScores []float64
	for rows.Next() {
		var userID int64
		var uniqueOpponents int
		var avgPerformance float64
		
		err := rows.Scan(&userID, &uniqueOpponents, &avgPerformance)
		if err != nil {
			continue
		}
		
		// Score based on diversity of opponents and performance
		score := math.Min(1.0, float64(uniqueOpponents)/10.0) * (avgPerformance/100.0)
		adaptabilityScores = append(adaptabilityScores, score)
	}

	if len(adaptabilityScores) == 0 {
		return 0.5, nil
	}

	sum := 0.0
	for _, score := range adaptabilityScores {
		sum += score
	}
	
	return sum / float64(len(adaptabilityScores)), nil
}

func (s *AdvancedAnalyticsService) storeAdvancedMetrics(gameID int, metrics *models.AdvancedGameMetrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	date := time.Now().Format("2006-01-02")
	
	// Store each metric
	metricsMap := map[string]float64{
		"player_efficiency":    metrics.PlayerEfficiency,
		"team_synergy":        metrics.TeamSynergy,
		"strategic_diversity": metrics.StrategicDiversity,
		"comeback_potential":  metrics.ComebackPotential,
		"clutch_performance":  metrics.ClutchPerformance,
		"consistency_index":   metrics.ConsistencyIndex,
		"adaptability_score":  metrics.AdaptabilityScore,
	}

	for metricType, value := range metricsMap {
		_, err = tx.Exec(`
			INSERT INTO game_analytics_advanced (game_id, date, metric_type, metric_value)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (game_id, date, metric_type)
			DO UPDATE SET metric_value = EXCLUDED.metric_value
		`, gameID, date, metricType, value)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *AdvancedAnalyticsService) GetAdvancedMetricsHistory(gameID int, days int) ([]models.GameAnalyticsAdvanced, error) {
	query := `
		SELECT id, game_id, date, metric_type, metric_value, metadata, created_at
		FROM game_analytics_advanced
		WHERE game_id = $1 AND date >= CURRENT_DATE - INTERVAL '%d days'
		ORDER BY date DESC, metric_type
	`
	
	rows, err := s.db.Query(fmt.Sprintf(query, days), gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []models.GameAnalyticsAdvanced
	for rows.Next() {
		var a models.GameAnalyticsAdvanced
		err := rows.Scan(&a.ID, &a.GameID, &a.Date, &a.MetricType, &a.MetricValue, &a.Metadata, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, a)
	}

	return analytics, nil
}

func (s *AdvancedAnalyticsService) GetGameComparison(gameIDs []int, metricType string, days int) (map[string]interface{}, error) {
	if len(gameIDs) == 0 {
		return nil, fmt.Errorf("no games provided for comparison")
	}

	// Build query for multiple games
	placeholders := make([]string, len(gameIDs))
	args := make([]interface{}, len(gameIDs))
	for i, gameID := range gameIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = gameID
	}

	query := fmt.Sprintf(`
		SELECT g.name, gaa.metric_value, gaa.date
		FROM game_analytics_advanced gaa
		JOIN games g ON gaa.game_id = g.id
		WHERE gaa.game_id IN (%s) AND gaa.metric_type = '%s'
		AND gaa.date >= CURRENT_DATE - INTERVAL '%d days'
		ORDER BY gaa.date DESC
	`, strings.Join(placeholders, ","), metricType, days)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comparison := map[string]interface{}{
		"metric_type": metricType,
		"games":      []map[string]interface{}{},
		"trends":     map[string][]float64{},
	}

	gameData := make(map[string][]map[string]interface{})
	
	for rows.Next() {
		var gameName string
		var metricValue float64
		var date time.Time
		
		err := rows.Scan(&gameName, &metricValue, &date)
		if err != nil {
			continue
		}
		
		if _, exists := gameData[gameName]; !exists {
			gameData[gameName] = []map[string]interface{}{}
		}
		
		gameData[gameName] = append(gameData[gameName], map[string]interface{}{
			"date":  date,
			"value": metricValue,
		})
	}

	// Process data for response
	for gameName, data := range gameData {
		var values []float64
		for _, point := range data {
			values = append(values, point["value"].(float64))
		}
		
		avg := 0.0
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			avg = sum / float64(len(values))
		}
		
		comparison["games"] = append(comparison["games"].([]map[string]interface{}), map[string]interface{}{
			"name":    gameName,
			"average": avg,
			"data":    data,
		})
		
		comparison["trends"].(map[string][]float64)[gameName] = values
	}

	return comparison, nil
}

// Helper function to calculate correlation coefficient
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}
	
	n := float64(len(x))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0
	
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}
	
	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
	
	if denominator == 0 {
		return 0.0
	}
	
	return numerator / denominator
}