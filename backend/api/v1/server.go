package v1

import (
	"database/sql"
	"net/http"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/pkg/cdn"
	"fantasy-esports-backend/api/v1/middleware"
	"fantasy-esports-backend/api/v1/handlers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	db     *sql.DB
	config *config.Config
	cdn    *cdn.CloudinaryClient
	upgrader websocket.Upgrader
}

func NewServer(db *sql.DB, cfg *config.Config) *Server {
	// Initialize CDN client
	cdnClient, err := cdn.NewCloudinaryClient(cfg.CloudinaryURL)
	if err != nil {
		panic("Failed to initialize CDN client: " + err.Error())
	}

	return &Server{
		db:     db,
		config: cfg,
		cdn:    cdnClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

func (s *Server) Start(addr string) error {
	gin.SetMode(s.config.GinMode)
	r := gin.Default()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(s.db, s.config, s.cdn)
	userHandler := handlers.NewUserHandler(s.db, s.config, s.cdn)
	gameHandler := handlers.NewGameHandler(s.db, s.config)
	contestHandler := handlers.NewContestHandler(s.db, s.config)
	walletHandler := handlers.NewWalletHandler(s.db, s.config)
	adminHandler := handlers.NewAdminHandler(s.db, s.config, s.cdn)

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorHandler())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "fantasy-esports-backend"})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/verify-mobile", authHandler.VerifyMobile)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(s.config.JWTSecret), authHandler.Logout)
			auth.POST("/social-login", authHandler.SocialLogin)
		}

		// User routes
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.POST("/kyc/upload", userHandler.UploadKYC)
			users.GET("/kyc/status", userHandler.GetKYCStatus)
			users.PUT("/preferences", userHandler.UpdatePreferences)
		}

		// Games and tournaments routes
		games := api.Group("/games")
		{
			games.GET("/", gameHandler.GetGames)
			games.GET("/:id", gameHandler.GetGameDetails)
		}

		tournaments := api.Group("/tournaments")
		{
			tournaments.GET("/", gameHandler.GetTournaments)
			tournaments.GET("/:id", gameHandler.GetTournamentDetails)
		}

		matches := api.Group("/matches")
		{
			matches.GET("/", gameHandler.GetMatches)
			matches.GET("/:id", gameHandler.GetMatchDetails)
			matches.GET("/:id/players", gameHandler.GetMatchPlayers)
			matches.GET("/:id/player-performance", gameHandler.GetPlayerPerformance)
		}

		// Contest routes
		contests := api.Group("/contests")
		contests.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			contests.GET("/", contestHandler.GetContests)
			contests.GET("/:id", contestHandler.GetContestDetails)
			contests.POST("/:id/join", contestHandler.JoinContest)
			contests.DELETE("/:id/leave", contestHandler.LeaveContest)
			contests.GET("/my-entries", contestHandler.GetMyEntries)
			contests.POST("/create-private", contestHandler.CreatePrivateContest)
		}

		// Fantasy team routes
		teams := api.Group("/teams")
		teams.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			teams.POST("/create", contestHandler.CreateTeam)
			teams.PUT("/:id", contestHandler.UpdateTeam)
			teams.GET("/my-teams", contestHandler.GetMyTeams)
			teams.GET("/:id", contestHandler.GetTeamDetails)
			teams.DELETE("/:id", contestHandler.DeleteTeam)
			teams.POST("/:id/clone", contestHandler.CloneTeam)
			teams.POST("/validate", contestHandler.ValidateTeam)
			teams.GET("/:id/performance", contestHandler.GetTeamPerformance)
		}

		// Leaderboard routes
		leaderboards := api.Group("/leaderboards")
		leaderboards.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			leaderboards.GET("/contests/:id", contestHandler.GetContestLeaderboard)
			leaderboards.GET("/live/:id", contestHandler.GetLiveLeaderboard)
			leaderboards.GET("/contests/:id/my-rank", contestHandler.GetMyRank)
			
			// Real-time leaderboard routes ⭐
			realtimeHandler := handlers.NewRealTimeLeaderboardHandler(s.db, s.config, contestHandler.GetLeaderboardService())
			leaderboards.GET("/real-time/:id", realtimeHandler.GetRealTimeLeaderboard)
			leaderboards.GET("/connections/:contest_id", realtimeHandler.GetActiveConnections)
			leaderboards.POST("/trigger-update/:contest_id", realtimeHandler.TriggerManualUpdate)
			
			// WebSocket endpoint for real-time updates
			leaderboards.GET("/ws/contest/:contest_id", realtimeHandler.HandleLeaderboardWebSocket)
		}

		// Wallet routes
		wallet := api.Group("/wallet")
		wallet.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			wallet.GET("/balance", walletHandler.GetBalance)
			wallet.POST("/deposit", walletHandler.Deposit)
			wallet.POST("/withdraw", walletHandler.Withdraw)
			wallet.GET("/transactions", walletHandler.GetTransactions)
			wallet.GET("/payment-methods", walletHandler.GetPaymentMethods)
			wallet.POST("/payment-methods", walletHandler.AddPaymentMethod)
		}

		// Payment routes
		payments := api.Group("/payments")
		payments.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			payments.GET("/:id/status", walletHandler.GetPaymentStatus)
		}

		// Referral routes
		referrals := api.Group("/referrals")
		referrals.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			referrals.GET("/my-stats", walletHandler.GetReferralStats)
			referrals.GET("/history", walletHandler.GetReferralHistory)
			referrals.POST("/apply", walletHandler.ApplyReferralCode)
			referrals.POST("/share", walletHandler.ShareReferral)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware(s.config.JWTSecret))
		{
			// Admin authentication
			admin.POST("/login", adminHandler.Login)

			// Match scoring routes
			admin.GET("/matches/live-scoring", adminHandler.GetLiveScoringMatches)
			admin.POST("/matches/:id/start-scoring", adminHandler.StartManualScoring)
			admin.POST("/matches/:id/events", adminHandler.AddMatchEvent)
			admin.PUT("/matches/:id/players/:player_id/stats", adminHandler.UpdatePlayerStats)
			admin.POST("/matches/:id/events/bulk", adminHandler.BulkUpdateEvents)
			admin.PUT("/matches/:id/score", adminHandler.UpdateMatchScore)
			admin.POST("/matches/:id/recalculate-points", adminHandler.RecalculatePoints)
			admin.GET("/matches/:id/dashboard", adminHandler.GetLiveDashboard)
			admin.POST("/matches/:id/complete", adminHandler.CompleteMatch)
			admin.GET("/matches/:id/events", adminHandler.GetMatchEvents)
			admin.PUT("/matches/:id/events/:event_id", adminHandler.EditMatchEvent)
			admin.DELETE("/matches/:id/events/:event_id", adminHandler.DeleteMatchEvent)

			// User management
			admin.GET("/users", adminHandler.GetUsers)
			admin.GET("/users/:id", adminHandler.GetUserDetails)
			admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			admin.PUT("/users/:id/kyc", adminHandler.ProcessKYC)

			// Game management
			admin.POST("/games", adminHandler.CreateGame)
			admin.PUT("/games/:id", adminHandler.UpdateGame)
			admin.DELETE("/games/:id", adminHandler.DeleteGame)

			// Tournament management
			admin.POST("/tournaments", adminHandler.CreateTournament)
			admin.PUT("/tournaments/:id", adminHandler.UpdateTournament)
			admin.POST("/matches", adminHandler.CreateMatch)
			admin.PUT("/matches/:id", adminHandler.UpdateMatch)

			// Contest management
			admin.POST("/contests", adminHandler.CreateContest)
			admin.PUT("/contests/:id", adminHandler.UpdateContest)
			admin.DELETE("/contests/:id", adminHandler.CancelContest)

			// Financial management
			admin.GET("/transactions", adminHandler.GetTransactions)
			admin.PUT("/withdrawals/:id/approve", adminHandler.ApproveWithdrawal)
			admin.PUT("/withdrawals/:id/reject", adminHandler.RejectWithdrawal)

			// System configuration
			admin.GET("/config", adminHandler.GetSystemConfig)
			admin.PUT("/config/:key", adminHandler.UpdateSystemConfig)
		}

		// WebSocket routes for live scoring
		api.GET("/admin/ws/live-scoring/:match_id", middleware.AdminWebSocketMiddleware(s.config.JWTSecret), adminHandler.HandleLiveScoringWebSocket)
	}

	return r.Run(addr)
}