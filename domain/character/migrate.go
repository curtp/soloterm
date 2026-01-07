package character

import (
	"soloterm/database"
)

func init() {
	// Register this package's migrations with the database package
	database.RegisterMigration(Migrate)
}

// Migrate runs all migrations for the games domain
func Migrate(dbStore *database.DBStore) error {
	// Migration: Create characters table
	if err := createCharactersTable(dbStore); err != nil {
		return err
	}

	// Migration: Create attributes table
	if err := createAttributesTable(dbStore); err != nil {
		return err
	}

	if err := addMissingColumns(dbStore); err != nil {
		return err
	}

	if err := removeExistingColumns(dbStore); err != nil {
		return err
	}

	return nil
}

// createCharactersTable creates the initial characters table and index
func createCharactersTable(dbStore *database.DBStore) error {
	schema := `
		CREATE TABLE IF NOT EXISTS characters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING NOT NULL,
			system STRING NOT NULL,
			role STRING NOT NULL,
			species STRING NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_character_by_name ON characters (name);
	`
	_, err := dbStore.Connection.Exec(schema)
	return err
}

func createAttributesTable(dbStore *database.DBStore) error {
	schema := `
		CREATE TABLE IF NOT EXISTS attributes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id INTEGER NOT NULL,
			name STRING NOT NULL,
			value STRING NOT NULL,
			attribute_group INTEGER NOT NULL DEFAULT 0,
			position_in_group INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_attribute_by_character_id ON attributes (character_id);
	`
	_, err := dbStore.Connection.Exec(schema)
	return err
}

func removeExistingColumns(dbStore *database.DBStore) error {
	return nil
}

func addMissingColumns(dbStore *database.DBStore) error {
	return nil
}
