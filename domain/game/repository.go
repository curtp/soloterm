package game

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"soloterm/database"
)

// Repository handles database operations for games
type Repository struct {
	db *database.DBStore
}

// NewRepository creates a new Repository
func NewRepository(db *database.DBStore) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a game
// Automatically manages created_at, and updated_at
// The game pointer is updated with the current values after save
func (r *Repository) Save(game *Game) error {
	log.Printf("Saving game: %+v", game)
	if game.ID == 0 {
		// INSERT - new game
		return r.insert(game)
	} else {
		// UPDATE - existing game
		return r.update(game)
	}
}

// Delete removes a game by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) Delete(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM games WHERE id = ?`

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

// GetByID retrieves a game by ID
func (r *Repository) GetByID(id int64) (*Game, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var game Game
	err := r.db.Connection.Get(&game, "SELECT * FROM games WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("game not found")
		}
		return nil, err
	}

	return &game, nil
}

// GetAll retrieves all games ordered by name
func (r *Repository) GetAll() ([]*Game, error) {
	var games []*Game
	err := r.db.Connection.Select(&games, "SELECT * FROM games ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (r *Repository) SaveNotes(gameID int64, notes string) error {
	log.Printf("Setting notes on game %d", gameID)
	query := `UPDATE games SET notes = ? WHERE id = ?`

	result, err := r.db.Connection.Exec(query, notes, gameID)

	if err != nil {
		return err
	}

	// Check if the game was updated
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("id '%d' not found", gameID)
	}

	return nil
}

// Inserts a new record
func (r *Repository) insert(game *Game) error {
	query := `
		INSERT INTO games (name, description, created_at, updated_at)
		VALUES (?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into game
	err := r.db.Connection.QueryRowx(query,
		game.Name,
		game.Description,
	).StructScan(game)

	return err
}

// Updates an existing record
func (r *Repository) update(game *Game) error {
	query := `
		UPDATE games SET name = ?, description = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	// Execute and scan the returned values back into game
	err := r.db.Connection.QueryRowx(query,
		game.Name,
		game.Description,
		game.ID,
	).StructScan(game)

	return err
}
