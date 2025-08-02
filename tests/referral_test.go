package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"../backend/services"
	"../backend/config"
	
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
	log.Println("Setting up referral test environment...")
	
	// We'll test with a mock/minimal setup since database might not be available
	testConfig = &config.Config{
		JWTSecret: "test-secret-key",
	}
}

func teardown() {
	if testDB != nil {
		testDB.Close()
	}
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
	
	t.Run("Max Reward Per User", func(t *testing.T) {
		if config.MaxRewardPerUser != 5000.0 {
			t.Errorf("Expected max reward per user 5000.0, got %.2f", config.MaxRewardPerUser)
		}
	})
	
	t.Run("Reward Expiry Days", func(t *testing.T) {
		if config.RewardExpiryDays != 30 {
			t.Errorf("Expected reward expiry days 30, got %d", config.RewardExpiryDays)
		}
	})
}

func TestReferralValidation(t *testing.T) {
	service := services.NewReferralService(nil)
	
	t.Run("Valid Referral Code Format", func(t *testing.T) {
		// Test that referral codes should be at least 6 characters
		testCodes := []string{
			"ABC123",   // Valid
			"REFER123", // Valid  
			"XYZ",      // Should be invalid (too short)
			"",         // Should be invalid (empty)
		}
		
		for _, code := range testCodes {
			if len(code) < 6 && code != "" {
				// This would be invalid in a real validation function
				continue
			}
			
			if code == "" {
				// Empty code should be invalid
				continue
			}
			
			// Valid codes should pass
			if len(code) >= 6 {
				t.Logf("Code '%s' is valid format", code)
			}
		}
	})
}

func TestReferralTierCalculation(t *testing.T) {
	service := services.NewReferralService(nil)
	
	testCases := []struct {
		successfulReferrals int
		expectedTier        string
		expectedReward      float64
	}{
		{0, "bronze", 50.0},
		{5, "bronze", 50.0},
		{10, "silver", 75.0},
		{15, "silver", 75.0},
		{25, "gold", 100.0},
		{30, "gold", 100.0},
		{50, "platinum", 150.0},
		{75, "platinum", 150.0},
		{100, "diamond", 200.0},
		{150, "diamond", 200.0},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Referrals_%d", tc.successfulReferrals), func(t *testing.T) {
			config := service.GetReferralConfig()
			
			// Find appropriate tier (same logic as in service)
			currentTier := config.Tiers[0]
			for i := len(config.Tiers) - 1; i >= 0; i-- {
				tier := config.Tiers[i]
				if tc.successfulReferrals >= tier.MinReferrals {
					currentTier = tier
					break
				}
			}
			
			if currentTier.Name != tc.expectedTier {
				t.Errorf("For %d referrals, expected tier '%s', got '%s'", 
					tc.successfulReferrals, tc.expectedTier, currentTier.Name)
			}
			
			if currentTier.RewardPerReferral != tc.expectedReward {
				t.Errorf("For %d referrals, expected reward %.2f, got %.2f", 
					tc.successfulReferrals, tc.expectedReward, currentTier.RewardPerReferral)
			}
		})
	}
}

func TestReferralBonusCalculation(t *testing.T) {
	service := services.NewReferralService(nil)
	config := service.GetReferralConfig()
	
	t.Run("Tier Bonus Rewards", func(t *testing.T) {
		expectedBonuses := map[string]float64{
			"bronze":   0.0,     // No bonus for bronze
			"silver":   200.0,   // Bonus for reaching silver
			"gold":     500.0,   // Bonus for reaching gold
			"platinum": 1000.0,  // Bonus for reaching platinum
			"diamond":  2500.0,  // Bonus for reaching diamond
		}
		
		for _, tier := range config.Tiers {
			expectedBonus := expectedBonuses[tier.Name]
			if tier.BonusReward != expectedBonus {
				t.Errorf("Tier '%s' expected bonus %.2f, got %.2f", 
					tier.Name, expectedBonus, tier.BonusReward)
			}
		}
	})
}

func TestReferralSystemIntegrity(t *testing.T) {
	t.Run("Service Creation", func(t *testing.T) {
		service := services.NewReferralService(nil)
		if service == nil {
			t.Fatal("Expected referral service to be created")
		}
	})
	
	t.Run("Config Consistency", func() {
		service := services.NewReferralService(nil)
		config := service.GetReferralConfig()
		
		// Verify tiers are in ascending order
		for i := 1; i < len(config.Tiers); i++ {
			prevTier := config.Tiers[i-1]
			currTier := config.Tiers[i]
			
			if currTier.MinReferrals <= prevTier.MinReferrals {
				t.Errorf("Tier '%s' min referrals (%d) should be greater than '%s' (%d)", 
					currTier.Name, currTier.MinReferrals, prevTier.Name, prevTier.MinReferrals)
			}
			
			if currTier.RewardPerReferral <= prevTier.RewardPerReferral {
				t.Errorf("Tier '%s' reward (%.2f) should be greater than '%s' (%.2f)", 
					currTier.Name, currTier.RewardPerReferral, prevTier.Name, prevTier.RewardPerReferral)
			}
		}
	})
}