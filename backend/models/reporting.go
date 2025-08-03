package models

import (
	"time"
)

// Report Types
type ReportType string

const (
	ReportTypeFinancial      ReportType = "financial"
	ReportTypeUser          ReportType = "user"
	ReportTypeContest       ReportType = "contest"
	ReportTypeGame          ReportType = "game"
	ReportTypePerformance   ReportType = "performance"
	ReportTypeCompliance    ReportType = "compliance"
	ReportTypeEngagement    ReportType = "engagement"
	ReportTypeReferral      ReportType = "referral"
)

// Report Status
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted ReportStatus = "completed"
	ReportStatusFailed    ReportStatus = "failed"
)

// Report Format
type ReportFormat string

const (
	ReportFormatJSON ReportFormat = "json"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatExcel ReportFormat = "excel"
)

// Report Request
type ReportRequest struct {
	ReportType  ReportType   `json:"report_type" validate:"required"`
	Format      ReportFormat `json:"format" validate:"required"`
	DateFrom    time.Time    `json:"date_from" validate:"required"`
	DateTo      time.Time    `json:"date_to" validate:"required"`
	Filters     ReportFilters `json:"filters"`
	EmailTo     []string     `json:"email_to,omitempty"`
	Description string       `json:"description,omitempty"`
}

// Report Filters
type ReportFilters struct {
	GameID       *int       `json:"game_id,omitempty"`
	UserID       *int64     `json:"user_id,omitempty"`
	ContestID    *int64     `json:"contest_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	MinAmount    *float64   `json:"min_amount,omitempty"`
	MaxAmount    *float64   `json:"max_amount,omitempty"`
	Region       *string    `json:"region,omitempty"`
	KYCStatus    *string    `json:"kyc_status,omitempty"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
}

// Generated Report
type GeneratedReport struct {
	ID          int64        `json:"id" db:"id"`
	ReportType  ReportType   `json:"report_type" db:"report_type"`
	Format      ReportFormat `json:"format" db:"format"`
	Status      ReportStatus `json:"status" db:"status"`
	Title       string       `json:"title" db:"title"`
	Description *string      `json:"description" db:"description"`
	FilePath    *string      `json:"file_path" db:"file_path"`
	FileSize    *int64       `json:"file_size" db:"file_size"`
	GeneratedBy int64        `json:"generated_by" db:"generated_by"`
	RequestData ReportRequest `json:"request_data" db:"request_data"`
	ResultData  interface{}  `json:"result_data" db:"result_data"`
	ErrorMessage *string     `json:"error_message" db:"error_message"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	CompletedAt *time.Time   `json:"completed_at" db:"completed_at"`
	ExpiresAt   *time.Time   `json:"expires_at" db:"expires_at"`
}

// Report Summary
type ReportSummary struct {
	ID          int64        `json:"id"`
	ReportType  ReportType   `json:"report_type"`
	Format      ReportFormat `json:"format"`
	Status      ReportStatus `json:"status"`
	Title       string       `json:"title"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at"`
	FileSize    *int64       `json:"file_size"`
}

// Financial Report Data
type FinancialReport struct {
	Summary          FinancialSummary     `json:"summary"`
	TransactionsByType []TransactionSummary `json:"transactions_by_type"`
	DailyRevenue     []DailyRevenue       `json:"daily_revenue"`
	TopSpenders      []TopSpender         `json:"top_spenders"`
	PaymentMethods   []PaymentMethodReport `json:"payment_methods"`
	TaxSummary       TaxSummary           `json:"tax_summary"`
}

type FinancialSummary struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalDeposits     float64 `json:"total_deposits"`
	TotalWithdrawals  float64 `json:"total_withdrawals"`
	PendingWithdrawals float64 `json:"pending_withdrawals"`
	NetRevenue        float64 `json:"net_revenue"`
	TransactionCount  int64   `json:"transaction_count"`
	AvgTransactionValue float64 `json:"avg_transaction_value"`
}

type TransactionSummary struct {
	Type        string  `json:"type"`
	Count       int64   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	AvgAmount   float64 `json:"avg_amount"`
}

type DailyRevenue struct {
	Date     time.Time `json:"date"`
	Revenue  float64   `json:"revenue"`
	Deposits float64   `json:"deposits"`
	Withdrawals float64 `json:"withdrawals"`
}

type TopSpender struct {
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	TotalSpent  float64 `json:"total_spent"`
	TransactionCount int64 `json:"transaction_count"`
}

type PaymentMethodReport struct {
	Method       string  `json:"method"`
	Count        int64   `json:"count"`
	Amount       float64 `json:"amount"`
	SuccessRate  float64 `json:"success_rate"`
}

type TaxSummary struct {
	TotalTDS        float64 `json:"total_tds"`
	TotalGST        float64 `json:"total_gst"`
	TaxableAmount   float64 `json:"taxable_amount"`
	TDSDeductions   int64   `json:"tds_deductions"`
}

// User Activity Report
type UserActivityReport struct {
	Summary         UserActivitySummary `json:"summary"`
	RegistrationTrend []DailyUserStats   `json:"registration_trend"`
	UsersByRegion   []RegionStats       `json:"users_by_region"`
	UsersByKYCStatus []KYCStats         `json:"users_by_kyc_status"`
	ActiveUserTrend []ActiveUserStats   `json:"active_user_trend"`
	UserEngagement  []UserEngagementStats `json:"user_engagement"`
}

type UserActivitySummary struct {
	TotalUsers        int64   `json:"total_users"`
	NewUsers          int64   `json:"new_users"`
	ActiveUsers       int64   `json:"active_users"`
	VerifiedUsers     int64   `json:"verified_users"`
	KYCCompletedUsers int64   `json:"kyc_completed_users"`
	RetentionRate     float64 `json:"retention_rate"`
	ChurnRate         float64 `json:"churn_rate"`
}

type DailyUserStats struct {
	Date            time.Time `json:"date"`
	NewRegistrations int64    `json:"new_registrations"`
	ActiveUsers     int64    `json:"active_users"`
	RetentionRate   float64  `json:"retention_rate"`
}

type RegionStats struct {
	Region     string `json:"region"`
	UserCount  int64  `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

type KYCStats struct {
	Status     string  `json:"status"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ActiveUserStats struct {
	Date           time.Time `json:"date"`
	DailyActive    int64     `json:"daily_active"`
	WeeklyActive   int64     `json:"weekly_active"`
	MonthlyActive  int64     `json:"monthly_active"`
}

type UserEngagementStats struct {
	UserID           int64   `json:"user_id"`
	Username         string  `json:"username"`
	ContestsJoined   int64   `json:"contests_joined"`
	TeamsCreated     int64   `json:"teams_created"`
	TotalSpent       float64 `json:"total_spent"`
	LastActivity     time.Time `json:"last_activity"`
	EngagementScore  float64 `json:"engagement_score"`
}

// Contest Performance Report
type ContestPerformanceReport struct {
	Summary           ContestSummary       `json:"summary"`
	ContestsByType    []ContestTypeStats   `json:"contests_by_type"`
	PopularContests   []PopularContestStats `json:"popular_contests"`
	ContestTrend      []DailyContestStats  `json:"contest_trend"`
	PrizeDistribution []PrizeDistributionStats `json:"prize_distribution"`
	ParticipationStats []ParticipationStats `json:"participation_stats"`
}

type ContestSummary struct {
	TotalContests        int64   `json:"total_contests"`
	CompletedContests    int64   `json:"completed_contests"`
	TotalParticipants    int64   `json:"total_participants"`
	TotalPrizePool       float64 `json:"total_prize_pool"`
	AvgParticipation     float64 `json:"avg_participation"`
	CompletionRate       float64 `json:"completion_rate"`
	FillRate            float64 `json:"fill_rate"`
}

type ContestTypeStats struct {
	ContestType      string  `json:"contest_type"`
	Count            int64   `json:"count"`
	TotalParticipants int64  `json:"total_participants"`
	AvgEntryFee      float64 `json:"avg_entry_fee"`
	TotalPrizePool   float64 `json:"total_prize_pool"`
}

type PopularContestStats struct {
	ContestID        int64   `json:"contest_id"`
	ContestName      string  `json:"contest_name"`
	GameName         string  `json:"game_name"`
	ParticipantCount int64   `json:"participant_count"`
	EntryFee         float64 `json:"entry_fee"`
	PrizePool        float64 `json:"prize_pool"`
	FillRate         float64 `json:"fill_rate"`
}

type DailyContestStats struct {
	Date              time.Time `json:"date"`
	ContestsCreated   int64     `json:"contests_created"`
	ContestsCompleted int64     `json:"contests_completed"`
	TotalParticipants int64     `json:"total_participants"`
	TotalPrizePool    float64   `json:"total_prize_pool"`
}

type PrizeDistributionStats struct {
	PrizeRange    string  `json:"prize_range"`
	ContestCount  int64   `json:"contest_count"`
	WinnerCount   int64   `json:"winner_count"`
	TotalPrizeDistributed float64 `json:"total_prize_distributed"`
}

type ParticipationStats struct {
	ParticipationRange string  `json:"participation_range"`
	ContestCount       int64   `json:"contest_count"`
	AvgFillRate        float64 `json:"avg_fill_rate"`
}

// Game Performance Report
type GamePerformanceReport struct {
	Summary         GameSummary         `json:"summary"`
	GameStats       []GamePerformanceStats `json:"game_stats"`
	PlayerStats     []PlayerPerformanceReport `json:"player_stats"`
	TeamStats       []TeamPerformanceStats `json:"team_stats"`
	MatchStats      []MatchPerformanceStats `json:"match_stats"`
	PopularityTrend []GamePopularityTrend `json:"popularity_trend"`
}

type GameSummary struct {
	TotalGames       int64 `json:"total_games"`
	ActiveGames      int64 `json:"active_games"`
	TotalMatches     int64 `json:"total_matches"`
	CompletedMatches int64 `json:"completed_matches"`
	TotalPlayers     int64 `json:"total_players"`
	ActivePlayers    int64 `json:"active_players"`
}

type GamePerformanceStats struct {
	GameID           int     `json:"game_id"`
	GameName         string  `json:"game_name"`
	TotalMatches     int64   `json:"total_matches"`
	TotalContests    int64   `json:"total_contests"`
	TotalParticipants int64  `json:"total_participants"`
	AvgParticipation float64 `json:"avg_participation"`
	PopularityScore  float64 `json:"popularity_score"`
}

type PlayerPerformanceReport struct {
	PlayerID        int64   `json:"player_id"`
	PlayerName      string  `json:"player_name"`
	GameName        string  `json:"game_name"`
	TeamName        string  `json:"team_name"`
	MatchesPlayed   int64   `json:"matches_played"`
	AvgPoints       float64 `json:"avg_points"`
	SelectionRate   float64 `json:"selection_rate"`
	CaptainRate     float64 `json:"captain_rate"`
	ViceCaptainRate float64 `json:"vice_captain_rate"`
}

type TeamPerformanceStats struct {
	TeamID          int64   `json:"team_id"`
	TeamName        string  `json:"team_name"`
	GameName        string  `json:"game_name"`
	MatchesPlayed   int64   `json:"matches_played"`
	WinRate         float64 `json:"win_rate"`
	AvgTeamPoints   float64 `json:"avg_team_points"`
	PopularityScore float64 `json:"popularity_score"`
}

type MatchPerformanceStats struct {
	MatchID         int64     `json:"match_id"`
	MatchName       string    `json:"match_name"`
	GameName        string    `json:"game_name"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	Status          string    `json:"status"`
	ContestCount    int64     `json:"contest_count"`
	ParticipantCount int64    `json:"participant_count"`
	TotalPrizePool  float64   `json:"total_prize_pool"`
}

type GamePopularityTrend struct {
	Date             time.Time `json:"date"`
	GameName         string    `json:"game_name"`
	MatchCount       int64     `json:"match_count"`
	ContestCount     int64     `json:"contest_count"`
	ParticipantCount int64     `json:"participant_count"`
}

// Compliance Report
type ComplianceReport struct {
	KYCCompliance     KYCComplianceStats     `json:"kyc_compliance"`
	TransactionCompliance TransactionComplianceStats `json:"transaction_compliance"`
	UserCompliance    UserComplianceStats    `json:"user_compliance"`
	AuditTrail        []AuditTrailEntry      `json:"audit_trail"`
	RiskAssessment    RiskAssessmentReport   `json:"risk_assessment"`
}

type KYCComplianceStats struct {
	TotalUsers           int64   `json:"total_users"`
	KYCVerifiedUsers     int64   `json:"kyc_verified_users"`
	PendingVerification  int64   `json:"pending_verification"`
	RejectedDocuments    int64   `json:"rejected_documents"`
	ComplianceRate       float64 `json:"compliance_rate"`
	AvgVerificationTime  float64 `json:"avg_verification_time"`
}

type TransactionComplianceStats struct {
	TotalTransactions    int64   `json:"total_transactions"`
	FlaggedTransactions  int64   `json:"flagged_transactions"`
	SuspiciousActivity   int64   `json:"suspicious_activity"`
	ComplianceRate       float64 `json:"compliance_rate"`
	LargeTransactions    int64   `json:"large_transactions"`
	FailedTransactions   int64   `json:"failed_transactions"`
}

type UserComplianceStats struct {
	ActiveUsers         int64   `json:"active_users"`
	SuspendedUsers      int64   `json:"suspended_users"`
	BannedUsers         int64   `json:"banned_users"`
	UnverifiedUsers     int64   `json:"unverified_users"`
	ComplianceRate      float64 `json:"compliance_rate"`
	RiskScore           float64 `json:"avg_risk_score"`
}

type AuditTrailEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	UserID      *int64    `json:"user_id"`
	AdminID     *int64    `json:"admin_id"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id"`
	Details     string    `json:"details"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Result      string    `json:"result"`
}

type RiskAssessmentReport struct {
	OverallRiskScore   float64            `json:"overall_risk_score"`
	RiskCategories     []RiskCategory     `json:"risk_categories"`
	HighRiskUsers      []ReportingHighRiskUser     `json:"high_risk_users"`
	RecommendedActions []RecommendedAction `json:"recommended_actions"`
}

type RiskCategory struct {
	Category    string  `json:"category"`
	RiskScore   float64 `json:"risk_score"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
}

type ReportingHighRiskUser struct {
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	RiskScore   float64 `json:"risk_score"`
	RiskFactors []string `json:"risk_factors"`
	LastActivity time.Time `json:"last_activity"`
}

type RecommendedAction struct {
	Priority    string `json:"priority"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// Referral Report
type ReferralReport struct {
	Summary          ReferralSummary         `json:"summary"`
	TierDistribution []TierDistributionStats `json:"tier_distribution"`
	TopReferrers     []TopReferrerStats      `json:"top_referrers"`
	ConversionFunnel []ConversionStep        `json:"conversion_funnel"`
	ReferralTrend    []DailyReferralStats    `json:"referral_trend"`
	EarningsStats    []ReferralEarningsStats `json:"earnings_stats"`
}

type ReferralSummary struct {
	TotalReferrals      int64   `json:"total_referrals"`
	SuccessfulReferrals int64   `json:"successful_referrals"`
	TotalEarnings       float64 `json:"total_earnings"`
	AvgEarningsPerReferral float64 `json:"avg_earnings_per_referral"`
	ConversionRate      float64 `json:"conversion_rate"`
	ActiveReferrers     int64   `json:"active_referrers"`
}

type TierDistributionStats struct {
	Tier        string  `json:"tier"`
	UserCount   int64   `json:"user_count"`
	Percentage  float64 `json:"percentage"`
	TotalEarnings float64 `json:"total_earnings"`
}

type TopReferrerStats struct {
	UserID          int64   `json:"user_id"`
	Username        string  `json:"username"`
	CurrentTier     string  `json:"current_tier"`
	TotalReferrals  int64   `json:"total_referrals"`
	SuccessfulReferrals int64 `json:"successful_referrals"`
	TotalEarnings   float64 `json:"total_earnings"`
	ConversionRate  float64 `json:"conversion_rate"`
}

type ConversionStep struct {
	Step           string  `json:"step"`
	Count          int64   `json:"count"`
	ConversionRate float64 `json:"conversion_rate"`
}

type DailyReferralStats struct {
	Date                time.Time `json:"date"`
	NewReferrals        int64     `json:"new_referrals"`
	CompletedReferrals  int64     `json:"completed_referrals"`
	TotalEarnings       float64   `json:"total_earnings"`
	ConversionRate      float64   `json:"conversion_rate"`
}

type ReferralEarningsStats struct {
	EarningsRange string  `json:"earnings_range"`
	UserCount     int64   `json:"user_count"`
	TotalEarnings float64 `json:"total_earnings"`
	AvgEarnings   float64 `json:"avg_earnings"`
}

// Report List Response
type ReportListResponse struct {
	Reports []ReportSummary `json:"reports"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Pages   int             `json:"pages"`
}

// Export Configuration
type ExportConfig struct {
	FileName    string                 `json:"file_name"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Columns     []ExportColumn         `json:"columns"`
	Formatting  map[string]interface{} `json:"formatting"`
	Filters     map[string]interface{} `json:"applied_filters"`
}

type ExportColumn struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // string, number, date, currency
	Format      string `json:"format,omitempty"`
	Width       int    `json:"width,omitempty"`
	Sortable    bool   `json:"sortable"`
	Filterable  bool   `json:"filterable"`
}