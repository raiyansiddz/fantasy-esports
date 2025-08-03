package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/errors"
	"fantasy-esports-backend/pkg/logger"
)

type ReportingService struct {
	db *sql.DB
}

func NewReportingService(db *sql.DB) *ReportingService {
	return &ReportingService{db: db}
}

// GenerateReport creates a new report request and starts generation
func (s *ReportingService) GenerateReport(request models.ReportRequest, generatedBy int64) (*models.GeneratedReport, error) {
	// Validate report request
	if err := s.validateReportRequest(request); err != nil {
		return nil, err
	}

	// Create report record
	report := &models.GeneratedReport{
		ReportType:  request.ReportType,
		Format:      request.Format,
		Status:      models.ReportStatusPending,
		Title:       s.generateReportTitle(request),
		Description: &request.Description,
		GeneratedBy: generatedBy,
		RequestData: request,
		CreatedAt:   time.Now(),
	}

	// Calculate expiry (reports expire after 30 days)
	expiryTime := time.Now().AddDate(0, 0, 30)
	report.ExpiresAt = &expiryTime

	// Insert report record
	query := `
		INSERT INTO generated_reports (report_type, format, status, title, description, generated_by, request_data, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	requestDataJSON, _ := json.Marshal(request)
	err := s.db.QueryRow(query,
		report.ReportType, report.Format, report.Status, report.Title,
		report.Description, report.GeneratedBy, requestDataJSON,
		report.CreatedAt, report.ExpiresAt,
	).Scan(&report.ID)

	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Start report generation asynchronously
	go s.processReport(report.ID)

	return report, nil
}

// processReport generates the actual report data
func (s *ReportingService) processReport(reportID int64) {
	// Update status to generating
	s.updateReportStatus(reportID, models.ReportStatusGenerating, nil)

	// Get report details
	report, err := s.GetReport(reportID)
	if err != nil {
		errMsg := err.Error()
		s.updateReportStatus(reportID, models.ReportStatusFailed, &errMsg)
		return
	}

	// Generate report data based on type
	var reportData interface{}
	switch report.ReportType {
	case models.ReportTypeFinancial:
		reportData, err = s.generateFinancialReport(report.RequestData)
	case models.ReportTypeUser:
		reportData, err = s.generateUserReport(report.RequestData)
	case models.ReportTypeContest:
		reportData, err = s.generateContestReport(report.RequestData)
	case models.ReportTypeGame:
		reportData, err = s.generateGameReport(report.RequestData)
	case models.ReportTypeCompliance:
		reportData, err = s.generateComplianceReport(report.RequestData)
	case models.ReportTypeReferral:
		reportData, err = s.generateReferralReport(report.RequestData)
	default:
		err = errors.NewError(errors.ErrInvalidRequest, "Unsupported report type")
	}

	if err != nil {
		errMsg := err.Error()
		s.updateReportStatus(reportID, models.ReportStatusFailed, &errMsg)
		return
	}

	// Store report data
	resultJSON, _ := json.Marshal(reportData)
	query := `
		UPDATE generated_reports 
		SET result_data = $1, status = $2, completed_at = $3
		WHERE id = $4
	`
	completedAt := time.Now()
	_, err = s.db.Exec(query, resultJSON, models.ReportStatusCompleted, completedAt, reportID)

	if err != nil {
		errMsg := err.Error()
		s.updateReportStatus(reportID, models.ReportStatusFailed, &errMsg)
		return
	}

	logger.Info("Report generation completed", map[string]interface{}{
		"report_id": reportID,
		"type":      report.ReportType,
	})
}

// generateFinancialReport creates financial analytics report
func (s *ReportingService) generateFinancialReport(request models.ReportRequest) (*models.FinancialReport, error) {
	report := &models.FinancialReport{}

	// Get financial summary
	summaryQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN transaction_type = 'deposit' AND status = 'completed' THEN amount END), 0) as total_deposits,
			COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal' AND status = 'completed' THEN amount END), 0) as total_withdrawals,
			COALESCE(SUM(CASE WHEN transaction_type = 'withdrawal' AND status = 'pending' THEN amount END), 0) as pending_withdrawals,
			COALESCE(SUM(CASE WHEN transaction_type = 'contest_fee' AND status = 'completed' THEN amount END), 0) as contest_revenue,
			COUNT(*) as transaction_count,
			AVG(amount) as avg_transaction_value
		FROM wallet_transactions
		WHERE created_at BETWEEN $1 AND $2
	`

	err := s.db.QueryRow(summaryQuery, request.DateFrom, request.DateTo).Scan(
		&report.Summary.TotalDeposits,
		&report.Summary.TotalWithdrawals,
		&report.Summary.PendingWithdrawals,
		&report.Summary.TotalRevenue,
		&report.Summary.TransactionCount,
		&report.Summary.AvgTransactionValue,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	report.Summary.NetRevenue = report.Summary.TotalRevenue - report.Summary.TotalWithdrawals

	// Get transactions by type
	transactionTypeQuery := `
		SELECT 
			transaction_type,
			COUNT(*) as count,
			SUM(amount) as total_amount,
			AVG(amount) as avg_amount
		FROM wallet_transactions
		WHERE created_at BETWEEN $1 AND $2 AND status = 'completed'
		GROUP BY transaction_type
		ORDER BY total_amount DESC
	`

	typeRows, err := s.db.Query(transactionTypeQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var summary models.TransactionSummary
		typeRows.Scan(&summary.Type, &summary.Count, &summary.TotalAmount, &summary.AvgAmount)
		report.TransactionsByType = append(report.TransactionsByType, summary)
	}

	// Get daily revenue trend
	dailyRevenueQuery := `
		SELECT 
			DATE(created_at) as date,
			SUM(CASE WHEN transaction_type = 'contest_fee' THEN amount ELSE 0 END) as revenue,
			SUM(CASE WHEN transaction_type = 'deposit' THEN amount ELSE 0 END) as deposits,
			SUM(CASE WHEN transaction_type = 'withdrawal' THEN amount ELSE 0 END) as withdrawals
		FROM wallet_transactions
		WHERE created_at BETWEEN $1 AND $2 AND status = 'completed'
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	dailyRows, err := s.db.Query(dailyRevenueQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer dailyRows.Close()

	for dailyRows.Next() {
		var daily models.DailyRevenue
		dailyRows.Scan(&daily.Date, &daily.Revenue, &daily.Deposits, &daily.Withdrawals)
		report.DailyRevenue = append(report.DailyRevenue, daily)
	}

	// Get top spenders
	topSpendersQuery := `
		SELECT 
			u.id,
			COALESCE(u.first_name || ' ' || u.last_name, u.mobile) as username,
			SUM(wt.amount) as total_spent,
			COUNT(*) as transaction_count
		FROM users u
		JOIN wallet_transactions wt ON wt.user_id = u.id
		WHERE wt.created_at BETWEEN $1 AND $2 
		AND wt.status = 'completed' 
		AND wt.transaction_type IN ('contest_fee', 'deposit')
		GROUP BY u.id, username
		ORDER BY total_spent DESC
		LIMIT 10
	`

	spenderRows, err := s.db.Query(topSpendersQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer spenderRows.Close()

	for spenderRows.Next() {
		var spender models.TopSpender
		spenderRows.Scan(&spender.UserID, &spender.Username, &spender.TotalSpent, &spender.TransactionCount)
		report.TopSpenders = append(report.TopSpenders, spender)
	}

	// Get payment method stats
	paymentMethodQuery := `
		SELECT 
			pt.gateway,
			COUNT(*) as count,
			SUM(pt.amount) as amount,
			(COUNT(CASE WHEN pt.status = 'success' THEN 1 END) * 100.0 / COUNT(*)) as success_rate
		FROM payment_transactions pt
		WHERE pt.created_at BETWEEN $1 AND $2
		GROUP BY pt.gateway
		ORDER BY amount DESC
	`

	paymentRows, err := s.db.Query(paymentMethodQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer paymentRows.Close()

	for paymentRows.Next() {
		var method models.PaymentMethodReport
		paymentRows.Scan(&method.Method, &method.Count, &method.Amount, &method.SuccessRate)
		report.PaymentMethods = append(report.PaymentMethods, method)
	}

	// Set placeholder tax data (would be calculated based on actual tax rules)
	report.TaxSummary = models.TaxSummary{
		TotalTDS:      report.Summary.TotalRevenue * 0.1, // 10% TDS
		TotalGST:      report.Summary.TotalRevenue * 0.18, // 18% GST
		TaxableAmount: report.Summary.TotalRevenue,
		TDSDeductions: report.Summary.TransactionCount / 10, // Estimated
	}

	return report, nil
}

// generateUserReport creates user activity report
func (s *ReportingService) generateUserReport(request models.ReportRequest) (*models.UserActivityReport, error) {
	report := &models.UserActivityReport{}

	// Get user summary
	summaryQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN created_at BETWEEN $1 AND $2 THEN 1 END) as new_users,
			COUNT(CASE WHEN last_login_at >= $1 THEN 1 END) as active_users,
			COUNT(CASE WHEN is_verified = true THEN 1 END) as verified_users,
			COUNT(CASE WHEN kyc_status = 'verified' THEN 1 END) as kyc_completed_users
		FROM users
		WHERE is_active = true
	`

	err := s.db.QueryRow(summaryQuery, request.DateFrom, request.DateTo).Scan(
		&report.Summary.TotalUsers,
		&report.Summary.NewUsers,
		&report.Summary.ActiveUsers,
		&report.Summary.VerifiedUsers,
		&report.Summary.KYCCompletedUsers,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate rates
	if report.Summary.TotalUsers > 0 {
		report.Summary.RetentionRate = float64(report.Summary.ActiveUsers) / float64(report.Summary.TotalUsers) * 100
		report.Summary.ChurnRate = 100 - report.Summary.RetentionRate
	}

	// Get registration trend
	registrationTrendQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as new_registrations,
			COUNT(CASE WHEN last_login_at >= created_at + INTERVAL '7 days' THEN 1 END) as active_users
		FROM users
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	trendRows, err := s.db.Query(registrationTrendQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer trendRows.Close()

	for trendRows.Next() {
		var trend models.DailyUserStats
		trendRows.Scan(&trend.Date, &trend.NewRegistrations, &trend.ActiveUsers)
		
		if trend.NewRegistrations > 0 {
			trend.RetentionRate = float64(trend.ActiveUsers) / float64(trend.NewRegistrations) * 100
		}
		
		report.RegistrationTrend = append(report.RegistrationTrend, trend)
	}

	// Get users by region
	regionQuery := `
		SELECT 
			COALESCE(state, 'Unknown') as region,
			COUNT(*) as user_count
		FROM users
		WHERE created_at BETWEEN $1 AND $2 AND is_active = true
		GROUP BY state
		ORDER BY user_count DESC
		LIMIT 10
	`

	regionRows, err := s.db.Query(regionQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer regionRows.Close()

	totalRegionUsers := int64(0)
	for regionRows.Next() {
		var region models.RegionStats
		regionRows.Scan(&region.Region, &region.UserCount)
		totalRegionUsers += region.UserCount
		report.UsersByRegion = append(report.UsersByRegion, region)
	}

	// Calculate percentages for regions
	for i := range report.UsersByRegion {
		if totalRegionUsers > 0 {
			report.UsersByRegion[i].Percentage = float64(report.UsersByRegion[i].UserCount) / float64(totalRegionUsers) * 100
		}
	}

	// Get KYC status distribution
	kycQuery := `
		SELECT 
			kyc_status,
			COUNT(*) as count
		FROM users
		WHERE created_at BETWEEN $1 AND $2 AND is_active = true
		GROUP BY kyc_status
		ORDER BY count DESC
	`

	kycRows, err := s.db.Query(kycQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer kycRows.Close()

	totalKYCUsers := int64(0)
	for kycRows.Next() {
		var kyc models.KYCStats
		kycRows.Scan(&kyc.Status, &kyc.Count)
		totalKYCUsers += kyc.Count
		report.UsersByKYCStatus = append(report.UsersByKYCStatus, kyc)
	}

	// Calculate percentages for KYC
	for i := range report.UsersByKYCStatus {
		if totalKYCUsers > 0 {
			report.UsersByKYCStatus[i].Percentage = float64(report.UsersByKYCStatus[i].Count) / float64(totalKYCUsers) * 100
		}
	}

	return report, nil
}

// generateContestReport creates contest performance report
func (s *ReportingService) generateContestReport(request models.ReportRequest) (*models.ContestPerformanceReport, error) {
	report := &models.ContestPerformanceReport{}

	// Get contest summary
	summaryQuery := `
		SELECT 
			COUNT(*) as total_contests,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_contests,
			SUM(current_participants) as total_participants,
			SUM(total_prize_pool) as total_prize_pool,
			AVG(current_participants) as avg_participation
		FROM contests
		WHERE created_at BETWEEN $1 AND $2
	`

	err := s.db.QueryRow(summaryQuery, request.DateFrom, request.DateTo).Scan(
		&report.Summary.TotalContests,
		&report.Summary.CompletedContests,
		&report.Summary.TotalParticipants,
		&report.Summary.TotalPrizePool,
		&report.Summary.AvgParticipation,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate rates
	if report.Summary.TotalContests > 0 {
		report.Summary.CompletionRate = float64(report.Summary.CompletedContests) / float64(report.Summary.TotalContests) * 100
	}

	// Calculate fill rate (avg participants / avg max participants)
	fillRateQuery := `
		SELECT AVG(CASE WHEN max_participants > 0 THEN (current_participants::float / max_participants::float) * 100 ELSE 0 END)
		FROM contests
		WHERE created_at BETWEEN $1 AND $2
	`
	s.db.QueryRow(fillRateQuery, request.DateFrom, request.DateTo).Scan(&report.Summary.FillRate)

	// Get contests by type
	contestTypeQuery := `
		SELECT 
			contest_type,
			COUNT(*) as count,
			SUM(current_participants) as total_participants,
			AVG(entry_fee) as avg_entry_fee,
			SUM(total_prize_pool) as total_prize_pool
		FROM contests
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY contest_type
		ORDER BY count DESC
	`

	typeRows, err := s.db.Query(contestTypeQuery, request.DateFrom, request.DateTo)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var contestType models.ContestTypeStats
		typeRows.Scan(&contestType.ContestType, &contestType.Count, &contestType.TotalParticipants, &contestType.AvgEntryFee, &contestType.TotalPrizePool)
		report.ContestsByType = append(report.ContestsByType, contestType)
	}

	return report, nil
}

// generateGameReport creates game performance report  
func (s *ReportingService) generateGameReport(request models.ReportRequest) (*models.GamePerformanceReport, error) {
	report := &models.GamePerformanceReport{}

	// Get game summary
	summaryQuery := `
		SELECT 
			COUNT(DISTINCT g.id) as total_games,
			COUNT(CASE WHEN g.is_active THEN 1 END) as active_games,
			COUNT(DISTINCT m.id) as total_matches,
			COUNT(CASE WHEN m.status = 'completed' THEN 1 END) as completed_matches,
			COUNT(DISTINCT p.id) as total_players,
			COUNT(CASE WHEN p.is_playing THEN 1 END) as active_players
		FROM games g
		LEFT JOIN matches m ON m.game_id = g.id AND m.created_at BETWEEN $1 AND $2
		LEFT JOIN players p ON p.game_id = g.id
	`

	err := s.db.QueryRow(summaryQuery, request.DateFrom, request.DateTo).Scan(
		&report.Summary.TotalGames,
		&report.Summary.ActiveGames,
		&report.Summary.TotalMatches,
		&report.Summary.CompletedMatches,
		&report.Summary.TotalPlayers,
		&report.Summary.ActivePlayers,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	return report, nil
}

// generateComplianceReport creates compliance report
func (s *ReportingService) generateComplianceReport(request models.ReportRequest) (*models.ComplianceReport, error) {
	report := &models.ComplianceReport{}

	// Get KYC compliance stats
	kycQuery := `
		SELECT 
			COUNT(DISTINCT u.id) as total_users,
			COUNT(CASE WHEN u.kyc_status = 'verified' THEN 1 END) as kyc_verified_users,
			COUNT(CASE WHEN u.kyc_status = 'pending' THEN 1 END) as pending_verification,
			COUNT(CASE WHEN kd.status = 'rejected' THEN 1 END) as rejected_documents
		FROM users u
		LEFT JOIN kyc_documents kd ON kd.user_id = u.id
		WHERE u.created_at BETWEEN $1 AND $2
	`

	err := s.db.QueryRow(kycQuery, request.DateFrom, request.DateTo).Scan(
		&report.KYCCompliance.TotalUsers,
		&report.KYCCompliance.KYCVerifiedUsers,
		&report.KYCCompliance.PendingVerification,
		&report.KYCCompliance.RejectedDocuments,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate compliance rate
	if report.KYCCompliance.TotalUsers > 0 {
		report.KYCCompliance.ComplianceRate = float64(report.KYCCompliance.KYCVerifiedUsers) / float64(report.KYCCompliance.TotalUsers) * 100
	}

	// Get transaction compliance
	transactionComplianceQuery := `
		SELECT 
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN amount > 10000 THEN 1 END) as large_transactions,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_transactions
		FROM wallet_transactions
		WHERE created_at BETWEEN $1 AND $2
	`

	err = s.db.QueryRow(transactionComplianceQuery, request.DateFrom, request.DateTo).Scan(
		&report.TransactionCompliance.TotalTransactions,
		&report.TransactionCompliance.LargeTransactions,
		&report.TransactionCompliance.FailedTransactions,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate transaction compliance rate
	if report.TransactionCompliance.TotalTransactions > 0 {
		suspiciousCount := report.TransactionCompliance.LargeTransactions + report.TransactionCompliance.FailedTransactions
		report.TransactionCompliance.FlaggedTransactions = suspiciousCount
		report.TransactionCompliance.ComplianceRate = float64(report.TransactionCompliance.TotalTransactions - suspiciousCount) / float64(report.TransactionCompliance.TotalTransactions) * 100
	}

	return report, nil
}

// generateReferralReport creates referral system report
func (s *ReportingService) generateReferralReport(request models.ReportRequest) (*models.ReferralReport, error) {
	report := &models.ReferralReport{}

	// Get referral summary
	summaryQuery := `
		SELECT 
			COUNT(*) as total_referrals,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_referrals,
			SUM(reward_amount) as total_earnings,
			COUNT(DISTINCT referrer_user_id) as active_referrers
		FROM referrals
		WHERE created_at BETWEEN $1 AND $2
	`

	err := s.db.QueryRow(summaryQuery, request.DateFrom, request.DateTo).Scan(
		&report.Summary.TotalReferrals,
		&report.Summary.SuccessfulReferrals,
		&report.Summary.TotalEarnings,
		&report.Summary.ActiveReferrers,
	)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Calculate conversion rate and avg earnings
	if report.Summary.TotalReferrals > 0 {
		report.Summary.ConversionRate = float64(report.Summary.SuccessfulReferrals) / float64(report.Summary.TotalReferrals) * 100
	}
	if report.Summary.SuccessfulReferrals > 0 {
		report.Summary.AvgEarningsPerReferral = report.Summary.TotalEarnings / float64(report.Summary.SuccessfulReferrals)
	}

	return report, nil
}

// Helper functions

func (s *ReportingService) validateReportRequest(request models.ReportRequest) error {
	if request.DateFrom.After(request.DateTo) {
		return errors.NewError(errors.ErrInvalidRequest, "Date from cannot be after date to")
	}

	if request.DateTo.Sub(request.DateFrom) > 365*24*time.Hour {
		return errors.NewError(errors.ErrInvalidRequest, "Date range cannot exceed 1 year")
	}

	supportedTypes := []models.ReportType{
		models.ReportTypeFinancial,
		models.ReportTypeUser,
		models.ReportTypeContest,
		models.ReportTypeGame,
		models.ReportTypeCompliance,
		models.ReportTypeReferral,
	}

	validType := false
	for _, t := range supportedTypes {
		if request.ReportType == t {
			validType = true
			break
		}
	}
	if !validType {
		return errors.NewError(errors.ErrInvalidRequest, "Unsupported report type")
	}

	return nil
}

func (s *ReportingService) generateReportTitle(request models.ReportRequest) string {
	typeNames := map[models.ReportType]string{
		models.ReportTypeFinancial:   "Financial Report",
		models.ReportTypeUser:       "User Activity Report",
		models.ReportTypeContest:    "Contest Performance Report",
		models.ReportTypeGame:       "Game Analytics Report",
		models.ReportTypeCompliance: "Compliance Report",
		models.ReportTypeReferral:   "Referral System Report",
	}

	baseName := typeNames[request.ReportType]
	dateRange := fmt.Sprintf("%s to %s", 
		request.DateFrom.Format("2006-01-02"), 
		request.DateTo.Format("2006-01-02"))

	return fmt.Sprintf("%s (%s)", baseName, dateRange)
}

func (s *ReportingService) updateReportStatus(reportID int64, status models.ReportStatus, errorMessage *string) {
	query := `UPDATE generated_reports SET status = $1, error_message = $2, updated_at = $3 WHERE id = $4`
	_, err := s.db.Exec(query, status, errorMessage, time.Now(), reportID)
	if err != nil {
		logger.Error("Failed to update report status", map[string]interface{}{
			"report_id": reportID,
			"status":    status,
			"error":     err.Error(),
		})
	}
}

// GetReport retrieves a specific report
func (s *ReportingService) GetReport(reportID int64) (*models.GeneratedReport, error) {
	query := `
		SELECT id, report_type, format, status, title, description, file_path, file_size,
		       generated_by, request_data, result_data, error_message, created_at, completed_at, expires_at
		FROM generated_reports
		WHERE id = $1
	`

	var report models.GeneratedReport
	var requestDataJSON, resultDataJSON []byte

	err := s.db.QueryRow(query, reportID).Scan(
		&report.ID, &report.ReportType, &report.Format, &report.Status, &report.Title,
		&report.Description, &report.FilePath, &report.FileSize, &report.GeneratedBy,
		&requestDataJSON, &resultDataJSON, &report.ErrorMessage,
		&report.CreatedAt, &report.CompletedAt, &report.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewError(errors.ErrResourceNotFound, "Report not found")
		}
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Unmarshal JSON data
	if requestDataJSON != nil {
		json.Unmarshal(requestDataJSON, &report.RequestData)
	}
	if resultDataJSON != nil {
		json.Unmarshal(resultDataJSON, &report.ResultData)
	}

	return &report, nil
}

// GetReports retrieves list of reports with pagination
func (s *ReportingService) GetReports(userID int64, page, limit int, reportType *models.ReportType) (*models.ReportListResponse, error) {
	offset := (page - 1) * limit

	// Build query conditions
	whereConditions := []string{"generated_by = $1"}
	args := []interface{}{userID}
	argIndex := 2

	if reportType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("report_type = $%d", argIndex))
		args = append(args, *reportType)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM generated_reports WHERE %s", whereClause)
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	// Get reports
	query := fmt.Sprintf(`
		SELECT id, report_type, format, status, title, created_at, completed_at, file_size
		FROM generated_reports
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}
	defer rows.Close()

	var reports []models.ReportSummary
	for rows.Next() {
		var report models.ReportSummary
		rows.Scan(&report.ID, &report.ReportType, &report.Format, &report.Status,
			&report.Title, &report.CreatedAt, &report.CompletedAt, &report.FileSize)
		reports = append(reports, report)
	}

	pages := int((total + int64(limit) - 1) / int64(limit))

	return &models.ReportListResponse{
		Reports: reports,
		Total:   total,
		Page:    page,
		Pages:   pages,
	}, nil
}

// DeleteReport removes a report
func (s *ReportingService) DeleteReport(reportID, userID int64) error {
	query := `DELETE FROM generated_reports WHERE id = $1 AND generated_by = $2`
	result, err := s.db.Exec(query, reportID, userID)
	if err != nil {
		return errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewError(errors.ErrDatabaseConnection, err.Error())
	}

	if rowsAffected == 0 {
		return errors.NewError(errors.ErrResourceNotFound, "Report not found or access denied")
	}

	return nil
}