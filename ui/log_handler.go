package ui

import (
	"fmt"
	sharedui "soloterm/shared/ui"
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
		if sharedui.HandleValidationError(err, lh.app.logForm) {
			return
		}

		// Other errors
		lh.app.notification.ShowError(fmt.Sprintf("Error saving log: %v", err))
		return
	}

	// Dispatch event with saved log
	lh.app.HandleEvent(&LogSavedEvent{
		BaseEvent: BaseEvent{action: LOG_SAVED},
		Log:       savedLog,
	})
}

// HandleCancel processes log form cancellation
func (lh *LogHandler) HandleCancel() {
	lh.app.HandleEvent(&LogCancelledEvent{
		BaseEvent: BaseEvent{action: LOG_CANCEL},
	})
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

	// Dispatch event to show confirmation
	lh.app.HandleEvent(&LogDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: LOG_DELETE_CONFIRM},
		Log:       lh.app.selectedLog,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (lh *LogHandler) ConfirmDelete(logID int64) {
	// Business logic: Delete the log
	err := lh.app.logService.Delete(logID)
	if err != nil {
		// Dispatch failure event with error
		lh.app.HandleEvent(&LogDeleteFailedEvent{
			BaseEvent: BaseEvent{action: LOG_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	lh.app.HandleEvent(&LogDeletedEvent{
		BaseEvent: BaseEvent{action: LOG_DELETED},
	})
}

// ShowModal displays the log form modal for creating a new log entry
func (lh *LogHandler) ShowModal() {
	// Check if we have a selected game
	if lh.app.selectedGame == nil {
		lh.app.notification.ShowError("Please select a game first")
		return
	}

	lh.app.HandleEvent(&LogShowNewEvent{
		BaseEvent: BaseEvent{action: LOG_SHOW_NEW},
	})
}

// ShowEditModal displays the log form modal for editing an existing log entry
func (lh *LogHandler) ShowEditModal(logID int64) {
	// Load the log entry from the database
	logEntry, err := lh.app.logService.GetByID(logID)
	if err != nil {
		lh.app.notification.ShowError("Error loading log entry: " + err.Error())
		return
	}

	lh.app.HandleEvent(&LogShowEditEvent{
		BaseEvent: BaseEvent{action: LOG_SHOW_EDIT},
		Log:       logEntry,
	})
}
