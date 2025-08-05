package v1

import (
	"database/sql"
	"fantasy-esports-backend/api/v1/handlers"
	"fantasy-esports-backend/api/v1/middleware"
	"fantasy-esports-backend/config"
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
	paymentService := services.NewPaymentService(s.db)
	
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
	paymentHandler := handlers.NewPaymentHandler(s.db, s.config, paymentService)

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

		// Admin login (public)
		v1.POST("/admin/login", adminHandler.Login)
	}

	// Protected user routes (require user authentication)
	userRoutes := v1.Group("")
	userRoutes.Use(middleware.AuthMiddleware(s.config.JWTSecret))
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
	}

	// WebSocket routes for real-time updates
	v1.GET("/ws/leaderboard/:contest_id", realtimeHandler.HandleLeaderboardWebSocket)
	adminRoutes.GET("/ws/live-scoring/:match_id", adminHandler.HandleLiveScoringWebSocket)
}

func (s *Server) Start(addr string) error {
	s.setupRoutes()
	return s.router.Run(addr)
}