package main

import (
	"fantasy-esports-backend/config"
	"fantasy-esports-backend/db"
	"fmt"
	"log"
)

func main() {
	cfg := config.Load()
	database, err := db.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Check contests for matches 1 and 3
	for _, matchID := range []int{1, 3} {
		fmt.Printf("=== CONTESTS FOR MATCH %d ===\n", matchID)
		rows, err := database.Query(`
			SELECT c.id, c.name, c.current_participants, c.max_participants, c.total_prize_pool
			FROM contests c 
			WHERE c.match_id = $1`, matchID)
		
		if err != nil {
			fmt.Printf("Error querying contests: %v\n", err)
			continue
		}
		defer rows.Close()
		
		contestCount := 0
		for rows.Next() {
			var contestID int64
			var contestName string 
			var currentParticipants, maxParticipants int
			var prizePool float64
			
			if err := rows.Scan(&contestID, &contestName, &currentParticipants, &maxParticipants, &prizePool); err == nil {
				fmt.Printf("Contest %d: %s (%d/%d participants, $%.2f prize pool)\n", 
					contestID, contestName, currentParticipants, maxParticipants, prizePool)
				contestCount++
			}
		}
		if contestCount == 0 {
			fmt.Printf("No contests found for match %d\n", matchID)
		}

		// Check fantasy teams for this match
		fmt.Printf("\n=== FANTASY TEAMS FOR MATCH %d ===\n", matchID)
		rows2, err := database.Query(`
			SELECT ut.id, ut.team_name, ut.total_points, u.mobile
			FROM user_teams ut
			JOIN users u ON ut.user_id = u.id
			WHERE ut.match_id = $1`, matchID)
		
		if err != nil {
			fmt.Printf("Error querying fantasy teams: %v\n", err)
			continue
		}
		defer rows2.Close()
		
		teamCount := 0
		for rows2.Next() {
			var teamID int64
			var teamName string
			var totalPoints float64
			var userMobile string
			
			if err := rows2.Scan(&teamID, &teamName, &totalPoints, &userMobile); err == nil {
				fmt.Printf("Team %d: %s by %s (%.2f points)\n", teamID, teamName, userMobile, totalPoints)
				teamCount++
			}
		}
		if teamCount == 0 {
			fmt.Printf("No fantasy teams found for match %d\n", matchID)
		}
		fmt.Println()
	}
}
