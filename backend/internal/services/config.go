package services

import (
	"database/sql"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/logger"
	"fmt"
	"time"
)

type ConfigService struct {
	db *sql.DB
}

func NewConfigService(db *sql.DB) *ConfigService {
	return &ConfigService{
		db: db,
	}
}

// GetGatewayConfig gets configuration for a specific gateway
func (s *ConfigService) GetGatewayConfig(gateway string) (*models.PaymentGatewayConfig, error) {
	query := `
		SELECT id, gateway, key1, key2, client_version, is_live, enabled, currency, created_at, updated_at
		FROM payment_gateway_configs
		WHERE gateway = $1`
	
	var config models.PaymentGatewayConfig
	err := s.db.QueryRow(query, gateway).Scan(
		&config.ID,
		&config.Gateway,
		&config.Key1,
		&config.Key2,
		&config.ClientVersion,
		&config.IsLive,
		&config.Enabled,
		&config.Currency,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gateway %s not configured", gateway)
		}
		return nil, fmt.Errorf("failed to get gateway config: %v", err)
	}
	
	return &config, nil
}

// GetAllGatewayConfigs gets all gateway configurations
func (s *ConfigService) GetAllGatewayConfigs() ([]models.PaymentGatewayConfig, error) {
	query := `
		SELECT id, gateway, key1, key2, client_version, is_live, enabled, currency, created_at, updated_at
		FROM payment_gateway_configs
		ORDER BY gateway`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway configs: %v", err)
	}
	defer rows.Close()
	
	var configs []models.PaymentGatewayConfig
	for rows.Next() {
		var config models.PaymentGatewayConfig
		err := rows.Scan(
			&config.ID,
			&config.Gateway,
			&config.Key1,
			&config.Key2,
			&config.ClientVersion,
			&config.IsLive,
			&config.Enabled,
			&config.Currency,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gateway config: %v", err)
		}
		configs = append(configs, config)
	}
	
	return configs, nil
}

// UpdateGatewayConfig updates or creates gateway configuration
func (s *ConfigService) UpdateGatewayConfig(config *models.PaymentGatewayConfig) error {
	// First check if config exists
	existing, err := s.GetGatewayConfig(config.Gateway)
	if err != nil && err.Error() != fmt.Sprintf("gateway %s not configured", config.Gateway) {
		return fmt.Errorf("failed to check existing config: %v", err)
	}
	
	var query string
	var args []interface{}
	
	if existing == nil {
		// Insert new config
		query = `
			INSERT INTO payment_gateway_configs (gateway, key1, key2, client_version, is_live, enabled, currency, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		args = []interface{}{
			config.Gateway,
			config.Key1,
			config.Key2,
			config.ClientVersion,
			config.IsLive,
			config.Enabled,
			config.Currency,
			time.Now(),
			time.Now(),
		}
	} else {
		// Update existing config
		query = `
			UPDATE payment_gateway_configs
			SET key1 = $2, key2 = $3, client_version = $4, is_live = $5, enabled = $6, currency = $7, updated_at = $8
			WHERE gateway = $1`
		args = []interface{}{
			config.Gateway,
			config.Key1,
			config.Key2,
			config.ClientVersion,
			config.IsLive,
			config.Enabled,
			config.Currency,
			time.Now(),
		}
	}
	
	_, err = s.db.Exec(query, args...)
	if err != nil {
		logger.Error("Failed to update gateway config", "gateway", config.Gateway, "error", err)
		return fmt.Errorf("failed to update gateway config: %v", err)
	}
	
	logger.Info("Gateway config updated successfully", "gateway", config.Gateway, "is_live", config.IsLive, "enabled", config.Enabled)
	return nil
}

// ToggleGatewayStatus enables/disables a gateway
func (s *ConfigService) ToggleGatewayStatus(gateway string, enabled bool) error {
	query := `UPDATE payment_gateway_configs SET enabled = $1, updated_at = $2 WHERE gateway = $3`
	
	result, err := s.db.Exec(query, enabled, time.Now(), gateway)
	if err != nil {
		return fmt.Errorf("failed to toggle gateway status: %v", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("gateway %s not found", gateway)
	}
	
	status := "disabled"
	if enabled {
		status = "enabled"
	}
	
	logger.Info("Gateway status updated", "gateway", gateway, "status", status)
	return nil
}

// InitializeDefaultConfigs creates default configurations for supported gateways
func (s *ConfigService) InitializeDefaultConfigs() error {
	// Default Razorpay config with test credentials
	razorpayConfig := &models.PaymentGatewayConfig{
		Gateway:       "razorpay",
		Key1:          "rzp_test_SvOV4KyH7o0FSg",
		Key2:          "pw7srjSx9oJeswvua7xhLPuk",
		ClientVersion: "1",
		IsLive:        false,
		Enabled:       true,
		Currency:      "INR",
	}
	
	// Default PhonePe config with test credentials
	phonepeConfig := &models.PaymentGatewayConfig{
		Gateway:       "phonepe",
		Key1:          "TEST-M22RDIMXCYCLN_25080",
		Key2:          "MmE5YTA4N2ItZDcwMy00MGYzLTljYzAtZmUwMjA0MTlhNzQ4",
		ClientVersion: "1",
		IsLive:        false,
		Enabled:       true,
		Currency:      "INR",
	}
	
	// Check if configs already exist
	existingRazorpay, _ := s.GetGatewayConfig("razorpay")
	if existingRazorpay == nil {
		if err := s.UpdateGatewayConfig(razorpayConfig); err != nil {
			logger.Error("Failed to initialize Razorpay config", "error", err)
		} else {
			logger.Info("Initialized default Razorpay configuration")
		}
	}
	
	existingPhonePe, _ := s.GetGatewayConfig("phonepe")
	if existingPhonePe == nil {
		if err := s.UpdateGatewayConfig(phonepeConfig); err != nil {
			logger.Error("Failed to initialize PhonePe config", "error", err)
		} else {
			logger.Info("Initialized default PhonePe configuration")
		}
	}
	
	return nil
}

// ValidateGatewayConfig validates gateway configuration
func (s *ConfigService) ValidateGatewayConfig(gateway string) error {
	config, err := s.GetGatewayConfig(gateway)
	if err != nil {
		return err
	}
	
	// Basic validation
	if config.Key1 == "" || config.Key2 == "" {
		return fmt.Errorf("missing required keys for gateway %s", gateway)
	}
	
	if config.Currency == "" {
		return fmt.Errorf("missing currency for gateway %s", gateway)
	}
	
	// Gateway-specific validation
	switch gateway {
	case "razorpay":
		// Validate Razorpay key format
		if config.IsLive && !isValidRazorpayLiveKey(config.Key1) {
			return fmt.Errorf("invalid Razorpay live key format")
		}
		if !config.IsLive && !isValidRazorpayTestKey(config.Key1) {
			return fmt.Errorf("invalid Razorpay test key format")
		}
		
	case "phonepe":
		// Validate PhonePe config
		if config.ClientVersion == "" {
			return fmt.Errorf("missing client version for PhonePe")
		}
	}
	
	return nil
}

// Helper functions for validation
func isValidRazorpayLiveKey(key string) bool {
	return len(key) > 0 && (key[:8] == "rzp_live" || key[:8] == "rzp_test") // Allow both for flexibility
}

func isValidRazorpayTestKey(key string) bool {
	return len(key) > 0 && key[:8] == "rzp_test"
}

// GetEnabledGateways returns list of enabled gateways
func (s *ConfigService) GetEnabledGateways() ([]string, error) {
	query := `SELECT gateway FROM payment_gateway_configs WHERE enabled = true ORDER BY gateway`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled gateways: %v", err)
	}
	defer rows.Close()
	
	var gateways []string
	for rows.Next() {
		var gateway string
		if err := rows.Scan(&gateway); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}
	
	return gateways, nil
}

// GetGatewayStats returns statistics for each gateway
func (s *ConfigService) GetGatewayStats() (map[string]interface{}, error) {
	query := `
		SELECT 
			pt.gateway,
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN pt.status = 'completed' THEN 1 END) as successful_transactions,
			COUNT(CASE WHEN pt.status = 'failed' THEN 1 END) as failed_transactions,
			COUNT(CASE WHEN pt.status = 'pending' THEN 1 END) as pending_transactions,
			COALESCE(SUM(CASE WHEN pt.status = 'completed' THEN pt.amount END), 0) as total_amount
		FROM payment_transactions pt
		WHERE pt.created_at >= NOW() - INTERVAL '30 days'
		GROUP BY pt.gateway
		ORDER BY pt.gateway`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway stats: %v", err)
	}
	defer rows.Close()
	
	stats := make(map[string]interface{})
	
	for rows.Next() {
		var gateway string
		var total, successful, failed, pending int64
		var totalAmount float64
		
		err := rows.Scan(&gateway, &total, &successful, &failed, &pending, &totalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gateway stats: %v", err)
		}
		
		successRate := 0.0
		if total > 0 {
			successRate = float64(successful) / float64(total) * 100
		}
		
		stats[gateway] = map[string]interface{}{
			"total_transactions":      total,
			"successful_transactions": successful,
			"failed_transactions":     failed,
			"pending_transactions":    pending,
			"total_amount":           totalAmount,
			"success_rate":           successRate,
		}
	}
	
	return stats, nil
}