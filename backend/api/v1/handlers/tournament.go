package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/services"
	"fantasy-esports-backend/pkg/cdn"
	
	"github.com/gin-gonic/gin"
)

type TournamentHandler struct {
	db                *sql.DB
	config            *config.Config
	cdn               *cdn.CloudinaryClient
	tournamentService *services.TournamentService
	liveStreamService *services.LiveStreamService
}

func NewTournamentHandler(db *sql.DB, cfg *config.Config, cdn *cdn.CloudinaryClient) *TournamentHandler {
	return &TournamentHandler{
		db:                db,
		config:            cfg,
		cdn:               cdn,
		tournamentService: services.NewTournamentService(db),
		liveStreamService: services.NewLiveStreamService(db),
	}
}

// @Summary Get tournament bracket
// @Description Get complete tournament bracket with stages and matches
// @Tags Tournaments
// @Accept json
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} services.TournamentBracket
// @Failure 404 {object} models.ErrorResponse
// @Router /tournaments/{id}/bracket [get]
func (h *TournamentHandler) GetTournamentBracket(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.ParseInt(tournamentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid tournament ID",
			Code:    "INVALID_TOURNAMENT_ID",
		})
		return
	}

	bracket, err := h.tournamentService.GetTournamentBracket(tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Tournament not found: " + err.Error(),
			Code:    "TOURNAMENT_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"bracket": bracket,
	})
}

// @Summary Get tournaments
// @Description Get list of tournaments with filtering options
// @Tags Tournaments
// @Accept json
// @Produce json
// @Param game_id query int false "Filter by game ID"
// @Param status query string false "Filter by status" Enums(upcoming, live, completed)
// @Param featured query bool false "Filter featured tournaments"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /tournaments [get]
func (h *TournamentHandler) GetTournaments(c *gin.Context) {
	// Parse query parameters
	gameIDStr := c.Query("game_id")
	status := c.Query("status")
	featuredStr := c.Query("featured")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	// Build query
	query := `SELECT t.id, t.name, t.game_id, t.description, t.start_date, t.end_date,
					 t.prize_pool, t.total_teams, t.status, t.is_featured, t.logo_url,
					 t.banner_url, t.created_at, g.name as game_name
			  FROM tournaments t
			  LEFT JOIN games g ON t.game_id = g.id
			  WHERE 1=1`
	
	args := []interface{}{}
	argCount := 1

	// Apply filters
	if gameIDStr != "" {
		query += " AND t.game_id = $" + strconv.Itoa(argCount)
		gameID, _ := strconv.Atoi(gameIDStr)
		args = append(args, gameID)
		argCount++
	}

	if status != "" {
		query += " AND t.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	if featuredStr == "true" {
		query += " AND t.is_featured = true"
	}

	query += " ORDER BY t.start_date DESC"
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch tournaments",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	var tournaments []map[string]interface{} = make([]map[string]interface{}, 0)
	for rows.Next() {
		var tournament models.Tournament
		var gameName sql.NullString
		
		err := rows.Scan(
			&tournament.ID, &tournament.Name, &tournament.GameID, &tournament.Description,
			&tournament.StartDate, &tournament.EndDate, &tournament.PrizePool,
			&tournament.TotalTeams, &tournament.Status, &tournament.IsFeatured,
			&tournament.LogoURL, &tournament.BannerURL, &tournament.CreatedAt, &gameName,
		)
		if err != nil {
			continue
		}

		tournamentData := map[string]interface{}{
			"id":          tournament.ID,
			"name":        tournament.Name,
			"game_id":     tournament.GameID,
			"game_name":   gameName.String,
			"description": tournament.Description,
			"start_date":  tournament.StartDate,
			"end_date":    tournament.EndDate,
			"prize_pool":  tournament.PrizePool,
			"total_teams": tournament.TotalTeams,
			"status":      tournament.Status,
			"is_featured": tournament.IsFeatured,
			"logo_url":    tournament.LogoURL,
			"banner_url":  tournament.BannerURL,
			"created_at":  tournament.CreatedAt,
		}

		tournaments = append(tournaments, tournamentData)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM tournaments t WHERE 1=1`
	countArgs := []interface{}{}
	countArgCount := 1

	if gameIDStr != "" {
		countQuery += " AND t.game_id = $" + strconv.Itoa(countArgCount)
		gameID, _ := strconv.Atoi(gameIDStr)
		countArgs = append(countArgs, gameID)
		countArgCount++
	}

	if status != "" {
		countQuery += " AND t.status = $" + strconv.Itoa(countArgCount)
		countArgs = append(countArgs, status)
		countArgCount++
	}

	if featuredStr == "true" {
		countQuery += " AND t.is_featured = true"
	}

	h.db.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"tournaments": tournaments,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  (total + limit - 1) / limit,
			"total_items":  total,
			"per_page":     limit,
		},
	})
}

// @Summary Get tournament details
// @Description Get detailed information about a specific tournament
// @Tags Tournaments
// @Accept json
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} models.Tournament
// @Failure 404 {object} models.ErrorResponse
// @Router /tournaments/{id} [get]
func (h *TournamentHandler) GetTournamentDetails(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.ParseInt(tournamentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid tournament ID",
			Code:    "INVALID_TOURNAMENT_ID",
		})
		return
	}

	// Get tournament details with game info
	var tournament models.Tournament
	var gameName string
	err = h.db.QueryRow(`
		SELECT t.id, t.name, t.game_id, t.description, t.start_date, t.end_date,
			   t.prize_pool, t.total_teams, t.status, t.is_featured, t.logo_url,
			   t.banner_url, t.created_at, g.name as game_name
		FROM tournaments t
		LEFT JOIN games g ON t.game_id = g.id
		WHERE t.id = $1`, tournamentID).Scan(
		&tournament.ID, &tournament.Name, &tournament.GameID, &tournament.Description,
		&tournament.StartDate, &tournament.EndDate, &tournament.PrizePool,
		&tournament.TotalTeams, &tournament.Status, &tournament.IsFeatured,
		&tournament.LogoURL, &tournament.BannerURL, &tournament.CreatedAt, &gameName,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Tournament not found",
			Code:    "TOURNAMENT_NOT_FOUND",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Database error",
			Code:    "DB_ERROR",
		})
		return
	}

	// Get tournament stages
	stageRows, err := h.db.Query(`
		SELECT id, name, stage_order, stage_type, start_date, end_date, max_teams
		FROM tournament_stages 
		WHERE tournament_id = $1 
		ORDER BY stage_order`, tournamentID)

	var stages []models.TournamentStage
	if err == nil {
		defer stageRows.Close()
		for stageRows.Next() {
			var stage models.TournamentStage
			stage.TournamentID = tournamentID
			stageRows.Scan(&stage.ID, &stage.Name, &stage.StageOrder, &stage.StageType,
				&stage.StartDate, &stage.EndDate, &stage.MaxTeams)
			stages = append(stages, stage)
		}
	}

	// Prepare response
	response := gin.H{
		"success":    true,
		"tournament": gin.H{
			"id":          tournament.ID,
			"name":        tournament.Name,
			"game_id":     tournament.GameID,
			"game_name":   gameName,
			"description": tournament.Description,
			"start_date":  tournament.StartDate,
			"end_date":    tournament.EndDate,
			"prize_pool":  tournament.PrizePool,
			"total_teams": tournament.TotalTeams,
			"status":      tournament.Status,
			"is_featured": tournament.IsFeatured,
			"logo_url":    tournament.LogoURL,
			"banner_url":  tournament.BannerURL,
			"created_at":  tournament.CreatedAt,
			"stages":      stages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// =====================
// ADMIN ENDPOINTS
// =====================

// @Summary Create tournament stage
// @Description Create a new stage for a tournament (Admin only)
// @Tags Admin Tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tournament ID"
// @Param request body models.TournamentStage true "Stage data"
// @Success 201 {object} models.TournamentStage
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/tournaments/{id}/stages [post]
func (h *TournamentHandler) CreateTournamentStage(c *gin.Context) {
	tournamentIDStr := c.Param("id")
	tournamentID, err := strconv.ParseInt(tournamentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid tournament ID",
			Code:    "INVALID_TOURNAMENT_ID",
		})
		return
	}

	var stage models.TournamentStage
	if err := c.ShouldBindJSON(&stage); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Create the stage
	newStage, err := h.tournamentService.CreateTournamentStage(tournamentID, stage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create stage: " + err.Error(),
			Code:    "STAGE_CREATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"stage":   newStage,
		"message": "Tournament stage created successfully",
	})
}

// @Summary Advance tournament stage
// @Description Advance teams from one stage to the next (Admin only)
// @Tags Admin Tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param stage_id path int true "Current Stage ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/tournaments/stages/{stage_id}/advance [post]
func (h *TournamentHandler) AdvanceToNextStage(c *gin.Context) {
	stageIDStr := c.Param("stage_id")
	stageID, err := strconv.ParseInt(stageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid stage ID",
			Code:    "INVALID_STAGE_ID",
		})
		return
	}

	err = h.tournamentService.AdvanceToNextStage(stageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Failed to advance to next stage: " + err.Error(),
			Code:    "ADVANCEMENT_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully advanced teams to next stage",
		"stage_id": stageID,
	})
}

// @Summary Set match live stream
// @Description Configure live stream for a match (Admin only)
// @Tags Admin Live Stream
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body map[string]interface{} true "Stream configuration"
// @Success 200 {object} services.LiveStream
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/matches/{id}/live-stream [post]
func (h *TournamentHandler) SetMatchLiveStream(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	var request struct {
		StreamURL         string  `json:"stream_url" binding:"required"`
		StreamTitle       *string `json:"stream_title"`
		StreamDescription *string `json:"stream_description"`
		AutoActivate      bool    `json:"auto_activate"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate stream URL
	err = h.liveStreamService.ValidateStreamURL(request.StreamURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid stream URL: " + err.Error(),
			Code:    "INVALID_STREAM_URL",
		})
		return
	}

	// Set live stream
	stream, err := h.liveStreamService.SetMatchLiveStream(
		matchID, request.StreamURL, request.StreamTitle, request.StreamDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to set live stream: " + err.Error(),
			Code:    "STREAM_SETUP_FAILED",
		})
		return
	}

	// Auto-activate if requested
	if request.AutoActivate {
		err = h.liveStreamService.ActivateMatchStream(matchID, true)
		if err != nil {
			// Log error but don't fail the request
			// Stream was set up successfully, activation just failed
		} else {
			stream.IsActive = true
			stream.StartedAt = &time.Time{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stream":  stream,
		"message": "Live stream configured successfully",
	})
}

// @Summary Activate/deactivate match live stream
// @Description Activate or deactivate live stream for a match (Admin only)
// @Tags Admin Live Stream
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Param request body map[string]interface{} true "Activation request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/matches/{id}/live-stream/activate [put]
func (h *TournamentHandler) ActivateMatchLiveStream(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	var request struct {
		Activate bool `json:"activate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	err = h.liveStreamService.ActivateMatchStream(matchID, request.Activate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update stream status: " + err.Error(),
			Code:    "STREAM_ACTIVATION_FAILED",
		})
		return
	}

	status := "deactivated"
	if request.Activate {
		status = "activated"
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"match_id":  matchID,
		"activated": request.Activate,
		"message":   fmt.Sprintf("Live stream %s successfully", status),
	})
}

// @Summary Get match live stream
// @Description Get live stream information for a match
// @Tags Live Stream
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} services.LiveStream
// @Failure 404 {object} models.ErrorResponse
// @Router /matches/{id}/live-stream [get]
func (h *TournamentHandler) GetMatchLiveStream(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	stream, err := h.liveStreamService.GetMatchLiveStream(matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Live stream not found: " + err.Error(),
			Code:    "STREAM_NOT_FOUND",
		})
		return
	}

	// Get additional stream stats
	stats, _ := h.liveStreamService.GetMatchStreamStats(matchID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stream":  stream,
		"stats":   stats,
	})
}

// @Summary Get active live streams
// @Description Get all currently active live streams
// @Tags Live Stream
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /live-streams/active [get]
func (h *TournamentHandler) GetActiveLiveStreams(c *gin.Context) {
	streams, err := h.liveStreamService.GetActiveLiveStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to get active streams: " + err.Error(),
			Code:    "STREAMS_FETCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"active_streams": streams,
		"count":         len(streams),
	})
}

// @Summary Remove match live stream
// @Description Remove live stream configuration for a match (Admin only)
// @Tags Admin Live Stream
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Match ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/matches/{id}/live-stream [delete]
func (h *TournamentHandler) RemoveMatchLiveStream(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_MATCH_ID",
		})
		return
	}

	err = h.liveStreamService.RemoveMatchLiveStream(matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Failed to remove live stream: " + err.Error(),
			Code:    "STREAM_REMOVAL_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"match_id": matchID,
		"message":  "Live stream configuration removed successfully",
	})
}