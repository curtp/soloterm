package ui

import (
	"fmt"
)

func (a *App) handleSessionShowNew(e *SessionShowNewEvent) {
	var gameID int64
	var gameName string
	if a.sessionView.currentSession != nil {
		gameID = a.sessionView.currentSession.GameID
		gameName = a.sessionView.currentSession.GameName
	} else {
		selectedGameState := a.gameView.GetCurrentSelection()
		if selectedGameState == nil {
			a.notification.ShowWarning("Select a game before adding a session.")
			return
		}
		gameID = *selectedGameState.GameID
		gameName = selectedGameState.GameName
	}

	a.sessionView.Form.Reset(gameID)
	a.sessionView.formModal.SetTitle(" New Session — " + gameName + " ")
	a.pages.ShowPage(SESSION_MODAL_ID)
	a.SetFocus(a.sessionView.Form)
}

func (a *App) handleSessionCancelled(e *SessionCancelledEvent) {
	a.pages.HidePage(SESSION_MODAL_ID)
	a.SetFocus(a.gameView.Tree)
}

func (a *App) handleGameNotesSelected(e *GameNotesSelectedEvent) {
	a.Autosave()
	if err := a.gameView.SetCurrentGame(e.GameID); err != nil {
		a.notification.ShowError(fmt.Sprintf("Error loading notes: %v", err))
		return
	}
	a.sessionView.SelectNotes()
}

func (a *App) handleSessionSelected(e *SessionSelectedEvent) {
	a.Autosave()
	if err := a.gameView.SetCurrentGame(e.GameID); err != nil {
		a.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
		return
	}
	a.sessionView.SelectSession(e.SessionID)
}

func (a *App) handleSessionSaved(e *SessionSavedEvent) {
	a.sessionView.Form.ClearFieldErrors()
	a.pages.HidePage(SESSION_MODAL_ID)
	a.sessionView.SelectSession(e.Session.ID)
	a.gameView.Refresh()
	a.gameView.SelectSession(e.Session.ID)
	a.SetFocus(a.sessionView.TextArea)
	a.notification.ShowSuccess("Session saved successfully")
}

func (a *App) handleSessionShowEdit(e *SessionShowEditEvent) {
	s := e.Session
	if s == nil && e.SessionID != nil {
		var err error
		s, err = a.sessionView.sessionService.GetByID(*e.SessionID)
		if err != nil {
			a.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
			return
		}
	}
	if s == nil {
		a.notification.ShowError("Please select a session to edit")
		return
	}

	a.sessionView.currentSessionID = &s.ID
	a.sessionView.currentSession = s
	a.sessionView.Form.PopulateForEdit(s)
	a.sessionView.formModal.SetTitle(" Edit Session ")
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
	a.sessionView.Reset()
	a.sessionView.Refresh()
	a.SetFocus(a.gameView.Tree)
	a.notification.ShowSuccess("Session deleted successfully")
}

func (a *App) handleSessionDeleteFailed(e *SessionDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete session: " + e.Error.Error())
}

func (a *App) handleSessionShowImport(_ *SessionShowImportEvent) {
	if a.sessionView.currentSessionID == nil && !a.sessionView.IsNotesMode() {
		return
	}
	a.fileView.ShowImport(a.sessionView, a.sessionView.TextArea)
}

func (a *App) handleSessionShowExport(_ *SessionShowExportEvent) {
	if a.sessionView.currentSessionID == nil && !a.sessionView.IsNotesMode() {
		return
	}
	a.fileView.ShowExport(a.sessionView, a.sessionView.TextArea)
}
