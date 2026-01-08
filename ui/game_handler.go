package ui

import (
	"fmt"
)

// GameHandler coordinates game-related UI operations
type GameHandler struct {
	app *App
}

// NewGameHandler creates a new game handler
func NewGameHandler(app *App) *GameHandler {
	return &GameHandler{
		app: app,
	}
}

// HandleSave processes game save operation
func (gh *GameHandler) HandleSave() {
	gameEntity := gh.app.gameForm.BuildDomain()

	savedGame, err := gh.app.gameService.Save(gameEntity)
	if err != nil {
		// Check if it's a validation error
		if handleValidationError(err, gh.app.gameForm) {
			return
		}

		// Other errors
		gh.app.notification.ShowError(fmt.Sprintf("Error saving game: %v", err))
		return
	}

	// Success - update domain state and orchestrate UI updates
	gh.app.selectedGame = savedGame
	gh.app.UpdateView(GAME_SAVED)
}

// HandleCancel processes game form cancellation
func (gh *GameHandler) HandleCancel() {
	gh.app.gameForm.Reset()
	gh.app.UpdateView(GAME_CANCEL)
}

// HandleEdit prepares form for editing selected game
func (gh *GameHandler) HandleEdit() {
	gh.app.UpdateView(GAME_SHOW_EDIT)
}

// HandleDelete processes game deletion with confirmation
func (gh *GameHandler) HandleDelete() {
	if gh.app.selectedGame == nil {
		gh.app.notification.ShowError("Please select a game to delete")
		return
	}

	// Get the game entity from the form
	gameEntity := gh.app.gameForm.BuildDomain()

	// Capture current focus to return to after cancel
	gh.app.confirmModal.SetReturnFocus(gh.app.GetFocus())

	// Show confirmation modal
	gh.app.confirmModal.Configure(
		"Are you sure you want to delete this game and all associated log entries?\n\nThis action cannot be undone.",
		func() {
			// Business logic: Delete the game
			err := gh.app.gameService.Delete(gameEntity.ID)
			if err != nil {
				gh.app.UpdateView(CONFIRM_CANCEL)
				gh.app.notification.ShowError(fmt.Sprintf("Error deleting game: %v", err))
				return
			}

			// Update domain state
			gh.app.selectedGame = nil

			// Orchestrate UI updates
			gh.app.UpdateView(GAME_DELETED)
		},
		func() {
			// Cancel deletion
			gh.app.UpdateView(CONFIRM_CANCEL)
		},
	)
	gh.app.UpdateView(CONFIRM_SHOW)
}

// ShowModal displays the game form modal for creating a new game
func (gh *GameHandler) ShowModal() {
	gh.app.UpdateView(GAME_SHOW_NEW)
}

// ShowEditModal displays the game form modal for editing an existing game
func (gh *GameHandler) ShowEditModal() {
	if gh.app.selectedGame == nil {
		gh.app.notification.ShowError("Please select a game to edit")
		return
	}

	gh.app.UpdateView(GAME_SHOW_EDIT)
}
