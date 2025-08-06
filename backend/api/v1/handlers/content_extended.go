package handlers

import (
	"fantasy-esports-backend/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ========================= SEO CONTENT MANAGEMENT =========================

// CreateSEOContent godoc
// @Summary Create SEO content
// @Description Create SEO content for a specific page
// @Tags Content Management - SEO
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param seo body models.SEOContentCreateRequest true "SEO content data"
// @Success 201 {object} models.SEOContent
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/seo [post]
func (h *ContentHandler) CreateSEOContent(c *gin.Context) {
	var req models.SEOContentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	seoContent, err := h.contentService.CreateSEOContent(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create SEO content",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "SEO content created successfully",
		"data":    seoContent,
	})
}

// UpdateSEOContent godoc
// @Summary Update SEO content
// @Description Update existing SEO content
// @Tags Content Management - SEO
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "SEO Content ID"
// @Param seo body models.SEOContentCreateRequest true "SEO content data"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/seo/{id} [put]
func (h *ContentHandler) UpdateSEOContent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid SEO content ID",
		})
		return
	}

	var req models.SEOContentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateSEOContent(id, &req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update SEO content",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "SEO content updated successfully",
	})
}

// GetSEOContent godoc
// @Summary Get SEO content
// @Description Get SEO content by ID
// @Tags Content Management - SEO
// @Produce json
// @Security BearerAuth
// @Param id path int true "SEO Content ID"
// @Success 200 {object} models.SEOContentResponse
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/seo/{id} [get]
func (h *ContentHandler) GetSEOContent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid SEO content ID",
		})
		return
	}

	seoContent, err := h.contentService.GetSEOContent(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "SEO content not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get SEO content",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    seoContent,
	})
}

// GetSEOContentBySlug godoc
// @Summary Get SEO content by slug
// @Description Get SEO content for a specific page by its slug (public endpoint)
// @Tags Content Management - SEO
// @Produce json
// @Param slug path string true "Page slug"
// @Success 200 {object} models.SEOContent
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /seo/{slug} [get]
func (h *ContentHandler) GetSEOContentBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Page slug is required",
		})
		return
	}

	seoContent, err := h.contentService.GetSEOContentBySlug(slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "SEO content not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get SEO content",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    seoContent,
	})
}

// ListSEOContent godoc
// @Summary List SEO content
// @Description Get a paginated list of SEO content
// @Tags Content Management - SEO
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param page_type query string false "Page type filter"
// @Param active query bool false "Filter by active status"
// @Success 200 {object} models.SEOContentListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/seo [get]
func (h *ContentHandler) ListSEOContent(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	pageType := c.Query("page_type")
	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	seoContents, err := h.contentService.ListSEOContent(page, limit, pageType, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list SEO content",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, seoContents)
}

// DeleteSEOContent godoc
// @Summary Delete SEO content
// @Description Delete SEO content
// @Tags Content Management - SEO
// @Produce json
// @Security BearerAuth
// @Param id path int true "SEO Content ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/seo/{id} [delete]
func (h *ContentHandler) DeleteSEOContent(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid SEO content ID",
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.DeleteSEOContent(id, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete SEO content",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "SEO content deleted successfully",
	})
}

// ========================= FAQ MANAGEMENT =========================

// CreateFAQSection godoc
// @Summary Create FAQ section
// @Description Create a new FAQ section/category
// @Tags Content Management - FAQ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param section body models.FAQSectionCreateRequest true "FAQ section data"
// @Success 201 {object} models.FAQSection
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/faq/sections [post]
func (h *ContentHandler) CreateFAQSection(c *gin.Context) {
	var req models.FAQSectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	section, err := h.contentService.CreateFAQSection(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create FAQ section",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "FAQ section created successfully",
		"data":    section,
	})
}

// UpdateFAQSection godoc
// @Summary Update FAQ section
// @Description Update an existing FAQ section
// @Tags Content Management - FAQ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "FAQ Section ID"
// @Param section body models.FAQSectionCreateRequest true "FAQ section data"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/faq/sections/{id} [put]
func (h *ContentHandler) UpdateFAQSection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid FAQ section ID",
		})
		return
	}

	var req models.FAQSectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateFAQSection(id, &req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update FAQ section",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FAQ section updated successfully",
	})
}

// ListFAQSections godoc
// @Summary List FAQ sections
// @Description Get a paginated list of FAQ sections
// @Tags Content Management - FAQ
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param active query bool false "Filter by active status"
// @Success 200 {object} models.FAQSectionListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /faq/sections [get]
func (h *ContentHandler) ListFAQSections(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	sections, err := h.contentService.ListFAQSections(page, limit, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list FAQ sections",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sections)
}

// CreateFAQItem godoc
// @Summary Create FAQ item
// @Description Create a new FAQ item/question
// @Tags Content Management - FAQ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item body models.FAQItemCreateRequest true "FAQ item data"
// @Success 201 {object} models.FAQItem
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/faq/items [post]
func (h *ContentHandler) CreateFAQItem(c *gin.Context) {
	var req models.FAQItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	item, err := h.contentService.CreateFAQItem(&req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "FAQ section not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create FAQ item",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "FAQ item created successfully",
		"data":    item,
	})
}

// UpdateFAQItem godoc
// @Summary Update FAQ item
// @Description Update an existing FAQ item
// @Tags Content Management - FAQ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "FAQ Item ID"
// @Param item body models.FAQItemCreateRequest true "FAQ item data"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/faq/items/{id} [put]
func (h *ContentHandler) UpdateFAQItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid FAQ item ID",
		})
		return
	}

	var req models.FAQItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateFAQItem(id, &req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update FAQ item",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "FAQ item updated successfully",
	})
}

// ListFAQItems godoc
// @Summary List FAQ items
// @Description Get a paginated list of FAQ items
// @Tags Content Management - FAQ
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param section_id query int false "Filter by FAQ section ID"
// @Param active query bool false "Filter by active status"
// @Success 200 {object} models.FAQItemListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /faq/items [get]
func (h *ContentHandler) ListFAQItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var sectionID *int64
	if c.Query("section_id") != "" {
		if id, err := strconv.ParseInt(c.Query("section_id"), 10, 64); err == nil {
			sectionID = &id
		}
	}

	var active *bool
	if c.Query("active") != "" {
		activeVal := c.Query("active") == "true"
		active = &activeVal
	}

	items, err := h.contentService.ListFAQItems(page, limit, sectionID, active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list FAQ items",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, items)
}

// TrackFAQView godoc
// @Summary Track FAQ view
// @Description Record a FAQ item view for analytics
// @Tags Content Management - FAQ
// @Produce json
// @Param id path int true "FAQ Item ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /faq/items/{id}/view [post]
func (h *ContentHandler) TrackFAQView(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid FAQ item ID",
		})
		return
	}

	// Record view analytics (asynchronous)
	go h.contentService.RecordContentView("faq", id)
	go h.contentService.IncrementFAQView(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "View recorded",
	})
}

// TrackFAQLike godoc
// @Summary Track FAQ like
// @Description Record a FAQ item like
// @Tags Content Management - FAQ
// @Produce json
// @Param id path int true "FAQ Item ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /faq/items/{id}/like [post]
func (h *ContentHandler) TrackFAQLike(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid FAQ item ID",
		})
		return
	}

	// Record like (asynchronous)
	go h.contentService.IncrementFAQLike(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Like recorded",
	})
}

// ========================= LEGAL DOCUMENT MANAGEMENT =========================

// CreateLegalDocument godoc
// @Summary Create legal document
// @Description Create a new legal document
// @Tags Content Management - Legal Documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param document body models.LegalDocumentCreateRequest true "Legal document data"
// @Success 201 {object} models.LegalDocument
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/legal [post]
func (h *ContentHandler) CreateLegalDocument(c *gin.Context) {
	var req models.LegalDocumentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	document, err := h.contentService.CreateLegalDocument(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create legal document",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Legal document created successfully",
		"data":    document,
	})
}

// UpdateLegalDocument godoc
// @Summary Update legal document
// @Description Update an existing legal document
// @Tags Content Management - Legal Documents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Legal Document ID"
// @Param document body models.LegalDocumentCreateRequest true "Legal document data"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/legal/{id} [put]
func (h *ContentHandler) UpdateLegalDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid legal document ID",
		})
		return
	}

	var req models.LegalDocumentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.UpdateLegalDocument(id, &req, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update legal document",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Legal document updated successfully",
	})
}

// PublishLegalDocument godoc
// @Summary Publish legal document
// @Description Publish a legal document (make it active)
// @Tags Content Management - Legal Documents
// @Produce json
// @Security BearerAuth
// @Param id path int true "Legal Document ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/legal/{id}/publish [patch]
func (h *ContentHandler) PublishLegalDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid legal document ID",
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.PublishLegalDocument(id, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to publish legal document",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Legal document published successfully",
	})
}

// GetActiveLegalDocument godoc
// @Summary Get active legal document
// @Description Get the currently active legal document by type (public endpoint)
// @Tags Content Management - Legal Documents
// @Produce json
// @Param type path string true "Document type" Enums(terms, privacy, refund, cookie, disclaimer)
// @Success 200 {object} models.LegalDocument
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /legal/{type} [get]
func (h *ContentHandler) GetActiveLegalDocument(c *gin.Context) {
	docType := c.Param("type")
	
	validTypes := map[string]bool{
		"terms": true, "privacy": true, "refund": true, "cookie": true, "disclaimer": true,
	}
	
	if !validTypes[docType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid document type",
		})
		return
	}

	document, err := h.contentService.GetActiveLegalDocument(docType)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Legal document not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to get legal document",
				"error":   err.Error(),
			})
		}
		return
	}

	// Record view analytics
	go h.contentService.RecordContentView("legal", document.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    document,
	})
}

// ListLegalDocuments godoc
// @Summary List legal documents
// @Description Get a paginated list of legal documents
// @Tags Content Management - Legal Documents
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Document type filter"
// @Param status query string false "Document status filter"
// @Success 200 {object} models.LegalDocumentListResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/legal [get]
func (h *ContentHandler) ListLegalDocuments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	docType := c.Query("type")
	status := c.Query("status")

	documents, err := h.contentService.ListLegalDocuments(page, limit, docType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list legal documents",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// DeleteLegalDocument godoc
// @Summary Delete legal document
// @Description Delete a legal document (only drafts can be deleted)
// @Tags Content Management - Legal Documents
// @Produce json
// @Security BearerAuth
// @Param id path int true "Legal Document ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/legal/{id} [delete]
func (h *ContentHandler) DeleteLegalDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid legal document ID",
		})
		return
	}

	adminID, _ := c.Get("admin_id")
	err = h.contentService.DeleteLegalDocument(id, adminID.(int64))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "cannot delete") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete legal document",
				"error":   err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Legal document deleted successfully",
	})
}

// ========================= CONTENT ANALYTICS =========================

// GetContentAnalytics godoc
// @Summary Get content analytics
// @Description Get analytics data for a specific content item
// @Tags Content Management - Analytics
// @Produce json
// @Security BearerAuth
// @Param content_type path string true "Content type" Enums(banner, campaign, seo, faq, legal)
// @Param content_id path int true "Content ID"
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /admin/content/analytics/{content_type}/{content_id} [get]
func (h *ContentHandler) GetContentAnalytics(c *gin.Context) {
	contentType := c.Param("content_type")
	contentID, err := strconv.ParseInt(c.Param("content_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid content ID",
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	validTypes := map[string]bool{
		"banner": true, "campaign": true, "seo": true, "faq": true, "legal": true,
	}
	
	if !validTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid content type",
		})
		return
	}

	analytics, err := h.contentService.GetContentAnalytics(contentType, contentID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get content analytics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     analytics,
		"period":   days,
		"type":     contentType,
		"id":       contentID,
	})
}