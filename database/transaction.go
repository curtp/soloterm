package database

import "github.com/jmoiron/sqlx"

// WithTx executes a function within a database transaction.
// Automatically commits on success or rolls back on error.
//
// Example usage:
//
//	err := database.WithTx(r.db, func(tx *sqlx.Tx) error {
//	    // Your transactional operations here
//	    _, err := tx.Exec("UPDATE ...")
//	    return err
//	})
func WithTx(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Execute the function
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
