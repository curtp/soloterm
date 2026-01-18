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

// RegisterMigration allows packages to register their migrations
func RegisterMigration(fn MigrationFunc) {
	migrations = append(migrations, fn)
}

// Setup connects to the database and runs all migrations
// Use ":memory:" for in-memory databases (useful for testing)
// Provide a path to the file to connect to.
// Leave path nil to use the DB_PATH environment variable
// If DB_PATH isn't found, it will use the default directory.
// Returns a ready-to-use DBStore.
// * Use dbStore.Connection to interact with the database
// * Use dbStore.Path to see what path is being used
func Setup(path *string) (*DBStore, error) {

	// If path isn't provided, try to get the path from either the
	// environment or the default location
	if path == nil {
		dbPath, err := getDBPath()
		if err != nil {
			return nil, err
		}
		path = &dbPath
	}

	// Connect to database
	dbStore, err := connectWithPath(*path)
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

	workDir := os.Getenv("SOLOTERM_WORK_DIR")

	err := os.MkdirAll(workDir, 0755)
	if err != nil {
		log.Printf("Error creating path for database: %s", err)
		return "", err
	}

	return filepath.Join(workDir, "soloterm.db"), nil
}
