package models

import (
	"time"
)

// Analytics Dashboard Models
type AnalyticsDashboard struct {
	UserMetrics      UserMetrics      `json:"user_metrics"`
	RevenueMetrics   RevenueMetrics   `json:"revenue_metrics"`
	ContestMetrics   ContestMetrics   `json:"contest_metrics"`
	GameMetrics      GameMetrics      `json:"game_metrics"`
	EngagementMetrics EngagementMetrics `json:"engagement_metrics"`
	SystemHealth     SystemHealth     `json:"system_health"`
	TopPerformers    TopPerformers    `json:"top_performers"`
}

// User Analytics
type UserMetrics struct {
	TotalUsers           int64           `json:"total_users"`
	ActiveUsers          int64           `json:"active_users"`
	NewUsersToday        int64           `json:"new_users_today"`
	NewUsersThisWeek     int64           `json:"new_users_this_week"`
	NewUsersThisMonth    int64           `json:"new_users_this_month"`
	VerifiedUsers        int64           `json:"verified_users"`
	KYCCompletedUsers    int64           `json:"kyc_completed_users"`
	UserGrowthRate       float64         `json:"user_growth_rate"`
	UserRetentionRate    float64         `json:"user_retention_rate"`
	UsersByState         []UsersByRegion `json:"users_by_state"`
	UserRegistrationTrend []DailyCount   `json:"user_registration_trend"`
}

type UsersByRegion struct {
	Region string `json:"region"`
	Count  int64  `json:"count"`
}

type DailyCount struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

// Revenue Analytics  
type RevenueMetrics struct {
	TotalRevenue          float64             `json:"total_revenue"`
	RevenueToday          float64             `json:"revenue_today"`
	RevenueThisWeek       float64             `json:"revenue_this_week"`
	RevenueThisMonth      float64             `json:"revenue_this_month"`
	RevenueGrowthRate     float64             `json:"revenue_growth_rate"`
	AvgRevenuePerUser     float64             `json:"avg_revenue_per_user"`
	TotalDeposits         float64             `json:"total_deposits"`
	TotalWithdrawals      float64             `json:"total_withdrawals"`
	PendingWithdrawals    float64             `json:"pending_withdrawals"`
	RevenueByGame         []RevenueByCategory `json:"revenue_by_game"`
	MonthlyRevenueTrend   []MonthlyRevenue    `json:"monthly_revenue_trend"`
	PaymentMethodDistribution []PaymentMethodStats `json:"payment_method_distribution"`
}

type RevenueByCategory struct {
	Category string  `json:"category"`
	Revenue  float64 `json:"revenue"`
}

type MonthlyRevenue struct {
	Month   time.Time `json:"month"`
	Revenue float64   `json:"revenue"`
}

type PaymentMethodStats struct {
	Method string  `json:"method"`
	Amount float64 `json:"amount"`
	Count  int64   `json:"count"`
}

// Contest Analytics
type ContestMetrics struct {
	TotalContests           int64                `json:"total_contests"`
	ActiveContests          int64                `json:"active_contests"`
	CompletedContests       int64                `json:"completed_contests"`
	TotalParticipations     int64                `json:"total_participations"`
	AvgParticipationsPerContest float64         `json:"avg_participations_per_contest"`
	TotalPrizeDistributed   float64              `json:"total_prize_distributed"`
	ContestsByEntryFee      []ContestsByCategory `json:"contests_by_entry_fee"`
	PopularContestTypes     []PopularContest     `json:"popular_contest_types"`
	ContestCompletionRate   float64              `json:"contest_completion_rate"`
}

type ContestsByCategory struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type PopularContest struct {
	ContestType     string `json:"contest_type"`
	ParticipantCount int64 `json:"participant_count"`
}

// Game Analytics
type GameMetrics struct {
	TotalGames            int64          `json:"total_games"`
	ActiveGames           int64          `json:"active_games"`
	MostPopularGame       string         `json:"most_popular_game"`
	GameParticipation     []GameStats    `json:"game_participation"`
	PlayerPerformance     []TopPlayer    `json:"player_performance"`
	MatchesCompleted      int64          `json:"matches_completed"`
	UpcomingMatches       int64          `json:"upcoming_matches"`
	LiveMatches           int64          `json:"live_matches"`
}

type GameStats struct {
	GameName        string  `json:"game_name"`
	TotalMatches    int64   `json:"total_matches"`
	TotalPlayers    int64   `json:"total_players"`
	AvgParticipation float64 `json:"avg_participation"`
}

type TopPlayer struct {
	PlayerName    string  `json:"player_name"`
	GameName      string  `json:"game_name"`
	TeamName      string  `json:"team_name"`
	AvgPoints     float64 `json:"avg_points"`
	TotalMatches  int64   `json:"total_matches"`
	PopularityScore float64 `json:"popularity_score"`
}

// Engagement Analytics
type EngagementMetrics struct {
	DailyActiveUsers       int64             `json:"daily_active_users"`
	WeeklyActiveUsers      int64             `json:"weekly_active_users"`
	MonthlyActiveUsers     int64             `json:"monthly_active_users"`
	AvgSessionDuration     float64           `json:"avg_session_duration"`
	UserRetentionDay7      float64           `json:"user_retention_day7"`
	UserRetentionDay30     float64           `json:"user_retention_day30"`
	FeatureUsage           []FeatureUsage    `json:"feature_usage"`
	PeakUsageHours         []HourlyUsage     `json:"peak_usage_hours"`
	ChurnRate              float64           `json:"churn_rate"`
}

type FeatureUsage struct {
	FeatureName string  `json:"feature_name"`
	UsageCount  int64   `json:"usage_count"`
	UsageRate   float64 `json:"usage_rate"`
}

type HourlyUsage struct {
	Hour  int   `json:"hour"`
	Users int64 `json:"users"`
}

// System Health
type SystemHealth struct {
	DatabaseHealth      string             `json:"database_health"`
	APIResponseTime     float64            `json:"api_response_time"`
	ErrorRate           float64            `json:"error_rate"`
	ActiveConnections   int64              `json:"active_connections"`
	SystemUptime        string             `json:"system_uptime"`
	MemoryUsage         float64            `json:"memory_usage"`
	CPUUsage            float64            `json:"cpu_usage"`
	RecentErrors        []SystemError      `json:"recent_errors"`
}

type SystemError struct {
	ErrorCode   string    `json:"error_code"`
	Message     string    `json:"message"`
	Count       int64     `json:"count"`
	LastOccurred time.Time `json:"last_occurred"`
}

// Top Performers
type TopPerformers struct {
	TopEarners         []TopEarner         `json:"top_earners"`
	TopReferrers       []TopReferrer       `json:"top_referrers"`
	MostActiveUsers    []MostActiveUser    `json:"most_active_users"`
	BiggestWinners     []BiggestWinner     `json:"biggest_winners"`
}

type TopEarner struct {
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	TotalEarnings float64 `json:"total_earnings"`
	ContestsWon int64   `json:"contests_won"`
}

type TopReferrer struct {
	UserID          int64   `json:"user_id"`
	Username        string  `json:"username"`
	ReferralsCount  int64   `json:"referrals_count"`
	EarningsFromReferrals float64 `json:"earnings_from_referrals"`
}

type MostActiveUser struct {
	UserID          int64 `json:"user_id"`
	Username        string `json:"username"`
	ContestsJoined  int64 `json:"contests_joined"`
	TeamsCreated    int64 `json:"teams_created"`
	LoginFrequency  float64 `json:"login_frequency"`
}

type BiggestWinner struct {
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	ContestName string  `json:"contest_name"`
	PrizeWon    float64 `json:"prize_won"`
	WinDate     time.Time `json:"win_date"`
}

// Analytics Filter Request
type AnalyticsFilters struct {
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	GameID      *int       `json:"game_id"`
	Period      string     `json:"period"` // day, week, month, year
	Granularity string     `json:"granularity"` // hour, day, week, month
}

// Detailed Analytics Requests
type UserAnalyticsRequest struct {
	Filters AnalyticsFilters `json:"filters"`
	Metrics []string         `json:"metrics"` // specific metrics to include
}

type RevenueAnalyticsRequest struct {
	Filters    AnalyticsFilters `json:"filters"`
	BreakdownBy string          `json:"breakdown_by"` // game, region, payment_method
}

type ContestAnalyticsRequest struct {
	Filters AnalyticsFilters `json:"filters"`
	ContestType *string       `json:"contest_type"`
	EntryFeeRange *struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"entry_fee_range"`
}

// Real-time Analytics
type RealTimeMetrics struct {
	ActiveUsers         int64             `json:"active_users"`
	LiveContests        int64             `json:"live_contests"`
	ActiveMatches       int64             `json:"active_matches"`
	TransactionsPerMinute int64           `json:"transactions_per_minute"`
	NewRegistrations    int64             `json:"new_registrations"`
	CurrentRevenue      float64           `json:"current_revenue"`
	SystemLoad          float64           `json:"system_load"`
	LastUpdated         time.Time         `json:"last_updated"`
}

// Performance Metrics
type PerformanceMetrics struct {
	APIEndpoints      []APIPerformance    `json:"api_endpoints"`
	DatabaseQueries   []QueryPerformance  `json:"database_queries"`
	CacheHitRate      float64             `json:"cache_hit_rate"`
	AverageResponseTime float64           `json:"average_response_time"`
	P95ResponseTime   float64             `json:"p95_response_time"`
	P99ResponseTime   float64             `json:"p99_response_time"`
	ErrorRate         float64             `json:"error_rate"`
	ThroughputPerSecond float64           `json:"throughput_per_second"`
}

type APIPerformance struct {
	Endpoint        string  `json:"endpoint"`
	Method          string  `json:"method"`
	AvgResponseTime float64 `json:"avg_response_time"`
	RequestCount    int64   `json:"request_count"`
	ErrorCount      int64   `json:"error_count"`
	ErrorRate       float64 `json:"error_rate"`
}

type QueryPerformance struct {
	QueryType       string  `json:"query_type"`
	AvgExecutionTime float64 `json:"avg_execution_time"`
	ExecutionCount  int64   `json:"execution_count"`
	SlowQueryCount  int64   `json:"slow_query_count"`
}

// Cohort Analysis
type CohortAnalysis struct {
	CohortData    []CohortData    `json:"cohort_data"`
	RetentionData []RetentionData `json:"retention_data"`
	Period        string          `json:"period"`
	GeneratedAt   time.Time       `json:"generated_at"`
}

type CohortData struct {
	CohortMonth   time.Time `json:"cohort_month"`
	UserCount     int64     `json:"user_count"`
	RevenuePerUser float64  `json:"revenue_per_user"`
}

type RetentionData struct {
	CohortMonth time.Time `json:"cohort_month"`
	Period      int       `json:"period"` // weeks or months since cohort start
	RetainedUsers int64   `json:"retained_users"`
	RetentionRate float64 `json:"retention_rate"`
}

// Funnel Analysis
type FunnelAnalysis struct {
	FunnelSteps []FunnelStep `json:"funnel_steps"`
	ConversionRate float64   `json:"overall_conversion_rate"`
	AnalysisPeriod string    `json:"analysis_period"`
	GeneratedAt    time.Time `json:"generated_at"`
}

type AnalyticsFunnelStep struct {
	StepName     string  `json:"step_name"`
	StepOrder    int     `json:"step_order"`
	UserCount    int64   `json:"user_count"`
	ConversionRate float64 `json:"conversion_rate"`
	DropoffRate  float64 `json:"dropoff_rate"`
}