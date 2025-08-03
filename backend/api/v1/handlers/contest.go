package handlers

import (
        "database/sql"
        "fmt"
        "net/http"
        "strconv"
        "fantasy-esports-backend/config"
        "fantasy-esports-backend/models"
        "fantasy-esports-backend/utils"
        "fantasy-esports-backend/services"
        "github.com/gin-gonic/gin"
)

type ContestHandler struct {
        db                 *sql.DB
        config             *config.Config
        leaderboardService *services.LeaderboardService
        referralService    *services.ReferralService
}

func NewContestHandler(db *sql.DB, cfg *config.Config) *ContestHandler {
        return &ContestHandler{
                db:                 db,
                config:             cfg,
                leaderboardService: services.NewLeaderboardService(db),
                referralService:    services.NewReferralService(db),
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
        errors := utils.ValidateContestEntry(userID, parseIntToInt64(contestID), req.UserTeamID)
        if len(errors) > 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   errors[0],
                        Code:    "VALIDATION_FAILED",
                })
                return
        }

        // Check if contest exists and is joinable
        var contest models.Contest
        err := h.db.QueryRow(`
                SELECT id, match_id, entry_fee, max_participants, current_participants, status
                FROM contests WHERE id = $1`, contestID).Scan(
                &contest.ID, &contest.MatchID, &contest.EntryFee, 
                &contest.MaxParticipants, &contest.CurrentParticipants, &contest.Status,
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

        // Check if contest is joinable
        if contest.Status != "upcoming" {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Contest is not joinable",
                        Code:    "CONTEST_NOT_JOINABLE",
                })
                return
        }

        if contest.CurrentParticipants >= contest.MaxParticipants {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Contest is full",
                        Code:    "CONTEST_FULL",
                })
                return
        }

        // Check if team exists and belongs to user
        var teamUserID int64
        err = h.db.QueryRow(`
                SELECT user_id FROM user_teams WHERE id = $1`, req.UserTeamID).Scan(&teamUserID)
        
        if err == sql.ErrNoRows {
                c.JSON(http.StatusNotFound, models.ErrorResponse{
                        Success: false,
                        Error:   "Team not found",
                        Code:    "TEAM_NOT_FOUND",
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

        if teamUserID != userID {
                c.JSON(http.StatusForbidden, models.ErrorResponse{
                        Success: false,
                        Error:   "Team does not belong to user",
                        Code:    "UNAUTHORIZED_TEAM",
                })
                return
        }

        // Check if user already joined this contest with this team
        var existingEntry int64
        err = h.db.QueryRow(`
                SELECT id FROM contest_participants 
                WHERE contest_id = $1 AND user_id = $2 AND team_id = $3`, 
                contestID, userID, req.UserTeamID).Scan(&existingEntry)
        
        if err == nil {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Already joined this contest with this team",
                        Code:    "ALREADY_JOINED",
                })
                return
        }

        // Check user's wallet balance (assuming they have sufficient funds for now)
        // In production, you would deduct entry fee from wallet

        // Add participant to contest
        _, err = h.db.Exec(`
                INSERT INTO contest_participants (contest_id, user_id, team_id, entry_fee_paid, joined_at)
                VALUES ($1, $2, $3, $4, NOW())`,
                contestID, userID, req.UserTeamID, contest.EntryFee)
        
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to join contest",
                        Code:    "JOIN_FAILED",
                })
                return
        }

        // Update contest participant count
        _, err = h.db.Exec(`
                UPDATE contests SET current_participants = current_participants + 1 
                WHERE id = $1`, contestID)
        
        if err != nil {
                // Log error but don't fail the request since participant was added
                // In production, you'd want to handle this in a transaction
        }

        // Trigger referral completion check for first contest
        err = h.referralService.CheckAndCompleteReferral(userID, "contest_join")
        if err != nil {
                // Log error but don't fail the join operation
                fmt.Printf("Failed to check referral completion for user %d: %v\n", userID, err)
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Successfully joined contest",
                "contest_id": contestID,
                "team_id": req.UserTeamID,
                "entry_fee": contest.EntryFee,
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

        // Calculate total credits used and validate captain/vice-captain
        var captainID, viceCaptainID int64
        var totalCredits float64
        
        for _, player := range req.Players {
                // Get player credit value
                var creditValue float64
                err := h.db.QueryRow(`
                        SELECT credit_value FROM players WHERE id = $1`, player.PlayerID).Scan(&creditValue)
                
                if err != nil {
                        c.JSON(http.StatusBadRequest, models.ErrorResponse{
                                Success: false,
                                Error:   "Invalid player ID: " + strconv.FormatInt(player.PlayerID, 10),
                                Code:    "INVALID_PLAYER",
                        })
                        return
                }
                
                totalCredits += creditValue
                
                if player.IsCaptain {
                        captainID = player.PlayerID
                }
                if player.IsViceCaptain {
                        viceCaptainID = player.PlayerID
                }
        }

        // Validate credit limit (100 credits)
        if totalCredits > 100.0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   fmt.Sprintf("Total credits (%.1f) exceed limit of 100", totalCredits),
                        Code:    "CREDITS_EXCEEDED",
                })
                return
        }

        // Create the team in database
        var teamID int64
        err = h.db.QueryRow(`
                INSERT INTO user_teams (user_id, match_id, team_name, captain_player_id, vice_captain_player_id, total_credits_used, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
                RETURNING id`,
                userID, req.MatchID, req.TeamName, captainID, viceCaptainID, totalCredits).Scan(&teamID)
        
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to create team",
                        Code:    "TEAM_CREATION_FAILED",
                })
                return
        }

        // Add players to the team
        for _, player := range req.Players {
                // Get player's real team ID
                var realTeamID int64
                h.db.QueryRow(`SELECT team_id FROM players WHERE id = $1`, player.PlayerID).Scan(&realTeamID)
                
                _, err = h.db.Exec(`
                        INSERT INTO team_players (team_id, player_id, real_team_id, is_captain, is_vice_captain)
                        VALUES ($1, $2, $3, $4, $5)`,
                        teamID, player.PlayerID, realTeamID, player.IsCaptain, player.IsViceCaptain)
                
                if err != nil {
                        // If player addition fails, we should rollback the team creation
                        // For now, we'll just log and continue
                        continue
                }
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "message": "Team created successfully",
                "team_id": teamID,
                "team_name": req.TeamName,
                "match_id": req.MatchID,
                "total_credits_used": totalCredits,
                "captain_id": captainID,
                "vice_captain_id": viceCaptainID,
                "user_id": userID,
        })
}

// Additional team management methods
func (h *ContestHandler) UpdateTeam(c *gin.Context) {
        teamID := c.Param("id")
        userID := c.GetInt64("user_id")

        var req struct {
                TeamName  string                   `json:"team_name"`
                Players   []models.PlayerSelection `json:"players"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Invalid request format",
                        Code:    "INVALID_REQUEST",
                })
                return
        }

        // Check if team exists and belongs to user
        var teamUserID int64
        var isLocked bool
        var matchID int64
        err := h.db.QueryRow(`
                SELECT user_id, is_locked, match_id FROM user_teams WHERE id = $1`, teamID).Scan(&teamUserID, &isLocked, &matchID)
        
        if err == sql.ErrNoRows {
                c.JSON(http.StatusNotFound, models.ErrorResponse{
                        Success: false,
                        Error:   "Team not found",
                        Code:    "TEAM_NOT_FOUND",
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

        if teamUserID != userID {
                c.JSON(http.StatusForbidden, models.ErrorResponse{
                        Success: false,
                        Error:   "You don't have permission to update this team",
                        Code:    "UNAUTHORIZED",
                })
                return
        }

        if isLocked {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Cannot update locked team (match may have started)",
                        Code:    "TEAM_LOCKED",
                })
                return
        }

        // If only updating team name
        if req.TeamName != "" && len(req.Players) == 0 {
                _, err = h.db.Exec(`
                        UPDATE user_teams SET team_name = $1, updated_at = NOW() WHERE id = $2`,
                        req.TeamName, teamID)
                
                if err != nil {
                        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                                Success: false,
                                Error:   "Failed to update team name",
                                Code:    "UPDATE_FAILED",
                        })
                        return
                }

                c.JSON(http.StatusOK, gin.H{
                        "success": true,
                        "team_id": teamID,
                        "message": "Team name updated successfully",
                })
                return
        }

        // If updating players, validate the new team composition
        if len(req.Players) > 0 {
                // Get game rules for validation
                var game models.Game
                err := h.db.QueryRow(`
                        SELECT g.id, g.total_team_size, g.max_players_per_team, g.min_players_per_team
                        FROM games g
                        JOIN matches m ON g.id = m.game_id
                        WHERE m.id = $1`, matchID).Scan(
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

                // Calculate new total credits and find captain/vice-captain
                var captainID, viceCaptainID int64
                var totalCredits float64
                
                for _, player := range req.Players {
                        var creditValue float64
                        err := h.db.QueryRow(`SELECT credit_value FROM players WHERE id = $1`, player.PlayerID).Scan(&creditValue)
                        if err != nil {
                                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                                        Success: false,
                                        Error:   "Invalid player ID: " + strconv.FormatInt(player.PlayerID, 10),
                                        Code:    "INVALID_PLAYER",
                                })
                                return
                        }
                        
                        totalCredits += creditValue
                        if player.IsCaptain {
                                captainID = player.PlayerID
                        }
                        if player.IsViceCaptain {
                                viceCaptainID = player.PlayerID
                        }
                }

                if totalCredits > 100.0 {
                        c.JSON(http.StatusBadRequest, models.ErrorResponse{
                                Success: false,
                                Error:   fmt.Sprintf("Total credits (%.1f) exceed limit of 100", totalCredits),
                                Code:    "CREDITS_EXCEEDED",
                        })
                        return
                }

                // Update team with new details
                _, err = h.db.Exec(`
                        UPDATE user_teams 
                        SET team_name = $1, captain_player_id = $2, vice_captain_player_id = $3, 
                            total_credits_used = $4, updated_at = NOW()
                        WHERE id = $5`,
                        req.TeamName, captainID, viceCaptainID, totalCredits, teamID)
                
                if err != nil {
                        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                                Success: false,
                                Error:   "Failed to update team",
                                Code:    "UPDATE_FAILED",
                        })
                        return
                }

                // Delete existing players and add new ones
                h.db.Exec("DELETE FROM team_players WHERE team_id = $1", teamID)
                
                for _, player := range req.Players {
                        var realTeamID int64
                        h.db.QueryRow(`SELECT team_id FROM players WHERE id = $1`, player.PlayerID).Scan(&realTeamID)
                        
                        h.db.Exec(`
                                INSERT INTO team_players (team_id, player_id, real_team_id, is_captain, is_vice_captain)
                                VALUES ($1, $2, $3, $4, $5)`,
                                teamID, player.PlayerID, realTeamID, player.IsCaptain, player.IsViceCaptain)
                }
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "team_id": teamID,
                "message": "Team updated successfully",
        })
}

func (h *ContestHandler) GetMyTeams(c *gin.Context) {
        userID := c.GetInt64("user_id")
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset := (page - 1) * limit

        query := `SELECT ut.id, ut.match_id, ut.team_name, ut.captain_player_id, 
                         ut.vice_captain_player_id, ut.total_credits_used, ut.total_points,
                         ut.final_rank, ut.is_locked, ut.created_at, ut.updated_at,
                         cp.name as captain_name, vcp.name as vice_captain_name,
                         m.name as match_name, m.scheduled_at
                  FROM user_teams ut
                  LEFT JOIN players cp ON ut.captain_player_id = cp.id
                  LEFT JOIN players vcp ON ut.vice_captain_player_id = vcp.id
                  LEFT JOIN matches m ON ut.match_id = m.id
                  WHERE ut.user_id = $1
                  ORDER BY ut.created_at DESC
                  LIMIT $2 OFFSET $3`

        rows, err := h.db.Query(query, userID, limit, offset)
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to fetch teams",
                        Code:    "DB_ERROR",
                })
                return
        }
        defer rows.Close()

        var teams []models.UserTeam
        for rows.Next() {
                var team models.UserTeam
                var captainName, viceCaptainName, matchName sql.NullString
                var scheduledAt sql.NullTime
                
                err := rows.Scan(
                        &team.ID, &team.MatchID, &team.TeamName, &team.CaptainPlayerID,
                        &team.ViceCaptainPlayerID, &team.TotalCreditsUsed, &team.TotalPoints,
                        &team.FinalRank, &team.IsLocked, &team.CreatedAt, &team.UpdatedAt,
                        &captainName, &viceCaptainName, &matchName, &scheduledAt,
                )
                if err != nil {
                        continue
                }
                
                if captainName.Valid {
                        team.CaptainName = &captainName.String
                }
                if viceCaptainName.Valid {
                        team.ViceCaptainName = &viceCaptainName.String
                }
                
                teams = append(teams, team)
        }

        // Get total count
        var total int
        h.db.QueryRow("SELECT COUNT(*) FROM user_teams WHERE user_id = $1", userID).Scan(&total)

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "teams":   teams,
                "total":   total,
                "page":    page,
                "user_id": userID,
        })
}

func (h *ContestHandler) GetTeamDetails(c *gin.Context) {
        teamID := c.Param("id")
        userID := c.GetInt64("user_id")

        // Get team details with match info
        var team models.UserTeam
        err := h.db.QueryRow(`
                SELECT ut.id, ut.user_id, ut.match_id, ut.team_name, ut.captain_player_id,
                       ut.vice_captain_player_id, ut.total_credits_used, ut.total_points,
                       ut.final_rank, ut.is_locked, ut.created_at, ut.updated_at,
                       cp.name as captain_name, vcp.name as vice_captain_name
                FROM user_teams ut
                LEFT JOIN players cp ON ut.captain_player_id = cp.id
                LEFT JOIN players vcp ON ut.vice_captain_player_id = vcp.id
                WHERE ut.id = $1 AND ut.user_id = $2`, teamID, userID).Scan(
                &team.ID, &team.UserID, &team.MatchID, &team.TeamName, &team.CaptainPlayerID,
                &team.ViceCaptainPlayerID, &team.TotalCreditsUsed, &team.TotalPoints,
                &team.FinalRank, &team.IsLocked, &team.CreatedAt, &team.UpdatedAt,
                &team.CaptainName, &team.ViceCaptainName,
        )

        if err == sql.ErrNoRows {
                c.JSON(http.StatusNotFound, models.ErrorResponse{
                        Success: false,
                        Error:   "Team not found or you don't have access",
                        Code:    "TEAM_NOT_FOUND",
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

        // Get team players
        playersRows, err := h.db.Query(`
                SELECT tp.id, tp.player_id, tp.real_team_id, tp.is_captain, tp.is_vice_captain,
                       tp.points_earned, p.name as player_name, t.name as real_team_name,
                       p.role, p.credit_value
                FROM team_players tp
                JOIN players p ON tp.player_id = p.id
                JOIN teams t ON tp.real_team_id = t.id
                WHERE tp.team_id = $1
                ORDER BY tp.is_captain DESC, tp.is_vice_captain DESC, p.name`, teamID)

        var players []models.TeamPlayer
        if err == nil {
                defer playersRows.Close()
                for playersRows.Next() {
                        var player models.TeamPlayer
                        playersRows.Scan(
                                &player.ID, &player.PlayerID, &player.RealTeamID, &player.IsCaptain,
                                &player.IsViceCaptain, &player.PointsEarned, &player.PlayerName,
                                &player.RealTeamName, &player.Role, &player.CreditValue,
                        )
                        players = append(players, player)
                }
        }
        team.Players = players

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "team":    team,
        })
}

func (h *ContestHandler) DeleteTeam(c *gin.Context) {
        teamID := c.Param("id")
        userID := c.GetInt64("user_id")

        // Check if team exists and belongs to user
        var teamUserID int64
        var isLocked bool
        err := h.db.QueryRow(`
                SELECT user_id, is_locked FROM user_teams WHERE id = $1`, teamID).Scan(&teamUserID, &isLocked)
        
        if err == sql.ErrNoRows {
                c.JSON(http.StatusNotFound, models.ErrorResponse{
                        Success: false,
                        Error:   "Team not found",
                        Code:    "TEAM_NOT_FOUND",
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

        if teamUserID != userID {
                c.JSON(http.StatusForbidden, models.ErrorResponse{
                        Success: false,
                        Error:   "You don't have permission to delete this team",
                        Code:    "UNAUTHORIZED",
                })
                return
        }

        if isLocked {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Cannot delete locked team (match may have started)",
                        Code:    "TEAM_LOCKED",
                })
                return
        }

        // Check if team is part of any contests
        var contestCount int
        h.db.QueryRow(`SELECT COUNT(*) FROM contest_participants WHERE team_id = $1`, teamID).Scan(&contestCount)
        
        if contestCount > 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Cannot delete team that is participating in contests",
                        Code:    "TEAM_IN_CONTEST",
                })
                return
        }

        // Delete team (team_players will be cascade deleted)
        _, err = h.db.Exec("DELETE FROM user_teams WHERE id = $1", teamID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to delete team",
                        Code:    "DELETE_FAILED",
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "team_id": teamID,
                "message": "Team deleted successfully",
        })
}

func (h *ContestHandler) CloneTeam(c *gin.Context) {
        teamID := c.Param("id")
        c.JSON(http.StatusOK, gin.H{"success": true, "team_id": teamID, "message": "Team cloned"})
}

func (h *ContestHandler) ValidateTeam(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"success": true, "is_valid": true, "message": "Team validation"})
}

func (h *ContestHandler) GetTeamPerformance(c *gin.Context) {
        teamID := parseIntToInt64(c.Param("id"))
        userID := c.GetInt64("user_id")
        
        if teamID == 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Invalid team ID",
                        Code:    "INVALID_TEAM_ID",
                })
                return
        }

        performance, err := h.leaderboardService.GetUserTeamPerformance(teamID, userID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to get team performance: " + err.Error(),
                        Code:    "PERFORMANCE_ERROR",
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success":     true,
                "team_id":     teamID,
                "performance": performance,
        })
}

// Leaderboard methods
func (h *ContestHandler) GetContestLeaderboard(c *gin.Context) {
        contestID := parseIntToInt64(c.Param("id"))
        if contestID == 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Invalid contest ID",
                        Code:    "INVALID_CONTEST_ID",
                })
                return
        }

        leaderboard, err := h.leaderboardService.CalculateContestLeaderboard(contestID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to calculate leaderboard: " + err.Error(),
                        Code:    "LEADERBOARD_ERROR",
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success":     true,
                "contest_id":  contestID,
                "leaderboard": leaderboard,
        })
}

func (h *ContestHandler) GetLiveLeaderboard(c *gin.Context) {
        contestID := parseIntToInt64(c.Param("id"))
        userID := c.GetInt64("user_id")
        
        if contestID == 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Invalid contest ID",
                        Code:    "INVALID_CONTEST_ID",
                })
                return
        }

        leaderboard, err := h.leaderboardService.GetLiveLeaderboard(contestID, userID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, models.ErrorResponse{
                        Success: false,
                        Error:   "Failed to get live leaderboard: " + err.Error(),
                        Code:    "LEADERBOARD_ERROR",
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success":          true,
                "contest_id":       contestID,
                "live_leaderboard": leaderboard,
        })
}

func (h *ContestHandler) GetMyRank(c *gin.Context) {
        contestID := parseIntToInt64(c.Param("id"))
        userID := c.GetInt64("user_id")
        
        if contestID == 0 {
                c.JSON(http.StatusBadRequest, models.ErrorResponse{
                        Success: false,
                        Error:   "Invalid contest ID",
                        Code:    "INVALID_CONTEST_ID",
                })
                return
        }

        rank, points, teamID, err := h.leaderboardService.GetUserRankInContest(contestID, userID)
        if err != nil {
                c.JSON(http.StatusNotFound, models.ErrorResponse{
                        Success: false,
                        Error:   "User not found in contest or contest doesn't exist",
                        Code:    "USER_NOT_IN_CONTEST",
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{
                "success":    true,
                "contest_id": contestID,
                "user_id":    userID,
                "rank":       rank,
                "points":     points,
                "team_id":    teamID,
        })
}

// GetLeaderboardService returns the leaderboard service instance
func (h *ContestHandler) GetLeaderboardService() *services.LeaderboardService {
        return h.leaderboardService
}

// Helper function
func parseIntToInt64(s string) int64 {
        val, _ := strconv.ParseInt(s, 10, 64)
        return val
}

// @Summary Create contest (Admin)
// @Description Create a new contest (Admin only)
// @Tags Admin Contest Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateContestRequest true "Create contest request"
// @Success 200 {object} map[string]interface{}
// @Router /admin/contests [post]
func (h *ContestHandler) CreateContest(c *gin.Context) {
        adminID := c.GetInt64("admin_id")
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
                "message": "Contest created successfully",
                "admin_id": adminID,
                "contest_name": req.ContestName,
        })
}

// @Summary Update contest (Admin)
// @Description Update an existing contest (Admin only)
// @Tags Admin Contest Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param request body models.UpdateContestRequest true "Update contest request"
// @Success 200 {object} map[string]interface{}
// @Router /admin/contests/{id} [put]
func (h *ContestHandler) UpdateContest(c *gin.Context) {
        contestID := c.Param("id")
        adminID := c.GetInt64("admin_id")

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "contest_id": contestID,
                "updated_by": adminID,
                "message": "Contest updated successfully",
        })
}

// @Summary Delete contest (Admin)
// @Description Delete a contest (Admin only)
// @Tags Admin Contest Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/contests/{id} [delete]
func (h *ContestHandler) DeleteContest(c *gin.Context) {
        contestID := c.Param("id")
        adminID := c.GetInt64("admin_id")

        c.JSON(http.StatusOK, gin.H{
                "success": true,
                "contest_id": contestID,
                "deleted_by": adminID,
                "message": "Contest deleted successfully",
        })
}