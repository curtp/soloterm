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
		// INSERT - new character
		return r.insert(character)
	} else {
		// UPDATE - existing character
		return r.update(character)
	}
}

func (r *Repository) SaveAttribute(attribute *Attribute) error {
	if attribute.ID == 0 {
		// INSERT - new attribute
		return r.insertAttribute(attribute)
	} else {
		// UPDATE - existing attribute
		return r.updateAttribute(attribute)
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

// Delete removes an attribute by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) DeleteAttribute(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM attributes WHERE id = ?`

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

// GetByID retrieves an attribute by ID
func (r *Repository) GetAttributeByID(id int64) (*Attribute, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var attribute Attribute
	err := r.db.Connection.Get(&attribute, "SELECT * FROM attributes WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("attribute not found")
		}
		return nil, err
	}

	return &attribute, nil
}

func (r *Repository) GetAttributesForCharacter(character_id int64) ([]*Attribute, error) {
	var attributes []*Attribute
	query := `
		SELECT *,
			COUNT(*) OVER (PARTITION BY character_id, attribute_group) as group_count,
			SUM(CASE WHEN position_in_group > 0 THEN 1 ELSE 0 END)
				OVER (PARTITION BY character_id, attribute_group) as group_count_after_zero
		FROM attributes
		WHERE character_id = ?
		ORDER BY attribute_group, position_in_group, created_at
	`
	err := r.db.Connection.Select(&attributes, query, character_id)
	if err != nil {
		return nil, err
	}
	return attributes, nil
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

// Inserts a new record
func (r *Repository) insert(character *Character) error {
	query := `
		INSERT INTO characters (name, system, role, species, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into character
	err := r.db.Connection.QueryRowx(query,
		character.Name,
		character.System,
		character.Role,
		character.Species,
	).Scan(&character.ID, &character.CreatedAt, &character.UpdatedAt)

	return err
}

// Updates an existing record
func (r *Repository) update(character *Character) error {
	query := `
		UPDATE characters SET name = ?, system = ?, role = ?, species = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	// Execute and scan the returned values back into character
	err := r.db.Connection.QueryRowx(query,
		character.Name,
		character.System,
		character.Role,
		character.Species,
		character.ID,
	).Scan(&character.CreatedAt, &character.UpdatedAt)

	return err
}

// Inserts a new record
func (r *Repository) insertAttribute(attribute *Attribute) error {
	query := `
		INSERT INTO attributes (character_id, attribute_group, position_in_group, name, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	// Execute and scan the returned values back into character
	err := r.db.Connection.QueryRowx(query,
		attribute.CharacterID,
		attribute.Group,
		attribute.PositionInGroup,
		attribute.Name,
		attribute.Value,
	).Scan(&attribute.ID, &attribute.CreatedAt, &attribute.UpdatedAt)

	return err
}

// Updates an existing record
func (r *Repository) updateAttribute(attribute *Attribute) error {
	query := `
		UPDATE attributes SET attribute_group = ?, position_in_group = ?, name = ?, value = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	// Execute and scan the returned values back into character
	err := r.db.Connection.QueryRowx(query,
		attribute.Group,
		attribute.PositionInGroup,
		attribute.Name,
		attribute.Value,
		attribute.ID,
	).Scan(&attribute.CreatedAt, &attribute.UpdatedAt)

	return err
}
