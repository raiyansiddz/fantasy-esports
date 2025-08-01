package utils

import (
	"regexp"
	"fantasy-esports-backend/models"
)

func ValidateMobile(mobile string) bool {
	// Indian mobile number validation
	matched, _ := regexp.MatchString(`^\+91[6-9]\d{9}$`, mobile)
	return matched
}

func ValidateEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

func ValidateTeamComposition(players []models.PlayerSelection, gameRules models.Game) []string {
	var errors []string
	
	if len(players) != gameRules.TotalTeamSize {
		errors = append(errors, "Invalid team size")
	}
	
	captainCount := 0
	viceCaptainCount := 0
	teamCount := make(map[int64]int)
	
	for _, player := range players {
		if player.IsCaptain {
			captainCount++
		}
		if player.IsViceCaptain {
			viceCaptainCount++
		}
		
		// Note: In real implementation, you'd fetch player details from DB
		// For now, we'll skip credit validation
	}
	
	if captainCount != 1 {
		errors = append(errors, "Must have exactly one captain")
	}
	if viceCaptainCount != 1 {
		errors = append(errors, "Must have exactly one vice captain")
	}
	
	// Validate max players per team (from team composition rules)
	for _, count := range teamCount {
		if count > gameRules.MaxPlayersPerTeam {
			errors = append(errors, "Too many players from same team")
			break
		}
	}
	
	return errors
}

func CalculateFantasyPoints(events []models.MatchEvent, isCaptain, isViceCaptain bool) float64 {
	var totalPoints float64
	
	for _, event := range events {
		totalPoints += event.Points
	}
	
	// Apply multipliers
	if isCaptain {
		totalPoints *= 2.0
	} else if isViceCaptain {
		totalPoints *= 1.5
	}
	
	return totalPoints
}

func ValidateContestEntry(userID int64, contestID int64, teamID int64) []string {
	var errors []string
	
	// Note: In real implementation, you'd check:
	// - Contest is not full
	// - User hasn't exceeded entry limit
	// - Team belongs to user
	// - Match hasn't started
	// - User has sufficient balance
	
	return errors
}