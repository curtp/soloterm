package ui

import (
	"fmt"
)

// LogHandler coordinates log-related UI operations
type LogHandler struct {
	app *App
}

// NewLogHandler creates a new log handler
func NewLogHandler(app *App) *LogHandler {
	return &LogHandler{
		app: app,
	}
}

// HandleSave processes log save operation
func (lh *LogHandler) HandleSave() {
	logEntity := lh.app.logForm.BuildDomain()

	savedLog, err := lh.app.logService.Save(logEntity)
	if err != nil {
		// Check if it's a validation error
		if handleValidationError(err, lh.app.logForm) {
			return
		}

		// Other errors
		lh.app.notification.ShowError(fmt.Sprintf("Error saving log: %v", err))
		return
	}

	// Success - update domain state and orchestrate UI updates
	lh.app.selectedLog = savedLog
	lh.app.UpdateView(LOG_SAVED)
}

// HandleCancel processes log form cancellation
func (lh *LogHandler) HandleCancel() {
	lh.app.logForm.Reset(lh.app.selectedGame.ID)
	lh.app.UpdateView(LOG_CANCEL)
}

// HandleDelete processes log deletion with confirmation
func (lh *LogHandler) HandleDelete() {
	if lh.app.selectedLog == nil {
		lh.app.notification.ShowError("Please select a log to delete")
		return
	}

	if lh.app.selectedLog.ID == 0 {
		lh.app.notification.ShowError("Only saved log entries can be deleted")
		return
	}

	// Get the log entity from the form
	logEntry := lh.app.logForm.BuildDomain()

	// Capture current focus to return to after cancel
	lh.app.confirmModal.SetReturnFocus(lh.app.GetFocus())

	// Show confirmation modal
	lh.app.confirmModal.Configure(
		"Are you sure you want to delete this log entry?\n\nThis action cannot be undone.",
		func() {
			// Business logic: Delete the log entry
			err := lh.app.logService.Delete(logEntry.ID)
			if err != nil {
				lh.app.UpdateView(CONFIRM_CANCEL)
				lh.app.notification.ShowError("Error deleting log entry: " + err.Error())
				return
			}

			// Orchestrate UI updates
			lh.app.UpdateView(LOG_DELETED)
		},
		func() {
			// Cancel deletion
			lh.app.UpdateView(CONFIRM_CANCEL)
		},
	)
	lh.app.UpdateView(CONFIRM_SHOW)
}

// ShowModal displays the log form modal for creating a new log entry
func (lh *LogHandler) ShowModal() {
	// Check if we have a selected game
	if lh.app.selectedGame == nil {
		lh.app.notification.ShowError("Please select a game first")
		return
	}

	lh.app.UpdateView(LOG_SHOW_NEW)
}

// ShowEditModal displays the log form modal for editing an existing log entry
func (lh *LogHandler) ShowEditModal(logID int64) {
	// Load the log entry from the database
	logEntry, err := lh.app.logService.GetByID(logID)
	if err != nil {
		lh.app.notification.ShowError("Error loading log entry: " + err.Error())
		return
	}

	// Set the selected log on the app
	lh.app.selectedLog = logEntry

	// Show the modal
	lh.app.UpdateView(LOG_SHOW_EDIT)
}
