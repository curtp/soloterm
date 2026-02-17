package character

import (
	"database/sql"
	"errors"
	"fmt"
	"soloterm/database"
)

// Repository handles database operations for characters
type Repository struct {
	db *database.DBStore
}

// NewRepository creates a new Repository
func NewRepository(db *database.DBStore) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a character
// Automatically manages created_at, and updated_at
// The character pointer is updated with the current values after save
func (r *Repository) Save(character *Character) error {
	if character.ID == 0 {
		return r.insert(character)
	} else {
		return r.update(character)
	}
}

// Delete removes a character by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) Delete(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM characters WHERE id = ?`

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

// GetByID retrieves a character by ID
func (r *Repository) GetByID(id int64) (*Character, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var character Character
	err := r.db.Connection.Get(&character, "SELECT * FROM characters WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("character not found")
		}
		return nil, err
	}

	return &character, nil
}

// GetAll retrieves all characters ordered by name
func (r *Repository) GetAll() ([]*Character, error) {
	var characters []*Character
	err := r.db.Connection.Select(&characters, "SELECT * FROM characters ORDER BY lower(system), lower(name) ASC")
	if err != nil {
		return nil, err
	}
	return characters, nil
}

// insert inserts a new character record
func (r *Repository) insert(character *Character) error {
	query := `
		INSERT INTO characters (name, system, role, species, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	err := r.db.Connection.QueryRowx(query,
		character.Name,
		character.System,
		character.Role,
		character.Species,
	).StructScan(character)

	return err
}

// update updates an existing character record
func (r *Repository) update(character *Character) error {
	query := `
		UPDATE characters SET name = ?, system = ?, role = ?, species = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	err := r.db.Connection.QueryRowx(query,
		character.Name,
		character.System,
		character.Role,
		character.Species,
		character.ID,
	).StructScan(character)

	return err
}
