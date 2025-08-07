package services

import (
	"database/sql"
	"encoding/json"
	"fantasy-esports-backend/models"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

type FraudDetectionService struct {
	db *sql.DB
}

func NewFraudDetectionService(db *sql.DB) *FraudDetectionService {
	return &FraudDetectionService{db: db}
}

// Real-time fraud detection during user actions
func (s *FraudDetectionService) CheckUserAction(userID int64, action string, contextData map[string]interface{}, ipAddress, userAgent string) error {
	// Log the behavior
	err := s.logUserBehavior(userID, "", action, contextData, ipAddress, userAgent)
	if err != nil {
		return err
	}

	// Run multiple fraud checks
	alerts := []models.FraudAlert{}

	// Check for multiple accounts from same IP
	if alert := s.checkMultipleAccounts(userID, ipAddress); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check for suspicious wallet activity
	if strings.Contains(action, "wallet") || strings.Contains(action, "payment") {
		if alert := s.checkSuspiciousWalletActivity(userID, contextData); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Check for bot-like behavior
	if alert := s.checkBotBehavior(userID, action, contextData); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check for rapid team creation
	if action == "team_created" {
		if alert := s.checkRapidTeamCreation(userID); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Check for contest manipulation
	if strings.Contains(action, "contest") {
		if alert := s.checkContestManipulation(userID, contextData); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Check for referral fraud
	if action == "referral_signup" {
		if alert := s.checkReferralFraud(userID, contextData); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	// Store alerts
	for _, alert := range alerts {
		err = s.createAlert(alert)
		if err != nil {
			continue // Don't fail the entire process for logging issues
		}
	}

	return nil
}

func (s *FraudDetectionService) logUserBehavior(userID int64, sessionID, action string, contextData map[string]interface{}, ipAddress, userAgent string) error {
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		contextJSON = []byte("{}")
	}

	var sessionIDPtr *string
	if sessionID != "" {
		sessionIDPtr = &sessionID
	}

	var userIDPtr *int64
	if userID != 0 {
		userIDPtr = &userID
	}

	var ipPtr *string
	if ipAddress != "" {
		ipPtr = &ipAddress
	}

	var uaPtr *string
	if userAgent != "" {
		uaPtr = &userAgent
	}

	_, err = s.db.Exec(`
		INSERT INTO user_behavior_logs (user_id, session_id, action, context_data, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userIDPtr, sessionIDPtr, action, contextJSON, ipPtr, uaPtr)

	return err
}

func (s *FraudDetectionService) checkMultipleAccounts(userID int64, ipAddress string) *models.FraudAlert {
	if ipAddress == "" {
		return nil
	}

	// Check for multiple accounts from same IP in last 24 hours
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id)
		FROM user_behavior_logs
		WHERE ip_address = $1 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
		AND user_id IS NOT NULL
	`, ipAddress).Scan(&count)

	if err != nil || count < 3 {
		return nil
	}

	detectionData := map[string]interface{}{
		"ip_address":    ipAddress,
		"account_count": count,
		"check_type":    "multiple_accounts",
	}

	severity := "medium"
	if count >= 5 {
		severity = "high"
	}
	if count >= 10 {
		severity = "critical"
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "multiple_accounts_same_ip",
		Severity:      severity,
		Description:   fmt.Sprintf("%d accounts detected from same IP address in 24 hours", count),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) checkSuspiciousWalletActivity(userID int64, contextData map[string]interface{}) *models.FraudAlert {
	// Check for large transactions
	var amount float64
	if amt, exists := contextData["amount"]; exists {
		if amtFloat, ok := amt.(float64); ok {
			amount = amtFloat
		}
	}

	// Check recent deposit/withdrawal pattern
	var totalDeposits, totalWithdrawals int
	var largeTransactionCount int

	s.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN transaction_type = 'deposit' THEN 1 END) as deposits,
			COUNT(CASE WHEN transaction_type = 'withdrawal' THEN 1 END) as withdrawals,
			COUNT(CASE WHEN amount >= 10000 THEN 1 END) as large_transactions
		FROM wallet_transactions
		WHERE user_id = $1 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
	`, userID).Scan(&totalDeposits, &totalWithdrawals, &largeTransactionCount)

	// Suspicious patterns
	suspiciousReasons := []string{}

	if amount >= 50000 {
		suspiciousReasons = append(suspiciousReasons, "Large single transaction")
	}

	if totalDeposits >= 10 {
		suspiciousReasons = append(suspiciousReasons, "High frequency deposits")
	}

	if totalWithdrawals >= 5 {
		suspiciousReasons = append(suspiciousReasons, "High frequency withdrawals")
	}

	if largeTransactionCount >= 3 {
		suspiciousReasons = append(suspiciousReasons, "Multiple large transactions")
	}

	if len(suspiciousReasons) == 0 {
		return nil
	}

	detectionData := map[string]interface{}{
		"current_amount":         amount,
		"deposits_24h":          totalDeposits,
		"withdrawals_24h":       totalWithdrawals,
		"large_transactions_24h": largeTransactionCount,
		"reasons":               suspiciousReasons,
	}

	severity := "medium"
	if len(suspiciousReasons) >= 3 || amount >= 100000 {
		severity = "high"
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "suspicious_wallet_activity",
		Severity:      severity,
		Description:   fmt.Sprintf("Suspicious wallet activity detected: %s", strings.Join(suspiciousReasons, ", ")),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) checkBotBehavior(userID int64, action string, contextData map[string]interface{}) *models.FraudAlert {
	// Check action frequency in last hour
	var actionCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM user_behavior_logs
		WHERE user_id = $1 AND action = $2 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
	`, userID, action).Scan(&actionCount)

	if err != nil {
		return nil
	}

	// Define thresholds for different actions
	thresholds := map[string]int{
		"team_created":    20,
		"contest_joined":  50,
		"profile_updated": 10,
		"api_call":       100,
	}

	threshold, exists := thresholds[action]
	if !exists {
		threshold = 30 // Default threshold
	}

	if actionCount < threshold {
		return nil
	}

	// Check for uniform timing patterns (bot indicator)
	var timestamps []time.Time
	rows, err := s.db.Query(`
		SELECT created_at
		FROM user_behavior_logs
		WHERE user_id = $1 AND action = $2 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
		ORDER BY created_at DESC
		LIMIT 10
	`, userID, action)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ts time.Time
			if rows.Scan(&ts) == nil {
				timestamps = append(timestamps, ts)
			}
		}
	}

	// Calculate timing uniformity
	uniformityScore := s.calculateTimingUniformity(timestamps)

	detectionData := map[string]interface{}{
		"action":           action,
		"count_1h":         actionCount,
		"threshold":        threshold,
		"uniformity_score": uniformityScore,
		"timestamps":       timestamps,
	}

	severity := "medium"
	if uniformityScore > 0.8 || actionCount > threshold*2 {
		severity = "high"
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "bot_like_behavior",
		Severity:      severity,
		Description:   fmt.Sprintf("Possible bot behavior: %d %s actions in 1 hour (threshold: %d)", actionCount, action, threshold),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) checkRapidTeamCreation(userID int64) *models.FraudAlert {
	// Check team creation rate in last 10 minutes
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM user_teams
		WHERE user_id = $1 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '10 minutes'
	`, userID).Scan(&count)

	if err != nil || count < 5 {
		return nil
	}

	// Check if teams are identical (copy-paste behavior)
	var identicalTeams int
	s.db.QueryRow(`
		SELECT COUNT(*)
		FROM (
			SELECT captain_player_id, vice_captain_player_id, total_credits_used, COUNT(*) as team_count
			FROM user_teams
			WHERE user_id = $1 AND created_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
			GROUP BY captain_player_id, vice_captain_player_id, total_credits_used
			HAVING COUNT(*) > 1
		) identical_teams
	`, userID).Scan(&identicalTeams)

	detectionData := map[string]interface{}{
		"teams_10min":      count,
		"identical_teams":  identicalTeams,
		"check_window":     "10 minutes",
	}

	severity := "medium"
	if count >= 10 || identicalTeams >= 3 {
		severity = "high"
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "rapid_team_creation",
		Severity:      severity,
		Description:   fmt.Sprintf("Rapid team creation detected: %d teams in 10 minutes", count),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) checkContestManipulation(userID int64, contextData map[string]interface{}) *models.FraudAlert {
	contestID, exists := contextData["contest_id"]
	if !exists {
		return nil
	}

	contestIDInt, ok := contestID.(float64)
	if !ok {
		return nil
	}

	// Check if user is creating many private contests with same friends
	var privateContestCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM contests
		WHERE created_by = $1 AND contest_type = 'private' AND created_at >= CURRENT_DATE - INTERVAL '1 day'
	`, userID).Scan(&privateContestCount)

	if err != nil {
		return nil
	}

	// Check for win trading patterns
	var winRate float64
	s.db.QueryRow(`
		SELECT 
			CASE WHEN COUNT(*) > 0 THEN
				CAST(SUM(CASE WHEN rank = 1 THEN 1 ELSE 0 END) AS FLOAT) / COUNT(*)
			ELSE 0 END as win_rate
		FROM contest_participants cp
		JOIN contests c ON cp.contest_id = c.id
		WHERE cp.user_id = $1 AND c.contest_type = 'private' AND cp.joined_at >= CURRENT_DATE - INTERVAL '7 days'
	`, userID).Scan(&winRate)

	suspiciousReasons := []string{}

	if privateContestCount >= 10 {
		suspiciousReasons = append(suspiciousReasons, "High private contest creation")
	}

	if winRate > 0.8 && privateContestCount >= 5 {
		suspiciousReasons = append(suspiciousReasons, "Unusually high win rate in private contests")
	}

	if len(suspiciousReasons) == 0 {
		return nil
	}

	detectionData := map[string]interface{}{
		"contest_id":            int64(contestIDInt),
		"private_contests_24h":  privateContestCount,
		"private_contest_win_rate": winRate,
		"reasons":              suspiciousReasons,
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "contest_manipulation",
		Severity:      "high",
		Description:   fmt.Sprintf("Possible contest manipulation: %s", strings.Join(suspiciousReasons, ", ")),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) checkReferralFraud(userID int64, contextData map[string]interface{}) *models.FraudAlert {
	referrerID, exists := contextData["referrer_id"]
	if !exists {
		return nil
	}

	referrerIDInt, ok := referrerID.(float64)
	if !ok {
		return nil
	}

	// Check referrer's recent referral activity
	var referralCount int
	var sameIPCount int

	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM referrals
		WHERE referrer_user_id = $1 AND created_at >= CURRENT_DATE
	`, int64(referrerIDInt)).Scan(&referralCount)

	if err != nil {
		return nil
	}

	// Check IP address patterns
	s.db.QueryRow(`
		SELECT COUNT(DISTINCT r.referred_user_id)
		FROM referrals r
		JOIN user_behavior_logs ubl1 ON r.referrer_user_id = ubl1.user_id
		JOIN user_behavior_logs ubl2 ON r.referred_user_id = ubl2.user_id
		WHERE r.referrer_user_id = $1 
		AND r.created_at >= CURRENT_DATE
		AND ubl1.ip_address = ubl2.ip_address
		AND ubl1.ip_address IS NOT NULL
	`, int64(referrerIDInt)).Scan(&sameIPCount)

	suspiciousReasons := []string{}

	if referralCount >= 20 {
		suspiciousReasons = append(suspiciousReasons, "High daily referral count")
	}

	if sameIPCount >= 3 {
		suspiciousReasons = append(suspiciousReasons, "Multiple referrals from same IP")
	}

	if len(suspiciousReasons) == 0 {
		return nil
	}

	detectionData := map[string]interface{}{
		"referrer_id":      int64(referrerIDInt),
		"referrals_today":  referralCount,
		"same_ip_count":    sameIPCount,
		"reasons":          suspiciousReasons,
	}

	severity := "medium"
	if referralCount >= 50 || sameIPCount >= 5 {
		severity = "high"
	}

	return &models.FraudAlert{
		UserID:        &userID,
		AlertType:     "referral_fraud",
		Severity:      severity,
		Description:   fmt.Sprintf("Suspicious referral activity: %s", strings.Join(suspiciousReasons, ", ")),
		DetectionData: mustMarshal(detectionData),
	}
}

func (s *FraudDetectionService) calculateTimingUniformity(timestamps []time.Time) float64 {
	if len(timestamps) < 3 {
		return 0.0
	}

	// Calculate intervals between consecutive timestamps
	var intervals []float64
	for i := 1; i < len(timestamps); i++ {
		interval := timestamps[i-1].Sub(timestamps[i]).Seconds()
		intervals = append(intervals, math.Abs(interval))
	}

	if len(intervals) == 0 {
		return 0.0
	}

	// Calculate coefficient of variation (lower = more uniform)
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	if mean == 0 {
		return 1.0 // Perfect uniformity if all intervals are 0
	}

	variance := 0.0
	for _, interval := range intervals {
		variance += math.Pow(interval-mean, 2)
	}
	variance /= float64(len(intervals))
	stdDev := math.Sqrt(variance)

	cv := stdDev / mean
	
	// Convert to uniformity score (higher = more uniform)
	// CV close to 0 means high uniformity
	uniformity := math.Max(0.0, 1.0-math.Min(1.0, cv))
	return uniformity
}

func (s *FraudDetectionService) createAlert(alert models.FraudAlert) error {
	detectionDataJSON := alert.DetectionData
	if detectionDataJSON == nil {
		detectionDataJSON = []byte("{}")
	}

	_, err := s.db.Exec(`
		INSERT INTO fraud_alerts (user_id, alert_type, severity, description, detection_data)
		VALUES ($1, $2, $3, $4, $5)
	`, alert.UserID, alert.AlertType, alert.Severity, alert.Description, detectionDataJSON)

	return err
}

// Admin methods for managing alerts
func (s *FraudDetectionService) GetAlerts(status string, severity string, limit int) ([]models.FraudAlert, error) {
	query := `
		SELECT fa.id, fa.user_id, fa.alert_type, fa.severity, fa.description, fa.detection_data,
			fa.status, fa.assigned_to, fa.resolved_at, fa.resolution_notes, fa.created_at,
			u.first_name || ' ' || u.last_name as user_name,
			au.full_name as assigned_name
		FROM fraud_alerts fa
		LEFT JOIN users u ON fa.user_id = u.id
		LEFT JOIN admin_users au ON fa.assigned_to = au.id
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		query += fmt.Sprintf(" AND fa.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if severity != "" {
		query += fmt.Sprintf(" AND fa.severity = $%d", argIndex)
		args = append(args, severity)
		argIndex++
	}

	query += " ORDER BY fa.created_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.FraudAlert
	for rows.Next() {
		var a models.FraudAlert
		err := rows.Scan(&a.ID, &a.UserID, &a.AlertType, &a.Severity, &a.Description,
			&a.DetectionData, &a.Status, &a.AssignedTo, &a.ResolvedAt, &a.ResolutionNotes,
			&a.CreatedAt, &a.UserName, &a.AssignedName)
		if err != nil {
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

func (s *FraudDetectionService) UpdateAlertStatus(alertID int64, status string, assignedTo *int64, resolutionNotes *string) error {
	query := `
		UPDATE fraud_alerts SET 
			status = $1, 
			assigned_to = $2, 
			resolution_notes = $3,
			resolved_at = CASE WHEN $1 = 'resolved' THEN CURRENT_TIMESTAMP ELSE resolved_at END
		WHERE id = $4
	`
	
	return s.db.QueryRow(query, status, assignedTo, resolutionNotes, alertID).Err()
}

func (s *FraudDetectionService) GetFraudStatistics(days int) (map[string]interface{}, error) {
	query := `
		SELECT 
			alert_type,
			severity,
			COUNT(*) as count,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved_count
		FROM fraud_alerts
		WHERE created_at >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY alert_type, severity
		ORDER BY count DESC
	`

	rows, err := s.db.Query(fmt.Sprintf(query, days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]interface{}{
		"by_type":     map[string]interface{}{},
		"by_severity": map[string]interface{}{},
		"resolution_rate": 0.0,
		"total_alerts":    0,
	}

	byType := make(map[string]int)
	bySeverity := make(map[string]int)
	totalAlerts := 0
	totalResolved := 0

	for rows.Next() {
		var alertType, severity string
		var count, resolvedCount int

		err := rows.Scan(&alertType, &severity, &count, &resolvedCount)
		if err != nil {
			continue
		}

		byType[alertType] += count
		bySeverity[severity] += count
		totalAlerts += count
		totalResolved += resolvedCount
	}

	stats["by_type"] = byType
	stats["by_severity"] = bySeverity
	stats["total_alerts"] = totalAlerts

	if totalAlerts > 0 {
		stats["resolution_rate"] = float64(totalResolved) / float64(totalAlerts)
	}

	return stats, nil
}

// Helper function for JSON marshaling
func mustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}