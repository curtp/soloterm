package testing

import (
	"soloterm/database"
	"testing"
)

// SetupTestDB creates an in-memory SQLite database for testing.
// Each test gets a fresh, isolated database.
//
// Note: This does NOT run migrations. Each domain should run its own
// Migrate() function after calling SetupTestDB().
//
// Example usage:
//
//	func TestSomething(t *testing.T) {
//	    db := testhelper.SetupTestDB(t)
//	    defer testhelper.TeardownTestDB(t, db)
//
//	    // Run your domain's migration
//	    if err := entry.Migrate(db); err != nil {
//	        t.Fatalf("Failed to migrate: %v", err)
//	    }
//
//	    // Use db for testing...
//	}
func SetupTestDB(t *testing.T) *database.DBStore {
	t.Helper() // Marks this as a test helper for better error reporting

	// Connect to in-memory database
	path := ":memory:"
	db, err := database.Setup(&path)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}

// CreateTestGame inserts a game row and returns its ID.
// Requires the games table to exist (blank-import soloterm/domain/game in your test file).
func CreateTestGame(t *testing.T, db *database.DBStore, name string) int64 {
	t.Helper()
	var id int64
	err := db.Connection.QueryRow(
		`INSERT INTO games (name, description, created_at, updated_at)
		 VALUES (?, '', datetime('now'), datetime('now')) RETURNING id`,
		name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("CreateTestGame: failed to create game %q: %v", name, err)
	}
	return id
}

// CreateTestSession inserts a session row and returns its ID.
// Requires the sessions table to exist (blank-import soloterm/domain/session in your test file).
func CreateTestSession(t *testing.T, db *database.DBStore, gameID int64, name string, content string) int64 {
	t.Helper()
	var id int64
	err := db.Connection.QueryRow(
		`INSERT INTO sessions (game_id, name, content, created_at, updated_at)
		 VALUES (?, ?, ?, datetime('now'), datetime('now')) RETURNING id`,
		gameID, name, content,
	).Scan(&id)
	if err != nil {
		t.Fatalf("CreateTestSession: failed to create session %q: %v", name, err)
	}
	return id
}

// TeardownTestDB closes the database connection.
// Use with defer: defer TeardownTestDB(t, db)
func TeardownTestDB(t *testing.T, db *database.DBStore) {
	t.Helper() // Marks this as a test helper for better error reporting

	if err := db.Connection.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}
