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

// TeardownTestDB closes the database connection.
// Use with defer: defer TeardownTestDB(t, db)
func TeardownTestDB(t *testing.T, db *database.DBStore) {
	t.Helper() // Marks this as a test helper for better error reporting

	if err := db.Connection.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// SetupTestDBWithMigration creates an in-memory SQLite database and runs the provided migration function.
// This is a convenience wrapper around SetupTestDB() for domain-specific test setup.
//
// Example usage:
//
//	func setupTestDB(t *testing.T) *sqlx.DB {
//	    return testhelper.SetupTestDBWithMigration(t, Migrate)
//	}
func SetupTestDBWithMigration(t *testing.T, migrateFn func(*database.DBStore) error) *database.DBStore {
	t.Helper()

	db := SetupTestDB(t)

	if err := migrateFn(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}
