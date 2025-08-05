package main

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/db"
	"fantasy-esports-backend/internal/services"
	"log"
)

func initializeTestUser(database *sql.DB) error {
	// Check if test user already exists
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE mobile = $1)", "+919876543210").Scan(&exists)
	if err != nil {
		return err
	}
	
	if exists {
		log.Println("Test user already exists, skipping creation")
		return nil
	}
	
	// Create test user with referral code and wallet
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Insert user
	var userID int64
	userQuery := `
		INSERT INTO users (mobile, first_name, last_name, email, is_verified, referral_code, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`
	
	err = tx.QueryRow(userQuery, "+919876543210", "Test", "User", "test@example.com", true, "TEST123").Scan(&userID)
	if err != nil {
		return err
	}
	
	// Create user wallet
	walletQuery := `
		INSERT INTO user_wallets (user_id, bonus_balance, deposit_balance, winning_balance, total_balance, updated_at)
		VALUES ($1, 0.0, 0.0, 0.0, 0.0, NOW())
	`
	
	_, err = tx.Exec(walletQuery, userID)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize database
	database, err := db.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()
	
	// Initialize config service
	configService := services.NewConfigService(database)
	
	// Initialize default gateway configurations
	log.Println("Initializing payment gateway configurations...")
	
	if err := configService.InitializeDefaultConfigs(); err != nil {
		log.Fatal("Failed to initialize gateway configs:", err)
	}
	
	// Initialize test user for payment testing
	log.Println("Initializing test user for payment testing...")
	
	if err := initializeTestUser(database); err != nil {
		log.Printf("Warning: Failed to initialize test user: %v", err)
	} else {
		log.Println("✅ Test user initialized successfully!")
	}
	
	log.Println("✅ Payment gateway configurations initialized successfully!")
	log.Println("Default configurations:")
	log.Println("- Razorpay: TEST environment with test credentials")
	log.Println("- PhonePe: TEST environment with test credentials")
	log.Println("- Test User: +919876543210 (for payment testing)")
	log.Println("")
	log.Println("You can now:")
	log.Println("1. Update configurations via Admin APIs")
	log.Println("2. Switch to production credentials when ready")
	log.Println("3. Enable/disable gateways as needed")
	log.Println("4. Test payments using mobile +919876543210 with OTP 123456")
}