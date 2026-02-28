// Package game provides core game domain logic, validation, and persistence.
// It defines the Game entity and its associated repository for database operations.
package game

import (
	"soloterm/shared/validation"
	"time"
)

const (
	MinNameLength        = 1
	MaxNameLength        = 50
	MinDescriptionLength = 3
	MaxDescriptionLength = 100
)

// Game represents a game in the system
type Game struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"` // May be nil
	Notes       string    `db:"notes"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func NewGame(name string) (*Game, error) {
	game := &Game{
		ID:   0,
		Name: name,
	}

	return game, nil
}

func (g *Game) Validate() *validation.Validator {
	v := validation.NewValidator()
	v.Check("name", g.Name != "", "is required")
	v.Check("name", len(g.Name) >= MinNameLength && len(g.Name) <= MaxNameLength, "must be between %d and %d characters", MinNameLength, MaxNameLength)
	if g.Description != nil {
		v.Check("description", len(*g.Description) >= MinDescriptionLength && len(*g.Description) <= MaxDescriptionLength, "must be between %d and %d characters", MinDescriptionLength, MaxDescriptionLength)
	}
	return v
}

func (g *Game) IsNew() bool {
	return (g.ID == 0)
}
