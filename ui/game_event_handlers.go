package ui

func (a *App) handleGameSaved(e *GameSavedEvent) {
	a.gameView.Form.Reset()
	a.pages.HidePage(GAME_MODAL_ID)
	a.gameView.SelectGame(&e.Game.ID)
	a.gameView.Refresh()
	a.SetFocus(a.gameView.Tree)
	a.notification.ShowSuccess("Game saved successfully")
}

func (a *App) handleGameCancel(_ *GameCancelledEvent) {
	a.pages.HidePage(GAME_MODAL_ID)
	a.SetFocus(a.gameView.Tree)
}

func (a *App) handleGameDeleteConfirm(e *GameDeleteConfirmEvent) {
	// Capture focus for restoration on cancel
	returnFocus := a.GetFocus()

	// Configure confirmation modal
	a.confirmModal.Configure(
		"Are you sure you want to delete this game and all associated sessions?\n\nThis action cannot be undone.",
		func() {
			// On confirm, call handler method to perform deletion
			a.gameView.ConfirmDelete(e.GameID)
		},
		func() {
			// On cancel, restore focus
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	// Show the confirmation modal
	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleGameDeleted(_ *GameDeletedEvent) {
	// Close modals
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(GAME_MODAL_ID)

	// Refresh and focus
	a.gameView.Refresh()
	a.SetFocus(a.gameView.Tree)

	// Show success notification
	a.notification.ShowSuccess("Game deleted successfully")
}

func (a *App) handleGameDeleteFailed(e *GameDeleteFailedEvent) {
	// Close confirmation modal
	a.pages.HidePage(CONFIRM_MODAL_ID)

	// Show error notification
	a.notification.ShowError("Error deleting game: " + e.Error.Error())
}

func (a *App) handleGameShowEdit(e *GameShowEditEvent) {
	if e.Game == nil {
		a.notification.ShowError("Please select a game to edit")
		return
	}

	a.gameView.Form.PopulateForEdit(e.Game)
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameView.Form)
}

func (a *App) handleGameShowNew(_ *GameShowNewEvent) {
	a.gameView.Form.Reset()
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameView.Form)
}

func (a *App) handleGameSelected(_ *GameSelectedEvent) {
}
