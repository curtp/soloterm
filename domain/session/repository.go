package session

import (
	"database/sql"
	"errors"
	"fmt"
	"soloterm/database"
)

// Repository handles database operations for sessions
type Repository struct {
	db *database.DBStore
}

// NewRepository creates a new Repository
func NewRepository(db *database.DBStore) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a session
// Automatically manages created_at, and updated_at
// The session pointer is updated with the current values after save
func (r *Repository) Save(session *Session) error {
	if session.ID == 0 {
		// INSERT - new session
		return r.insert(session)
	} else {
		// UPDATE - existing session
		return r.update(session)
	}
}

// Delete removes a session by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) Delete(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM sessions WHERE id = ?`

	result, err := r.db.Connection.Exec(query, id)

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

	query := `DELETE FROM sessions WHERE game_id = ?`

	result, err := r.db.Connection.Exec(query, gameID)

	if err != nil {
		return 0, err
	}

	// Check if a row was actually deleted
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rows == 0 {
		return 0, fmt.Errorf("No sessions for game '%d' found", gameID)
	}

	return rows, nil

}

// GetByID retrieves a session by ID
func (r *Repository) GetByID(id int64) (*Session, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var session Session
	query := `SELECT s.*, g.name AS game_name FROM sessions s
		JOIN games g ON s.game_id = g.id
		WHERE s.id = ?`
	err := r.db.Connection.Get(&session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}

// GetAllForGame retrieves all sessions for the game ordered by created_at. Excludes content for performance reasons
func (r *Repository) GetAllForGame(gameID int64) ([]*Session, error) {
	var sessions []*Session
	query := `SELECT s.id, s.game_id, s.name, s.created_at, s.updated_at, g.name AS game_name
		FROM sessions s
		JOIN games g ON s.game_id = g.id
		WHERE s.game_id = ? ORDER BY s.created_at ASC`
	err := r.db.Connection.Select(&sessions, query, gameID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetAllContentForGame retrieves content from all sessions for a game
func (r *Repository) GetAllContentForGame(gameID int64) ([]string, error) {
	var contents []string
	err := r.db.Connection.Select(&contents, "SELECT content FROM sessions WHERE game_id = ?", gameID)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *Repository) SearchByGame(gameID int64, term string) ([]*Session, error) {
	var sessions []*Session
	query := `
		SELECT s.*, g.name AS game_name FROM sessions s
		JOIN games g ON s.game_id = g.id
		WHERE s.game_id = ? AND LOWER(s.content) LIKE LOWER('%' || ? || '%')
		ORDER BY s.created_at
	`
	err := r.db.Connection.Select(&sessions, query, gameID, term)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Inserts a new record
func (r *Repository) insert(session *Session) error {
	query := `
		INSERT INTO sessions (game_id, name, content, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into session
	err := r.db.Connection.QueryRowx(query,
		session.GameID,
		session.Name,
		session.Content,
	).StructScan(session)

	return err
}

// Updates an existing record
func (r *Repository) update(session *Session) error {
	query := `
		UPDATE sessions SET game_id = ?, name = ?, content = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	// Execute and scan the returned values back into session
	err := r.db.Connection.QueryRowx(query,
		session.GameID,
		session.Name,
		session.Content,
		session.ID,
	).StructScan(session)

	return err
}
