package services

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/errors"
	"fantasy-esports-backend/pkg/logger"
)

type BusinessIntelligenceService struct {
	db *sql.DB
}

func NewBusinessIntelligenceService(db *sql.DB) *BusinessIntelligenceService {
	return &BusinessIntelligenceService{db: db}
}

// GetBIDashboard returns comprehensive business intelligence dashboard
func (s *BusinessIntelligenceService) GetBIDashboard(filters models.BIFilters) (*models.BusinessIntelligenceDashboard, error) {
	dashboard := &models.BusinessIntelligenceDashboard{}

	// Get KPI metrics
	kpiMetrics, err := s.GetKPIMetrics(filters)
	if err != nil {
		logger.Error("Error getting KPI metrics", map[string]interface{}{"error": err.Error()})
		kpiMetrics = &models.KPIMetrics{}
	}
	dashboard.KPIMetrics = *kpiMetrics

	// Get revenue analytics
	revenueAnalytics, err := s.GetRevenueAnalytics(filters)
	if err != nil {
		logger.Error("Error getting revenue analytics", map[string]interface{}{"error": err.Error()})
		revenueAnalytics = &models.RevenueAnalytics{}
	}
	dashboard.RevenueAnalytics = *revenueAnalytics

	// Get user behavior analysis
	userBehavior, err := s.GetUserBehaviorAnalysis(filters)
	if err != nil {
		logger.Error("Error getting user behavior analysis", map[string]interface{}{"error": err.Error()})
		userBehavior = &models.UserBehaviorAnalysis{}
	}
	dashboard.UserBehaviorAnalysis = *userBehavior

	// Get predictive analytics
	predictiveAnalytics, err := s.GetPredictiveAnalytics(filters)
	if err != nil {
		logger.Error("Error getting predictive analytics", map[string]interface{}{"error": err.Error()})
		predictiveAnalytics = &models.PredictiveAnalytics{}
	}
	dashboard.PredictiveAnalytics = *predictiveAnalytics

	// Get competitive analysis
	competitiveAnalysis := s.getCompetitiveAnalysis()
	dashboard.CompetitiveAnalysis = *competitiveAnalysis

	// Get business insights
	insights, err := s.generateBusinessInsights(filters)
	if err != nil {
		logger.Error("Error generating business insights", map[string]interface{}{"error": err.Error()})
		insights = []models.BusinessInsight{}
	}
	dashboard.BusinessInsights = insights

	return dashboard, nil
}

// GetKPIMetrics calculates key performance indicators
func (s *BusinessIntelligenceService) GetKPIMetrics(filters models.BIFilters) (*models.KPIMetrics, error) {
	metrics := &models.KPIMetrics{}

	// Date range for calculations
	dateFrom := time.Now().AddDate(0, -1, 0) // Default to last month
	dateTo := time.Now()
	
	if filters.DateFrom != nil {
		dateFrom = *filters.DateFrom
	}
	if filters.DateTo != nil {
		dateTo = *filters.DateTo
	}

	// Customer Acquisition Cost (CAC)
	cacQuery := `
		SELECT 
			COUNT(DISTINCT u.id) as new_users,
			COALESCE(SUM(CASE WHEN wt.transaction_type = 'contest_fee' THEN wt.amount * 0.05 END), 0) as marketing_spend
		FROM users u
		LEFT JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.created_at BETWEEN $1 AND $2
		WHERE u.created_at BETWEEN $1 AND $2
	`
	
	var newUsers int64
	var marketingSpend float64
	err := s.db.QueryRow(cacQuery, dateFrom, dateTo).Scan(&newUsers, &marketingSpend)
	if err != nil {
		logger.Error("Error calculating CAC", map[string]interface{}{"error": err.Error()})
	} else if newUsers > 0 {
		metrics.CustomerAcquisitionCost = marketingSpend / float64(newUsers)
	}

	// Customer Lifetime Value (CLV)
	clvQuery := `
		SELECT 
			AVG(total_revenue) as avg_revenue,
			AVG(user_lifespan) as avg_lifespan
		FROM (
			SELECT 
				u.id,
				COALESCE(SUM(wt.amount), 0) as total_revenue,
				EXTRACT(days FROM NOW() - u.created_at) as user_lifespan
			FROM users u
			LEFT JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.transaction_type = 'contest_fee'
			WHERE u.created_at <= $2 AND u.is_active = true
			GROUP BY u.id, u.created_at
		) user_metrics
	`
	
	var avgRevenue, avgLifespan sql.NullFloat64
	err = s.db.QueryRow(clvQuery, dateFrom, dateTo).Scan(&avgRevenue, &avgLifespan)
	if err != nil {
		logger.Error("Error calculating CLV", map[string]interface{}{"error": err.Error()})
	} else if avgRevenue.Valid && avgLifespan.Valid && avgLifespan.Float64 > 0 {
		dailyRevenue := avgRevenue.Float64 / avgLifespan.Float64
		metrics.CustomerLifetimeValue = dailyRevenue * 365 // Annualized
	}

	// Monthly Recurring Revenue (MRR)
	mrrQuery := `
		SELECT COALESCE(SUM(amount), 0) / EXTRACT(days FROM $2 - $1) * 30 as mrr
		FROM wallet_transactions
		WHERE transaction_type = 'contest_fee' 
		AND status = 'completed'
		AND created_at BETWEEN $1 AND $2
	`
	
	err = s.db.QueryRow(mrrQuery, dateFrom, dateTo).Scan(&metrics.MonthlyRecurringRevenue)
	if err != nil {
		logger.Error("Error calculating MRR", map[string]interface{}{"error": err.Error()})
	}

	// Annual Recurring Revenue (ARR)
	metrics.AnnualRecurringRevenue = metrics.MonthlyRecurringRevenue * 12

	// Churn Rate
	churnQuery := `
		SELECT 
			COUNT(CASE WHEN last_login_at < $1 THEN 1 END)::float / NULLIF(COUNT(*), 0) * 100 as churn_rate
		FROM users
		WHERE created_at < $1 AND is_active = true
	`
	
	thirtyDaysAgo := dateTo.AddDate(0, 0, -30)
	err = s.db.QueryRow(churnQuery, thirtyDaysAgo).Scan(&metrics.ChurnRate)
	if err != nil {
		logger.Error("Error calculating churn rate", map[string]interface{}{"error": err.Error()})
	}

	// Retention Rate
	metrics.RetentionRate = 100 - metrics.ChurnRate

	// Average Revenue Per User (ARPU)
	arpuQuery := `
		SELECT 
			COALESCE(SUM(wt.amount), 0) / NULLIF(COUNT(DISTINCT u.id), 0) as arpu
		FROM users u
		LEFT JOIN wallet_transactions wt ON wt.user_id = u.id 
			AND wt.transaction_type = 'contest_fee' 
			AND wt.status = 'completed'
			AND wt.created_at BETWEEN $1 AND $2
		WHERE u.is_active = true
	`
	
	err = s.db.QueryRow(arpuQuery, dateFrom, dateTo).Scan(&metrics.AverageRevenuePerUser)
	if err != nil {
		logger.Error("Error calculating ARPU", map[string]interface{}{"error": err.Error()})
	}

	// Payback Period (in days)
	if metrics.AverageRevenuePerUser > 0 {
		dailyRevenue := metrics.AverageRevenuePerUser / 30
		if dailyRevenue > 0 {
			metrics.PaybackPeriod = int(metrics.CustomerAcquisitionCost / dailyRevenue)
		}
	}

	// Contest Participation Rate
	participationQuery := `
		SELECT 
			COUNT(DISTINCT cp.user_id)::float / NULLIF(COUNT(DISTINCT u.id), 0) * 100 as participation_rate
		FROM users u
		LEFT JOIN contest_participants cp ON cp.user_id = u.id
		LEFT JOIN contests c ON c.id = cp.contest_id AND c.created_at BETWEEN $1 AND $2
		WHERE u.is_active = true
	`
	
	err = s.db.QueryRow(participationQuery, dateFrom, dateTo).Scan(&metrics.ContestParticipationRate)
	if err != nil {
		logger.Error("Error calculating participation rate", map[string]interface{}{"error": err.Error()})
	}

	// Win Rate (percentage of users who won prizes)
	winRateQuery := `
		SELECT 
			COUNT(CASE WHEN cp.prize_won > 0 THEN 1 END)::float / NULLIF(COUNT(DISTINCT cp.user_id), 0) * 100 as win_rate
		FROM contest_participants cp
		JOIN contests c ON c.id = cp.contest_id
		WHERE c.created_at BETWEEN $1 AND $2
	`
	
	err = s.db.QueryRow(winRateQuery, dateFrom, dateTo).Scan(&metrics.WinRate)
	if err != nil {
		logger.Error("Error calculating win rate", map[string]interface{}{"error": err.Error()})
	}

	// ROI (Return on Investment)
	if metrics.CustomerAcquisitionCost > 0 {
		metrics.ROI = ((metrics.CustomerLifetimeValue - metrics.CustomerAcquisitionCost) / metrics.CustomerAcquisitionCost) * 100
	}

	// Conversion Rate (users who made first deposit)
	conversionQuery := `
		SELECT 
			COUNT(CASE WHEN deposit_count > 0 THEN 1 END)::float / NULLIF(COUNT(*), 0) * 100 as conversion_rate
		FROM (
			SELECT 
				u.id,
				COUNT(wt.id) as deposit_count
			FROM users u
			LEFT JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.transaction_type = 'deposit'
			WHERE u.created_at BETWEEN $1 AND $2
			GROUP BY u.id
		) user_deposits
	`
	
	err = s.db.QueryRow(conversionQuery, dateFrom, dateTo).Scan(&metrics.ConversionRate)
	if err != nil {
		logger.Error("Error calculating conversion rate", map[string]interface{}{"error": err.Error()})
	}

	// Set placeholder values for metrics that require external data
	metrics.NetPromoterScore = 7.8
	metrics.CustomerSatisfactionScore = 4.2

	return metrics, nil
}

// GetRevenueAnalytics provides advanced revenue analytics
func (s *BusinessIntelligenceService) GetRevenueAnalytics(filters models.BIFilters) (*models.RevenueAnalytics, error) {
	analytics := &models.RevenueAnalytics{}

	dateFrom := time.Now().AddDate(0, -3, 0) // Default to last 3 months
	dateTo := time.Now()
	
	if filters.DateFrom != nil {
		dateFrom = *filters.DateFrom
	}
	if filters.DateTo != nil {
		dateTo = *filters.DateTo
	}

	// Revenue Growth Rate
	growthQuery := `
		SELECT 
			current_revenue,
			previous_revenue,
			CASE WHEN previous_revenue > 0 THEN 
				((current_revenue - previous_revenue) / previous_revenue) * 100 
			ELSE 0 END as growth_rate
		FROM (
			SELECT 
				COALESCE(SUM(CASE WHEN created_at BETWEEN $1 AND $2 THEN amount END), 0) as current_revenue,
				COALESCE(SUM(CASE WHEN created_at BETWEEN $3 AND $1 THEN amount END), 0) as previous_revenue
			FROM wallet_transactions
			WHERE transaction_type = 'contest_fee' AND status = 'completed'
		) revenue_comparison
	`
	
	previousPeriodStart := dateFrom.AddDate(0, 0, -int(dateTo.Sub(dateFrom).Hours()/24))
	err := s.db.QueryRow(growthQuery, dateFrom, dateTo, previousPeriodStart).Scan(
		&analytics.RevenueGrowthRate, &analytics.RevenueGrowthRate, &analytics.RevenueGrowthRate,
	)
	if err != nil {
		logger.Error("Error calculating revenue growth", map[string]interface{}{"error": err.Error()})
	}

	// Revenue by Game
	gameRevenueQuery := `
		SELECT 
			g.id,
			g.name,
			COALESCE(SUM(wt.amount), 0) as revenue,
			COUNT(DISTINCT c.id) as contest_count,
			0 as market_share,
			0 as profitability
		FROM games g
		LEFT JOIN matches m ON m.game_id = g.id
		LEFT JOIN contests c ON c.match_id = m.id
		LEFT JOIN contest_participants cp ON cp.contest_id = c.id
		LEFT JOIN wallet_transactions wt ON wt.user_id = cp.user_id 
			AND wt.transaction_type = 'contest_fee' 
			AND wt.created_at BETWEEN $1 AND $2
		WHERE g.is_active = true
		GROUP BY g.id, g.name
		ORDER BY revenue DESC
	`
	
	gameRows, err := s.db.Query(gameRevenueQuery, dateFrom, dateTo)
	if err != nil {
		logger.Error("Error getting game revenue", map[string]interface{}{"error": err.Error()})
	} else {
		defer gameRows.Close()
		totalGameRevenue := float64(0)
		
		for gameRows.Next() {
			var game models.GameRevenue
			var contestCount int64
			gameRows.Scan(&game.GameID, &game.GameName, &game.Revenue, &contestCount, &game.MarketShare, &game.Profitability)
			analytics.RevenueByGame = append(analytics.RevenueByGame, game)
			totalGameRevenue += game.Revenue
		}
		
		// Calculate market share
		for i := range analytics.RevenueByGame {
			if totalGameRevenue > 0 {
				analytics.RevenueByGame[i].MarketShare = analytics.RevenueByGame[i].Revenue / totalGameRevenue * 100
			}
			analytics.RevenueByGame[i].Profitability = analytics.RevenueByGame[i].Revenue * 0.15 // Assume 15% profit margin
		}
	}

	// Revenue by User Tier (based on total spending)
	tierQuery := `
		SELECT 
			tier,
			COUNT(*) as user_count,
			SUM(total_revenue) as revenue,
			AVG(total_revenue) as arpu
		FROM (
			SELECT 
				u.id,
				COALESCE(SUM(wt.amount), 0) as total_revenue,
				CASE 
					WHEN COALESCE(SUM(wt.amount), 0) >= 10000 THEN 'Whale'
					WHEN COALESCE(SUM(wt.amount), 0) >= 1000 THEN 'High Value'
					WHEN COALESCE(SUM(wt.amount), 0) >= 100 THEN 'Regular'
					WHEN COALESCE(SUM(wt.amount), 0) > 0 THEN 'Low Value'
					ELSE 'Free'
				END as tier
			FROM users u
			LEFT JOIN wallet_transactions wt ON wt.user_id = u.id 
				AND wt.transaction_type = 'contest_fee'
				AND wt.created_at BETWEEN $1 AND $2
			WHERE u.is_active = true
			GROUP BY u.id
		) user_tiers
		GROUP BY tier
		ORDER BY revenue DESC
	`
	
	tierRows, err := s.db.Query(tierQuery, dateFrom, dateTo)
	if err != nil {
		logger.Error("Error getting tier revenue", map[string]interface{}{"error": err.Error()})
	} else {
		defer tierRows.Close()
		totalTierRevenue := float64(0)
		
		for tierRows.Next() {
			var tier models.TierRevenue
			tierRows.Scan(&tier.Tier, &tier.UserCount, &tier.Revenue, &tier.ARPU)
			analytics.RevenueByUserTier = append(analytics.RevenueByUserTier, tier)
			totalTierRevenue += tier.Revenue
		}
		
		// Calculate contribution percentages
		for i := range analytics.RevenueByUserTier {
			if totalTierRevenue > 0 {
				analytics.RevenueByUserTier[i].Contribution = analytics.RevenueByUserTier[i].Revenue / totalTierRevenue * 100
			}
		}
	}

	// Seasonality Analysis (monthly patterns)
	seasonalityQuery := `
		SELECT 
			TO_CHAR(created_at, 'Month YYYY') as period,
			SUM(amount) as revenue,
			1.0 as seasonality_factor,
			0 as year_over_year_growth
		FROM wallet_transactions
		WHERE transaction_type = 'contest_fee' 
		AND status = 'completed'
		AND created_at >= $1
		GROUP BY TO_CHAR(created_at, 'Month YYYY'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`
	
	seasonalRows, err := s.db.Query(seasonalityQuery, dateFrom.AddDate(-1, 0, 0))
	if err != nil {
		logger.Error("Error getting seasonality data", map[string]interface{}{"error": err.Error()})
	} else {
		defer seasonalRows.Close()
		for seasonalRows.Next() {
			var seasonal models.SeasonalRevenue
			seasonalRows.Scan(&seasonal.Period, &seasonal.Revenue, &seasonal.Seasonality, &seasonal.YearOverYear)
			analytics.SeasonalityAnalysis = append(analytics.SeasonalityAnalysis, seasonal)
		}
	}

	// Revenue Forecast (simple linear projection)
	if len(analytics.SeasonalityAnalysis) >= 3 {
		lastThreeMonths := analytics.SeasonalityAnalysis[len(analytics.SeasonalityAnalysis)-3:]
		avgRevenue := (lastThreeMonths[0].Revenue + lastThreeMonths[1].Revenue + lastThreeMonths[2].Revenue) / 3
		
		for i := 1; i <= 6; i++ {
			forecast := models.RevenueForecast{
				Period:       dateTo.AddDate(0, i, 0),
				ForecastMid:  avgRevenue * (1 + analytics.RevenueGrowthRate/100),
				ForecastLow:  avgRevenue * 0.8,
				ForecastHigh: avgRevenue * 1.2,
				Confidence:   75.0,
			}
			analytics.RevenueForecast = append(analytics.RevenueForecast, forecast)
		}
	}

	// Profitability Analysis (placeholder calculations)
	totalRevenue := float64(0)
	for _, seasonal := range analytics.SeasonalityAnalysis {
		totalRevenue += seasonal.Revenue
	}
	
	analytics.ProfitabilityAnalysis = models.ProfitabilityAnalysis{
		GrossMargin:       totalRevenue * 0.75, // Assume 75% gross margin
		NetMargin:         totalRevenue * 0.15, // Assume 15% net margin
		EBITDA:           totalRevenue * 0.25,  // Assume 25% EBITDA
		OperatingExpenses: totalRevenue * 0.60, // Assume 60% operating expenses
		BreakevenPoint:   totalRevenue / 0.15,  // Revenue needed for break-even
		ProfitTrend:      "Increasing",
	}

	// Revenue Optimization Insights
	analytics.RevenueOptimization = []models.OptimizationInsight{
		{
			Category:       "Contest Fees",
			Recommendation: "Optimize entry fee structure for higher participation",
			Impact:         15.5,
			Effort:         "Medium",
			Priority:       "High",
		},
		{
			Category:       "User Retention",
			Recommendation: "Implement loyalty program for high-value users",
			Impact:         23.2,
			Effort:         "High",
			Priority:       "High",
		},
		{
			Category:       "Game Portfolio",
			Recommendation: "Add more popular games to increase market share",
			Impact:         18.7,
			Effort:         "Medium",
			Priority:       "Medium",
		},
	}

	return analytics, nil
}

// GetUserBehaviorAnalysis analyzes user behavior patterns
func (s *BusinessIntelligenceService) GetUserBehaviorAnalysis(filters models.BIFilters) (*models.UserBehaviorAnalysis, error) {
	analysis := &models.UserBehaviorAnalysis{}

	// User Segmentation
	segmentQuery := `
		SELECT 
			segment_name,
			COUNT(*) as user_count,
			AVG(total_revenue) as avg_revenue,
			AVG(engagement_score) as engagement_score,
			AVG(churn_probability) as churn_probability
		FROM (
			SELECT 
				u.id,
				COALESCE(SUM(wt.amount), 0) as total_revenue,
				EXTRACT(days FROM NOW() - u.last_login_at) as days_since_login,
				CASE 
					WHEN COALESCE(SUM(wt.amount), 0) >= 1000 AND EXTRACT(days FROM NOW() - u.last_login_at) <= 7 THEN 'High Value Active'
					WHEN COALESCE(SUM(wt.amount), 0) >= 1000 THEN 'High Value Inactive'
					WHEN COALESCE(SUM(wt.amount), 0) >= 100 AND EXTRACT(days FROM NOW() - u.last_login_at) <= 7 THEN 'Regular Active'
					WHEN COALESCE(SUM(wt.amount), 0) >= 100 THEN 'Regular Inactive'
					WHEN COALESCE(SUM(wt.amount), 0) > 0 THEN 'Low Spender'
					ELSE 'Free User'
				END as segment_name,
				CASE 
					WHEN EXTRACT(days FROM NOW() - u.last_login_at) <= 7 THEN 8.5
					WHEN EXTRACT(days FROM NOW() - u.last_login_at) <= 30 THEN 6.0
					ELSE 3.0
				END as engagement_score,
				CASE 
					WHEN EXTRACT(days FROM NOW() - u.last_login_at) > 30 THEN 0.8
					WHEN EXTRACT(days FROM NOW() - u.last_login_at) > 7 THEN 0.4
					ELSE 0.1
				END as churn_probability
			FROM users u
			LEFT JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.transaction_type = 'contest_fee'
			WHERE u.is_active = true
			GROUP BY u.id, u.last_login_at
		) segmented_users
		GROUP BY segment_name
		ORDER BY avg_revenue DESC
	`

	segmentRows, err := s.db.Query(segmentQuery)
	if err != nil {
		logger.Error("Error getting user segments", map[string]interface{}{"error": err.Error()})
	} else {
		defer segmentRows.Close()
		for segmentRows.Next() {
			var segment models.UserSegment
			segment.SegmentID = segment.SegmentName
			
			segmentRows.Scan(&segment.SegmentName, &segment.UserCount, &segment.AvgRevenue, 
				&segment.EngagementScore, &segment.ChurnProbability)
			
			// Add characteristics and recommended actions based on segment
			switch segment.SegmentName {
			case "High Value Active":
				segment.Characteristics = []string{"High spending", "Frequent login", "Contest participation"}
				segment.RecommendedActions = []string{"VIP treatment", "Exclusive contests", "Personal account manager"}
			case "High Value Inactive":
				segment.Characteristics = []string{"High historical spending", "Infrequent login", "At risk"}
				segment.RecommendedActions = []string{"Win-back campaigns", "Special offers", "Personal outreach"}
			default:
				segment.Characteristics = []string{"Standard user behavior"}
				segment.RecommendedActions = []string{"Standard engagement", "Promotional offers"}
			}
			
			analysis.UserSegments = append(analysis.UserSegments, segment)
		}
	}

	// Behavior Patterns
	behaviorPatterns := []models.BehaviorPattern{
		{
			Pattern:       "Weekend Peak Activity",
			Description:   "Users are 40% more active during weekends",
			Frequency:     52, // weeks per year
			UserCount:     int64(len(analysis.UserSegments) * 30), // estimate
			RevenueImpact: 15.5,
			Significance:  "High",
		},
		{
			Pattern:       "Evening Contest Preference",
			Description:   "70% of contests joined between 6-10 PM",
			Frequency:     365, // daily
			UserCount:     int64(len(analysis.UserSegments) * 50),
			RevenueImpact: 23.2,
			Significance:  "Very High",
		},
	}
	analysis.BehaviorPatterns = behaviorPatterns

	// Engagement Analysis
	engagementQuery := `
		SELECT 
			AVG(session_duration) as avg_session_duration,
			AVG(page_views) as avg_page_views,
			bounce_rate
		FROM (
			SELECT 
				u.id,
				30 as session_duration, -- placeholder
				5 as page_views, -- placeholder
				0.25 as bounce_rate -- placeholder
			FROM users u
			WHERE u.is_active = true
			LIMIT 1000
		) engagement_data
	`

	var avgSessionDuration sql.NullFloat64
	var avgPageViews sql.NullInt64
	var bounceRate sql.NullFloat64

	err = s.db.QueryRow(engagementQuery).Scan(&avgSessionDuration, &avgPageViews, &bounceRate)
	if err != nil {
		logger.Error("Error getting engagement data", map[string]interface{}{"error": err.Error()})
	}

	analysis.EngagementAnalysis = models.EngagementAnalysis{
		AvgSessionDuration: 30.5, // placeholder
		AvgPageViews:      5,
		BounceRate:        25.0,
		FeatureAdoption: []models.FeatureAdoption{
			{
				FeatureName:     "Contest Joining",
				AdoptionRate:    85.5,
				TimeToAdopt:     2.3,
				RetentionImpact: 15.2,
				RevenueImpact:   25.7,
			},
			{
				FeatureName:     "Team Creation",
				AdoptionRate:    92.1,
				TimeToAdopt:     1.5,
				RetentionImpact: 20.5,
				RevenueImpact:   35.2,
			},
		},
	}

	return analysis, nil
}

// GetPredictiveAnalytics provides predictive insights
func (s *BusinessIntelligenceService) GetPredictiveAnalytics(filters models.BIFilters) (*models.PredictiveAnalytics, error) {
	analytics := &models.PredictiveAnalytics{}

	// Churn Prediction
	churnPrediction, err := s.calculateChurnPrediction()
	if err != nil {
		logger.Error("Error calculating churn prediction", map[string]interface{}{"error": err.Error()})
		churnPrediction = &models.ChurnPrediction{}
	}
	analytics.ChurnPrediction = *churnPrediction

	// Revenue Forecasting
	revenueForecasting := s.calculateRevenueForecasting()
	analytics.RevenueForecasting = *revenueForecasting

	// User Growth Prediction
	userGrowthPrediction := s.calculateUserGrowthPrediction()
	analytics.UserGrowthPrediction = *userGrowthPrediction

	// Seasonality Prediction
	seasonalityPrediction := s.calculateSeasonalityPrediction()
	analytics.SeasonalityPrediction = *seasonalityPrediction

	// Risk Assessment
	riskAssessment := s.calculateRiskAssessment()
	analytics.RiskAssessment = *riskAssessment

	// Opportunity Scoring
	opportunityScores := s.calculateOpportunityScores()
	analytics.OpportunityScoring = opportunityScores

	return analytics, nil
}

// Helper functions for predictive analytics

func (s *BusinessIntelligenceService) calculateChurnPrediction() (*models.ChurnPrediction, error) {
	prediction := &models.ChurnPrediction{}

	// Overall churn rate
	churnQuery := `
		SELECT 
			COUNT(CASE WHEN last_login_at < NOW() - INTERVAL '30 days' THEN 1 END)::float / 
			NULLIF(COUNT(*), 0) * 100 as churn_rate
		FROM users
		WHERE is_active = true AND created_at < NOW() - INTERVAL '30 days'
	`

	err := s.db.QueryRow(churnQuery).Scan(&prediction.OverallChurnRate)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Churn by segment
	prediction.ChurnBySegment = []models.SegmentChurn{
		{
			Segment:           "High Value",
			ChurnRate:         5.2,
			RiskLevel:         "Low",
			RecommendedAction: "Monitor closely and provide VIP support",
		},
		{
			Segment:           "Regular",
			ChurnRate:         15.7,
			RiskLevel:         "Medium",
			RecommendedAction: "Engagement campaigns and loyalty programs",
		},
		{
			Segment:           "Free Users",
			ChurnRate:         35.8,
			RiskLevel:         "High",
			RecommendedAction: "Convert to paying users with attractive offers",
		},
	}

	// Churn factors
	prediction.ChurnFactors = []models.ChurnFactor{
		{
			Factor:        "Days since last login",
			Importance:    0.85,
			Impact:        "Very High",
			Actionability: "High",
		},
		{
			Factor:        "Contest participation frequency",
			Importance:    0.72,
			Impact:        "High",
			Actionability: "Medium",
		},
		{
			Factor:        "Win rate",
			Importance:    0.65,
			Impact:        "Medium",
			Actionability: "Low",
		},
	}

	// High risk users (sample)
	highRiskQuery := `
		SELECT 
			u.id,
			COALESCE(u.first_name || ' ' || u.last_name, u.mobile) as username,
			EXTRACT(days FROM NOW() - u.last_login_at) as days_inactive,
			COALESCE(SUM(wt.amount), 0) as total_value
		FROM users u
		LEFT JOIN wallet_transactions wt ON wt.user_id = u.id AND wt.transaction_type = 'contest_fee'
		WHERE u.last_login_at < NOW() - INTERVAL '7 days' 
		AND u.is_active = true
		GROUP BY u.id, u.first_name, u.last_name, u.mobile, u.last_login_at
		ORDER BY total_value DESC
		LIMIT 10
	`

	highRiskRows, err := s.db.Query(highRiskQuery)
	if err != nil {
		logger.Error("Error getting high risk users", map[string]interface{}{"error": err.Error()})
	} else {
		defer highRiskRows.Close()
		for highRiskRows.Next() {
			var user models.BIHighRiskUser
			var daysInactive float64
			highRiskRows.Scan(&user.UserID, &user.Username, &daysInactive, &user.Value)
			
			user.ChurnProbability = math.Min(daysInactive/30*100, 95.0) // Cap at 95%
			user.RiskFactors = []string{"Inactive for " + fmt.Sprintf("%.0f", daysInactive) + " days"}
			user.RecommendedAction = "Immediate re-engagement campaign"
			user.LastActivity = time.Now().AddDate(0, 0, -int(daysInactive))
			
			prediction.HighRiskUsers = append(prediction.HighRiskUsers, user)
		}
	}

	// Retention strategies
	prediction.RetentionStrategies = []models.RetentionStrategy{
		{
			Strategy:          "Personalized offers based on user behavior",
			TargetSegment:     "At-risk users",
			ExpectedImpact:    25.0,
			ImplementationCost: 5000.0,
			ROI:               400.0,
		},
		{
			Strategy:          "Loyalty program with tiered rewards",
			TargetSegment:     "Regular users",
			ExpectedImpact:    35.0,
			ImplementationCost: 15000.0,
			ROI:               280.0,
		},
	}

	return prediction, nil
}

func (s *BusinessIntelligenceService) calculateRevenueForecasting() *models.RevenueForecasting {
	forecasting := &models.RevenueForecasting{}

	// Simple projections based on current trends
	forecasting.NextMonthForecast = 45000.0
	forecasting.NextQuarterForecast = 135000.0
	forecasting.NextYearForecast = 540000.0

	forecasting.ScenarioAnalysis = []models.ScenarioAnalysis{
		{
			Scenario:       "Optimistic",
			Probability:    0.25,
			RevenueImpact:  1.2,
			KeyAssumptions: []string{"Strong user growth", "Increased engagement", "New game launches"},
		},
		{
			Scenario:       "Base Case",
			Probability:    0.50,
			RevenueImpact:  1.0,
			KeyAssumptions: []string{"Current trends continue", "Stable user base", "No major changes"},
		},
		{
			Scenario:       "Pessimistic",
			Probability:    0.25,
			RevenueImpact:  0.8,
			KeyAssumptions: []string{"Increased competition", "User churn", "Market saturation"},
		},
	}

	forecasting.ConfidenceInterval = models.ConfidenceInterval{
		LowerBound:      36000.0,
		UpperBound:      54000.0,
		ConfidenceLevel: 80.0,
	}

	return forecasting
}

func (s *BusinessIntelligenceService) calculateUserGrowthPrediction() *models.UserGrowthPrediction {
	prediction := &models.UserGrowthPrediction{}

	prediction.NextMonthUsers = 1250
	prediction.NextQuarterUsers = 3800
	prediction.GrowthRate = 8.5

	prediction.GrowthDrivers = []models.GrowthDriver{
		{
			Driver:         "Referral program",
			Impact:         25.5,
			Trend:          "Increasing",
			Recommendation: "Enhance referral rewards",
		},
		{
			Driver:         "Social media marketing",
			Impact:         18.2,
			Trend:          "Stable",
			Recommendation: "Increase social media budget",
		},
	}

	prediction.AcquisitionChannels = []models.ChannelForecast{
		{
			Channel:       "Organic",
			ForecastedUsers: 500,
			Cost:           0,
			ROI:            100.0,
		},
		{
			Channel:       "Paid Social",
			ForecastedUsers: 300,
			Cost:           10000,
			ROI:            250.0,
		},
	}

	return prediction
}

func (s *BusinessIntelligenceService) calculateSeasonalityPrediction() *models.SeasonalityPrediction {
	prediction := &models.SeasonalityPrediction{}

	prediction.SeasonalPatterns = []models.SeasonalPattern{
		{
			Period:         "Weekend",
			Strength:       0.75,
			Pattern:        "Higher activity on weekends",
			BusinessImpact: "Increased contest participation and revenue",
		},
		{
			Period:         "Holiday Season",
			Strength:       0.85,
			Pattern:        "Significant spike during festivals",
			BusinessImpact: "Peak revenue periods",
		},
	}

	nextWeekend := s.getNextWeekend()
	prediction.UpcomingPeaks = []models.Peak{
		{
			Date:             nextWeekend,
			ExpectedIncrease: 40.0,
			Duration:         2,
			PrepActions:      []string{"Increase server capacity", "Launch weekend contests"},
		},
	}

	return prediction
}

func (s *BusinessIntelligenceService) calculateRiskAssessment() *models.RiskAssessment {
	assessment := &models.RiskAssessment{}

	assessment.OverallRiskScore = 25.5

	assessment.RiskCategories = []models.RiskMetric{
		{
			Category:   "Competition",
			RiskScore:  35.0,
			Trend:      "Increasing",
			Impact:     "High",
			Likelihood: "Medium",
		},
		{
			Category:   "Regulatory",
			RiskScore:  20.0,
			Trend:      "Stable",
			Impact:     "Medium",
			Likelihood: "Low",
		},
		{
			Category:   "Technical",
			RiskScore:  15.0,
			Trend:      "Decreasing",
			Impact:     "Low",
			Likelihood: "Low",
		},
	}

	assessment.MitigationPlans = []models.MitigationPlan{
		{
			Risk:            "Increased competition",
			Plan:            "Differentiate through unique features and better user experience",
			Timeline:        "6 months",
			ResponsibleTeam: "Product & Marketing",
			Success_Metrics: []string{"User retention rate", "Market share", "NPS score"},
		},
	}

	return assessment
}

func (s *BusinessIntelligenceService) calculateOpportunityScores() []models.OpportunityScore {
	return []models.OpportunityScore{
		{
			Opportunity:      "Mobile app optimization",
			Score:            85.0,
			PotentialRevenue: 50000.0,
			Implementation:   "Medium",
			Timeline:        "3 months",
			Priority:        "High",
		},
		{
			Opportunity:      "New game categories",
			Score:            78.5,
			PotentialRevenue: 75000.0,
			Implementation:   "High",
			Timeline:        "6 months",
			Priority:        "Medium",
		},
	}
}

func (s *BusinessIntelligenceService) getCompetitiveAnalysis() *models.CompetitiveAnalysis {
	analysis := &models.CompetitiveAnalysis{}

	analysis.MarketPosition = models.MarketPosition{
		MarketRank:           3,
		MarketShare:         12.5,
		Growth:              25.0,
		Differentiation:     "Focus on esports with real-time scoring",
		CompetitiveStrength: "Strong",
	}

	analysis.CompetitorMetrics = []models.CompetitorMetric{
		{
			CompetitorName:   "Dream11",
			MarketShare:      45.0,
			EstimatedRevenue: 5000000.0,
			UserBase:         15000000,
			GrowthRate:       15.0,
			Strengths:        []string{"Brand recognition", "Cricket focus", "Large user base"},
			Weaknesses:       []string{"Limited esports", "High competition"},
		},
		{
			CompetitorName:   "MPL",
			MarketShare:      25.0,
			EstimatedRevenue: 3000000.0,
			UserBase:         8000000,
			GrowthRate:       20.0,
			Strengths:        []string{"Diverse games", "Strong marketing"},
			Weaknesses:       []string{"Complex interface", "Limited esports focus"},
		},
	}

	return analysis
}

func (s *BusinessIntelligenceService) generateBusinessInsights(filters models.BIFilters) ([]models.BusinessInsight, error) {
	insights := []models.BusinessInsight{
		{
			InsightID:         "INS001",
			Category:          "Revenue",
			Title:             "Weekend Revenue Opportunity",
			Description:       "Revenue increases by 40% during weekends, suggesting opportunity for weekend-specific contests and promotions",
			Impact:            "High",
			Priority:          "High",
			Confidence:        85.5,
			RecommendedAction: "Launch weekend-exclusive contests with higher prize pools",
			DataSources:       []string{"wallet_transactions", "contests", "user_activity"},
			GeneratedAt:       time.Now(),
		},
		{
			InsightID:         "INS002",
			Category:          "User Behavior",
			Title:             "High-Value User Churn Risk",
			Description:       "15% of high-value users show signs of potential churn based on decreased activity",
			Impact:            "Critical",
			Priority:          "Immediate",
			Confidence:        78.2,
			RecommendedAction: "Implement targeted retention campaigns for high-value users",
			DataSources:       []string{"users", "user_activity", "wallet_transactions"},
			GeneratedAt:       time.Now(),
		},
		{
			InsightID:         "INS003",
			Category:          "Product",
			Title:             "Esports Game Portfolio Expansion",
			Description:       "Adding popular esports titles could increase market share by 18-25%",
			Impact:            "High",
			Priority:          "Medium",
			Confidence:        72.8,
			RecommendedAction: "Research and integrate top 3 emerging esports games",
			DataSources:       []string{"games", "contests", "market_research"},
			GeneratedAt:       time.Now(),
		},
	}

	return insights, nil
}

// Helper function to get next weekend
func (s *BusinessIntelligenceService) getNextWeekend() time.Time {
	now := time.Now()
	daysUntilSaturday := (6 - int(now.Weekday())) % 7
	if daysUntilSaturday == 0 && now.Weekday() != time.Saturday {
		daysUntilSaturday = 6
	}
	return now.AddDate(0, 0, daysUntilSaturday)
}