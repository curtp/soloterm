package oracle

import (
	"soloterm/database"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the oracle domain
func Migrate(dbStore *database.DBStore) error {
	// Migration: Create oracles table
	if err := createOraclesTable(dbStore); err != nil {
		return err
	}
	return nil
}

// createOraclesTable creates the initial oracles table and index
func createOraclesTable(dbStore *database.DBStore) error {
	schema := `
		CREATE TABLE IF NOT EXISTS oracles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category STRING NOT NULL,
			name STRING NOT NULL,
			content TEXT default '',
			category_position INTEGER NOT NULL default 0,
			position_in_category INTEGER NOT NULL default 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_oracles_by_name ON oracles (name);
	`
	_, err := dbStore.Connection.Exec(schema)
	return err
}
