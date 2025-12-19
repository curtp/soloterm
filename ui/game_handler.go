package ui

import (
	"soloterm/domain/game"

	"github.com/jmoiron/sqlx"
)

// GameHandler contains handlers for game-related operations
type GameHandler struct {
	db       *sqlx.DB
	gameRepo *game.Repository
}

// NewGameHandler creates a new GameHandlers instance
func NewGameHandler(db *sqlx.DB) *GameHandler {
	return &GameHandler{
		db:       db,
		gameRepo: game.NewRepository(db),
	}
}

// SaveGame saves the game and returns any validation errors
func (h *GameHandler) SaveGame(id *int64, name string, description string) (*game.Game, error) {
	// Description could be nil, so account for that
	var desc *string
	if description != "" {
		desc = &description
	}

	g := &game.Game{
		Name:        name,
		Description: desc,
	}

	// If an ID was provided, then set it on the game. That means it's an update
	if id != nil {
		g.Id = *id
	}

	// Validate the data
	validator := g.Validate()
	if validator.HasErrors() {
		// Return validation error (we'll handle this specially)
		return nil, &ValidationError{Validator: validator}
	}

	// Save to database
	err := h.gameRepo.Save(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (h *GameHandler) GetAll() ([]*game.Game, error) {
	return h.gameRepo.GetAll()
}

func (h *GameHandler) GetByID(id int64) (*game.Game, error) {
	return h.gameRepo.GetByID(id)
}

// Delete deletes a game by ID
func (h *GameHandler) Delete(id int64) error {
	_, err := h.gameRepo.Delete(id)
	return err
}
