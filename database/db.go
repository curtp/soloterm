// Package database provides database connection management and migration support.
// It handles SQLite database initialization, connection pooling, and schema migrations.
package database

import (
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

// RegisterMigration allows packages to register their migrations
func RegisterMigration(fn MigrationFunc) {
	migrations = append(migrations, fn)
}

// Setup connects to the database and runs all migrations.
// Use ":memory:" for in-memory databases (useful for testing).
// Returns a ready-to-use DBStore.
// * Use dbStore.Connection to interact with the database
// * Use dbStore.Path to see what path is being used
func Setup(path string) (*DBStore, error) {
	// Connect to database
	dbStore, err := connectWithPath(path)
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

// connectWithPath establishes a connection to the SQLite database at the specified path
// Use ":memory:" for in-memory databases (useful for testing)
func connectWithPath(dbPath string) (*DBStore, error) {
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
	}

	// Always enforce foreign keys, including in-memory test databases
	_, err = dbStore.Connection.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	return &dbStore, nil
}

// ResolveDBPath returns the database file path using the following precedence:
// DB_PATH env var → configuredDir (database_dir in config.yaml) → defaultDir (platform data dir)
func ResolveDBPath(configuredDir, defaultDir string) string {
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}
	if configuredDir != "" {
		return filepath.Join(configuredDir, "soloterm.db")
	}
	return filepath.Join(defaultDir, "soloterm.db")
}
