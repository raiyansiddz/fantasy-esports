package models

import (
	"time"
)

// Business Intelligence Dashboard
type BusinessIntelligenceDashboard struct {
	KPIMetrics          KPIMetrics          `json:"kpi_metrics"`
	RevenueAnalytics    RevenueAnalytics    `json:"revenue_analytics"`
	UserBehaviorAnalysis UserBehaviorAnalysis `json:"user_behavior_analysis"`
	PredictiveAnalytics PredictiveAnalytics `json:"predictive_analytics"`
	CompetitiveAnalysis CompetitiveAnalysis `json:"competitive_analysis"`
	BusinessInsights    []BusinessInsight   `json:"business_insights"`
}

// Key Performance Indicators
type KPIMetrics struct {
	CustomerAcquisitionCost    float64 `json:"customer_acquisition_cost"`
	CustomerLifetimeValue      float64 `json:"customer_lifetime_value"`
	MonthlyRecurringRevenue    float64 `json:"monthly_recurring_revenue"`
	AnnualRecurringRevenue     float64 `json:"annual_recurring_revenue"`
	ChurnRate                  float64 `json:"churn_rate"`
	RetentionRate              float64 `json:"retention_rate"`
	AverageRevenuePerUser      float64 `json:"average_revenue_per_user"`
	PaybackPeriod              int     `json:"payback_period_days"`
	NetPromoterScore           float64 `json:"net_promoter_score"`
	CustomerSatisfactionScore  float64 `json:"customer_satisfaction_score"`
	ContestParticipationRate   float64 `json:"contest_participation_rate"`
	WinRate                    float64 `json:"win_rate"`
	ROI                        float64 `json:"roi"`
	ConversionRate             float64 `json:"conversion_rate"`
}

// Advanced Revenue Analytics
type RevenueAnalytics struct {
	RevenueGrowthRate       float64                `json:"revenue_growth_rate"`
	RevenueBySegment        []SegmentRevenue       `json:"revenue_by_segment"`
	RevenueByGame           []GameRevenue          `json:"revenue_by_game"`
	RevenueByUserTier       []TierRevenue          `json:"revenue_by_user_tier"`
	SeasonalityAnalysis     []SeasonalRevenue      `json:"seasonality_analysis"`
	RevenueForecast         []RevenueForecast      `json:"revenue_forecast"`
	ProfitabilityAnalysis   ProfitabilityAnalysis  `json:"profitability_analysis"`
	RevenueOptimization     []OptimizationInsight  `json:"revenue_optimization"`
}

type SegmentRevenue struct {
	Segment       string    `json:"segment"`
	Revenue       float64   `json:"revenue"`
	Growth        float64   `json:"growth"`
	UserCount     int64     `json:"user_count"`
	ARPU          float64   `json:"arpu"`
	Trend         string    `json:"trend"`
}

type GameRevenue struct {
	GameID        int       `json:"game_id"`
	GameName      string    `json:"game_name"`
	Revenue       float64   `json:"revenue"`
	Growth        float64   `json:"growth"`
	MarketShare   float64   `json:"market_share"`
	Profitability float64   `json:"profitability"`
}

type TierRevenue struct {
	Tier          string    `json:"tier"`
	Revenue       float64   `json:"revenue"`
	UserCount     int64     `json:"user_count"`
	ARPU          float64   `json:"arpu"`
	Contribution  float64   `json:"contribution_percentage"`
}

type SeasonalRevenue struct {
	Period        string    `json:"period"`
	Revenue       float64   `json:"revenue"`
	Seasonality   float64   `json:"seasonality_factor"`
	YearOverYear  float64   `json:"year_over_year_growth"`
}

type RevenueForecast struct {
	Period        time.Time `json:"period"`
	ForecastLow   float64   `json:"forecast_low"`
	ForecastHigh  float64   `json:"forecast_high"`
	ForecastMid   float64   `json:"forecast_mid"`
	Confidence    float64   `json:"confidence_level"`
}

type ProfitabilityAnalysis struct {
	GrossMargin       float64 `json:"gross_margin"`
	NetMargin         float64 `json:"net_margin"`
	EBITDA           float64 `json:"ebitda"`
	OperatingExpenses float64 `json:"operating_expenses"`
	BreakevenPoint    float64 `json:"breakeven_point"`
	ProfitTrend       string  `json:"profit_trend"`
}

type OptimizationInsight struct {
	Category      string  `json:"category"`
	Recommendation string `json:"recommendation"`
	Impact        float64 `json:"potential_impact"`
	Effort        string  `json:"effort_level"`
	Priority      string  `json:"priority"`
}

// User Behavior Analysis
type UserBehaviorAnalysis struct {
	UserSegments        []UserSegment        `json:"user_segments"`
	BehaviorPatterns    []BehaviorPattern    `json:"behavior_patterns"`
	EngagementAnalysis  EngagementAnalysis   `json:"engagement_analysis"`
	ConversionFunnels   []ConversionFunnel   `json:"conversion_funnels"`
	CohortAnalysis      []CohortInsight      `json:"cohort_analysis"`
	PersonalizationData PersonalizationData  `json:"personalization_data"`
}

type UserSegment struct {
	SegmentID          string    `json:"segment_id"`
	SegmentName        string    `json:"segment_name"`
	UserCount          int64     `json:"user_count"`
	AvgRevenue         float64   `json:"avg_revenue"`
	EngagementScore    float64   `json:"engagement_score"`
	ChurnProbability   float64   `json:"churn_probability"`
	Characteristics    []string  `json:"characteristics"`
	RecommendedActions []string  `json:"recommended_actions"`
}

type BehaviorPattern struct {
	Pattern           string    `json:"pattern"`
	Description       string    `json:"description"`
	Frequency         int64     `json:"frequency"`
	UserCount         int64     `json:"user_count"`
	RevenueImpact     float64   `json:"revenue_impact"`
	Significance      string    `json:"significance"`
}

type EngagementAnalysis struct {
	AvgSessionDuration    float64              `json:"avg_session_duration"`
	AvgPageViews         int                  `json:"avg_page_views"`
	BounceRate           float64              `json:"bounce_rate"`
	FeatureAdoption      []FeatureAdoption    `json:"feature_adoption"`
	UserJourney          []UserJourneyStep    `json:"user_journey"`
	EngagementTrend      []EngagementTrend    `json:"engagement_trend"`
}

type FeatureAdoption struct {
	FeatureName       string  `json:"feature_name"`
	AdoptionRate      float64 `json:"adoption_rate"`
	TimeToAdopt       float64 `json:"time_to_adopt_days"`
	RetentionImpact   float64 `json:"retention_impact"`
	RevenueImpact     float64 `json:"revenue_impact"`
}

type UserJourneyStep struct {
	Step              string  `json:"step"`
	StepOrder         int     `json:"step_order"`
	UserCount         int64   `json:"user_count"`
	ConversionRate    float64 `json:"conversion_rate"`
	DropOffRate       float64 `json:"drop_off_rate"`
	AvgTimeSpent      float64 `json:"avg_time_spent"`
}

type EngagementTrend struct {
	Date              time.Time `json:"date"`
	EngagementScore   float64   `json:"engagement_score"`
	ActiveUsers       int64     `json:"active_users"`
	SessionDuration   float64   `json:"session_duration"`
}

type ConversionFunnel struct {
	FunnelName        string         `json:"funnel_name"`
	Steps             []BIFunnelStep   `json:"steps"`
	OverallConversion float64        `json:"overall_conversion_rate"`
	BottleneckStep    string         `json:"bottleneck_step"`
}

type BIFunnelStep struct {
	StepName          string  `json:"step_name"`
	StepOrder         int     `json:"step_order"`
	Users             int64   `json:"users"`
	ConversionRate    float64 `json:"conversion_rate"`
	DropOffRate       float64 `json:"drop_off_rate"`
}

type CohortInsight struct {
	CohortPeriod      string    `json:"cohort_period"`
	CohortSize        int64     `json:"cohort_size"`
	RetentionRates    []float64 `json:"retention_rates"`
	RevenuePerCohort  []float64 `json:"revenue_per_cohort"`
	LifetimeValue     float64   `json:"lifetime_value"`
}

type PersonalizationData struct {
	PersonalizationScore float64                    `json:"personalization_score"`
	UserPreferences      []UserPreference           `json:"user_preferences"`
	RecommendationEngine RecommendationEngineMetrics `json:"recommendation_engine"`
	ContentPerformance   []ContentPerformance       `json:"content_performance"`
}

type UserPreference struct {
	PreferenceType    string  `json:"preference_type"`
	PreferenceValue   string  `json:"preference_value"`
	UserCount         int64   `json:"user_count"`
	EngagementImpact  float64 `json:"engagement_impact"`
}

type RecommendationEngineMetrics struct {
	ClickThroughRate  float64 `json:"click_through_rate"`
	ConversionRate    float64 `json:"conversion_rate"`
	RevenueImpact     float64 `json:"revenue_impact"`
	AccuracyScore     float64 `json:"accuracy_score"`
}

type ContentPerformance struct {
	ContentType       string  `json:"content_type"`
	EngagementRate    float64 `json:"engagement_rate"`
	ConversionRate    float64 `json:"conversion_rate"`
	ShareRate         float64 `json:"share_rate"`
	Performance       string  `json:"performance"`
}

// Predictive Analytics
type PredictiveAnalytics struct {
	ChurnPrediction     ChurnPrediction     `json:"churn_prediction"`
	RevenueForecasting  RevenueForecasting  `json:"revenue_forecasting"`
	UserGrowthPrediction UserGrowthPrediction `json:"user_growth_prediction"`
	SeasonalityPrediction SeasonalityPrediction `json:"seasonality_prediction"`
	RiskAssessment      RiskAssessment      `json:"risk_assessment"`
	OpportunityScoring  []OpportunityScore  `json:"opportunity_scoring"`
}

type ChurnPrediction struct {
	OverallChurnRate    float64           `json:"overall_churn_rate"`
	ChurnBySegment      []SegmentChurn    `json:"churn_by_segment"`
	ChurnFactors        []ChurnFactor     `json:"churn_factors"`
	HighRiskUsers       []BIHighRiskUser    `json:"high_risk_users"`
	RetentionStrategies []RetentionStrategy `json:"retention_strategies"`
}

type SegmentChurn struct {
	Segment           string  `json:"segment"`
	ChurnRate         float64 `json:"churn_rate"`
	RiskLevel         string  `json:"risk_level"`
	RecommendedAction string  `json:"recommended_action"`
}

type ChurnFactor struct {
	Factor            string  `json:"factor"`
	Importance        float64 `json:"importance_score"`
	Impact            string  `json:"impact"`
	Actionability     string  `json:"actionability"`
}

type BIHighRiskUser struct {
	UserID            int64   `json:"user_id"`
	ChurnProbability  float64 `json:"churn_probability"`
	RiskFactors       []string `json:"risk_factors"`
	RecommendedAction string  `json:"recommended_action"`
	Value             float64 `json:"user_value"`
}

type RetentionStrategy struct {
	Strategy          string  `json:"strategy"`
	TargetSegment     string  `json:"target_segment"`
	ExpectedImpact    float64 `json:"expected_impact"`
	ImplementationCost float64 `json:"implementation_cost"`
	ROI               float64 `json:"roi"`
}

type RevenueForecasting struct {
	NextMonthForecast  float64            `json:"next_month_forecast"`
	NextQuarterForecast float64           `json:"next_quarter_forecast"`
	NextYearForecast   float64            `json:"next_year_forecast"`
	ScenarioAnalysis   []ScenarioAnalysis `json:"scenario_analysis"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
}

type ScenarioAnalysis struct {
	Scenario          string  `json:"scenario"`
	Probability       float64 `json:"probability"`
	RevenueImpact     float64 `json:"revenue_impact"`
	KeyAssumptions    []string `json:"key_assumptions"`
}

type ConfidenceInterval struct {
	LowerBound        float64 `json:"lower_bound"`
	UpperBound        float64 `json:"upper_bound"`
	ConfidenceLevel   float64 `json:"confidence_level"`
}

type UserGrowthPrediction struct {
	NextMonthUsers    int64              `json:"next_month_users"`
	NextQuarterUsers  int64              `json:"next_quarter_users"`
	GrowthRate        float64            `json:"growth_rate"`
	GrowthDrivers     []GrowthDriver     `json:"growth_drivers"`
	AcquisitionChannels []ChannelForecast `json:"acquisition_channels"`
}

type GrowthDriver struct {
	Driver            string  `json:"driver"`
	Impact            float64 `json:"impact_score"`
	Trend             string  `json:"trend"`
	Recommendation    string  `json:"recommendation"`
}

type ChannelForecast struct {
	Channel           string  `json:"channel"`
	ForecastedUsers   int64   `json:"forecasted_users"`
	Cost              float64 `json:"cost"`
	ROI               float64 `json:"roi"`
}

type SeasonalityPrediction struct {
	SeasonalPatterns  []SeasonalPattern  `json:"seasonal_patterns"`
	UpcomingPeaks     []Peak             `json:"upcoming_peaks"`
	UpcomingTroughs   []Trough           `json:"upcoming_troughs"`
	Recommendations   []SeasonalRecommendation `json:"recommendations"`
}

type SeasonalPattern struct {
	Period            string  `json:"period"`
	Strength          float64 `json:"strength"`
	Pattern           string  `json:"pattern"`
	BusinessImpact    string  `json:"business_impact"`
}

type Peak struct {
	Date              time.Time `json:"date"`
	ExpectedIncrease  float64   `json:"expected_increase"`
	Duration          int       `json:"duration_days"`
	PrepActions       []string  `json:"prep_actions"`
}

type Trough struct {
	Date              time.Time `json:"date"`
	ExpectedDecrease  float64   `json:"expected_decrease"`
	Duration          int       `json:"duration_days"`
	MitigationActions []string  `json:"mitigation_actions"`
}

type SeasonalRecommendation struct {
	Period            string  `json:"period"`
	Recommendation    string  `json:"recommendation"`
	ExpectedImpact    float64 `json:"expected_impact"`
	Priority          string  `json:"priority"`
}

type RiskAssessment struct {
	OverallRiskScore  float64      `json:"overall_risk_score"`
	RiskCategories    []RiskMetric `json:"risk_categories"`
	MitigationPlans   []MitigationPlan `json:"mitigation_plans"`
	MonitoringAlerts  []MonitoringAlert `json:"monitoring_alerts"`
}

type RiskMetric struct {
	Category          string  `json:"category"`
	RiskScore         float64 `json:"risk_score"`
	Trend             string  `json:"trend"`
	Impact            string  `json:"impact"`
	Likelihood        string  `json:"likelihood"`
}

type MitigationPlan struct {
	Risk              string   `json:"risk"`
	Plan              string   `json:"plan"`
	Timeline          string   `json:"timeline"`
	ResponsibleTeam   string   `json:"responsible_team"`
	Success_Metrics   []string `json:"success_metrics"`
}

type MonitoringAlert struct {
	Metric            string  `json:"metric"`
	Threshold         float64 `json:"threshold"`
	CurrentValue      float64 `json:"current_value"`
	AlertLevel        string  `json:"alert_level"`
	RecommendedAction string  `json:"recommended_action"`
}

type OpportunityScore struct {
	Opportunity       string  `json:"opportunity"`
	Score             float64 `json:"score"`
	PotentialRevenue  float64 `json:"potential_revenue"`
	Implementation    string  `json:"implementation_difficulty"`
	Timeline          string  `json:"timeline"`
	Priority          string  `json:"priority"`
}

// Competitive Analysis
type CompetitiveAnalysis struct {
	MarketPosition      MarketPosition       `json:"market_position"`
	CompetitorMetrics   []CompetitorMetric   `json:"competitor_metrics"`
	FeatureComparison   []FeatureComparison  `json:"feature_comparison"`
	PricingAnalysis     PricingAnalysis      `json:"pricing_analysis"`
	MarketTrends        []MarketTrend        `json:"market_trends"`
	CompetitiveAdvantages []CompetitiveAdvantage `json:"competitive_advantages"`
}

type MarketPosition struct {
	MarketRank        int     `json:"market_rank"`
	MarketShare       float64 `json:"market_share"`
	Growth            float64 `json:"growth"`
	Differentiation   string  `json:"differentiation"`
	CompetitiveStrength string `json:"competitive_strength"`
}

type CompetitorMetric struct {
	CompetitorName    string  `json:"competitor_name"`
	MarketShare       float64 `json:"market_share"`
	EstimatedRevenue  float64 `json:"estimated_revenue"`
	UserBase          int64   `json:"user_base"`
	GrowthRate        float64 `json:"growth_rate"`
	Strengths         []string `json:"strengths"`
	Weaknesses        []string `json:"weaknesses"`
}

type FeatureComparison struct {
	Feature           string              `json:"feature"`
	OurImplementation string              `json:"our_implementation"`
	Competitors       []CompetitorFeature `json:"competitors"`
	Advantage         string              `json:"advantage"`
}

type CompetitorFeature struct {
	CompetitorName    string `json:"competitor_name"`
	Implementation    string `json:"implementation"`
	Quality           string `json:"quality"`
}

type PricingAnalysis struct {
	OurPricing        PricingModel        `json:"our_pricing"`
	CompetitorPricing []CompetitorPricing `json:"competitor_pricing"`
	PricePosition     string              `json:"price_position"`
	ValueScore        float64             `json:"value_score"`
	Recommendations   []PricingRecommendation `json:"recommendations"`
}

type PricingModel struct {
	Model             string  `json:"model"`
	BasePrice         float64 `json:"base_price"`
	PremiumFeatures   []string `json:"premium_features"`
	ValueProposition  string  `json:"value_proposition"`
}

type CompetitorPricing struct {
	CompetitorName    string  `json:"competitor_name"`
	PricingModel      string  `json:"pricing_model"`
	Price             float64 `json:"price"`
	ValueScore        float64 `json:"value_score"`
}

type PricingRecommendation struct {
	Recommendation    string  `json:"recommendation"`
	ExpectedImpact    float64 `json:"expected_impact"`
	RiskLevel         string  `json:"risk_level"`
	Timeline          string  `json:"timeline"`
}

type MarketTrend struct {
	Trend             string  `json:"trend"`
	Impact            float64 `json:"impact_score"`
	Timeline          string  `json:"timeline"`
	OpportunityThreat string  `json:"opportunity_or_threat"`
	RecommendedAction string  `json:"recommended_action"`
}

type CompetitiveAdvantage struct {
	Advantage         string  `json:"advantage"`
	Strength          string  `json:"strength"`
	Sustainability    string  `json:"sustainability"`
	LeverageStrategy  string  `json:"leverage_strategy"`
}

// Business Insight
type BusinessInsight struct {
	InsightID         string    `json:"insight_id"`
	Category          string    `json:"category"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Impact            string    `json:"impact"`
	Priority          string    `json:"priority"`
	Confidence        float64   `json:"confidence_score"`
	RecommendedAction string    `json:"recommended_action"`
	DataSources       []string  `json:"data_sources"`
	GeneratedAt       time.Time `json:"generated_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// Business Intelligence Filters
type BIFilters struct {
	DateFrom          *time.Time `json:"date_from,omitempty"`
	DateTo            *time.Time `json:"date_to,omitempty"`
	UserSegment       *string    `json:"user_segment,omitempty"`
	GameID            *int       `json:"game_id,omitempty"`
	RevenueThreshold  *float64   `json:"revenue_threshold,omitempty"`
	ConfidenceLevel   *float64   `json:"confidence_level,omitempty"`
	IncludeForecasts  bool       `json:"include_forecasts"`
	IncludePredictions bool      `json:"include_predictions"`
}

// Advanced Query Request
type AdvancedAnalyticsRequest struct {
	QueryType         string                 `json:"query_type"`
	Filters           BIFilters              `json:"filters"`
	Metrics           []string               `json:"metrics"`
	Dimensions        []string               `json:"dimensions"`
	Aggregations      map[string]string      `json:"aggregations"`
	CustomParameters  map[string]interface{} `json:"custom_parameters"`
}

// Custom Metric Definition
type CustomMetric struct {
	MetricID          string                 `json:"metric_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Formula           string                 `json:"formula"`
	Parameters        map[string]interface{} `json:"parameters"`
	Category          string                 `json:"category"`
	DataSources       []string               `json:"data_sources"`
	UpdateFrequency   string                 `json:"update_frequency"`
	IsActive          bool                   `json:"is_active"`
	CreatedBy         int64                  `json:"created_by"`
	CreatedAt         time.Time              `json:"created_at"`
}

// Alert Configuration
type AlertConfiguration struct {
	AlertID           string                 `json:"alert_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	MetricID          string                 `json:"metric_id"`
	Threshold         float64                `json:"threshold"`
	Condition         string                 `json:"condition"` // greater_than, less_than, equal_to, etc.
	Severity          string                 `json:"severity"`
	NotificationChannels []string            `json:"notification_channels"`
	Recipients        []string               `json:"recipients"`
	IsActive          bool                   `json:"is_active"`
	CreatedBy         int64                  `json:"created_by"`
	CreatedAt         time.Time              `json:"created_at"`
}

// Dashboard Configuration  
type DashboardConfiguration struct {
	DashboardID       string                 `json:"dashboard_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Layout            DashboardLayout        `json:"layout"`
	Widgets           []DashboardWidget      `json:"widgets"`
	Filters           map[string]interface{} `json:"filters"`
	RefreshRate       int                    `json:"refresh_rate_seconds"`
	IsPublic          bool                   `json:"is_public"`
	SharedWith        []int64                `json:"shared_with"`
	CreatedBy         int64                  `json:"created_by"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type DashboardLayout struct {
	Columns           int    `json:"columns"`
	ResponsiveBreakpoints map[string]int `json:"responsive_breakpoints"`
}

type DashboardWidget struct {
	WidgetID          string                 `json:"widget_id"`
	Type              string                 `json:"type"`
	Title             string                 `json:"title"`
	Position          WidgetPosition         `json:"position"`
	Configuration     map[string]interface{} `json:"configuration"`
	DataSource        string                 `json:"data_source"`
	RefreshRate       int                    `json:"refresh_rate_seconds"`
}

type WidgetPosition struct {
	X                 int `json:"x"`
	Y                 int `json:"y"`
	Width             int `json:"width"`
	Height            int `json:"height"`
}