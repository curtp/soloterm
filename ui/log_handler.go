package ui

import (
	"fmt"
	"soloterm/domain/log"
	"soloterm/shared/validation"
	"strings"
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
		if validator, ok := err.(*validation.Validator); ok {
			fieldErrors := make(map[string]string)
			for _, fieldError := range validator.Errors {
				fieldErrors[fieldError.Field] = fieldError.FormattedErrorMessage()
			}
			lh.app.logForm.SetFieldErrors(fieldErrors)
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

// HandleEdit prepares form for editing selected log
func (lh *LogHandler) HandleEdit() {
	lh.app.UpdateView(LOG_SHOW_EDIT)
}

// HandleDelete processes log deletion with confirmation
func (lh *LogHandler) HandleDelete() {
	if lh.app.selectedLog == nil {
		lh.app.notification.ShowError("Please select a log to delete")
		return
	}

	// Get the log entity from the form
	logEntry := lh.app.logForm.BuildDomain()

	// Show confirmation modal
	lh.app.confirmModal.Show(
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

// LoadLogsForGame loads all logs for a game and displays them in the log view
func (lh *LogHandler) LoadLogsForGame(gameID int64) {
	// Clear the log view
	lh.app.logView.Clear()

	// Load logs from database
	logs, err := lh.app.logService.GetAllForGame(gameID)
	if err != nil {
		lh.app.logView.SetText("[red]Error loading logs: " + err.Error())
		return
	}

	if len(logs) == 0 {
		lh.app.logView.SetText("[gray]No logs yet for this game. Press Ctrl+L to create one.")
		return
	}

	lh.DisplayLogs(logs)
}

// LoadLogsForSession loads logs for a specific session and displays them in the log view
func (lh *LogHandler) LoadLogsForSession(gameID int64, sessionDate string) {
	// Clear the log view
	lh.app.logView.Clear()

	// Load logs for this session from database
	logs, err := lh.app.logService.GetLogsForSession(gameID, sessionDate)
	if err != nil {
		lh.app.logView.SetText("[red]Error loading session logs: " + err.Error())
		return
	}

	lh.DisplayLogs(logs)
}

// DisplayLogs renders logs in the log view panel
func (lh *LogHandler) DisplayLogs(logs []*log.Log) {
	// Display logs grouped by time blocks
	var output string
	var lastBlockLabel string

	for _, l := range logs {
		// Format timestamp in local timezone
		localTime := l.CreatedAt.Local()

		// Create time block label (date + hour)
		blockLabel := localTime.Format("2006-01-02 3PM")

		// Print time block header if it changed
		if blockLabel != lastBlockLabel {
			output += "[lime::b]===  " + blockLabel + "  ===[-::-]\n\n"
			lastBlockLabel = blockLabel
		}

		// Print the log entry
		timestamp := localTime.Format("03:04 PM")
		output += "[\"" + fmt.Sprintf("%d", l.ID) + "\"][::i][aqua::b]" + timestamp + "[-::-] "
		output += "[::i][yellow::b]" + l.LogType.LogTypeDisplayName() + "[-::-]\n"
		if len(l.Description) > 0 {
			output += "[::i][yellow::b]Description:[-::-] " + l.Description + "\n"
		}
		if len(l.Result) > 0 {
			output += "[::i][yellow::b]Result:[-::-] " + l.Result + "\n"
		}
		if len(l.Narrative) > 0 {
			output += "[::i][yellow::b]Narrative:[-::-]\n" + l.Narrative + "[\"\"]\n"
		}
		output += "[-::-][\"\"]\n"
		_, _, w, _ := lh.app.logView.GetInnerRect()
		output += strings.Repeat("─", w) + "\n\n"
	}

	lh.app.logView.SetText(output)
	lh.app.logView.SetTitle(" Session Logs - Click To Edit, Ctrl+L To Add ")
}
