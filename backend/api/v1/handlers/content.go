package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/pkg/cdn"
	"fantasy-esports-backend/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContentHandler struct {
	contentService *services.ContentService
	db             *sql.DB
	config         *config.Config
	cdnClient      *cdn.CloudinaryClient
}

func NewContentHandler(db *sql.DB, cfg *config.Config, cdnClient cdn.Client) *ContentHandler {
	return &ContentHandler{
		contentService: services.NewContentService(db),
		db:             db,
		config:         cfg,
		cdnClient:      cdnClient,
	}
}

// ========================= BANNER MANAGEMENT =========================

// CreateBanner godoc
// @Summary Create a new banner
// @Description Create a new promotional banner
// @Tags Content Management - Banners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param banner body models.BannerCreateRequest true "Banner data"
// @Success 201 {object} models.Banner
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners [post]
func (h *ContentHandler) CreateBanner(c *gin.Context) {
	var req models.BannerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "End date must be after start date",
		})
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Admin ID not found in context",
		})
		return
	}

	banner, err := h.contentService.CreateBanner(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create banner",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Banner created successfully",
		"data":    banner,
	})
}

// UpdateBanner godoc
// @Summary Update a banner
// @Description Update an existing promotional banner
// @Tags Content Management - Banners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Param banner body models.BannerCreateRequest true "Banner data"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners/{id} [put]
func (h *ContentHandler) UpdateBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid banner ID",
		})
		return
	}

	var req models.BannerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateBanner(id, &req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update banner",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Banner updated successfully",
	})
}

// GetBanner godoc
// @Summary Get banner details
// @Description Get details of a specific banner
// @Tags Content Management - Banners
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Success 200 {object} models.BannerResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners/{id} [get]
func (h *ContentHandler) GetBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid banner ID",
		})
		return
	}

	banner, err := h.contentService.GetBanner(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Banner not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get banner",
				"error":   err.Error(),
			})
		}
		return
	}

	// Record view analytics
	go h.contentService.RecordContentView("banner", id)
	go h.contentService.IncrementBannerView(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    banner,
	})
}

// ListBanners godoc
// @Summary List all banners
// @Description Get a paginated list of banners with filtering options
// @Tags Content Management - Banners
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param position query string false "Banner position" Enums(top, middle, bottom, sidebar)
// @Param type query string false "Banner type" Enums(promotion, announcement, sponsored)
// @Param status query string false "Banner status" Enums(active, inactive)
// @Success 200 {object} models.BannerListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners [get]
func (h *ContentHandler) ListBanners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	position := c.Query("position")
	bannerType := c.Query("type")
	status := c.Query("status")

	banners, err := h.contentService.ListBanners(page, limit, position, bannerType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list banners",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, banners)
}

// DeleteBanner godoc
// @Summary Delete a banner
// @Description Delete a promotional banner
// @Tags Content Management - Banners
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners/{id} [delete]
func (h *ContentHandler) DeleteBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid banner ID",
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.DeleteBanner(id, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete banner",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Banner deleted successfully",
	})
}

// ToggleBannerStatus godoc
// @Summary Toggle banner status
// @Description Toggle active/inactive status of a banner
// @Tags Content Management - Banners
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/banners/{id}/toggle [patch]
func (h *ContentHandler) ToggleBannerStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid banner ID",
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.ToggleBannerStatus(id, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to toggle banner status",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Banner status toggled successfully",
	})
}

// TrackBannerClick godoc
// @Summary Track banner click
// @Description Record a banner click for analytics
// @Tags Content Management - Banners
// @Produce json
// @Param id path int true "Banner ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /banners/{id}/click [post]
func (h *ContentHandler) TrackBannerClick(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid banner ID",
		})
		return
	}

	// Record click analytics (asynchronous)
	go h.contentService.RecordContentClick("banner", id)
	go h.contentService.IncrementBannerClick(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Click recorded",
	})
}

// GetActiveBanners godoc
// @Summary Get active banners
// @Description Get currently active banners for display (public endpoint)
// @Tags Content Management - Banners
// @Produce json
// @Param position query string false "Banner position" Enums(top, middle, bottom, sidebar)
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /banners/active [get]
func (h *ContentHandler) GetActiveBanners(c *gin.Context) {
	position := c.Query("position")

	banners, err := h.contentService.ListBanners(1, 50, position, "", "active")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get active banners",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    banners.Banners,
		"count":   len(banners.Banners),
	})
}

// ========================= EMAIL TEMPLATE MANAGEMENT =========================

// CreateEmailTemplate godoc
// @Summary Create email template
// @Description Create a new email template for marketing campaigns
// @Tags Content Management - Email Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} models.EmailTemplate
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/email-templates [post]
func (h *ContentHandler) CreateEmailTemplate(c *gin.Context) {
	var req struct {
		Name        string          `json:"name" binding:"required,max=200"`
		Description string          `json:"description" binding:"max=500"`
		Subject     string          `json:"subject" binding:"required,max=200"`
		HTMLContent string          `json:"html_content" binding:"required"`
		TextContent string          `json:"text_content"`
		Category    string          `json:"category" binding:"required,oneof=welcome promotional transactional newsletter"`
		Variables   models.JSONMap  `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	template, err := h.contentService.CreateEmailTemplate(
		req.Name, req.Description, req.Subject, req.HTMLContent,
		req.TextContent, req.Category, req.Variables, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create email template",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Email template created successfully",
		"data":    template,
	})
}

// ListEmailTemplates godoc
// @Summary List email templates
// @Description Get a paginated list of email templates
// @Tags Content Management - Email Templates
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Template category"
// @Param active query bool false "Filter by active status"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/email-templates [get]
func (h *ContentHandler) ListEmailTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	category := c.Query("category")
	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	templates, total, err := h.contentService.ListEmailTemplates(page, limit, category, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list email templates",
			"error":   err.Error(),
		})
		return
	}

	pages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"templates": templates,
		"total":     total,
		"page":      page,
		"pages":     pages,
	})
}

// ========================= MARKETING CAMPAIGN MANAGEMENT =========================

// CreateMarketingCampaign godoc
// @Summary Create marketing campaign
// @Description Create a new email marketing campaign
// @Tags Content Management - Marketing Campaigns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param campaign body models.MarketingCampaignCreateRequest true "Campaign data"
// @Success 201 {object} models.MarketingCampaign
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/campaigns [post]
func (h *ContentHandler) CreateMarketingCampaign(c *gin.Context) {
	var req models.MarketingCampaignCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	campaign, err := h.contentService.CreateMarketingCampaign(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create marketing campaign",
			"error":   err.Error(),
		})
		return
	}

	// Calculate estimated recipients
	if recipients, err := h.contentService.CalculateCampaignRecipients(req.TargetSegment, req.TargetCriteria); err == nil {
		// Update campaign with recipient count (you might want to add this to the service)
		c.JSON(http.StatusCreated, gin.H{
			"success":             true,
			"message":             "Marketing campaign created successfully",
			"data":                campaign,
			"estimated_recipients": recipients,
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Marketing campaign created successfully",
			"data":    campaign,
		})
	}
}

// ListMarketingCampaigns godoc
// @Summary List marketing campaigns
// @Description Get a paginated list of marketing campaigns
// @Tags Content Management - Marketing Campaigns
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Campaign status"
// @Param segment query string false "Target segment"
// @Success 200 {object} models.CampaignListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/campaigns [get]
func (h *ContentHandler) ListMarketingCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	status := c.Query("status")
	segment := c.Query("segment")

	campaigns, err := h.contentService.ListMarketingCampaigns(page, limit, status, segment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list marketing campaigns",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, campaigns)
}

// UpdateCampaignStatus godoc
// @Summary Update campaign status
// @Description Update the status of a marketing campaign
// @Tags Content Management - Marketing Campaigns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Campaign ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/campaigns/{id}/status [patch]
func (h *ContentHandler) UpdateCampaignStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid campaign ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=draft scheduled sending sent cancelled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateCampaignStatus(id, req.Status, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update campaign status",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Campaign status updated successfully",
	})
}