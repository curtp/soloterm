package database

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// MigrationFunc is a function that runs a migration
type MigrationFunc func(*sqlx.DB) error

var migrations []MigrationFunc

// RegisterMigration allows packages to register their migrations
func RegisterMigration(fn MigrationFunc) {
	migrations = append(migrations, fn)
}

// Setup connects to the database and runs all migrations
// Returns a ready-to-use database connection
func Setup() (*sqlx.DB, error) {
	// Connect to database
	db, err := Connect()
	if err != nil {
		return nil, err
	}

	// Run all registered migrations
	if len(migrations) > 0 {
		log.Println("Running migrations...")
		for _, migrate := range migrations {
			if err := migrate(db); err != nil {
				db.Close()
				return nil, err
			}
		}
		log.Println("Migrations complete")
	}

	return db, nil
}
