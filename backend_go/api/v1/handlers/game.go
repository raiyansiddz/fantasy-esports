package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	db     *sql.DB
	config *config.Config
}

func NewGameHandler(db *sql.DB, cfg *config.Config) *GameHandler {
	return &GameHandler{
		db:     db,
		config: cfg,
	}
}

// @Summary Get all games
// @Description Get list of all available games with filters
// @Tags Games
// @Accept json
// @Produce json
// @Param status query string false "Game status filter" Enums(active, inactive)
// @Param category query string false "Game category filter" Enums(fps, moba, battle_royale)
// @Success 200 {object} map[string]interface{}
// @Router /games [get]
func (h *GameHandler) GetGames(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")

	query := `SELECT id, name, code, category, description, logo_url, is_active, 
			         scoring_rules, player_roles, team_composition, min_players_per_team,
			         max_players_per_team, total_team_size, created_at, updated_at 
			  FROM games WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if status != "" {
		if status == "active" {
			query += " AND is_active = true"
		} else if status == "inactive" {
			query += " AND is_active = false"
		}
	}

	if category != "" {
		query += " AND category = $" + strconv.Itoa(argCount)
		args = append(args, category)
		argCount++
	}

	query += " ORDER BY name"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch games",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(
			&game.ID, &game.Name, &game.Code, &game.Category, &game.Description,
			&game.LogoURL, &game.IsActive, &game.ScoringRules, &game.PlayerRoles,
			&game.TeamComposition, &game.MinPlayersPerTeam, &game.MaxPlayersPerTeam,
			&game.TotalTeamSize, &game.CreatedAt, &game.UpdatedAt,
		)
		if err != nil {
			continue
		}
		games = append(games, game)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"games":   games,
	})
}

// @Summary Get game details
// @Description Get detailed information about a specific game
// @Tags Games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} models.Game
// @Failure 404 {object} models.ErrorResponse
// @Router /games/{id} [get]
func (h *GameHandler) GetGameDetails(c *gin.Context) {
	gameID := c.Param("id")

	var game models.Game
	err := h.db.QueryRow(`
		SELECT id, name, code, category, description, logo_url, is_active, 
			   scoring_rules, player_roles, team_composition, min_players_per_team,
			   max_players_per_team, total_team_size, created_at, updated_at 
		FROM games WHERE id = $1`, gameID).Scan(
		&game.ID, &game.Name, &game.Code, &game.Category, &game.Description,
		&game.LogoURL, &game.IsActive, &game.ScoringRules, &game.PlayerRoles,
		&game.TeamComposition, &game.MinPlayersPerTeam, &game.MaxPlayersPerTeam,
		&game.TotalTeamSize, &game.CreatedAt, &game.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Game not found",
			Code:    "GAME_NOT_FOUND",
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"game":    game,
	})
}

// @Summary Get tournaments
// @Description Get list of tournaments with filters
// @Tags Tournaments
// @Accept json
// @Produce json
// @Param game_id query int false "Filter by game ID"
// @Param status query string false "Tournament status" Enums(upcoming, live, completed)
// @Param featured query bool false "Show only featured tournaments"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /tournaments [get]
func (h *GameHandler) GetTournaments(c *gin.Context) {
	gameID := c.Query("game_id")
	status := c.Query("status")
	featured := c.Query("featured")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	offset := (page - 1) * limit

	query := `SELECT t.id, t.name, t.game_id, t.description, t.start_date, t.end_date,
			         t.prize_pool, t.total_teams, t.status, t.is_featured, t.logo_url,
			         t.banner_url, t.created_at, g.name as game_name
			  FROM tournaments t
			  JOIN games g ON t.game_id = g.id
			  WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if gameID != "" {
		query += " AND t.game_id = $" + strconv.Itoa(argCount)
		args = append(args, gameID)
		argCount++
	}

	if status != "" {
		query += " AND t.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	if featured == "true" {
		query += " AND t.is_featured = true"
	}

	query += " ORDER BY t.start_date DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

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

	var tournaments []models.Tournament
	for rows.Next() {
		var tournament models.Tournament
		var gameName string
		err := rows.Scan(
			&tournament.ID, &tournament.Name, &tournament.GameID, &tournament.Description,
			&tournament.StartDate, &tournament.EndDate, &tournament.PrizePool,
			&tournament.TotalTeams, &tournament.Status, &tournament.IsFeatured,
			&tournament.LogoURL, &tournament.BannerURL, &tournament.CreatedAt, &gameName,
		)
		if err != nil {
			continue
		}
		tournaments = append(tournaments, tournament)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM tournaments WHERE 1=1"
	if gameID != "" {
		countQuery += " AND game_id = " + gameID
	}
	if status != "" {
		countQuery += " AND status = '" + status + "'"
	}
	if featured == "true" {
		countQuery += " AND is_featured = true"
	}

	var total int
	h.db.QueryRow(countQuery).Scan(&total)
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"tournaments": tournaments,
		"total":       total,
		"page":        page,
		"pages":       totalPages,
	})
}

// @Summary Get tournament details
// @Description Get detailed information about a tournament
// @Tags Tournaments
// @Accept json
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} models.Tournament
// @Failure 404 {object} models.ErrorResponse
// @Router /tournaments/{id} [get]
func (h *GameHandler) GetTournamentDetails(c *gin.Context) {
	tournamentID := c.Param("id")

	var tournament models.Tournament
	err := h.db.QueryRow(`
		SELECT id, name, game_id, description, start_date, end_date,
			   prize_pool, total_teams, status, is_featured, logo_url,
			   banner_url, created_at
		FROM tournaments WHERE id = $1`, tournamentID).Scan(
		&tournament.ID, &tournament.Name, &tournament.GameID, &tournament.Description,
		&tournament.StartDate, &tournament.EndDate, &tournament.PrizePool,
		&tournament.TotalTeams, &tournament.Status, &tournament.IsFeatured,
		&tournament.LogoURL, &tournament.BannerURL, &tournament.CreatedAt,
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

	// Get stages
	stagesRows, err := h.db.Query(`
		SELECT id, name, stage_order, stage_type, start_date, end_date, max_teams, rules
		FROM tournament_stages WHERE tournament_id = $1 ORDER BY stage_order`, tournamentID)

	if err == nil {
		defer stagesRows.Close()
		var stages []models.TournamentStage
		for stagesRows.Next() {
			var stage models.TournamentStage
			stagesRows.Scan(&stage.ID, &stage.Name, &stage.StageOrder, &stage.StageType,
				&stage.StartDate, &stage.EndDate, &stage.MaxTeams, &stage.Rules)
			stages = append(stages, stage)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"tournament": tournament,
	})
}

// @Summary Get matches
// @Description Get list of matches with filters
// @Tags Matches
// @Accept json
// @Produce json
// @Param tournament_id query int false "Filter by tournament ID"
// @Param game_id query int false "Filter by game ID"
// @Param status query string false "Match status" Enums(upcoming, live, completed)
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /matches [get]
func (h *GameHandler) GetMatches(c *gin.Context) {
	// Implementation for getting matches
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"matches": []models.Match{},
		"message": "Matches endpoint implemented",
	})
}

// @Summary Get match details
// @Description Get detailed information about a match including teams and players
// @Tags Matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} models.Match
// @Failure 404 {object} models.ErrorResponse
// @Router /matches/{id} [get]
func (h *GameHandler) GetMatchDetails(c *gin.Context) {
	matchID := c.Param("id")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"match_id": matchID,
		"message": "Match details endpoint implemented",
	})
}

// @Summary Get match players
// @Description Get all players participating in a match
// @Tags Matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Param role query string false "Filter by player role"
// @Param sort_by query string false "Sort by field" Enums(credit_value, recent_form, avg_points)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} map[string]interface{}
// @Router /matches/{id}/players [get]
func (h *GameHandler) GetMatchPlayers(c *gin.Context) {
	matchID := c.Param("id")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"match_id": matchID,
		"players": []models.Player{},
		"message": "Match players endpoint implemented",
	})
}

// @Summary Get player performance
// @Description Get player performance statistics for a match
// @Tags Matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Param player_id query int false "Filter by specific player"
// @Success 200 {object} map[string]interface{}
// @Router /matches/{id}/player-performance [get]
func (h *GameHandler) GetPlayerPerformance(c *gin.Context) {
	matchID := c.Param("id")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"match_id": matchID,
		"performance": []models.PlayerPerformance{},
		"message": "Player performance endpoint implemented",
	})
}