package ui

import (
	"fmt"
	"soloterm/domain/game"
	sharedui "soloterm/shared/ui"
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
		if sharedui.HandleValidationError(err, gh.app.gameForm) {
			return
		}

		// Other errors
		gh.app.notification.ShowError(fmt.Sprintf("Error saving game: %v", err))
		return
	}

	gh.app.HandleEvent(&GameSavedEvent{
		BaseEvent: BaseEvent{action: GAME_SAVED},
		Game:      savedGame,
	})

}

// HandleCancel processes game form cancellation
func (gh *GameHandler) HandleCancel() {
	gh.app.HandleEvent(&GameCancelledEvent{
		BaseEvent: BaseEvent{action: GAME_CANCEL},
	})
}

// HandleDelete processes game deletion with confirmation
func (gh *GameHandler) HandleDelete() {
	if gh.app.selectedGame == nil {
		gh.app.notification.ShowError("Please select a game to delete")
		return
	}

	// Dispatch event to show confirmation
	gh.app.HandleEvent(&GameDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: GAME_DELETE_CONFIRM},
		Game:      gh.app.selectedGame,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (gh *GameHandler) ConfirmDelete(gameID int64) {
	// Business logic: Delete the game
	err := gh.app.gameService.Delete(gameID)
	if err != nil {
		// Dispatch failure event with error
		gh.app.HandleEvent(&GameDeleteFailedEvent{
			BaseEvent: BaseEvent{action: GAME_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	gh.app.HandleEvent(&GameDeletedEvent{
		BaseEvent: BaseEvent{action: GAME_DELETED},
		Game:      gh.app.selectedGame,
	})
}

// ShowNewModal displays the game form modal for creating a new game
func (gh *GameHandler) ShowNewModal() {
	gh.app.HandleEvent(&GameShowNewEvent{
		BaseEvent: BaseEvent{action: GAME_SHOW_NEW},
	})
}

// ShowEditModal displays the game form modal for editing an existing game
func (gh *GameHandler) ShowEditModal(game *game.Game) {
	gh.app.HandleEvent(&GameShowEditEvent{
		BaseEvent: BaseEvent{action: GAME_SHOW_EDIT},
		Game:      game,
	})

}
