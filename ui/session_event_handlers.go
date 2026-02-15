package ui

func (a *App) handleSessionShowNew(e *SessionShowNewEvent) {
	a.sessionView.Modal.SetTitle(" New Session ")
	selectedGameState := a.GetSelectedGameState()
	if selectedGameState == nil {
		a.notification.ShowWarning("Select a game before adding a session.")
		return
	}
	a.sessionView.Form.Reset(*selectedGameState.GameID)
	a.pages.ShowPage(SESSION_MODAL_ID)
	a.SetFocus(a.sessionView.Form)
}

func (a *App) handleSessionCancelled(e *SessionCancelledEvent) {
	a.pages.HidePage(SESSION_MODAL_ID)
}

func (a *App) handleSessionSelected(e *SessionSelectedEvent) {
	a.sessionView.currentSessionID = &e.SessionID
	a.sessionView.Refresh()
}

func (a *App) handleSessionSaved(e *SessionSavedEvent) {
	a.sessionView.Form.ClearFieldErrors()
	a.pages.HidePage(SESSION_MODAL_ID)
	a.sessionView.currentSessionID = &e.Session.ID
	a.sessionView.Refresh()
	a.gameView.Refresh()
	a.gameView.SelectSession(e.Session.ID)
	a.SetFocus(a.sessionView.TextArea)
	a.notification.ShowSuccess("Session saved successfully")
}

func (a *App) handleSessionShowEdit(e *SessionShowEditEvent) {
	if e.Session == nil {
		a.notification.ShowError("Please select a session to edit")
		return
	}

	a.sessionView.Form.PopulateForEdit(e.Session)
	a.pages.ShowPage(SESSION_MODAL_ID)
	a.SetFocus(a.sessionView.Form)
}

func (a *App) handleSessionDeleteConfirm(e *SessionDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this session?",
		func() {
			a.sessionView.ConfirmDelete(e.Session.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleSessionDeleted(_ *SessionDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(SESSION_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.gameView.Refresh()
	a.sessionView.Refresh()
	a.SetFocus(a.gameView.Tree)
	a.notification.ShowSuccess("Session deleted successfully")
}

func (a *App) handleSessionDeleteFailed(e *SessionDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete session: " + e.Error.Error())
}

