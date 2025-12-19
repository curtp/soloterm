package game

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for the K/V store
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new Repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a game
// Automatically manages created_at, and updated_at
// The game pointer is updated with the current values after save
func (r *Repository) Save(game *Game) error {
	if game.Id == 0 {
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

// GetByID retrieves an account by ID
func (r *Repository) GetByID(id int64) (*Game, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var game Game
	err := r.db.Get(&game, "SELECT * FROM games WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("gamd not found")
		}
		return nil, err
	}

	return &game, nil
}

// GetAll retrieves all games ordered by name
func (r *Repository) GetAll() ([]*Game, error) {
	var games []*Game
	err := r.db.Select(&games, "SELECT * FROM games ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	return games, nil
}

// Inserts a enw record
func (r *Repository) insert(game *Game) error {
	query := `
		INSERT INTO games (name, description, created_at, updated_at)
		VALUES (?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into account
	err := r.db.QueryRowx(query,
		game.Name,
		game.Description,
	).Scan(&game.Id, &game.CreatedAt, &game.UpdatedAt)

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
	err := r.db.QueryRowx(query,
		game.Name,
		game.Description,
		game.Id,
	).Scan(&game.CreatedAt, &game.UpdatedAt)

	return err
}
