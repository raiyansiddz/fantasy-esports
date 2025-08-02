package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Initialize(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings with proper timeouts
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour) // Prevent connection timeouts

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createKYCDocumentsTable,
		createGamesTable,
		createTeamsTable,
		createPlayersTable,
		createTournamentsTable,
		createTournamentStagesTable,
		createMatchesTable,
		createMatchParticipantsTable,
		createMatchEventsTable,
		createContestsTable,
		createUserTeamsTable,
		createTeamPlayersTable,
		createContestParticipantsTable,
		createUserWalletsTable,
		createWalletTransactionsTable,
		createPaymentTransactionsTable,
		createReferralsTable,
		createAdminUsersTable,
		createSystemConfigTable,
		insertDefaultConfigs,
		insertSampleData,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}