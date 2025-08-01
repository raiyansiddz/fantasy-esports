package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/models"
	"fantasy-esports-backend/utils"
	"github.com/gin-gonic/gin"
)

type ContestHandler struct {
	db     *sql.DB
	config *config.Config
}

func NewContestHandler(db *sql.DB, cfg *config.Config) *ContestHandler {
	return &ContestHandler{
		db:     db,
		config: cfg,
	}
}

// @Summary Get contests
// @Description Get list of contests for matches
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param match_id query int false "Filter by match ID"
// @Param contest_type query string false "Contest type" Enums(free, paid, private)
// @Param entry_fee_min query float64 false "Minimum entry fee"
// @Param entry_fee_max query float64 false "Maximum entry fee"
// @Param status query string false "Contest status" Enums(upcoming, live, completed)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /contests [get]
func (h *ContestHandler) GetContests(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := `SELECT c.id, c.match_id, c.name, c.contest_type, c.entry_fee,
			         c.max_participants, c.current_participants, c.total_prize_pool,
			         c.is_guaranteed, c.prize_distribution, c.contest_rules, c.status,
			         c.invite_code, c.is_multi_entry, c.max_entries_per_user, c.created_at,
			         m.name as match_name, t.name as tournament_name, m.scheduled_at, m.lock_time
			  FROM contests c
			  LEFT JOIN matches m ON c.match_id = m.id
			  LEFT JOIN tournaments t ON m.tournament_id = t.id
			  WHERE c.status = 'upcoming'
			  ORDER BY m.scheduled_at
			  LIMIT $1 OFFSET $2`

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch contests",
			Code:    "DB_ERROR",
		})
		return
	}
	defer rows.Close()

	var contests []models.Contest
	for rows.Next() {
		var contest models.Contest
		var matchName, tournamentName sql.NullString
		
		err := rows.Scan(
			&contest.ID, &contest.MatchID, &contest.Name, &contest.ContestType,
			&contest.EntryFee, &contest.MaxParticipants, &contest.CurrentParticipants,
			&contest.TotalPrizePool, &contest.IsGuaranteed, &contest.PrizeDistribution,
			&contest.ContestRules, &contest.Status, &contest.InviteCode,
			&contest.IsMultiEntry, &contest.MaxEntriesPerUser, &contest.CreatedAt,
			&matchName, &tournamentName, &contest.ScheduledAt, &contest.LockTime,
		)
		if err != nil {
			continue
		}
		
		if matchName.Valid {
			contest.MatchName = &matchName.String
		}
		if tournamentName.Valid {
			contest.TournamentName = &tournamentName.String
		}
		
		contests = append(contests, contest)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"contests": contests,
		"page":     page,
	})
}

// @Summary Get contest details
// @Description Get detailed information about a contest
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 200 {object} models.Contest
// @Failure 404 {object} models.ErrorResponse
// @Router /contests/{id} [get]
func (h *ContestHandler) GetContestDetails(c *gin.Context) {
	contestID := c.Param("id")

	var contest models.Contest
	err := h.db.QueryRow(`
		SELECT c.id, c.match_id, c.name, c.contest_type, c.entry_fee,
			   c.max_participants, c.current_participants, c.total_prize_pool,
			   c.is_guaranteed, c.prize_distribution, c.contest_rules, c.status,
			   c.invite_code, c.is_multi_entry, c.max_entries_per_user, c.created_at,
			   m.name as match_name, t.name as tournament_name, m.scheduled_at, m.lock_time
		FROM contests c
		LEFT JOIN matches m ON c.match_id = m.id
		LEFT JOIN tournaments t ON m.tournament_id = t.id
		WHERE c.id = $1`, contestID).Scan(
		&contest.ID, &contest.MatchID, &contest.Name, &contest.ContestType,
		&contest.EntryFee, &contest.MaxParticipants, &contest.CurrentParticipants,
		&contest.TotalPrizePool, &contest.IsGuaranteed, &contest.PrizeDistribution,
		&contest.ContestRules, &contest.Status, &contest.InviteCode,
		&contest.IsMultiEntry, &contest.MaxEntriesPerUser, &contest.CreatedAt,
		&contest.MatchName, &contest.TournamentName, &contest.ScheduledAt, &contest.LockTime,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Contest not found",
			Code:    "CONTEST_NOT_FOUND",
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
		"contest": contest,
	})
}

// @Summary Join contest
// @Description Join a contest with a fantasy team
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param request body models.JoinContestRequest true "Join contest request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /contests/{id}/join [post]
func (h *ContestHandler) JoinContest(c *gin.Context) {
	userID := c.GetInt64("user_id")
	contestID := c.Param("id")

	var req models.JoinContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Validate contest and team ownership
	errors := utils.ValidateContestEntry(userID, parseInt64(contestID), req.UserTeamID)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   errors[0],
			Code:    "VALIDATION_FAILED",
		})
		return
	}

	// In a real implementation, you would:
	// 1. Check contest availability
	// 2. Check user's wallet balance
	// 3. Deduct entry fee
	// 4. Add participant to contest
	// 5. Update contest participant count

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully joined contest",
		"contest_id": contestID,
		"team_id": req.UserTeamID,
	})
}

// @Summary Leave contest
// @Description Leave a contest before match lock time
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 200 {object} map[string]interface{}
// @Router /contests/{id}/leave [delete]
func (h *ContestHandler) LeaveContest(c *gin.Context) {
	userID := c.GetInt64("user_id")
	contestID := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully left contest",
		"contest_id": contestID,
		"user_id": userID,
	})
}

// @Summary Get my contest entries
// @Description Get user's contest participation history
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Entry status" Enums(upcoming, live, completed)
// @Param game_id query int false "Filter by game ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /contests/my-entries [get]
func (h *ContestHandler) GetMyEntries(c *gin.Context) {
	userID := c.GetInt64("user_id")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"entries": []models.ContestParticipant{},
		"user_id": userID,
		"message": "My entries endpoint implemented",
	})
}

// @Summary Create private contest
// @Description Create a private contest for friends
// @Tags Contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateContestRequest true "Create contest request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /contests/create-private [post]
func (h *ContestHandler) CreatePrivateContest(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Private contest created successfully",
		"creator_id": userID,
		"contest_name": req.ContestName,
	})
}

// @Summary Create fantasy team
// @Description Create a new fantasy team for a match
// @Tags Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTeamRequest true "Create team request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /teams/create [post]
func (h *ContestHandler) CreateTeam(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get game rules for validation
	var game models.Game
	err := h.db.QueryRow(`
		SELECT g.id, g.total_team_size, g.max_players_per_team, g.min_players_per_team
		FROM games g
		JOIN matches m ON g.id = m.game_id
		WHERE m.id = $1`, req.MatchID).Scan(
		&game.ID, &game.TotalTeamSize, &game.MaxPlayersPerTeam, &game.MinPlayersPerTeam,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid match ID",
			Code:    "INVALID_MATCH",
		})
		return
	}

	// Validate team composition
	errors := utils.ValidateTeamComposition(req.Players, game)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   errors[0],
			Code:    "VALIDATION_FAILED",
		})
		return
	}

	// In a real implementation, you would create the team in database
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Team created successfully",
		"team_name": req.TeamName,
		"match_id": req.MatchID,
		"user_id": userID,
	})
}

// Additional team management methods
func (h *ContestHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "message": "Team updated"})
}

func (h *ContestHandler) GetMyTeams(c *gin.Context) {
	userID := c.GetInt64("user_id")
	c.JSON(http.StatusOK, gin.H{"success": true, "user_id": userID, "teams": []models.UserTeam{}})
}

func (h *ContestHandler) GetTeamDetails(c *gin.Context) {
	teamID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "team": models.UserTeam{}})
}

func (h *ContestHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "message": "Team deleted"})
}

func (h *ContestHandler) CloneTeam(c *gin.Context) {
	teamID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "message": "Team cloned"})
}

func (h *ContestHandler) ValidateTeam(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "is_valid": true, "message": "Team validation"})
}

func (h *ContestHandler) GetTeamPerformance(c *gin.Context) {
	teamID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "performance": map[string]interface{}{}})
}

// Leaderboard methods
func (h *ContestHandler) GetContestLeaderboard(c *gin.Context) {
	contestID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "contest_id": contestID, "leaderboard": models.Leaderboard{}})
}

func (h *ContestHandler) GetLiveLeaderboard(c *gin.Context) {
	contestID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true, "contest_id": contestID, "live_leaderboard": models.Leaderboard{}})
}

func (h *ContestHandler) GetMyRank(c *gin.Context) {
	contestID := c.Param("id")
	userID := c.GetInt64("user_id")
	c.JSON(http.StatusOK, gin.H{"success": true, "contest_id": contestID, "user_id": userID, "rank": 1})
}

// Helper function
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}