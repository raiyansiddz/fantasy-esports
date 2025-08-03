package handlers

import (
	"net/http"
	"strconv"
	"time"

	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/errors"
	"fantasy-esports-backend/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	biService       *services.BusinessIntelligenceService
	reportingService *services.ReportingService
}

func NewAnalyticsHandler(analyticsService *services.AnalyticsService, biService *services.BusinessIntelligenceService, reportingService *services.ReportingService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		biService:       biService,
		reportingService: reportingService,
	}
}

// GetAnalyticsDashboard returns comprehensive analytics dashboard
// @Summary Get analytics dashboard
// @Description Get comprehensive analytics dashboard with all metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param game_id query int false "Filter by game ID"
// @Param period query string false "Period (day/week/month/year)"
// @Security BearerAuth
// @Success 200 {object} models.AnalyticsDashboard
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/dashboard [get]
func (h *AnalyticsHandler) GetAnalyticsDashboard(c *gin.Context) {
	filters := h.parseAnalyticsFilters(c)

	dashboard, err := h.analyticsService.GetDashboard(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetAnalyticsDashboard",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// GetUserMetrics returns detailed user analytics
// @Summary Get user metrics
// @Description Get detailed user analytics and metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.UserMetrics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/users [get]
func (h *AnalyticsHandler) GetUserMetrics(c *gin.Context) {
	filters := h.parseAnalyticsFilters(c)

	metrics, err := h.analyticsService.GetUserMetrics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetUserMetrics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetRevenueMetrics returns detailed revenue analytics
// @Summary Get revenue metrics
// @Description Get detailed revenue analytics and financial metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.RevenueMetrics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/revenue [get]
func (h *AnalyticsHandler) GetRevenueMetrics(c *gin.Context) {
	filters := h.parseAnalyticsFilters(c)

	metrics, err := h.analyticsService.GetRevenueMetrics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetRevenueMetrics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetContestMetrics returns contest performance analytics
// @Summary Get contest metrics
// @Description Get contest performance analytics and metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.ContestMetrics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/contests [get]
func (h *AnalyticsHandler) GetContestMetrics(c *gin.Context) {
	filters := h.parseAnalyticsFilters(c)

	metrics, err := h.analyticsService.GetContestMetrics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetContestMetrics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetGameMetrics returns game performance analytics
// @Summary Get game metrics
// @Description Get game performance analytics and metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.GameMetrics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/games [get]
func (h *AnalyticsHandler) GetGameMetrics(c *gin.Context) {
	filters := h.parseAnalyticsFilters(c)

	metrics, err := h.analyticsService.GetGameMetrics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetGameMetrics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetRealTimeMetrics returns current real-time metrics
// @Summary Get real-time metrics
// @Description Get current real-time system metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.RealTimeMetrics
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/realtime [get]
func (h *AnalyticsHandler) GetRealTimeMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetRealTimeMetrics()
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetRealTimeMetrics",
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetPerformanceMetrics returns system performance metrics
// @Summary Get performance metrics
// @Description Get system performance and API metrics
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.PerformanceMetrics
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/analytics/performance [get]
func (h *AnalyticsHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetPerformanceMetrics()
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetPerformanceMetrics",
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetBIDashboard returns business intelligence dashboard
// @Summary Get business intelligence dashboard
// @Description Get comprehensive business intelligence dashboard with KPIs and insights
// @Tags Business Intelligence
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param user_segment query string false "Filter by user segment"
// @Param confidence_level query number false "Minimum confidence level"
// @Security BearerAuth
// @Success 200 {object} models.BusinessIntelligenceDashboard
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/bi/dashboard [get]
func (h *AnalyticsHandler) GetBIDashboard(c *gin.Context) {
	filters := h.parseBIFilters(c)

	dashboard, err := h.biService.GetBIDashboard(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetBIDashboard",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// GetKPIMetrics returns key performance indicators
// @Summary Get KPI metrics
// @Description Get key performance indicators and business metrics
// @Tags Business Intelligence
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.KPIMetrics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/bi/kpis [get]
func (h *AnalyticsHandler) GetKPIMetrics(c *gin.Context) {
	filters := h.parseBIFilters(c)

	metrics, err := h.biService.GetKPIMetrics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetKPIMetrics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetRevenueAnalytics returns advanced revenue analytics
// @Summary Get revenue analytics
// @Description Get advanced revenue analytics with forecasting and segmentation
// @Tags Business Intelligence
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.RevenueAnalytics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/bi/revenue [get]
func (h *AnalyticsHandler) GetRevenueAnalytics(c *gin.Context) {
	filters := h.parseBIFilters(c)

	analytics, err := h.biService.GetRevenueAnalytics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetRevenueAnalytics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetUserBehaviorAnalysis returns user behavior analysis
// @Summary Get user behavior analysis
// @Description Get user behavior patterns, segmentation and engagement analysis
// @Tags Business Intelligence
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.UserBehaviorAnalysis
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/bi/user-behavior [get]
func (h *AnalyticsHandler) GetUserBehaviorAnalysis(c *gin.Context) {
	filters := h.parseBIFilters(c)

	analysis, err := h.biService.GetUserBehaviorAnalysis(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetUserBehaviorAnalysis",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// GetPredictiveAnalytics returns predictive analytics
// @Summary Get predictive analytics
// @Description Get predictive analytics including churn prediction and forecasting
// @Tags Business Intelligence
// @Accept json
// @Produce json
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.PredictiveAnalytics
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/bi/predictive [get]
func (h *AnalyticsHandler) GetPredictiveAnalytics(c *gin.Context) {
	filters := h.parseBIFilters(c)

	analytics, err := h.biService.GetPredictiveAnalytics(filters)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetPredictiveAnalytics",
			"filters": filters,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GenerateReport creates a new report
// @Summary Generate report
// @Description Generate a new report with specified parameters
// @Tags Reporting
// @Accept json
// @Produce json
// @Param request body models.ReportRequest true "Report request parameters"
// @Security BearerAuth
// @Success 200 {object} models.GeneratedReport
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/reports/generate [post]
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	var request models.ReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		appErr := errors.ValidationError(map[string]string{
			"request": "Invalid request format",
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		appErr := errors.NewError(errors.ErrUnauthorized, "Admin ID not found in context")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	report, err := h.reportingService.GenerateReport(request, adminID.(int64))
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GenerateReport",
			"request": request,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetReports returns list of generated reports
// @Summary Get reports
// @Description Get list of generated reports with pagination
// @Tags Reporting
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param report_type query string false "Filter by report type"
// @Security BearerAuth
// @Success 200 {object} models.ReportListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/reports [get]
func (h *AnalyticsHandler) GetReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	var reportType *models.ReportType
	if rt := c.Query("report_type"); rt != "" {
		reportTypeValue := models.ReportType(rt)
		reportType = &reportTypeValue
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		appErr := errors.NewError(errors.ErrUnauthorized, "Admin ID not found in context")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	reports, err := h.reportingService.GetReports(adminID.(int64), page, limit, reportType)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler": "GetReports",
			"page":    page,
			"limit":   limit,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reports,
	})
}

// GetReport returns a specific report
// @Summary Get report
// @Description Get a specific report by ID
// @Tags Reporting
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Security BearerAuth
// @Success 200 {object} models.GeneratedReport
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/reports/{id} [get]
func (h *AnalyticsHandler) GetReport(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		appErr := errors.NewError(errors.ErrInvalidRequest, "Invalid report ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	report, err := h.reportingService.GetReport(reportID)
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler":   "GetReport",
			"report_id": reportID,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// DeleteReport deletes a specific report
// @Summary Delete report
// @Description Delete a specific report by ID
// @Tags Reporting
// @Accept json
// @Produce json
// @Param id path int true "Report ID"
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/reports/{id} [delete]
func (h *AnalyticsHandler) DeleteReport(c *gin.Context) {
	reportID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		appErr := errors.NewError(errors.ErrInvalidRequest, "Invalid report ID")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		appErr := errors.NewError(errors.ErrUnauthorized, "Admin ID not found in context")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	err = h.reportingService.DeleteReport(reportID, adminID.(int64))
	if err != nil {
		appErr := err.(*errors.AppError)
		appErr.LogError(map[string]interface{}{
			"handler":   "DeleteReport",
			"report_id": reportID,
		})
		c.JSON(appErr.HTTPStatus, appErr.ToResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Report deleted successfully",
	})
}

// Helper functions

func (h *AnalyticsHandler) parseAnalyticsFilters(c *gin.Context) models.AnalyticsFilters {
	filters := models.AnalyticsFilters{}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dt, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters.DateFrom = &dt
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if dt, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters.DateTo = &dt
		}
	}

	if gameID := c.Query("game_id"); gameID != "" {
		if id, err := strconv.Atoi(gameID); err == nil {
			filters.GameID = &id
		}
	}

	filters.Period = c.DefaultQuery("period", "month")
	filters.Granularity = c.DefaultQuery("granularity", "day")

	return filters
}

func (h *AnalyticsHandler) parseBIFilters(c *gin.Context) models.BIFilters {
	filters := models.BIFilters{}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dt, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filters.DateFrom = &dt
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if dt, err := time.Parse("2006-01-02", dateTo); err == nil {
			filters.DateTo = &dt
		}
	}

	if userSegment := c.Query("user_segment"); userSegment != "" {
		filters.UserSegment = &userSegment
	}

	if gameID := c.Query("game_id"); gameID != "" {
		if id, err := strconv.Atoi(gameID); err == nil {
			filters.GameID = &id
		}
	}

	if revenueThreshold := c.Query("revenue_threshold"); revenueThreshold != "" {
		if threshold, err := strconv.ParseFloat(revenueThreshold, 64); err == nil {
			filters.RevenueThreshold = &threshold
		}
	}

	if confidenceLevel := c.Query("confidence_level"); confidenceLevel != "" {
		if level, err := strconv.ParseFloat(confidenceLevel, 64); err == nil {
			filters.ConfidenceLevel = &level
		}
	}

	filters.IncludeForecasts = c.DefaultQuery("include_forecasts", "true") == "true"
	filters.IncludePredictions = c.DefaultQuery("include_predictions", "true") == "true"

	return filters
}