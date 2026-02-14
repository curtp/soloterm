package session

import (
	"soloterm/database"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the logs domain
func Migrate(db *database.DBStore) error {
	// Migration: Create log table
	if err := createTable(db); err != nil {
		return err
	}

	return nil
}

// createTable creates the initial logs table and index
func createTable(db *database.DBStore) error {
	schema := `
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			name STRING NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_by_game_id ON sessions (game_id);
	`
	_, err := db.Connection.Exec(schema)
	return err
}
