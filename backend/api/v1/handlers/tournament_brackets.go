package handlers

import (
	"database/sql"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TournamentBracketHandler struct {
	db                       *sql.DB
	cfg                      *config.Config
	tournamentBracketService *services.TournamentBracketService
}

func NewTournamentBracketHandler(db *sql.DB, cfg *config.Config) *TournamentBracketHandler {
	return &TournamentBracketHandler{
		db:                       db,
		cfg:                      cfg,
		tournamentBracketService: services.NewTournamentBracketService(db),
	}
}

func (h *TournamentBracketHandler) CreateBracket(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	var req models.CreateBracketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	// Validate bracket type
	validTypes := map[string]bool{
		"single_elimination": true,
		"double_elimination": true,
		"round_robin":        true,
		"swiss":              true,
	}

	if !validTypes[req.BracketType] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid bracket type",
			Code:    "INVALID_BRACKET_TYPE",
		})
		return
	}

	bracket, err := h.tournamentBracketService.CreateBracket(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "CREATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"bracket": bracket,
	})
}

func (h *TournamentBracketHandler) GetBracket(c *gin.Context) {
	bracketID, err := strconv.ParseInt(c.Param("bracket_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid bracket ID",
			Code:    "INVALID_ID",
		})
		return
	}

	bracket, err := h.tournamentBracketService.GetBracket(bracketID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Success: false,
				Error:   "Bracket not found",
				Code:    "NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch bracket",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"bracket": bracket,
	})
}

func (h *TournamentBracketHandler) GetTournamentBrackets(c *gin.Context) {
	tournamentID, err := strconv.ParseInt(c.Param("tournament_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid tournament ID",
			Code:    "INVALID_ID",
		})
		return
	}

	brackets, err := h.tournamentBracketService.GetTournamentBrackets(tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch brackets",
			Code:    "FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"brackets": brackets,
	})
}

func (h *TournamentBracketHandler) AdvanceBracket(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	bracketID, err := strconv.ParseInt(c.Param("bracket_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid bracket ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req struct {
		MatchResults map[string]interface{} `json:"match_results" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	err = h.tournamentBracketService.AdvanceBracket(bracketID, req.MatchResults)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    "ADVANCE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bracket advanced successfully",
	})
}

func (h *TournamentBracketHandler) UpdateBracketStatus(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	bracketID, err := strconv.ParseInt(c.Param("bracket_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid bracket ID",
			Code:    "INVALID_ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request data",
			Code:    "INVALID_DATA",
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"setup":     true,
		"active":    true,
		"completed": true,
		"cancelled": true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid status",
			Code:    "INVALID_STATUS",
		})
		return
	}

	_, err = h.db.Exec(`
		UPDATE tournament_brackets 
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.Status, bracketID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update bracket status",
			Code:    "UPDATE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bracket status updated successfully",
	})
}

func (h *TournamentBracketHandler) DeleteBracket(c *gin.Context) {
	// Verify admin access
	_, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Admin access required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	bracketID, err := strconv.ParseInt(c.Param("bracket_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid bracket ID",
			Code:    "INVALID_ID",
		})
		return
	}

	_, err = h.db.Exec("DELETE FROM tournament_brackets WHERE id = $1", bracketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to delete bracket",
			Code:    "DELETE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bracket deleted successfully",
	})
}

// Helper endpoint to get bracket types and their descriptions
func (h *TournamentBracketHandler) GetBracketTypes(c *gin.Context) {
	bracketTypes := []map[string]interface{}{
		{
			"type":        "single_elimination",
			"name":        "Single Elimination",
			"description": "Teams are eliminated after losing one match. Simple tournament format.",
			"min_teams":   2,
			"max_teams":   128,
		},
		{
			"type":        "double_elimination",
			"name":        "Double Elimination",
			"description": "Teams must lose twice to be eliminated. Winners and losers brackets.",
			"min_teams":   4,
			"max_teams":   64,
		},
		{
			"type":        "round_robin",
			"name":        "Round Robin",
			"description": "Every team plays every other team once. Best for smaller groups.",
			"min_teams":   3,
			"max_teams":   16,
		},
		{
			"type":        "swiss",
			"name":        "Swiss System",
			"description": "Teams are paired based on performance. No elimination until final cut.",
			"min_teams":   8,
			"max_teams":   32,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"bracket_types": bracketTypes,
	})
}