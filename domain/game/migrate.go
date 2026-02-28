package game

import (
	"soloterm/database"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the games domain
func Migrate(dbStore *database.DBStore) error {
	// Migration: Create games table
	if err := createGamesTable(dbStore); err != nil {
		return err
	}

	if err := addNotesToGamesTable(dbStore); err != nil {
		return err
	}

	return nil
}

// createGamesTable creates the initial games table and index
func createGamesTable(dbStore *database.DBStore) error {
	schema := `
		CREATE TABLE IF NOT EXISTS games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING NOT NULL,
			description TEXT,
			notes TEXT default '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_games_by_name ON games (name);
	`
	_, err := dbStore.Connection.Exec(schema)
	return err
}

func addNotesToGamesTable(dbStore *database.DBStore) error {
	defaultValue := "''"
	return database.AddColumn(dbStore.Connection, "games", "notes", "text", false, &defaultValue)
}
