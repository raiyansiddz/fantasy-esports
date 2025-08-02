// distributePrizes distributes prizes to winners - FIXED VERSION
func (h *AdminHandler) distributePrizes(tx *sql.Tx, matchID string) (map[string]interface{}, error) {
        prizeDistribution := make(map[string]interface{})
        
        // First, check if there are any contest participants for this match
        var participantCount int
        err := tx.QueryRow(`
                SELECT COUNT(*)
                FROM contest_participants cp
                JOIN contests c ON cp.contest_id = c.id
                WHERE c.match_id = $1`, matchID).Scan(&participantCount)
        
        if err != nil {
                return nil, err
        }
        
        // Handle the case where no participants exist - return success with zero distributions
        if participantCount == 0 {
                prizeDistribution["total_amount"] = 0.0
                prizeDistribution["contests_processed"] = 0
                prizeDistribution["winners_rewarded"] = 0
                prizeDistribution["distribution_timestamp"] = time.Now()
                prizeDistribution["message"] = "No contest participants found - prize distribution completed with zero distributions"
                return prizeDistribution, nil
        }
        
        // Get all contests for this match that have prizes
        rows, err := tx.Query(`
                SELECT id, total_prize_pool, prize_distribution
                FROM contests 
                WHERE match_id = $1 AND total_prize_pool > 0`, matchID)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        totalPrizesDistributed := 0.0
        contestsWithPrizes := 0
        winnersRewarded := 0
        
        for rows.Next() {
                var contestID int64
                var prizePool float64
                var prizeDistributionJSON string
                
                if err := rows.Scan(&contestID, &prizePool, &prizeDistributionJSON); err != nil {
                        continue
                }
                
                // Parse prize distribution JSON to get percentages
                var prizeDistribution map[string]interface{}
                if err := json.Unmarshal([]byte(prizeDistributionJSON), &prizeDistribution); err != nil {
                        // If JSON parsing fails, use default percentages
                        winnerPct := 50.0
                        runnerUpPct := 30.0
                        
                        // Process with defaults
                        h.processPrizeDistributionForContest(tx, contestID, prizePool, winnerPct, runnerUpPct, &totalPrizesDistributed, &winnersRewarded)
                        contestsWithPrizes++
                        continue
                }
                
                // Extract winner and runner-up percentages (with defaults)
                winnerPct := 50.0  // Default 50% for winner
                runnerUpPct := 30.0 // Default 30% for runner-up
                
                if positions, ok := prizeDistribution["positions"].([]interface{}); ok && len(positions) >= 2 {
                        if pos1, ok := positions[0].(map[string]interface{}); ok {
                                if pct, ok := pos1["percentage"].(float64); ok {
                                        winnerPct = pct
                                }
                        }
                        if pos2, ok := positions[1].(map[string]interface{}); ok {
                                if pct, ok := pos2["percentage"].(float64); ok {
                                        runnerUpPct = pct
                                }
                        }
                }
                
                // Process prize distribution for this contest
                h.processPrizeDistributionForContest(tx, contestID, prizePool, winnerPct, runnerUpPct, &totalPrizesDistributed, &winnersRewarded)
                contestsWithPrizes++
        }
        
        prizeDistribution["total_amount"] = totalPrizesDistributed
        prizeDistribution["contests_processed"] = contestsWithPrizes
        prizeDistribution["winners_rewarded"] = winnersRewarded
        prizeDistribution["distribution_timestamp"] = time.Now()
        
        return prizeDistribution, nil
}

// Helper function to process prize distribution for a single contest
func (h *AdminHandler) processPrizeDistributionForContest(tx *sql.Tx, contestID int64, prizePool, winnerPct, runnerUpPct float64, totalPrizesDistributed *float64, winnersRewarded *int) {
        // Check if this specific contest has participants with ranks
        var contestParticipantCount int
        err := tx.QueryRow(`
                SELECT COUNT(*)
                FROM contest_participants
                WHERE contest_id = $1 AND rank <= 3`, contestID).Scan(&contestParticipantCount)
        
        if err != nil || contestParticipantCount == 0 {
                // No winners for this contest, skip
                return
        }
        
        // Get top 3 winners for this contest
        winnerRows, err := tx.Query(`
                SELECT cp.user_id, cp.rank, ut.total_points
                FROM contest_participants cp
                JOIN user_teams ut ON cp.team_id = ut.id
                WHERE cp.contest_id = $1 AND cp.rank <= 3
                ORDER BY cp.rank`, contestID)
        
        if err != nil {
                return
        }
        defer winnerRows.Close()
        
        // Distribute prizes to winners
        for winnerRows.Next() {
                var userID int64
                var rank int
                var totalPoints float64
                
                if err := winnerRows.Scan(&userID, &rank, &totalPoints); err != nil {
                        continue
                }
                
                var prizeAmount float64
                switch rank {
                case 1:
                        prizeAmount = prizePool * (winnerPct / 100.0)
                case 2:
                        prizeAmount = prizePool * (runnerUpPct / 100.0)
                case 3:
                        prizeAmount = prizePool * 0.1 // 10% for third place
                }
                
                if prizeAmount > 0 {
                        // Add prize to user's wallet
                        _, err = tx.Exec(`
                                INSERT INTO wallet_transactions (user_id, transaction_type, amount, description, status, created_at)
                                VALUES ($1, 'prize_credit', $2, $3, 'completed', NOW())`,
                                userID, prizeAmount, fmt.Sprintf("Prize for contest %d (Rank %d)", contestID, rank))
                        
                        if err == nil {
                                // Update user's wallet balance
                                tx.Exec(`
                                        UPDATE wallets 
                                        SET balance = balance + $1, updated_at = NOW()
                                        WHERE user_id = $2`, prizeAmount, userID)
                                
                                *totalPrizesDistributed += prizeAmount
                                *winnersRewarded++
                        }
                }
        }
}