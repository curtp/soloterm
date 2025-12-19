package log

import (
	"soloterm/database"

	"github.com/jmoiron/sqlx"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the logs domain
func Migrate(db *sqlx.DB) error {
	// Migration: Create log table
	if err := createTable(db); err != nil {
		return err
	}

	return nil
}

// createTable creates the initial logs table and index
func createTable(db *sqlx.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			log_type STRING NOT NULL,
			description TEXT NOT NULL,
			result STRING NOT NULL,
			narrative STRING NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_logs_by_game_id ON logs (game_id);
	`
	_, err := db.Exec(schema)
	return err
}
