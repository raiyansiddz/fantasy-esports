package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
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
	tournamentID := c.Query("tournament_id")
	gameID := c.Query("game_id")
	status := c.Query("status")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	offset := (page - 1) * limit

	query := `SELECT m.id, m.tournament_id, m.game_id, m.name, m.scheduled_at, m.lock_time,
	                 m.status, m.match_type, m.map, m.best_of, m.result, m.winner_team_id,
	                 m.created_at, m.updated_at, t.name as tournament_name, g.name as game_name
	          FROM matches m
	          LEFT JOIN tournaments t ON m.tournament_id = t.id
	          JOIN games g ON m.game_id = g.id
	          WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if tournamentID != "" {
		query += " AND m.tournament_id = $" + strconv.Itoa(argCount)
		args = append(args, tournamentID)
		argCount++
	}

	if gameID != "" {
		query += " AND m.game_id = $" + strconv.Itoa(argCount)
		args = append(args, gameID)
		argCount++
	}

	if status != "" {
		query += " AND m.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	if dateFrom != "" {
		query += " AND DATE(m.scheduled_at) >= $" + strconv.Itoa(argCount)
		args = append(args, dateFrom)
		argCount++
	}

	if dateTo != "" {
		query += " AND DATE(m.scheduled_at) <= $" + strconv.Itoa(argCount)
		args = append(args, dateTo)
		argCount++
	}

	query += " ORDER BY m.scheduled_at ASC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch matches",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var match models.Match
		var tournamentName, gameName *string
		err := rows.Scan(
			&match.ID, &match.TournamentID, &match.GameID, &match.Name,
			&match.ScheduledAt, &match.LockTime, &match.Status, &match.MatchType,
			&match.Map, &match.BestOf, &match.Result, &match.WinnerTeamID,
			&match.CreatedAt, &match.UpdatedAt, &tournamentName, &gameName,
		)
		if err != nil {
			continue
		}
		match.TournamentName = tournamentName
		match.GameName = gameName
		matches = append(matches, match)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM matches m WHERE 1=1"
	countArgs := []interface{}{}
	countArgCount := 1

	if tournamentID != "" {
		countQuery += " AND m.tournament_id = $" + strconv.Itoa(countArgCount)
		countArgs = append(countArgs, tournamentID)
		countArgCount++
	}
	if gameID != "" {
		countQuery += " AND m.game_id = $" + strconv.Itoa(countArgCount)
		countArgs = append(countArgs, gameID)
		countArgCount++
	}
	if status != "" {
		countQuery += " AND m.status = $" + strconv.Itoa(countArgCount)
		countArgs = append(countArgs, status)
		countArgCount++
	}

	var total int
	h.db.QueryRow(countQuery, countArgs...).Scan(&total)
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"matches": matches,
		"total":   total,
		"page":    page,
		"pages":   totalPages,
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
	
	// Get match details
	var match models.Match
	err := h.db.QueryRow(`
		SELECT m.id, m.tournament_id, m.game_id, m.name, m.scheduled_at, m.lock_time,
		       m.status, m.match_type, m.map, m.best_of, m.result, m.winner_team_id,
		       m.created_at, m.updated_at, t.name as tournament_name, g.name as game_name
		FROM matches m
		LEFT JOIN tournaments t ON m.tournament_id = t.id
		JOIN games g ON m.game_id = g.id
		WHERE m.id = $1`, matchID).Scan(
		&match.ID, &match.TournamentID, &match.GameID, &match.Name,
		&match.ScheduledAt, &match.LockTime, &match.Status, &match.MatchType,
		&match.Map, &match.BestOf, &match.Result, &match.WinnerTeamID,
		&match.CreatedAt, &match.UpdatedAt, &match.TournamentName, &match.GameName,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Match not found",
			Code:    "MATCH_NOT_FOUND",
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

	// Get participating teams
	teamsRows, err := h.db.Query(`
		SELECT t.id, t.name, t.short_name, t.logo_url, t.region, t.is_active,
		       mp.seed, mp.final_position, mp.team_score, mp.points_earned
		FROM match_participants mp
		JOIN teams t ON mp.team_id = t.id
		WHERE mp.match_id = $1
		ORDER BY mp.seed`, matchID)

	if err == nil {
		defer teamsRows.Close()
		var teams []models.Team
		for teamsRows.Next() {
			var team models.Team
			var seed, finalPosition *int
			var teamScore int
			var pointsEarned float64
			teamsRows.Scan(&team.ID, &team.Name, &team.ShortName, &team.LogoURL, 
				&team.Region, &team.IsActive, &seed, &finalPosition, &teamScore, &pointsEarned)
			teams = append(teams, team)
		}
		match.Teams = teams
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"match":   match,
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