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

	// Check contests table schema
	fmt.Println("=== CONTESTS TABLE SCHEMA ===")
	rows, err := database.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'contests' 
		ORDER BY ordinal_position`)
	
	if err != nil {
		fmt.Printf("Error querying schema: %v\n", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var columnName, dataType, isNullable string
		if err := rows.Scan(&columnName, &dataType, &isNullable); err == nil {
			fmt.Printf("Column: %s, Type: %s, Nullable: %s\n", columnName, dataType, isNullable)
		}
	}

	// Check if contests have the expected columns
	fmt.Println("\n=== SAMPLE CONTEST DATA ===")
	contestRows, err := database.Query(`
		SELECT id, name, prize_pool, winner_percentage, runner_up_percentage 
		FROM contests 
		WHERE match_id = 3 
		LIMIT 3`)
	
	if err != nil {
		fmt.Printf("Error querying contests: %v\n", err)
		return
	}
	defer contestRows.Close()
	
	for contestRows.Next() {
		var id int64
		var name string
		var prizePool, winnerPct, runnerUpPct interface{}
		
		if err := contestRows.Scan(&id, &name, &prizePool, &winnerPct, &runnerUpPct); err != nil {
			fmt.Printf("Error scanning contest: %v\n", err)
			continue
		}
		
		fmt.Printf("Contest %d: %s, Prize: %v, Winner%%: %v, RunnerUp%%: %v\n", 
			id, name, prizePool, winnerPct, runnerUpPct)
	}
}