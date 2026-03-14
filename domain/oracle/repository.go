package oracle

import (
	"database/sql"
	"errors"
	"fmt"
	"soloterm/database"
	"strings"
)

// Repository handles database operations for oracles
type Repository struct {
	db *database.DBStore
}

// NewRepository creates a new Repository
func NewRepository(db *database.DBStore) *Repository {
	return &Repository{db: db}
}

// Save creates or updates a oracle
// Automatically manages created_at, and updated_at
// The oracle pointer is updated with the current values after save
func (r *Repository) Save(oracle *Oracle) error {
	if oracle.ID == 0 {
		// INSERT - new oracle
		return r.insert(oracle)
	} else {
		// UPDATE - existing oracle
		return r.update(oracle)
	}
}

// Delete removes a oracle by id
// Returns the number of rows deleted and an error if the id doesn't exist
func (r *Repository) Delete(id int64) (int64, error) {
	if id == 0 {
		return 0, errors.New("id cannot be empty")
	}

	query := `DELETE FROM oracles WHERE id = ?`

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

// GetByID retrieves a oracle by ID
func (r *Repository) GetByID(id int64) (*Oracle, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var oracle Oracle
	err := r.db.Connection.Get(&oracle, "SELECT * FROM oracles WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("oracle not found")
		}
		return nil, err
	}

	return &oracle, nil
}

// GetAll retrieves all oracles ordered by category_position then position_in_category
func (r *Repository) GetAll() ([]*Oracle, error) {
	var oracles []*Oracle
	err := r.db.Connection.Select(&oracles, "SELECT * FROM oracles ORDER BY category_position, position_in_category")
	if err != nil {
		return nil, err
	}
	return oracles, nil
}

// GetCategoryInfo returns one row per distinct category with its sort position and max item position.
func (r *Repository) GetCategoryInfo() ([]*CategoryInfo, error) {
	var info []*CategoryInfo
	err := r.db.Connection.Select(&info, `
		SELECT category AS name, category_position, MAX(position_in_category) AS max_position_in_category
		FROM oracles
		GROUP BY category, category_position
		ORDER BY category_position
	`)
	return info, err
}

func (r *Repository) GetByName(name string) (*Oracle, error) {
	var oracle Oracle
	err := r.db.Connection.Get(&oracle, "SELECT * FROM oracles WHERE lower(name) = lower(?) LIMIT 1", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oracle named '%s' not found.", name)
		}
		return nil, err
	}
	return &oracle, nil
}

// GetAllByName returns all oracles whose name matches case-insensitively, ordered by sort position.
func (r *Repository) GetAllByName(name string) ([]*Oracle, error) {
	var oracles []*Oracle
	err := r.db.Connection.Select(&oracles,
		"SELECT * FROM oracles WHERE lower(name) = lower(?) ORDER BY category_position, position_in_category",
		name)
	return oracles, err
}

// GetByCategoryAndName returns the oracle matching both category and name, case-insensitively.
func (r *Repository) GetByCategoryAndName(category, name string) (*Oracle, error) {
	var oracle Oracle
	err := r.db.Connection.Get(&oracle,
		"SELECT * FROM oracles WHERE lower(category) = lower(?) AND lower(name) = lower(?)",
		category, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oracle '%s/%s' not found", category, name)
		}
		return nil, err
	}
	return &oracle, nil
}

// FindByPrefix returns oracles for the table hint display. The prefix is the
// text the user has typed after "@":
//   - "fan"          → category OR name starts with "fan" (case-insensitive)
//   - "fantasy/desc" → category starts with "fantasy" AND name starts with "desc"
//
// Results are capped at 20 and ordered by sort position.
func (r *Repository) FindByPrefix(prefix string) ([]*Oracle, error) {
	var oracles []*Oracle
	var err error
	if idx := strings.Index(prefix, "/"); idx != -1 {
		catPfx := strings.ToLower(prefix[:idx]) + "%"
		namePfx := strings.ToLower(prefix[idx+1:]) + "%"
		err = r.db.Connection.Select(&oracles,
			`SELECT * FROM oracles WHERE lower(category) LIKE ? AND lower(name) LIKE ?
			 ORDER BY category_position, position_in_category LIMIT 20`,
			catPfx, namePfx)
	} else {
		pfx := strings.ToLower(prefix) + "%"
		err = r.db.Connection.Select(&oracles,
			`SELECT * FROM oracles WHERE lower(category) LIKE ? OR lower(name) LIKE ?
			 ORDER BY category_position, position_in_category LIMIT 20`,
			pfx, pfx)
	}
	return oracles, err
}

func (r *Repository) SaveContent(oracleID int64, content string) error {
	query := `UPDATE oracles SET content = ? WHERE id = ?`

	result, err := r.db.Connection.Exec(query, content, oracleID)

	if err != nil {
		return err
	}

	// Check if the oracle was updated
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("id '%d' not found", oracleID)
	}

	return nil
}

// Inserts a new record using positions already set on the oracle struct
func (r *Repository) insert(oracle *Oracle) error {
	query := `
		INSERT INTO oracles (category, name, content, category_position, position_in_category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'subsec'), datetime('now', 'subsec'))
		RETURNING id, created_at, updated_at
	`

	return r.db.Connection.QueryRowx(query,
		oracle.Category,
		oracle.Name,
		oracle.Content,
		oracle.CategoryPosition,
		oracle.PositionInCategory,
	).StructScan(oracle)
}

// Updates an existing record including sort positions
func (r *Repository) update(oracle *Oracle) error {
	query := `
		UPDATE oracles SET category = ?, name = ?, category_position = ?, position_in_category = ?, updated_at = datetime('now','subsec')
		WHERE id = ?
		RETURNING created_at, updated_at
	`

	return r.db.Connection.QueryRowx(query,
		oracle.Category,
		oracle.Name,
		oracle.CategoryPosition,
		oracle.PositionInCategory,
		oracle.ID,
	).StructScan(oracle)
}

// SwapCategoryPositions atomically moves all oracles in two categories to each other's position
func (r *Repository) SwapCategoryPositions(catA, catB string, posA, posB int) error {
	query := `UPDATE oracles
	          SET category_position = CASE
	              WHEN category = ? THEN ?
	              WHEN category = ? THEN ?
	          END,
	          updated_at = datetime('now','subsec')
	          WHERE category IN (?, ?)`
	_, err := r.db.Connection.Exec(query, catA, posB, catB, posA, catA, catB)
	return err
}

// SwapPositionsInCategory atomically swaps the position_in_category of two oracles
func (r *Repository) SwapPositionsInCategory(idA int64, posA int, idB int64, posB int) error {
	query := `UPDATE oracles
	          SET position_in_category = CASE
	              WHEN id = ? THEN ?
	              WHEN id = ? THEN ?
	          END,
	          updated_at = datetime('now','subsec')
	          WHERE id IN (?, ?)`
	_, err := r.db.Connection.Exec(query, idA, posB, idB, posA, idA, idB)
	return err
}
