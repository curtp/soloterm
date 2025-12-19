package database

import (
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Connect establishes a connection to the SQLite database
// Returns a sqlx.DB connection or an error
func Connect() (*sqlx.DB, error) {
	return ConnectWithPath(getDBPath())
}

// ConnectWithPath establishes a connection to the SQLite database at the specified path
// Use ":memory:" for in-memory databases (useful for testing)
func ConnectWithPath(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrency
	// Skip for in-memory databases as WAL doesn't apply
	if dbPath != ":memory:" {
		_, err = db.Exec("PRAGMA journal_mode=WAL")
		if err != nil {
			return nil, err
		}

		_, err = db.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			return nil, err
		}

	}

	return db, nil
}

// getDBPath returns the path to the database file
// Checks environment variable first, then falls back to ./data/kvstore.db
func getDBPath() string {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath
	}

	os.MkdirAll("./data", 0755)
	return filepath.Join("data", "soloterm.db")
}
