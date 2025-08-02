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

	// Check match 1 details
	fmt.Println("=== MATCH 1 DETAILS ===")
	var id, gameID, bestOf int64
	var status, matchType, name string
	err = database.QueryRow(`
		SELECT id, name, status, match_type, game_id, best_of
		FROM matches WHERE id = 1`).Scan(&id, &name, &status, &matchType, &gameID, &bestOf)
	
	if err != nil {
		fmt.Printf("Match 1 not found: %v\n", err)
	} else {
		fmt.Printf("Match 1: ID=%d, Name=%s, Status=%s, Type=%s, Game=%d, BestOf=%d\n", id, name, status, matchType, gameID, bestOf)
	}

	// Check match participants for match 1
	fmt.Println("\n=== MATCH 1 PARTICIPANTS ===")
	rows, err := database.Query(`
		SELECT mp.team_id, t.name, mp.team_score 
		FROM match_participants mp 
		JOIN teams t ON mp.team_id = t.id 
		WHERE mp.match_id = 1`)
	
	if err != nil {
		fmt.Printf("Error querying participants: %v\n", err)
	} else {
		defer rows.Close()
		participantCount := 0
		for rows.Next() {
			var teamID int64
			var teamName string
			var teamScore int
			if err := rows.Scan(&teamID, &teamName, &teamScore); err == nil {
				fmt.Printf("Team %d: %s (Score: %d)\n", teamID, teamName, teamScore)
				participantCount++
			}
		}
		if participantCount == 0 {
			fmt.Println("No participants found for match 1")
		}
	}

	// Check if there are any matches with participants
	fmt.Println("\n=== MATCHES WITH PARTICIPANTS ===")
	rows2, err := database.Query(`
		SELECT m.id, m.name, COUNT(mp.team_id) as participant_count
		FROM matches m
		LEFT JOIN match_participants mp ON m.id = mp.match_id
		GROUP BY m.id, m.name
		ORDER BY m.id
		LIMIT 5`)
	
	if err != nil {
		fmt.Printf("Error querying matches: %v\n", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var matchID int64
			var matchName string
			var count int
			if err := rows2.Scan(&matchID, &matchName, &count); err == nil {
				fmt.Printf("Match %d: %s (%d participants)\n", matchID, matchName, count)
			}
		}
	}
}
