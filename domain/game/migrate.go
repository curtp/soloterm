package game

import (
	"soloterm/database"

	"github.com/jmoiron/sqlx"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the games domain
func Migrate(db *sqlx.DB) error {
	// Migration: Create games table
	if err := createGamesTable(db); err != nil {
		return err
	}

	return nil
}

// createGamesTable creates the initial games table and index
func createGamesTable(db *sqlx.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_games_by_name ON games (name);
	`
	_, err := db.Exec(schema)
	return err
}
