package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/api/v1/handlers"
	
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Test database configuration
var testDB *sql.DB
var testConfig *config.Config

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	log.Println("Setting up test environment...")
	
	// Setup test database connection
	testConfig = &config.Config{
		DatabaseURL: "postgres://postgres:password@localhost:5432/fantasy_esports_test?sslmode=disable",
		JWTSecret:   "test-secret-key",
	}
	
	var err error
	testDB, err = sql.Open("postgres", testConfig.DatabaseURL)
	if err != nil {
		log.Printf("Failed to connect to test database: %v", err)
		log.Println("Continuing with in-memory testing...")
	} else {
		err = testDB.Ping()
		if err != nil {
			log.Printf("Failed to ping test database: %v", err)
			testDB = nil
		}
	}
	
	// Create test tables if database is available
	if testDB != nil {
		createTestTables()
	}
}

func teardown() {
	if testDB != nil {
		testDB.Close()
	}
}

func createTestTables() {
	// Create minimal test tables for referral testing
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			mobile VARCHAR(20) UNIQUE NOT NULL,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			email VARCHAR(100),
			is_verified BOOLEAN DEFAULT TRUE,
			is_active BOOLEAN DEFAULT TRUE,
			referral_code VARCHAR(20) UNIQUE,
			referred_by_code VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS referrals (
			id SERIAL PRIMARY KEY,
			referrer_user_id INTEGER NOT NULL,
			referred_user_id INTEGER NOT NULL,
			referral_code VARCHAR(20) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			reward_amount DECIMAL(10,2) DEFAULT 0,
			completion_criteria VARCHAR(50),
			completed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS user_wallets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER UNIQUE NOT NULL,
			bonus_balance DECIMAL(12,2) DEFAULT 0,
			deposit_balance DECIMAL(12,2) DEFAULT 0,
			winning_balance DECIMAL(12,2) DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS wallet_transactions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			transaction_type VARCHAR(20) NOT NULL,
			amount DECIMAL(12,2) NOT NULL,
			balance_type VARCHAR(10) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'completed',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		)`,
	}
	
	for _, query := range queries {
		_, err := testDB.Exec(query)
		if err != nil {
			log.Printf("Failed to create test table: %v", err)
		}
	}
}

func TestReferralService_ApplyReferralCode(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available, skipping integration tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Clear test data
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	// Create test users
	referrerID := createTestUser(t, "+919876543210", "REFER123")
	referredID := createTestUser(t, "+919876543211", "")
	
	t.Run("Apply Valid Referral Code", func(t *testing.T) {
		err := service.ApplyReferralCode(referredID, "REFER123")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Verify referral record was created
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM referrals WHERE referrer_user_id = $1 AND referred_user_id = $2", 
			referrerID, referredID).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("Expected 1 referral record, found %d", count)
		}
	})
	
	t.Run("Apply Invalid Referral Code", func(t *testing.T) {
		newUserID := createTestUser(t, "+919876543212", "")
		err := service.ApplyReferralCode(newUserID, "INVALID123")
		if err == nil {
			t.Fatal("Expected error for invalid referral code")
		}
	})
	
	t.Run("Self Referral Prevention", func(t *testing.T) {
		err := service.ApplyReferralCode(referrerID, "REFER123")
		if err == nil {
			t.Fatal("Expected error for self-referral")
		}
	})
	
	t.Run("Duplicate Referral Prevention", func(t *testing.T) {
		err := service.ApplyReferralCode(referredID, "REFER123")
		if err == nil {
			t.Fatal("Expected error for duplicate referral")
		}
	})
}

func TestReferralService_CheckAndCompleteReferral(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available, skipping integration tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Clear test data
	testDB.Exec("DELETE FROM wallet_transactions")
	testDB.Exec("DELETE FROM user_wallets")
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	// Create test users
	referrerID := createTestUser(t, "+919876543220", "COMP123")
	referredID := createTestUser(t, "+919876543221", "")
	
	// Apply referral code
	err := service.ApplyReferralCode(referredID, "COMP123")
	if err != nil {
		t.Fatalf("Failed to apply referral code: %v", err)
	}
	
	t.Run("Complete Referral on First Deposit", func(t *testing.T) {
		err := service.CheckAndCompleteReferral(referredID, "deposit")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Verify referral was completed
		var status string
		err = testDB.QueryRow("SELECT status FROM referrals WHERE referred_user_id = $1", referredID).Scan(&status)
		if err != nil || status != "completed" {
			t.Fatalf("Expected referral status 'completed', got '%s'", status)
		}
		
		// Verify referrer got bonus balance
		var bonusBalance float64
		err = testDB.QueryRow("SELECT bonus_balance FROM user_wallets WHERE user_id = $1", referrerID).Scan(&bonusBalance)
		if err != nil || bonusBalance <= 0 {
			t.Fatalf("Expected positive bonus balance, got %.2f", bonusBalance)
		}
	})
	
	t.Run("No Completion on Wrong Action", func(t *testing.T) {
		// Create another referral
		newReferredID := createTestUser(t, "+919876543222", "")
		err := service.ApplyReferralCode(newReferredID, "COMP123")
		if err != nil {
			t.Fatalf("Failed to apply referral code: %v", err)
		}
		
		err = service.CheckAndCompleteReferral(newReferredID, "profile_update")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		// Verify referral is still pending
		var status string
		err = testDB.QueryRow("SELECT status FROM referrals WHERE referred_user_id = $1", newReferredID).Scan(&status)
		if err != nil || status != "pending" {
			t.Fatalf("Expected referral status 'pending', got '%s'", status)
		}
	})
}

func TestReferralService_GetUserReferralStats(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available, skipping integration tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Clear test data
	testDB.Exec("DELETE FROM wallet_transactions")
	testDB.Exec("DELETE FROM user_wallets")
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	// Create test user
	userID := createTestUser(t, "+919876543230", "STATS123")
	
	// Create some referral records
	referred1 := createTestUser(t, "+919876543231", "")
	referred2 := createTestUser(t, "+919876543232", "")
	
	service.ApplyReferralCode(referred1, "STATS123")
	service.ApplyReferralCode(referred2, "STATS123")
	
	// Complete one referral
	service.CheckAndCompleteReferral(referred1, "deposit")
	
	t.Run("Get Referral Statistics", func(t *testing.T) {
		stats, err := service.GetUserReferralStats(userID)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if stats.ReferralCode != "STATS123" {
			t.Errorf("Expected referral code 'STATS123', got '%s'", stats.ReferralCode)
		}
		
		if stats.TotalReferrals != 2 {
			t.Errorf("Expected 2 total referrals, got %d", stats.TotalReferrals)
		}
		
		if stats.SuccessfulReferrals != 1 {
			t.Errorf("Expected 1 successful referral, got %d", stats.SuccessfulReferrals)
		}
		
		if stats.CurrentTier != "bronze" {
			t.Errorf("Expected 'bronze' tier, got '%s'", stats.CurrentTier)
		}
		
		if stats.NextTierRequirement != 9 { // 10 - 1 = 9 more needed for silver
			t.Errorf("Expected 9 for next tier requirement, got %d", stats.NextTierRequirement)
		}
	})
}

func TestReferralService_GetUserTierReward(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available, skipping integration tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Clear test data
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	// Create test user
	userID := createTestUser(t, "+919876543240", "TIER123")
	
	t.Run("Bronze Tier Reward", func(t *testing.T) {
		reward, err := service.GetUserTierReward(userID)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if reward != 50.0 {
			t.Errorf("Expected bronze tier reward 50.0, got %.2f", reward)
		}
	})
	
	t.Run("Silver Tier Reward", func(t *testing.T) {
		// Create 10 completed referrals to reach silver tier
		for i := 0; i < 10; i++ {
			referredID := createTestUser(t, fmt.Sprintf("+91987654324%d", i+1), "")
			service.ApplyReferralCode(referredID, "TIER123")
			
			// Mark as completed
			testDB.Exec("UPDATE referrals SET status = 'completed', completed_at = NOW() WHERE referred_user_id = $1", referredID)
		}
		
		reward, err := service.GetUserTierReward(userID)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if reward != 75.0 {
			t.Errorf("Expected silver tier reward 75.0, got %.2f", reward)
		}
	})
}

func TestReferralHTTPHandlers(t *testing.T) {
	if testDB == nil {
		t.Skip("Database not available, skipping HTTP handler tests")
	}
	
	// Setup test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	walletHandler := handlers.NewWalletHandler(testDB, testConfig)
	
	// Setup routes
	api := router.Group("/api/v1")
	referrals := api.Group("/referrals")
	{
		referrals.GET("/my-stats", mockAuthMiddleware(1), walletHandler.GetReferralStats)
		referrals.GET("/history", mockAuthMiddleware(1), walletHandler.GetReferralHistory)
		referrals.POST("/apply", mockAuthMiddleware(2), walletHandler.ApplyReferralCode)
		referrals.POST("/share", mockAuthMiddleware(1), walletHandler.ShareReferral)
		referrals.GET("/leaderboard", mockAuthMiddleware(1), walletHandler.GetReferralLeaderboard)
	}
	
	// Clear test data
	testDB.Exec("DELETE FROM wallet_transactions")
	testDB.Exec("DELETE FROM user_wallets")
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	// Create test users
	user1ID := createTestUser(t, "+919876543250", "HTTP123")
	user2ID := createTestUser(t, "+919876543251", "")
	
	t.Run("GET /referrals/my-stats", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/referrals/my-stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response["success"].(bool) {
			t.Error("Expected success to be true")
		}
		
		stats := response["referral_stats"].(map[string]interface{})
		if stats["referral_code"] != "HTTP123" {
			t.Errorf("Expected referral code 'HTTP123', got %v", stats["referral_code"])
		}
	})
	
	t.Run("POST /referrals/apply", func(t *testing.T) {
		body := map[string]string{"referral_code": "HTTP123"}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("POST", "/api/v1/referrals/apply", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response["success"].(bool) {
			t.Error("Expected success to be true")
		}
	})
	
	t.Run("POST /referrals/share", func(t *testing.T) {
		body := map[string]interface{}{
			"method": "whatsapp",
			"message": "Join me on Fantasy Esports!",
		}
		jsonBody, _ := json.Marshal(body)
		
		req, _ := http.NewRequest("POST", "/api/v1/referrals/share", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response["success"].(bool) {
			t.Error("Expected success to be true")
		}
		
		if response["method"] != "whatsapp" {
			t.Errorf("Expected method 'whatsapp', got %v", response["method"])
		}
		
		if response["whatsapp_url"] == nil {
			t.Error("Expected whatsapp_url to be present")
		}
	})
	
	t.Run("GET /referrals/leaderboard", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/referrals/leaderboard?limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response["success"].(bool) {
			t.Error("Expected success to be true")
		}
		
		if response["limit"].(float64) != 10 {
			t.Errorf("Expected limit 10, got %v", response["limit"])
		}
	})
}

func TestReferralConfig(t *testing.T) {
	service := services.NewReferralService(nil) // Can test config without DB
	
	config := service.GetReferralConfig()
	
	t.Run("Config Values", func(t *testing.T) {
		if config.BaseReward != 50.0 {
			t.Errorf("Expected base reward 50.0, got %.2f", config.BaseReward)
		}
		
		if config.CompletionCriteria != "first_deposit" {
			t.Errorf("Expected completion criteria 'first_deposit', got '%s'", config.CompletionCriteria)
		}
		
		if len(config.Tiers) != 5 {
			t.Errorf("Expected 5 tiers, got %d", len(config.Tiers))
		}
		
		// Check tier progression
		expectedTiers := []struct {
			name        string
			minReferrals int
			reward      float64
		}{
			{"bronze", 0, 50.0},
			{"silver", 10, 75.0},
			{"gold", 25, 100.0},
			{"platinum", 50, 150.0},
			{"diamond", 100, 200.0},
		}
		
		for i, expected := range expectedTiers {
			tier := config.Tiers[i]
			if tier.Name != expected.name {
				t.Errorf("Expected tier %d name '%s', got '%s'", i, expected.name, tier.Name)
			}
			if tier.MinReferrals != expected.minReferrals {
				t.Errorf("Expected tier %d min referrals %d, got %d", i, expected.minReferrals, tier.MinReferrals)
			}
			if tier.RewardPerReferral != expected.reward {
				t.Errorf("Expected tier %d reward %.2f, got %.2f", i, expected.reward, tier.RewardPerReferral)
			}
		}
	})
}

// Helper functions

func createTestUser(t *testing.T, mobile, referralCode string) int64 {
	var userID int64
	query := `INSERT INTO users (mobile, first_name, last_name, referral_code, created_at, updated_at) 
			  VALUES ($1, 'Test', 'User', $2, NOW(), NOW()) RETURNING id`
	
	if referralCode == "" {
		referralCode = generateTestReferralCode()
	}
	
	err := testDB.QueryRow(query, mobile, referralCode).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	return userID
}

func generateTestReferralCode() string {
	return fmt.Sprintf("TEST%d", time.Now().UnixNano()%100000)
}

func mockAuthMiddleware(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// Benchmark tests

func BenchmarkApplyReferralCode(b *testing.B) {
	if testDB == nil {
		b.Skip("Database not available, skipping benchmark tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Setup test data
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	referrerID := createBenchmarkUser(b, "+919876543260", "BENCH123")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		userID := createBenchmarkUser(b, fmt.Sprintf("+91987654326%d", i+1), "")
		service.ApplyReferralCode(userID, "BENCH123")
	}
}

func BenchmarkGetReferralStats(b *testing.B) {
	if testDB == nil {
		b.Skip("Database not available, skipping benchmark tests")
	}
	
	service := services.NewReferralService(testDB)
	
	// Setup test data
	testDB.Exec("DELETE FROM referrals")
	testDB.Exec("DELETE FROM users")
	
	userID := createBenchmarkUser(b, "+919876543270", "BENCH456")
	
	// Create some referrals
	for i := 0; i < 50; i++ {
		referredID := createBenchmarkUser(b, fmt.Sprintf("+91987654327%d", i+1), "")
		service.ApplyReferralCode(referredID, "BENCH456")
		if i%2 == 0 {
			testDB.Exec("UPDATE referrals SET status = 'completed' WHERE referred_user_id = $1", referredID)
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		service.GetUserReferralStats(userID)
	}
}

func createBenchmarkUser(b *testing.B, mobile, referralCode string) int64 {
	var userID int64
	query := `INSERT INTO users (mobile, first_name, last_name, referral_code, created_at, updated_at) 
			  VALUES ($1, 'Bench', 'User', $2, NOW(), NOW()) RETURNING id`
	
	if referralCode == "" {
		referralCode = fmt.Sprintf("BENCH%d", time.Now().UnixNano()%100000)
	}
	
	err := testDB.QueryRow(query, mobile, referralCode).Scan(&userID)
	if err != nil {
		b.Fatalf("Failed to create benchmark user: %v", err)
	}
	
	return userID
}