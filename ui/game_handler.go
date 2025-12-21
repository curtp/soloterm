package ui

import (
	"fmt"
	"soloterm/shared/validation"
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
		if validator, ok := err.(*validation.Validator); ok {
			fieldErrors := make(map[string]string)
			for _, fieldError := range validator.Errors {
				fieldErrors[fieldError.Field] = fieldError.FormattedErrorMessage()
			}
			gh.app.gameForm.SetFieldErrors(fieldErrors)
			return
		}

		// Other errors
		gh.app.notification.ShowError(fmt.Sprintf("Error saving game: %v", err))
		return
	}

	// Success - update selected game and refresh UI
	gh.app.selectedGame = savedGame
	gh.app.gameForm.ClearFieldErrors()
	gh.app.pages.HidePage(GAME_MODAL_ID)
	gh.app.SetFocus(gh.app.gameTree)
	gh.app.refreshGameTree()
	gh.app.notification.ShowSuccess("Game saved successfully")
}

// HandleCancel processes game form cancellation
func (gh *GameHandler) HandleCancel() {
	gh.app.gameForm.Reset()
	gh.app.gameForm.ClearFieldErrors()
	gh.app.pages.HidePage(GAME_MODAL_ID)
	gh.app.SetFocus(gh.app.gameTree)
}

// HandleEdit prepares form for editing selected game
func (gh *GameHandler) HandleEdit() {
	gh.app.gameForm.PopulateForEdit(gh.app.selectedGame)
	gh.app.pages.ShowPage(GAME_MODAL_ID)
	gh.app.SetFocus(gh.app.gameForm)
}

// HandleDelete processes game deletion with confirmation
func (gh *GameHandler) HandleDelete() {
	if gh.app.selectedGame == nil {
		gh.app.notification.ShowError("Please select a game to delete")
		return
	}

	// Get the game entity from the form
	gameEntity := gh.app.gameForm.BuildDomain()

	// Show confirmation modal
	gh.app.confirmModal.Show(
		"Are you sure you want to delete this game and all associated log entries?\n\nThis action cannot be undone.",
		func() {
			// Delete the game
			err := gh.app.gameService.Delete(gameEntity.ID)
			if err != nil {
				gh.app.pages.HidePage(CONFIRM_MODAL_ID)
				gh.app.notification.ShowError(fmt.Sprintf("Error deleting game: %v", err))
				return
			}

			// Close modals
			gh.app.pages.HidePage(CONFIRM_MODAL_ID)
			gh.app.pages.HidePage(GAME_MODAL_ID)
			gh.app.pages.SwitchToPage(MAIN_PAGE_ID)

			// Refresh the UI
			gh.app.logView.Clear()
			gh.app.refreshGameTree()
			gh.app.SetFocus(gh.app.gameTree)
			gh.app.selectedGame = nil

			// Show success notification
			gh.app.notification.ShowSuccess("Game deleted successfully")
		},
		func() {
			// Cancel - just hide the confirmation modal
			gh.app.pages.HidePage(CONFIRM_MODAL_ID)
		},
	)
	gh.app.pages.ShowPage(CONFIRM_MODAL_ID)
}

// ShowModal displays the game form modal for creating a new game
func (gh *GameHandler) ShowModal() {
	gh.app.gameForm.Reset()
	gh.app.pages.ShowPage(GAME_MODAL_ID)
	gh.app.SetFocus(gh.app.gameForm)
}

// ShowEditModal displays the game form modal for editing an existing game
func (gh *GameHandler) ShowEditModal() {
	if gh.app.selectedGame == nil {
		gh.app.notification.ShowError("Please select a game to edit")
		return
	}

	gh.app.gameForm.PopulateForEdit(gh.app.selectedGame)
	gh.app.pages.ShowPage(GAME_MODAL_ID)
	gh.app.SetFocus(gh.app.gameForm)
}
