package log

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for logs
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new Repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a log
// Automatically manages created_at, and updated_at
// The log pointer is updated with the current values after save
func (r *Repository) Save(log *Log) error {
	if log.ID == 0 {
		// INSERT - new game
		return r.insert(log)
	} else {
		// UPDATE - existing log
		return r.update(log)
	}
}

// Delete removes a log by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) Delete(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM logs WHERE id = ?`

	result, err := r.db.Exec(query, id)

	if err != nil {
		return 0, err
	}

	// Check if a row was actually deleted
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, fmt.Errorf("id '%d' not found", id)
	}

	return rows, nil
}

func (r *Repository) DeleteAllForGame(gameID int64) (int64, error) {
	if gameID == 0 {
		return 0, errors.New("gameID cannot be empty")
	}

	query := `DELETE FROM logs WHERE game_id = ?`

	result, err := r.db.Exec(query, gameID)

	if err != nil {
		return 0, err
	}

	// Check if a row was actually deleted
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, fmt.Errorf("id '%d' not found", gameID)
	}

	return rows, nil

}

// GetByID retrieves a log by ID
func (r *Repository) GetByID(id int64) (*Log, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var log Log
	err := r.db.Get(&log, "SELECT * FROM logs WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("log not found")
		}
		return nil, err
	}

	return &log, nil
}

// GetAllForGame retrieves all logs for the game ordered by created_at
func (r *Repository) GetAllForGame(gameID int64) ([]*Log, error) {
	var logs []*Log
	err := r.db.Select(&logs, "SELECT * FROM logs WHERE game_id = ? ORDER BY created_at ASC", gameID)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// Session represents a logical grouping of logs by date
type Session struct {
	GameID   int64
	Date     string `db:"session_date"`
	LogCount int    `db:"log_count"`
}

// GetSessionsForGame retrieves all unique session dates for a game
// Sessions are grouped by date (excluding time)
func (r *Repository) GetSessionsForGame(gameID int64) ([]*Session, error) {
	query := `
		SELECT
			DATE(created_at, 'localtime') as session_date,
			COUNT(*) as log_count
		FROM logs
		WHERE game_id = ?
		GROUP BY DATE(created_at, 'localtime')
		ORDER BY session_date ASC
	`

	var sessions []*Session
	err := r.db.Select(&sessions, query, gameID)
	if err != nil {
		return nil, err
	}

	// Populate the GameID for each session
	for _, session := range sessions {
		session.GameID = gameID
	}

	return sessions, nil
}

// GetLogsForSession retrieves all logs for a specific session date
func (r *Repository) GetLogsForSession(gameID int64, sessionDate string) ([]*Log, error) {
	query := `
		SELECT * FROM logs
		WHERE game_id = ?
		AND DATE(created_at, 'localtime') >= ?
		ORDER BY created_at ASC
	`
	var logs []*Log
	err := r.db.Select(&logs, query, gameID, sessionDate)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// Inserts a new record
func (r *Repository) insert(log *Log) error {
	query := `
		INSERT INTO logs (game_id, log_type, description, result, narrative, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into log
	err := r.db.QueryRowx(query,
		log.GameID,
		log.LogType,
		log.Description,
		log.Result,
		log.Narrative,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)

	return err
}

// Updates an existing record
func (r *Repository) update(log *Log) error {
	query := `
		UPDATE logs SET game_id = ?, log_type = ?, description = ?, result = ?, narrative = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	// Execute and scan the returned values back into log
	err := r.db.QueryRowx(query,
		log.GameID,
		log.LogType,
		log.Description,
		log.Result,
		log.Narrative,
		log.ID,
	).Scan(&log.CreatedAt, &log.UpdatedAt)

	return err
}
