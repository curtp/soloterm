package character

import (
	"database/sql"
	"errors"
	"fmt"
	"soloterm/database"
)

// AttributeRepository handles database operations for attributes
type AttributeRepository struct {
	db *database.DBStore
}

// NewAttributeRepository creates a new AttributeRepository
func NewAttributeRepository(db *database.DBStore) *AttributeRepository {
	return &AttributeRepository{db: db}
}

// Save creates or updates an attribute
func (r *AttributeRepository) Save(attribute *Attribute) error {
	if attribute.ID == 0 {
		return r.insert(attribute)
	} else {
		return r.update(attribute)
	}
}

// Delete removes an attribute by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *AttributeRepository) Delete(id int64) (int64, error) {
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

// GetByID retrieves an attribute by ID
func (r *AttributeRepository) GetByID(id int64) (*Attribute, error) {
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

// GetForCharacter retrieves all attributes for a character
func (r *AttributeRepository) GetForCharacter(character_id int64) ([]*Attribute, error) {
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

// Swaps two groups position
func (r *AttributeRepository) SwapGroups(characterID int64, groupA, groupB int) error {
	query := `UPDATE attributes
              SET attribute_group = CASE
                  WHEN attribute_group = ? THEN ?
                  WHEN attribute_group = ? THEN ?
              END,
              updated_at = datetime('now','subsec')
              WHERE attribute_group IN (?, ?) AND character_id = ?`
	_, err := r.db.Connection.Exec(query, groupA, groupB, groupB, groupA, groupA, groupB, characterID)
	return err
}

// Moves an attribute to a new position
func (r *AttributeRepository) UpdatePosition(id int64, group, position int) error {
	query := `UPDATE attributes SET attribute_group = ?, position_in_group = ?, updated_at = datetime('now','subsec') WHERE id = ?`
	_, err := r.db.Connection.Exec(query, group, position, id)
	return err
}

// insert inserts a new attribute record
func (r *AttributeRepository) insert(attribute *Attribute) error {
	query := `
		INSERT INTO attributes (character_id, attribute_group, position_in_group, name, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	err := r.db.Connection.QueryRowx(query,
		attribute.CharacterID,
		attribute.Group,
		attribute.PositionInGroup,
		attribute.Name,
		attribute.Value,
	).StructScan(attribute)

	return err
}

// update updates an existing attribute record
func (r *AttributeRepository) update(attribute *Attribute) error {
	query := `
		UPDATE attributes SET attribute_group = ?, position_in_group = ?, name = ?, value = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	err := r.db.Connection.QueryRowx(query,
		attribute.Group,
		attribute.PositionInGroup,
		attribute.Name,
		attribute.Value,
		attribute.ID,
	).StructScan(attribute)

	return err
}
