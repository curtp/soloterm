package database

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func AddColumn(db *sqlx.DB, tableName string, column string, columnType string, notnull bool, defaultValue *string) error {
	exists, err := columnExists(db, tableName, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, column, columnType)

	// Prevent Nulls
	if notnull {
		query += " NOT NULL"
	}

	// Set a default value
	if defaultValue != nil {
		query += fmt.Sprintf(" DEFAULT %s", *defaultValue)
	}

	// Execute the query
	_, err = db.Exec(query)
	return err
}

func RemoveColumn(db *sqlx.DB, tableName string, column string) error {
	exists, err := columnExists(db, tableName, column)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, column)

	// Execute the query
	_, err = db.Exec(query)
	return err
}

// columnExists checks if a column exists in a table
func columnExists(db *sqlx.DB, table string, column string) (bool, error) {
	var exists bool

	// Note: pragma_table_info requires string interpolation, not ? placeholders
	// The column name uses ? placeholder for the WHERE clause
	query := fmt.Sprintf("SELECT COUNT(*) > 0 FROM pragma_table_info('%s') WHERE name = ?", table)

	err := db.Get(&exists, query, column)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}
