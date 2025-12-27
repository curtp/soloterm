// Package database provides database connection management and migration support.
// It handles SQLite database initialization, connection pooling, and schema migrations.
package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type DBStore struct {
	Connection *sqlx.DB
	Path       *string
}

// MigrationFunc is a function that runs a migration
type MigrationFunc func(*DBStore) error

var migrations []MigrationFunc

// Connect establishes a connection to the SQLite database
// Returns a DBStore or an error
func Connect() (*DBStore, error) {
	path, err := getDBPath()
	if err != nil {
		return nil, err
	}

	log.Printf("Connecting to path: %s", path)

	return ConnectWithPath(path)
}

// ConnectWithPath establishes a connection to the SQLite database at the specified path
// Use ":memory:" for in-memory databases (useful for testing)
func ConnectWithPath(dbPath string) (*DBStore, error) {
	dbStore := DBStore{}

	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	dbStore.Connection = db
	dbStore.Path = &dbPath

	// Enable WAL mode for better concurrency
	// Skip for in-memory databases as WAL doesn't apply
	if dbPath != ":memory:" {
		_, err = dbStore.Connection.Exec("PRAGMA journal_mode=WAL")
		if err != nil {
			return nil, err
		}

		_, err = dbStore.Connection.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			return nil, err
		}

	}

	return &dbStore, nil
}

// getDBPath returns the path to the database file
// Checks environment variable first, then falls back to ./data/kvstore.db
func getDBPath() (string, error) {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath, nil
	}

	home, _ := os.UserHomeDir()
	path := home + "/soloterm"

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Error creating path for database: %s", err)
		return "", err
	}

	return filepath.Join(path, "soloterm.db"), nil
}

// RegisterMigration allows packages to register their migrations
func RegisterMigration(fn MigrationFunc) {
	migrations = append(migrations, fn)
}

// Setup connects to the database and runs all migrations
// Returns a ready-to-use database connection
func Setup() (*DBStore, error) {
	// Connect to database
	dbStore, err := Connect()
	if err != nil {
		return nil, err
	}

	// Run all registered migrations
	if len(migrations) > 0 {
		for _, migrate := range migrations {
			if err := migrate(dbStore); err != nil {
				dbStore.Connection.Close()
				return nil, err
			}
		}
	}

	return dbStore, nil
}
