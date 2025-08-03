package services

import (
	"database/sql"
	"fmt"
	"time"
	"strings"

	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/errors"
	"fantasy-esports-backend/pkg/logger"
)

type AnalyticsService struct {
	db *sql.DB
}

func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetDashboard returns comprehensive analytics dashboard
func (s *AnalyticsService) GetDashboard(filters models.AnalyticsFilters) (*models.AnalyticsDashboard, error) {
	dashboard := &models.AnalyticsDashboard{}

	// Get all metrics in parallel
	userMetricsChan := make(chan models.UserMetrics, 1)
	revenueMetricsChan := make(chan models.RevenueMetrics, 1)
	contestMetricsChan := make(chan models.ContestMetrics, 1)
	gameMetricsChan := make(chan models.GameMetrics, 1)
	engagementMetricsChan := make(chan models.EngagementMetrics, 1)
	systemHealthChan := make(chan models.SystemHealth, 1)
	topPerformersChan := make(chan models.TopPerformers, 1)

	// Execute analytics queries concurrently
	go s.getUserMetricsAsync(filters, userMetricsChan)
	go s.getRevenueMetricsAsync(filters, revenueMetricsChan)
	go s.getContestMetricsAsync(filters, contestMetricsChan)
	go s.getGameMetricsAsync(filters, gameMetricsChan)
	go s.getEngagementMetricsAsync(filters, engagementMetricsChan)
	go s.getSystemHealthAsync(systemHealthChan)
	go s.getTopPerformersAsync(filters, topPerformersChan)

	// Collect results
	dashboard.UserMetrics = <-userMetricsChan
	dashboard.RevenueMetrics = <-revenueMetricsChan
	dashboard.ContestMetrics = <-contestMetricsChan
	dashboard.GameMetrics = <-gameMetricsChan
	dashboard.EngagementMetrics = <-engagementMetricsChan
	dashboard.SystemHealth = <-systemHealthChan
	dashboard.TopPerformers = <-topPerformersChan

	return dashboard, nil
}

// GetUserMetrics returns detailed user analytics
func (s *AnalyticsService) GetUserMetrics(filters models.AnalyticsFilters) (*models.UserMetrics, error) {
	metrics := models.UserMetrics{}

	// Get basic user counts
	basicQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_verified = true THEN 1 END) as verified_users,
			COUNT(CASE WHEN kyc_status = 'verified' THEN 1 END) as kyc_completed_users,
			COUNT(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 END) as new_users_today,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as new_users_this_week,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as new_users_this_month,
			COUNT(CASE WHEN last_login_at >= CURRENT_DATE - INTERVAL '1 day' THEN 1 END) as active_users
		FROM users 
		WHERE is_active = true
	`

	err := s.db.QueryRow(basicQuery).Scan(
		&metrics.TotalUsers,
		&metrics.VerifiedUsers,
		&metrics.KYCCompletedUsers,
		&metrics.NewUsersToday,
		&metrics.NewUsersThisWeek,
		&metrics.NewUsersThisMonth,
		&metrics.ActiveUsers,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate growth rate
	if metrics.TotalUsers > 0 {
		metrics.UserGrowthRate = float64(metrics.NewUsersThisMonth) / float64(metrics.TotalUsers) * 100
	}

	// Get users by state
	stateQuery := `
		SELECT state, COUNT(*) as count
		FROM users 
		WHERE state IS NOT NULL AND is_active = true
		GROUP BY state 
		ORDER BY count DESC 
		LIMIT 10
	`
	stateRows, err := s.db.Query(stateQuery)
	if err != nil {
		logger.Error("Error fetching users by state", map[string]interface{}{"error": err.Error()})
	} else {
		defer stateRows.Close()
		for stateRows.Next() {
			var region models.UsersByRegion
			stateRows.Scan(&region.Region, &region.Count)
			metrics.UsersByState = append(metrics.UsersByState, region)
		}
	}

	// Get registration trend (last 30 days)
	trendQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users 
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days' 
		GROUP BY DATE(created_at) 
		ORDER BY date
	`
	trendRows, err := s.db.Query(trendQuery)
	if err != nil {
		logger.Error("Error fetching registration trend", map[string]interface{}{"error": err.Error()})
	} else {
		defer trendRows.Close()
		for trendRows.Next() {
			var trend models.DailyCount
			trendRows.Scan(&trend.Date, &trend.Count)
			metrics.UserRegistrationTrend = append(metrics.UserRegistrationTrend, trend)
		}
	}

	// Calculate retention rate (simplified)
	retentionQuery := `
		SELECT COUNT(*) * 100.0 / NULLIF(
			(SELECT COUNT(*) FROM users WHERE created_at <= CURRENT_DATE - INTERVAL '7 days'), 0
		) as retention_rate
		FROM users 
		WHERE created_at <= CURRENT_DATE - INTERVAL '7 days' 
		AND last_login_at >= CURRENT_DATE - INTERVAL '7 days'
	`
	err = s.db.QueryRow(retentionQuery).Scan(&metrics.UserRetentionRate)
	if err != nil {
		logger.Error("Error calculating retention rate", map[string]interface{}{"error": err.Error()})
		metrics.UserRetentionRate = 0
	}

	return &metrics, nil
}

// GetRevenueMetrics returns detailed revenue analytics
func (s *AnalyticsService) GetRevenueMetrics(filters models.AnalyticsFilters) (*models.RevenueMetrics, error) {
	metrics := models.RevenueMetrics{}

	// Get basic revenue metrics
	revenueQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN transaction_type = 'deposit' AND status = 'completed' THEN amount END), 0) as total_deposits,
			COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal' AND status = 'completed' THEN amount END), 0) as total_withdrawals,
			COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal' AND status = 'pending' THEN amount END), 0) as pending_withdrawals,
			COALESCE(SUM(CASE WHEN transaction_type = 'contest_fee' AND status = 'completed' THEN amount END), 0) as contest_revenue,
			COALESCE(SUM(CASE WHEN DATE(created_at) = CURRENT_DATE AND status = 'completed' THEN amount END), 0) as revenue_today,
			COALESCE(SUM(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' AND status = 'completed' THEN amount END), 0) as revenue_this_week,
			COALESCE(SUM(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' AND status = 'completed' THEN amount END), 0) as revenue_this_month
		FROM wallet_transactions
	`

	var contestRevenue float64
	err := s.db.QueryRow(revenueQuery).Scan(
		&metrics.TotalDeposits,
		&metrics.TotalWithdrawals,
		&metrics.PendingWithdrawals,
		&contestRevenue,
		&metrics.RevenueToday,
		&metrics.RevenueThisWeek,
		&metrics.RevenueThisMonth,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate total revenue (deposits - withdrawals)
	metrics.TotalRevenue = contestRevenue + (metrics.TotalDeposits * 0.05) // Assuming 5% platform fee

	// Calculate average revenue per user
	userCountQuery := `SELECT COUNT(*) FROM users WHERE is_active = true`
	var totalUsers int64
	s.db.QueryRow(userCountQuery).Scan(&totalUsers)
	if totalUsers > 0 {
		metrics.AvgRevenuePerUser = metrics.TotalRevenue / float64(totalUsers)
	}

	// Calculate revenue growth rate
	lastMonthRevenueQuery := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM wallet_transactions 
		WHERE created_at >= CURRENT_DATE - INTERVAL '60 days' 
		AND created_at < CURRENT_DATE - INTERVAL '30 days' 
		AND status = 'completed'
		AND transaction_type = 'contest_fee'
	`
	var lastMonthRevenue float64
	s.db.QueryRow(lastMonthRevenueQuery).Scan(&lastMonthRevenue)
	if lastMonthRevenue > 0 {
		metrics.RevenueGrowthRate = ((metrics.RevenueThisMonth - lastMonthRevenue) / lastMonthRevenue) * 100
	}

	// Get revenue by game
	gameRevenueQuery := `
		SELECT g.name, COALESCE(SUM(wt.amount), 0) as revenue
		FROM games g
		LEFT JOIN contests c ON c.match_id IN (SELECT id FROM matches WHERE game_id = g.id)
		LEFT JOIN contest_participants cp ON cp.contest_id = c.id
		LEFT JOIN wallet_transactions wt ON wt.reference_id = cp.contest_id::text AND wt.transaction_type = 'contest_fee'
		WHERE wt.status = 'completed'
		GROUP BY g.id, g.name
		ORDER BY revenue DESC
		LIMIT 10
	`
	gameRevenueRows, err := s.db.Query(gameRevenueQuery)
	if err != nil {
		logger.Error("Error fetching revenue by game", map[string]interface{}{"error": err.Error()})
	} else {
		defer gameRevenueRows.Close()
		for gameRevenueRows.Next() {
			var gameRevenue models.RevenueByCategory
			gameRevenueRows.Scan(&gameRevenue.Category, &gameRevenue.Revenue)
			metrics.RevenueByGame = append(metrics.RevenueByGame, gameRevenue)
		}
	}

	// Get monthly revenue trend (last 12 months)
	monthlyTrendQuery := `
		SELECT DATE_TRUNC('month', created_at) as month, SUM(amount) as revenue
		FROM wallet_transactions 
		WHERE created_at >= CURRENT_DATE - INTERVAL '12 months' 
		AND status = 'completed'
		AND transaction_type = 'contest_fee'
		GROUP BY DATE_TRUNC('month', created_at) 
		ORDER BY month
	`
	trendRows, err := s.db.Query(monthlyTrendQuery)
	if err != nil {
		logger.Error("Error fetching monthly revenue trend", map[string]interface{}{"error": err.Error()})
	} else {
		defer trendRows.Close()
		for trendRows.Next() {
			var trend models.MonthlyRevenue
			trendRows.Scan(&trend.Month, &trend.Revenue)
			metrics.MonthlyRevenueTrend = append(metrics.MonthlyRevenueTrend, trend)
		}
	}

	// Get payment method distribution
	paymentMethodQuery := `
		SELECT pt.gateway, SUM(pt.amount) as amount, COUNT(*) as count
		FROM payment_transactions pt
		WHERE pt.status = 'success' AND pt.type = 'deposit'
		GROUP BY pt.gateway
		ORDER BY amount DESC
	`
	paymentRows, err := s.db.Query(paymentMethodQuery)
	if err != nil {
		logger.Error("Error fetching payment method distribution", map[string]interface{}{"error": err.Error()})
	} else {
		defer paymentRows.Close()
		for paymentRows.Next() {
			var payment models.PaymentMethodStats
			paymentRows.Scan(&payment.Method, &payment.Amount, &payment.Count)
			metrics.PaymentMethodDistribution = append(metrics.PaymentMethodDistribution, payment)
		}
	}

	return &metrics, nil
}

// GetContestMetrics returns detailed contest analytics
func (s *AnalyticsService) GetContestMetrics(filters models.AnalyticsFilters) (*models.ContestMetrics, error) {
	metrics := models.ContestMetrics{}

	// Get basic contest metrics
	contestQuery := `
		SELECT 
			COUNT(*) as total_contests,
			COUNT(CASE WHEN status = 'live' OR status = 'upcoming' THEN 1 END) as active_contests,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_contests,
			COALESCE(SUM(current_participants), 0) as total_participations,
			COALESCE(SUM(total_prize_pool), 0) as total_prize_distributed
		FROM contests
	`

	err := s.db.QueryRow(contestQuery).Scan(
		&metrics.TotalContests,
		&metrics.ActiveContests,
		&metrics.CompletedContests,
		&metrics.TotalParticipations,
		&metrics.TotalPrizeDistributed,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate average participations per contest
	if metrics.TotalContests > 0 {
		metrics.AvgParticipationsPerContest = float64(metrics.TotalParticipations) / float64(metrics.TotalContests)
	}

	// Calculate contest completion rate
	if metrics.TotalContests > 0 {
		metrics.ContestCompletionRate = (float64(metrics.CompletedContests) / float64(metrics.TotalContests)) * 100
	}

	// Get contests by entry fee
	entryFeeQuery := `
		SELECT 
			CASE 
				WHEN entry_fee = 0 THEN 'Free'
				WHEN entry_fee <= 50 THEN 'Low (₹1-50)'
				WHEN entry_fee <= 200 THEN 'Medium (₹51-200)'
				WHEN entry_fee <= 1000 THEN 'High (₹201-1000)'
				ELSE 'Premium (₹1000+)'
			END as category,
			COUNT(*) as count
		FROM contests
		GROUP BY 
			CASE 
				WHEN entry_fee = 0 THEN 'Free'
				WHEN entry_fee <= 50 THEN 'Low (₹1-50)'
				WHEN entry_fee <= 200 THEN 'Medium (₹51-200)'
				WHEN entry_fee <= 1000 THEN 'High (₹201-1000)'
				ELSE 'Premium (₹1000+)'
			END
		ORDER BY count DESC
	`
	entryFeeRows, err := s.db.Query(entryFeeQuery)
	if err != nil {
		logger.Error("Error fetching contests by entry fee", map[string]interface{}{"error": err.Error()})
	} else {
		defer entryFeeRows.Close()
		for entryFeeRows.Next() {
			var category models.ContestsByCategory
			entryFeeRows.Scan(&category.Category, &category.Count)
			metrics.ContestsByEntryFee = append(metrics.ContestsByEntryFee, category)
		}
	}

	// Get popular contest types
	contestTypeQuery := `
		SELECT contest_type, SUM(current_participants) as participant_count
		FROM contests
		GROUP BY contest_type
		ORDER BY participant_count DESC
		LIMIT 5
	`
	typeRows, err := s.db.Query(contestTypeQuery)
	if err != nil {
		logger.Error("Error fetching popular contest types", map[string]interface{}{"error": err.Error()})
	} else {
		defer typeRows.Close()
		for typeRows.Next() {
			var contestType models.PopularContest
			typeRows.Scan(&contestType.ContestType, &contestType.ParticipantCount)
			metrics.PopularContestTypes = append(metrics.PopularContestTypes, contestType)
		}
	}

	return &metrics, nil
}

// GetGameMetrics returns detailed game analytics
func (s *AnalyticsService) GetGameMetrics(filters models.AnalyticsFilters) (*models.GameMetrics, error) {
	metrics := models.GameMetrics{}

	// Get basic game metrics
	gameQuery := `
		SELECT 
			COUNT(*) as total_games,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_games
		FROM games
	`

	err := s.db.QueryRow(gameQuery).Scan(
		&metrics.TotalGames,
		&metrics.ActiveGames,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Get match statistics
	matchQuery := `
		SELECT 
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as matches_completed,
			COUNT(CASE WHEN status = 'upcoming' THEN 1 END) as upcoming_matches,
			COUNT(CASE WHEN status = 'live' THEN 1 END) as live_matches
		FROM matches
	`

	err = s.db.QueryRow(matchQuery).Scan(
		&metrics.MatchesCompleted,
		&metrics.UpcomingMatches,
		&metrics.LiveMatches,
	)
	if err != nil {
		logger.Error("Error fetching match statistics", map[string]interface{}{"error": err.Error()})
	}

	// Get most popular game
	popularGameQuery := `
		SELECT g.name
		FROM games g
		JOIN matches m ON m.game_id = g.id
		JOIN contests c ON c.match_id = m.id
		GROUP BY g.id, g.name
		ORDER BY SUM(c.current_participants) DESC
		LIMIT 1
	`
	err = s.db.QueryRow(popularGameQuery).Scan(&metrics.MostPopularGame)
	if err != nil {
		logger.Error("Error fetching most popular game", map[string]interface{}{"error": err.Error()})
		metrics.MostPopularGame = "N/A"
	}

	// Get game participation statistics
	gameStatsQuery := `
		SELECT 
			g.name,
			COUNT(DISTINCT m.id) as total_matches,
			COUNT(DISTINCT p.id) as total_players,
			AVG(c.current_participants) as avg_participation
		FROM games g
		LEFT JOIN matches m ON m.game_id = g.id
		LEFT JOIN contests c ON c.match_id = m.id
		LEFT JOIN players p ON p.game_id = g.id
		WHERE g.is_active = true
		GROUP BY g.id, g.name
		ORDER BY avg_participation DESC
	`
	gameStatsRows, err := s.db.Query(gameStatsQuery)
	if err != nil {
		logger.Error("Error fetching game participation stats", map[string]interface{}{"error": err.Error()})
	} else {
		defer gameStatsRows.Close()
		for gameStatsRows.Next() {
			var gameStats models.GameStats
			var avgParticipation sql.NullFloat64
			gameStatsRows.Scan(&gameStats.GameName, &gameStats.TotalMatches, &gameStats.TotalPlayers, &avgParticipation)
			if avgParticipation.Valid {
				gameStats.AvgParticipation = avgParticipation.Float64
			}
			metrics.GameParticipation = append(metrics.GameParticipation, gameStats)
		}
	}

	// Get top player performance
	topPlayersQuery := `
		SELECT 
			p.name,
			g.name,
			t.name,
			AVG(tp.points_earned) as avg_points,
			COUNT(DISTINCT ut.match_id) as total_matches,
			COUNT(DISTINCT tp.team_id) as popularity_score
		FROM players p
		JOIN games g ON g.id = p.game_id
		JOIN teams t ON t.id = p.team_id
		LEFT JOIN team_players tp ON tp.player_id = p.id
		LEFT JOIN user_teams ut ON ut.id = tp.team_id
		WHERE p.is_playing = true
		GROUP BY p.id, p.name, g.name, t.name
		HAVING COUNT(DISTINCT tp.team_id) > 0
		ORDER BY avg_points DESC
		LIMIT 10
	`
	topPlayersRows, err := s.db.Query(topPlayersQuery)
	if err != nil {
		logger.Error("Error fetching top players", map[string]interface{}{"error": err.Error()})
	} else {
		defer topPlayersRows.Close()
		for topPlayersRows.Next() {
			var player models.TopPlayer
			var avgPoints sql.NullFloat64
			topPlayersRows.Scan(
				&player.PlayerName,
				&player.GameName,
				&player.TeamName,
				&avgPoints,
				&player.TotalMatches,
				&player.PopularityScore,
			)
			if avgPoints.Valid {
				player.AvgPoints = avgPoints.Float64
			}
			metrics.PlayerPerformance = append(metrics.PlayerPerformance, player)
		}
	}

	return &metrics, nil
}

// Helper functions for async operations
func (s *AnalyticsService) getUserMetricsAsync(filters models.AnalyticsFilters, ch chan models.UserMetrics) {
	metrics, err := s.GetUserMetrics(filters)
	if err != nil {
		logger.Error("Error getting user metrics", map[string]interface{}{"error": err.Error()})
		ch <- models.UserMetrics{}
		return
	}
	ch <- *metrics
}

func (s *AnalyticsService) getRevenueMetricsAsync(filters models.AnalyticsFilters, ch chan models.RevenueMetrics) {
	metrics, err := s.GetRevenueMetrics(filters)
	if err != nil {
		logger.Error("Error getting revenue metrics", map[string]interface{}{"error": err.Error()})
		ch <- models.RevenueMetrics{}
		return
	}
	ch <- *metrics
}

func (s *AnalyticsService) getContestMetricsAsync(filters models.AnalyticsFilters, ch chan models.ContestMetrics) {
	metrics, err := s.GetContestMetrics(filters)
	if err != nil {
		logger.Error("Error getting contest metrics", map[string]interface{}{"error": err.Error()})
		ch <- models.ContestMetrics{}
		return
	}
	ch <- *metrics
}

func (s *AnalyticsService) getGameMetricsAsync(filters models.AnalyticsFilters, ch chan models.GameMetrics) {
	metrics, err := s.GetGameMetrics(filters)
	if err != nil {
		logger.Error("Error getting game metrics", map[string]interface{}{"error": err.Error()})
		ch <- models.GameMetrics{}
		return
	}
	ch <- *metrics
}

func (s *AnalyticsService) getEngagementMetricsAsync(filters models.AnalyticsFilters, ch chan models.EngagementMetrics) {
	// Implementation for engagement metrics
	metrics := models.EngagementMetrics{}
	
	// Get daily, weekly, monthly active users
	activeUsersQuery := `
		SELECT 
			COUNT(CASE WHEN last_login_at >= CURRENT_DATE - INTERVAL '1 day' THEN 1 END) as dau,
			COUNT(CASE WHEN last_login_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as wau,
			COUNT(CASE WHEN last_login_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as mau
		FROM users
		WHERE is_active = true
	`
	
	err := s.db.QueryRow(activeUsersQuery).Scan(
		&metrics.DailyActiveUsers,
		&metrics.WeeklyActiveUsers,
		&metrics.MonthlyActiveUsers,
	)
	if err != nil {
		logger.Error("Error getting active users", map[string]interface{}{"error": err.Error()})
	}

	// Calculate retention rates (simplified)
	if metrics.MonthlyActiveUsers > 0 {
		metrics.UserRetentionDay7 = float64(metrics.WeeklyActiveUsers) / float64(metrics.MonthlyActiveUsers) * 100
		metrics.UserRetentionDay30 = 85.0 // Placeholder
		metrics.ChurnRate = 15.0          // Placeholder
	}

	ch <- metrics
}

func (s *AnalyticsService) getSystemHealthAsync(ch chan models.SystemHealth) {
	health := models.SystemHealth{}
	
	// Check database health
	err := s.db.Ping()
	if err != nil {
		health.DatabaseHealth = "Unhealthy"
	} else {
		health.DatabaseHealth = "Healthy"
	}
	
	// Set default values (in production, these would come from monitoring systems)
	health.APIResponseTime = 150.5
	health.ErrorRate = 0.02
	health.ActiveConnections = 45
	health.SystemUptime = "15 days 4 hours"
	health.MemoryUsage = 65.4
	health.CPUUsage = 23.8
	
	ch <- health
}

func (s *AnalyticsService) getTopPerformersAsync(filters models.AnalyticsFilters, ch chan models.TopPerformers) {
	performers := models.TopPerformers{}
	
	// Get top earners
	earnersQuery := `
		SELECT u.id, COALESCE(u.first_name, 'User') || ' ' || COALESCE(u.last_name, u.id::text) as username,
		       SUM(wt.amount) as total_earnings, COUNT(cp.contest_id) as contests_won
		FROM users u
		JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.transaction_type = 'prize_credit'
		JOIN contest_participants cp ON cp.user_id = u.id AND cp.prize_won > 0
		WHERE wt.status = 'completed'
		GROUP BY u.id
		ORDER BY total_earnings DESC
		LIMIT 10
	`
	earnersRows, err := s.db.Query(earnersQuery)
	if err != nil {
		logger.Error("Error fetching top earners", map[string]interface{}{"error": err.Error()})
	} else {
		defer earnersRows.Close()
		for earnersRows.Next() {
			var earner models.TopEarner
			earnersRows.Scan(&earner.UserID, &earner.Username, &earner.TotalEarnings, &earner.ContestsWon)
			performers.TopEarners = append(performers.TopEarners, earner)
		}
	}
	
	ch <- performers
}

// GetRealTimeMetrics returns current real-time metrics
func (s *AnalyticsService) GetRealTimeMetrics() (*models.RealTimeMetrics, error) {
	metrics := models.RealTimeMetrics{}
	
	// Get current active users (last 5 minutes)
	activeUsersQuery := `
		SELECT COUNT(*) FROM users 
		WHERE last_login_at >= NOW() - INTERVAL '5 minutes'
	`
	s.db.QueryRow(activeUsersQuery).Scan(&metrics.ActiveUsers)
	
	// Get live contests
	liveContestsQuery := `SELECT COUNT(*) FROM contests WHERE status = 'live'`
	s.db.QueryRow(liveContestsQuery).Scan(&metrics.LiveContests)
	
	// Get active matches
	activeMatchesQuery := `SELECT COUNT(*) FROM matches WHERE status = 'live'`
	s.db.QueryRow(activeMatchesQuery).Scan(&metrics.ActiveMatches)
	
	// Get transactions per minute (last minute)
	transactionsQuery := `
		SELECT COUNT(*) FROM wallet_transactions 
		WHERE created_at >= NOW() - INTERVAL '1 minute'
	`
	s.db.QueryRow(transactionsQuery).Scan(&metrics.TransactionsPerMinute)
	
	// Get new registrations (today)
	registrationsQuery := `
		SELECT COUNT(*) FROM users 
		WHERE DATE(created_at) = CURRENT_DATE
	`
	s.db.QueryRow(registrationsQuery).Scan(&metrics.NewRegistrations)
	
	// Get current revenue (today)
	revenueQuery := `
		SELECT COALESCE(SUM(amount), 0) FROM wallet_transactions 
		WHERE DATE(created_at) = CURRENT_DATE AND status = 'completed' 
		AND transaction_type = 'contest_fee'
	`
	s.db.QueryRow(revenueQuery).Scan(&metrics.CurrentRevenue)
	
	metrics.SystemLoad = 0.45 // Placeholder
	metrics.LastUpdated = time.Now()
	
	return &metrics, nil
}

// GetPerformanceMetrics returns system performance metrics
func (s *AnalyticsService) GetPerformanceMetrics() (*models.PerformanceMetrics, error) {
	metrics := models.PerformanceMetrics{}
	
	// In a real system, these would come from APM tools
	// Setting placeholder values for now
	metrics.CacheHitRate = 89.5
	metrics.AverageResponseTime = 145.2
	metrics.P95ResponseTime = 285.7
	metrics.P99ResponseTime = 567.3
	metrics.ErrorRate = 0.02
	metrics.ThroughputPerSecond = 125.8
	
	// Placeholder API performance data
	metrics.APIEndpoints = []models.APIPerformance{
		{
			Endpoint:        "/api/v1/contests",
			Method:          "GET",
			AvgResponseTime: 85.2,
			RequestCount:    15420,
			ErrorCount:      12,
			ErrorRate:       0.08,
		},
		{
			Endpoint:        "/api/v1/teams/create",
			Method:          "POST",
			AvgResponseTime: 165.7,
			RequestCount:    8750,
			ErrorCount:      25,
			ErrorRate:       0.29,
		},
	}
	
	// Placeholder query performance data
	metrics.DatabaseQueries = []models.QueryPerformance{
		{
			QueryType:       "SELECT_CONTESTS",
			AvgExecutionTime: 45.2,
			ExecutionCount:  25430,
			SlowQueryCount:  15,
		},
		{
			QueryType:       "INSERT_TEAM",
			AvgExecutionTime: 12.8,
			ExecutionCount:  8750,
			SlowQueryCount:  2,
		},
	}
	
	return &metrics, nil
}

// applyFilters applies date and other filters to queries
func (s *AnalyticsService) applyFilters(baseQuery string, filters models.AnalyticsFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}
	
	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}
	
	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		if strings.Contains(baseQuery, "WHERE") {
			whereClause = " AND " + strings.Join(conditions, " AND ")
		}
		baseQuery += whereClause
	}
	
	return baseQuery, args
}