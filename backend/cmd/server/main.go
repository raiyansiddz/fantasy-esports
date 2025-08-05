// @title Fantasy Esports API
// @version 1.0
// @description Backend API for fantasy esports platform with full admin control and manual scoring
// @termsOfService http://yourdomain.com/terms/

// @contact.name Raiyan Siddique
// @contact.email support@yourdomain.com

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/db"
	"fantasy-esports-backend/api/v1"
	"fantasy-esports-backend/pkg/logger"
	_ "fantasy-esports-backend/docs"
	"log"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize database
	database, err := db.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()
	
	// Run database migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	
	// Initialize and start server
	server := v1.NewServer(database, cfg)
	log.Printf("🚀 Fantasy Esports Backend Server starting on port %s", cfg.Port)
	log.Fatal(server.Start(":" + cfg.Port))
}