package main

import (
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/db"
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
	log.Println("Running database migrations...")
	if err := db.RunMigrations(database); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	
	log.Println("✅ Database migrations completed successfully!")
}