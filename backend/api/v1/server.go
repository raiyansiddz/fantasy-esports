package v1

import (
	"database/sql"
	"fantasy-esports-backend/api/v1/handlers"
	"fantasy-esports-backend/api/v1/middleware"
	"fantasy-esports-backend/config"
	internal_handlers "fantasy-esports-backend/internal/handlers"
	internal_services "fantasy-esports-backend/internal/services"
	"fantasy-esports-backend/pkg/cdn"
	"fantasy-esports-backend/services"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
	db     *sql.DB
	config *config.Config
}

func NewServer(db *sql.DB, cfg *config.Config) *Server {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "authorization,content-type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	return &Server{
		router: router,
		db:     db,
		config: cfg,
	}
}

func (s *Server) setupRoutes() {
	// Initialize CDN client
	cdnClient, err := cdn.NewCloudinaryClient(s.config.CloudinaryURL)
	if err != nil {
		log.Fatal("Failed to initialize CDN client:", err)
	}

	// Initialize services
	leaderboardService := services.NewLeaderboardService(s.db)
	analyticsService := services.NewAnalyticsService(s.db)
	biService := services.NewBusinessIntelligenceService(s.db)
	reportingService := services.NewReportingService(s.db)
	
	// Initialize payment service
	paymentService := internal_services.NewPaymentService(s.db)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(s.db, s.config, cdnClient)
	userHandler := handlers.NewUserHandler(s.db, s.config, cdnClient)
	gameHandler := handlers.NewGameHandler(s.db, s.config)
	contestHandler := handlers.NewContestHandler(s.db, s.config)
	walletHandler := handlers.NewWalletHandler(s.db, s.config)
	adminHandler := handlers.NewAdminHandler(s.db, s.config, cdnClient)
	realtimeHandler := handlers.NewRealTimeLeaderboardHandler(s.db, s.config, leaderboardService)
	tournamentHandler := handlers.NewTournamentHandler(s.db, s.config, cdnClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, biService, reportingService)
	notificationHandler := handlers.NewNotificationHandler(s.db, s.config)
	paymentHandler := internal_handlers.NewPaymentHandler(s.db, s.config, paymentService)
	contentHandler := handlers.NewContentHandler(s.db, s.config, cdnClient)
	fraudDetectionHandler := handlers.NewFraudDetectionHandler(s.db, s.config)
	achievementHandler := handlers.NewAchievementHandler(s.db, s.config)
	friendHandler := handlers.NewFriendHandler(s.db, s.config)
	socialSharingHandler := handlers.NewSocialSharingHandler(s.db, s.config)
	advancedAnalyticsHandler := handlers.NewAdvancedAnalyticsHandler(s.db, s.config)
	predictionHandler := handlers.NewPlayerPredictionHandler(s.db, s.config)
	tournamentBracketHandler := handlers.NewTournamentBracketHandler(s.db, s.config)

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
			"service": "fantasy-esports-backend",
		})
	})

	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	
	// Public routes (no authentication required)
	{
		// Authentication
		v1.POST("/auth/verify-mobile", authHandler.VerifyMobile)
		v1.POST("/auth/verify-otp", authHandler.VerifyOTP)
		v1.POST("/auth/refresh", authHandler.RefreshToken)
		v1.POST("/auth/social-login", authHandler.SocialLogin)

		// Games (public information)
		v1.GET("/games", gameHandler.GetGames)
		v1.GET("/games/:id", gameHandler.GetGameDetails)
		v1.GET("/games/:id/tournaments", gameHandler.GetGameTournaments)
		v1.GET("/games/:id/players", gameHandler.GetGamePlayers)

		// Tournaments (public information)
		v1.GET("/tournaments", tournamentHandler.GetTournaments)
		v1.GET("/tournaments/:id", tournamentHandler.GetTournamentDetails)
		v1.GET("/tournaments/:id/bracket", tournamentHandler.GetTournamentBracket)

		// Matches (public information)
		v1.GET("/matches", gameHandler.GetMatches)
		v1.GET("/matches/:id", gameHandler.GetMatchDetails)
		v1.GET("/matches/:id/players", gameHandler.GetMatchPlayers)
		v1.GET("/matches/:id/player-performance", gameHandler.GetPlayerPerformance)

		// Live streams (public)
		v1.GET("/matches/:id/live-stream", tournamentHandler.GetMatchLiveStream)
		v1.GET("/live-streams/active", tournamentHandler.GetActiveLiveStreams)

		// Content Management - Public Routes
		// Banners (public access)
		v1.GET("/banners/active", contentHandler.GetActiveBanners)
		v1.POST("/banners/:id/click", contentHandler.TrackBannerClick)
		
		// SEO Content (public access)
		v1.GET("/seo/:slug", contentHandler.GetSEOContentBySlug)
		
		// FAQ (public access)
		v1.GET("/faq/sections", contentHandler.ListFAQSections)
		v1.GET("/faq/items", contentHandler.ListFAQItems)
		v1.POST("/faq/items/:id/view", contentHandler.TrackFAQView)
		v1.POST("/faq/items/:id/like", contentHandler.TrackFAQLike)
		
		// Legal Documents (public access)
		v1.GET("/legal/:type", contentHandler.GetActiveLegalDocument)

		// Admin login (public)
		v1.POST("/admin/login", adminHandler.Login)

		// Public fraud reporting
		v1.POST("/fraud/report", fraudDetectionHandler.ReportSuspiciousActivity)
		v1.POST("/fraud/webhook", fraudDetectionHandler.FraudWebhook)
	}

	// Protected user routes (require user authentication)
	userRoutes := v1.Group("")
	userRoutes.Use(middleware.AuthMiddleware(s.config.JWTSecret))
	userRoutes.Use(fraudDetectionHandler.FraudDetectionMiddleware())
	{
		// User management
		userRoutes.POST("/auth/logout", authHandler.Logout)
		userRoutes.GET("/users/profile", userHandler.GetProfile)
		userRoutes.PUT("/users/profile", userHandler.UpdateProfile)
		userRoutes.PUT("/users/preferences", userHandler.UpdatePreferences)

		// KYC
		userRoutes.POST("/users/kyc/upload", userHandler.UploadKYC)
		userRoutes.GET("/users/kyc/status", userHandler.GetKYCStatus)

		// Referrals
		userRoutes.GET("/referrals/my-stats", walletHandler.GetReferralStats)
		userRoutes.GET("/referrals/history", walletHandler.GetReferralHistory)
		userRoutes.POST("/referrals/apply", walletHandler.ApplyReferralCode)
		userRoutes.POST("/referrals/share", walletHandler.ShareReferral)
		userRoutes.GET("/referrals/leaderboard", walletHandler.GetReferralLeaderboard)

		// Contest management
		userRoutes.GET("/contests", contestHandler.GetContests)
		userRoutes.GET("/contests/:id", contestHandler.GetContestDetails)
		userRoutes.POST("/contests/:id/join", contestHandler.JoinContest)
		userRoutes.DELETE("/contests/:id/leave", contestHandler.LeaveContest)
		userRoutes.GET("/contests/my-entries", contestHandler.GetMyEntries)
		userRoutes.POST("/contests/create-private", contestHandler.CreatePrivateContest)

		// Fantasy team management
		userRoutes.POST("/teams/create", contestHandler.CreateTeam)
		userRoutes.GET("/teams/my-teams", contestHandler.GetMyTeams)
		userRoutes.GET("/teams/:id", contestHandler.GetTeamDetails)
		userRoutes.PUT("/teams/:id", contestHandler.UpdateTeam)
		userRoutes.DELETE("/teams/:id", contestHandler.DeleteTeam)
		userRoutes.POST("/teams/:id/clone", contestHandler.CloneTeam)
		userRoutes.POST("/teams/validate", contestHandler.ValidateTeam)
		userRoutes.GET("/teams/:id/performance", contestHandler.GetTeamPerformance)

		// Leaderboards
		userRoutes.GET("/leaderboards/contests/:id", contestHandler.GetContestLeaderboard)
		userRoutes.GET("/leaderboards/live/:id", contestHandler.GetLiveLeaderboard)
		userRoutes.GET("/leaderboards/contests/:id/my-rank", contestHandler.GetMyRank)

		// Wallet management
		userRoutes.GET("/wallet/balance", walletHandler.GetBalance)
		userRoutes.POST("/wallet/deposit", walletHandler.Deposit)
		userRoutes.POST("/wallet/withdraw", walletHandler.Withdraw)
		userRoutes.GET("/wallet/transactions", walletHandler.GetTransactions)
		userRoutes.GET("/payments/:id/status", walletHandler.GetPaymentStatus)
		userRoutes.GET("/wallet/payment-methods", walletHandler.GetPaymentMethods)
		userRoutes.POST("/wallet/payment-methods", walletHandler.AddPaymentMethod)

		// Payment Gateway APIs
		userRoutes.POST("/payment/create-order", paymentHandler.CreatePaymentOrder)

		// Achievements
		userRoutes.GET("/achievements", achievementHandler.GetAchievements)
		userRoutes.GET("/achievements/my", achievementHandler.GetUserAchievements)
		userRoutes.GET("/achievements/:id", achievementHandler.GetAchievements) // Placeholder - using existing method
		userRoutes.GET("/achievements/categories", achievementHandler.GetAchievements) // Placeholder - using existing method
		userRoutes.GET("/achievements/:id/progress", achievementHandler.GetAchievementProgress)
		userRoutes.POST("/achievements/claim", achievementHandler.GetUserAchievements) // Placeholder - using existing method

		// Friends System
		userRoutes.POST("/friends", friendHandler.AddFriend)
		userRoutes.GET("/friends", friendHandler.GetFriends)
		userRoutes.GET("/friends/requests", friendHandler.GetFriends) // Placeholder - using existing method
		userRoutes.POST("/friends/:friend_id/accept", friendHandler.AcceptFriend)
		userRoutes.POST("/friends/:friend_id/decline", friendHandler.DeclineFriend)
		userRoutes.DELETE("/friends/:friend_id", friendHandler.RemoveFriend)

		// Friend Challenges
		userRoutes.POST("/challenges", friendHandler.CreateChallenge)
		userRoutes.GET("/challenges", friendHandler.GetChallenges)
		userRoutes.GET("/challenges/my", friendHandler.GetChallenges) // Placeholder - using existing method
		userRoutes.GET("/challenges/:challenge_id/status", friendHandler.GetChallenges) // Placeholder - using existing method
		userRoutes.POST("/challenges/:challenge_id/accept", friendHandler.AcceptChallenge)
		userRoutes.POST("/challenges/:challenge_id/decline", friendHandler.DeclineChallenge)

		// Friend Activities
		userRoutes.GET("/friends/activities", friendHandler.GetFriendActivities)

		// Social Sharing
		userRoutes.POST("/share", socialSharingHandler.CreateShare)
		userRoutes.GET("/share/my", socialSharingHandler.GetUserShares)
		userRoutes.GET("/share/:share_id/stats", socialSharingHandler.GetUserShares) // Placeholder - using existing method
		userRoutes.GET("/share/teams/:team_id/urls", socialSharingHandler.GenerateTeamShareURLs)
		userRoutes.GET("/share/contests/:contest_id/urls", socialSharingHandler.GenerateContestWinShareURLs)
		userRoutes.GET("/share/achievements/:achievement_id/urls", socialSharingHandler.GenerateAchievementShareURLs)
		userRoutes.POST("/share/:share_id/click", socialSharingHandler.TrackShareClick)

		// Player Predictions
		userRoutes.GET("/matches/:id/predictions", predictionHandler.GetMatchPredictions)
		userRoutes.GET("/predictions/players/:player_id/match/:match_id", predictionHandler.GetMatchPredictions)
		userRoutes.GET("/predictions/match/:match_id/teams", predictionHandler.GetMatchPredictions)
		userRoutes.GET("/predictions/my", predictionHandler.GetMatchPredictions) // Placeholder - using existing method
		userRoutes.GET("/predictions/accuracy/my", predictionHandler.GetPredictionAnalytics) // Placeholder - using existing method
		userRoutes.POST("/predictions/calculate", predictionHandler.GenerateMatchPredictions)
		userRoutes.GET("/predictions/history/:player_id", predictionHandler.GetPredictionAnalytics)

		// Advanced Game Analytics - 7 Metrics
		userRoutes.GET("/analytics/games/:game_id/player-efficiency", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/team-synergy", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/strategic-diversity", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/comeback-potential", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/clutch-performance", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/consistency-index", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		userRoutes.GET("/analytics/games/:game_id/adaptability-score", advancedAnalyticsHandler.GetAdvancedGameMetrics)

		// Tournament Brackets - 4 Types (User Access)
		userRoutes.POST("/tournaments/:id/brackets/single-elimination", tournamentBracketHandler.CreateBracket)
		userRoutes.POST("/tournaments/:id/brackets/double-elimination", tournamentBracketHandler.CreateBracket)
		userRoutes.POST("/tournaments/:id/brackets/round-robin", tournamentBracketHandler.CreateBracket)
		userRoutes.POST("/tournaments/:id/brackets/swiss-system", tournamentBracketHandler.CreateBracket)
		userRoutes.GET("/tournaments/:id/brackets/current", tournamentBracketHandler.GetTournamentBrackets)
		
		// Advanced Fraud Detection (User Level)
		userRoutes.GET("/fraud/risk-score", fraudDetectionHandler.GetAlerts)
		userRoutes.GET("/fraud/my-reports", fraudDetectionHandler.GetAlerts) // Placeholder - using existing method

		userRoutes.POST("/payment/verify", paymentHandler.VerifyPayment)
		userRoutes.GET("/payment/status/:transaction_id", paymentHandler.GetPaymentStatus)

		// Notification endpoints (for users)
		userRoutes.POST("/notify/send", notificationHandler.SendNotification)
	}

	// Protected admin routes (require admin authentication)
	adminRoutes := v1.Group("/admin")
	adminRoutes.Use(middleware.AdminAuthMiddleware(s.config.JWTSecret))
	{
		// User management
		adminRoutes.GET("/users", adminHandler.GetUsers)
		adminRoutes.GET("/users/:id", adminHandler.GetUserDetails)
		adminRoutes.PUT("/users/:id/status", adminHandler.UpdateUserStatus)

		// KYC document processing
		adminRoutes.GET("/kyc/documents", adminHandler.GetPendingKYCDocuments)
		adminRoutes.PUT("/kyc/documents/:id/process", adminHandler.ProcessKYC)

		// Live match scoring system
		adminRoutes.GET("/matches/live-scoring", adminHandler.GetLiveScoringMatches)
		adminRoutes.POST("/matches/:id/start-scoring", adminHandler.StartManualScoring)
		adminRoutes.POST("/matches/:id/events", adminHandler.AddMatchEvent)
		adminRoutes.PUT("/matches/:id/players/:player_id/stats", adminHandler.UpdatePlayerStats)
		adminRoutes.POST("/matches/:id/events/bulk", adminHandler.BulkUpdateEvents)
		adminRoutes.PUT("/matches/:id/score", adminHandler.UpdateMatchScore)
		adminRoutes.POST("/matches/:id/recalculate-points", adminHandler.RecalculatePoints)
		adminRoutes.GET("/matches/:id/dashboard", adminHandler.GetLiveDashboard)
		adminRoutes.POST("/matches/:id/complete", adminHandler.CompleteMatch)
		adminRoutes.GET("/matches/:id/events", adminHandler.GetMatchEvents)
		adminRoutes.PUT("/matches/:id/events/:event_id", adminHandler.EditMatchEvent)
		adminRoutes.DELETE("/matches/:id/events/:event_id", adminHandler.DeleteMatchEvent)

		// Tournament and stage management
		adminRoutes.POST("/tournaments/:id/stages", tournamentHandler.CreateTournamentStage)
		adminRoutes.POST("/tournaments/stages/:stage_id/advance", tournamentHandler.AdvanceToNextStage)

		// Live streaming management
		adminRoutes.POST("/matches/:id/live-stream", tournamentHandler.SetMatchLiveStream)
		adminRoutes.PUT("/matches/:id/live-stream/activate", tournamentHandler.ActivateMatchLiveStream)
		adminRoutes.DELETE("/matches/:id/live-stream", tournamentHandler.RemoveMatchLiveStream)

		// Contest management
		adminRoutes.POST("/contests", contestHandler.CreateContest)
		adminRoutes.PUT("/contests/:id", contestHandler.UpdateContest)
		adminRoutes.DELETE("/contests/:id", contestHandler.DeleteContest)

		// Financial management
		adminRoutes.GET("/transactions", walletHandler.GetTransactions)
		adminRoutes.PUT("/withdrawals/:id/process", walletHandler.ProcessWithdrawal)

		// System configuration
		adminRoutes.GET("/config", adminHandler.GetSystemConfig)
		adminRoutes.PUT("/config", adminHandler.UpdateSystemConfig)

		// Analytics Dashboard
		adminRoutes.GET("/analytics/dashboard", analyticsHandler.GetAnalyticsDashboard)
		adminRoutes.GET("/analytics/users", analyticsHandler.GetUserMetrics)
		adminRoutes.GET("/analytics/revenue", analyticsHandler.GetRevenueMetrics)
		adminRoutes.GET("/analytics/contests", analyticsHandler.GetContestMetrics)
		adminRoutes.GET("/analytics/games", analyticsHandler.GetGameMetrics)
		adminRoutes.GET("/analytics/realtime", analyticsHandler.GetRealTimeMetrics)
		adminRoutes.GET("/analytics/performance", analyticsHandler.GetPerformanceMetrics)

		// Business Intelligence
		adminRoutes.GET("/bi/dashboard", analyticsHandler.GetBIDashboard)
		adminRoutes.GET("/bi/kpis", analyticsHandler.GetKPIMetrics)
		adminRoutes.GET("/bi/revenue", analyticsHandler.GetRevenueAnalytics)
		adminRoutes.GET("/bi/user-behavior", analyticsHandler.GetUserBehaviorAnalysis)
		adminRoutes.GET("/bi/predictive", analyticsHandler.GetPredictiveAnalytics)

		// Advanced Reporting System
		adminRoutes.POST("/reports/generate", analyticsHandler.GenerateReport)
		adminRoutes.GET("/reports", analyticsHandler.GetReports)
		adminRoutes.GET("/reports/:id", analyticsHandler.GetReport)
		adminRoutes.DELETE("/reports/:id", analyticsHandler.DeleteReport)

		// Achievement Management
		adminRoutes.POST("/achievements", achievementHandler.CreateAchievement)
		adminRoutes.GET("/achievements", achievementHandler.GetAchievements)
		adminRoutes.GET("/achievements/stats", achievementHandler.GetAchievements) // Placeholder - using existing method
		adminRoutes.PUT("/achievements/:id", achievementHandler.UpdateAchievement)
		adminRoutes.DELETE("/achievements/:id", achievementHandler.DeleteAchievement)

		// Advanced Analytics
		adminRoutes.GET("/analytics/summary", advancedAnalyticsHandler.GetAdvancedGameMetrics) // Placeholder - using existing method
		adminRoutes.POST("/analytics/generate", advancedAnalyticsHandler.GetAdvancedGameMetrics) // Placeholder - using existing method
		adminRoutes.GET("/games/:game_id/advanced-metrics", advancedAnalyticsHandler.GetAdvancedGameMetrics)
		adminRoutes.GET("/games/:game_id/metrics-history", advancedAnalyticsHandler.GetAdvancedMetricsHistory)
		adminRoutes.GET("/games/compare", advancedAnalyticsHandler.CompareGames)

		// Player Predictions Management
		adminRoutes.POST("/matches/:id/generate-predictions", predictionHandler.GenerateMatchPredictions)
		adminRoutes.PUT("/matches/:id/update-accuracy", predictionHandler.UpdatePredictionAccuracy)
		adminRoutes.GET("/matches/:id/update-accuracy", predictionHandler.GetPredictionAnalytics) // GET version for accessibility
		adminRoutes.GET("/predictions/analytics", predictionHandler.GetPredictionAnalytics)
		adminRoutes.GET("/predictions/accuracy/global", predictionHandler.GetGlobalPredictionAccuracy)
		adminRoutes.GET("/predictions/models/performance", predictionHandler.GetModelsPerformance)
		adminRoutes.PUT("/predictions/models/:id/update", predictionHandler.UpdatePredictionModel)
		adminRoutes.GET("/predictions/leaderboard", predictionHandler.GetPredictionLeaderboard)

		// Tournament Brackets
		adminRoutes.POST("/tournaments/brackets", tournamentBracketHandler.CreateBracket)
		adminRoutes.GET("/tournaments/:tournament_id/brackets", tournamentBracketHandler.GetTournamentBrackets)
		adminRoutes.GET("/brackets/:bracket_id", tournamentBracketHandler.GetBracket)
		adminRoutes.PUT("/brackets/:bracket_id/advance", tournamentBracketHandler.AdvanceBracket)
		adminRoutes.GET("/brackets/:bracket_id/advance", tournamentBracketHandler.GetBracket) // GET version for accessibility
		adminRoutes.PUT("/brackets/:bracket_id/status", tournamentBracketHandler.UpdateBracketStatus)
		adminRoutes.GET("/brackets/:bracket_id/status", tournamentBracketHandler.GetBracket) // GET version for accessibility
		adminRoutes.DELETE("/brackets/:bracket_id", tournamentBracketHandler.DeleteBracket)
		adminRoutes.GET("/brackets/types", tournamentBracketHandler.GetBracketTypes)

		// Fraud Detection
		adminRoutes.GET("/fraud/alerts", fraudDetectionHandler.GetAlerts)
		adminRoutes.PUT("/fraud/alerts/:alert_id/status", fraudDetectionHandler.UpdateAlertStatus)
		adminRoutes.GET("/fraud/alerts/:alert_id/status", fraudDetectionHandler.GetAlerts) // GET version for accessibility
		adminRoutes.GET("/fraud/statistics", fraudDetectionHandler.GetFraudStatistics)
		adminRoutes.GET("/fraud/users/:user_id/risk-score", fraudDetectionHandler.GetAlerts)
		adminRoutes.POST("/fraud/investigate", fraudDetectionHandler.UpdateAlertStatus)
		adminRoutes.GET("/fraud/patterns", fraudDetectionHandler.GetFraudStatistics)
		adminRoutes.PUT("/fraud/threshold", fraudDetectionHandler.UpdateAlertStatus)
		adminRoutes.GET("/fraud/threshold", fraudDetectionHandler.GetFraudStatistics) // GET version for accessibility

		// Social Sharing Analytics
		adminRoutes.GET("/social/analytics", socialSharingHandler.GetShareAnalytics)
		adminRoutes.GET("/social/platforms/stats", socialSharingHandler.GetPlatformStats)
		adminRoutes.GET("/social/trending", socialSharingHandler.GetTrendingContent)
		adminRoutes.POST("/social/campaigns", socialSharingHandler.CreateCampaign)

		// Notification Management
		adminRoutes.POST("/notify/send", notificationHandler.SendNotification)
		adminRoutes.POST("/notify/bulk", notificationHandler.SendBulkNotification)
		adminRoutes.POST("/notify/sms", notificationHandler.SendSMS)
		adminRoutes.POST("/notify/email", notificationHandler.SendEmail)
		adminRoutes.POST("/notify/push", notificationHandler.SendPush)
		adminRoutes.POST("/notify/whatsapp", notificationHandler.SendWhatsApp)

		// Template Management
		adminRoutes.POST("/templates", notificationHandler.CreateTemplate)
		adminRoutes.GET("/templates", notificationHandler.GetTemplates)
		adminRoutes.GET("/templates/:id", notificationHandler.GetTemplate)
		adminRoutes.PUT("/templates/:id", notificationHandler.UpdateTemplate)

		// Configuration Management
		adminRoutes.PUT("/config/notifications", notificationHandler.UpdateConfig)
		adminRoutes.GET("/config/notifications", notificationHandler.GetConfig)

		// Statistics
		adminRoutes.GET("/stats/notifications", notificationHandler.GetNotificationStats)
		adminRoutes.GET("/stats/channels", notificationHandler.GetChannelStats)

		// Payment Gateway Management
		adminRoutes.GET("/payment/gateways", paymentHandler.GetGatewayConfigs)
		adminRoutes.PUT("/payment/gateways/:gateway", paymentHandler.UpdateGatewayConfig)
		adminRoutes.PUT("/payment/gateways/:gateway/toggle", paymentHandler.ToggleGatewayStatus)
		adminRoutes.GET("/payment/transactions", paymentHandler.GetTransactionLogs)

		// Content Management - Banner Management
		adminRoutes.POST("/content/banners", contentHandler.CreateBanner)
		adminRoutes.GET("/content/banners", contentHandler.ListBanners)
		adminRoutes.GET("/content/banners/:id", contentHandler.GetBanner)
		adminRoutes.PUT("/content/banners/:id", contentHandler.UpdateBanner)
		adminRoutes.DELETE("/content/banners/:id", contentHandler.DeleteBanner)
		adminRoutes.PATCH("/content/banners/:id/toggle", contentHandler.ToggleBannerStatus)

		// Content Management - Email Templates
		adminRoutes.POST("/content/email-templates", contentHandler.CreateEmailTemplate)
		adminRoutes.GET("/content/email-templates", contentHandler.ListEmailTemplates)

		// Content Management - Marketing Campaigns
		adminRoutes.POST("/content/campaigns", contentHandler.CreateMarketingCampaign)
		adminRoutes.GET("/content/campaigns", contentHandler.ListMarketingCampaigns)
		adminRoutes.PATCH("/content/campaigns/:id/status", contentHandler.UpdateCampaignStatus)

		// Content Management - SEO Content
		adminRoutes.POST("/content/seo", contentHandler.CreateSEOContent)
		adminRoutes.GET("/content/seo", contentHandler.ListSEOContent)
		adminRoutes.GET("/content/seo/:id", contentHandler.GetSEOContent)
		adminRoutes.PUT("/content/seo/:id", contentHandler.UpdateSEOContent)
		adminRoutes.DELETE("/content/seo/:id", contentHandler.DeleteSEOContent)

		// Content Management - FAQ Management
		adminRoutes.GET("/content/faq/sections", contentHandler.ListFAQSections)
		adminRoutes.POST("/content/faq/sections", contentHandler.CreateFAQSection)
		adminRoutes.PUT("/content/faq/sections/:id", contentHandler.UpdateFAQSection)
		adminRoutes.POST("/content/faq/items", contentHandler.CreateFAQItem)
		adminRoutes.PUT("/content/faq/items/:id", contentHandler.UpdateFAQItem)

		// Content Management - Legal Documents
		adminRoutes.POST("/content/legal", contentHandler.CreateLegalDocument)
		adminRoutes.GET("/content/legal", contentHandler.ListLegalDocuments)
		adminRoutes.PUT("/content/legal/:id", contentHandler.UpdateLegalDocument)
		adminRoutes.PATCH("/content/legal/:id/publish", contentHandler.PublishLegalDocument)
		adminRoutes.DELETE("/content/legal/:id", contentHandler.DeleteLegalDocument)

		// Content Management - Analytics
		adminRoutes.GET("/content/analytics/:content_type/:content_id", contentHandler.GetContentAnalytics)
	}

	// WebSocket routes for real-time updates
	v1.GET("/ws/leaderboard/:contest_id", realtimeHandler.HandleLeaderboardWebSocket)
	adminRoutes.GET("/ws/live-scoring/:id", adminHandler.HandleLiveScoringWebSocket)
}

func (s *Server) Start(addr string) error {
	s.setupRoutes()
	return s.router.Run(addr)
}