package ui

import (
	"fmt"
	syslog "log"
	"slices"
	"soloterm/domain/log"
	sharedui "soloterm/shared/ui"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogView provides log-specific UI operations
type LogView struct {
	app           *App
	helper        *LogViewHelper
	loadedLogIDs  []int64
	selectedLogID *int64
}

// NewLogViewHelper creates a new log view
func NewLogView(app *App) *LogView {
	return &LogView{app: app, helper: NewLogViewHelper(app.logService)}
}

// Setup initializes all log UI components
func (lv *LogView) Setup() {
	lv.setupTextView()
	lv.setupModal()
	lv.setupKeyBindings()
	lv.setupFocusHandlers()
}

// setupTextView configures the log text view
func (lv *LogView) setupTextView() {
	lv.app.logTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetChangedFunc(func() {
			lv.app.Draw()
		})

	lv.app.logTextView.SetBorder(true).
		SetTitle(" [::b]Select Game To View ").
		SetTitleAlign(tview.AlignLeft)
}

// setupModal configures the log form modal
func (lv *LogView) setupModal() {
	// Create help text view at modal level (no border, will be inside form container)
	helpTextView := tview.NewTextView()
	helpTextView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true).
		SetBorder(false)

	// Create container with border that holds both form and help
	lv.app.logModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(lv.app.logForm, 0, 1, true).
		AddItem(helpTextView, 3, 0, false) // Fixed 3-line height for help
	lv.app.logModalContent.SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	// Subscribe to form's help text changes
	lv.app.logForm.SetHelpTextChangeHandler(func(text string) {
		helpTextView.SetText(text)
	})

	// Set up handlers
	lv.app.logForm.SetupHandlers(
		lv.HandleSave,
		lv.HandleCancel,
		lv.HandleDelete,
	)

	// Center the modal on screen
	lv.app.logModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(lv.app.logModalContent, 23, 2, true). // Increased height for help text
				AddItem(nil, 0, 1, false),
			100, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)

	lv.app.logModal.SetFocusFunc(func() {
		lv.app.SetModalHelpMessage(*lv.app.logForm.DataForm)
	})

}

// setupKeyBindings configures keyboard shortcuts for the log view
func (lv *LogView) setupKeyBindings() {
	lv.app.logTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			lv.highlightPreviousLog()
			return nil
		case tcell.KeyDown:
			lv.highlightNextLog()
			return nil
		case tcell.KeyPgUp:
			// Scroll up within the text view
			row, col := lv.app.logTextView.GetScrollOffset()
			lv.app.logTextView.ScrollTo(row-20, col)
			return nil
		case tcell.KeyPgDn:
			// Scroll down within the text view
			row, col := lv.app.logTextView.GetScrollOffset()
			lv.app.logTextView.ScrollTo(row+20, col)
			return nil
		case tcell.KeyCtrlE:
			// Get currently highlighted region and open edit modal
			if lv.selectedLogID != nil {
				lv.ShowEditModal(*lv.selectedLogID)
			}
			return nil
		case tcell.KeyCtrlN:
			lv.ShowModal()
			return nil
		}

		return event
	})

	lv.app.logModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlT:
			// Dispatch event with saved log
			lv.app.HandleEvent(&TagShowEvent{
				BaseEvent: BaseEvent{action: TAG_SHOW},
			})
			return nil
		}
		return event
	})

}

// setupFocusHandlers configures focus event handlers
func (lv *LogView) setupFocusHandlers() {
	lv.app.logTextView.SetFocusFunc(func() {
		lv.app.updateFooterHelp("[aqua::b]Session Logs[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]PgUp/PgDn[white] Scroll  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
	})
}

// highlightPreviousLog navigates to the previous log entry
func (lv *LogView) highlightPreviousLog() {
	// Nothing to do if there are no logs loaded
	if len(lv.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := lv.app.logTextView.GetHighlights()

	// If nothing is highlighted, highlight the last region
	if len(currentHighlights) == 0 {
		lv.selectedLogID = &lv.loadedLogIDs[len(lv.loadedLogIDs)-1]
		lv.app.logTextView.Highlight(fmt.Sprintf("%d", *lv.selectedLogID))
		lv.app.logTextView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]

	// Convert the current selection to an integer
	id, err := strconv.ParseInt(currentRegion, 10, 64)
	if err != nil {
		syslog.Printf("Error: %s", err)
		return
	}

	// Search for it in the array and then select the next appropriate log
	ndx := slices.Index(lv.loadedLogIDs, id)
	if ndx > -1 {
		if ndx > 0 {
			lv.selectedLogID = &lv.loadedLogIDs[ndx-1]
		} else {
			lv.selectedLogID = &lv.loadedLogIDs[len(lv.loadedLogIDs)-1]
		}
	} else {
		return
	}

	// Do the highlighting
	lv.app.logTextView.Highlight(fmt.Sprintf("%d", *lv.selectedLogID))
	lv.app.logTextView.ScrollToHighlight()
}

// highlightNextLog navigates to the next log entry
func (lv *LogView) highlightNextLog() {
	if len(lv.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := lv.app.logTextView.GetHighlights()

	// If nothing is highlighted, highlight the first region
	if len(currentHighlights) == 0 {
		lv.selectedLogID = &lv.loadedLogIDs[0]
		lv.app.logTextView.Highlight(fmt.Sprintf("%d", *lv.selectedLogID))
		lv.app.logTextView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]

	// Convert the current selection to an integer
	id, err := strconv.ParseInt(currentRegion, 10, 64)
	if err != nil {
		syslog.Printf("Error: %s", err)
		return
	}

	// Search for it in the array and then select the next appropriate log
	ndx := slices.Index(lv.loadedLogIDs, id)
	if ndx > -1 {
		if ndx < len(lv.loadedLogIDs)-1 {
			lv.selectedLogID = &lv.loadedLogIDs[ndx+1]
		} else {
			lv.selectedLogID = &lv.loadedLogIDs[0]
		}
	} else {
		return
	}

	// Do the highlighting
	lv.app.logTextView.Highlight(fmt.Sprintf("%d", *lv.selectedLogID))
	lv.app.logTextView.ScrollToHighlight()
}

// Refresh loads logs for the currently selected game or session
func (lv *LogView) Refresh() {
	// Clear the highlight
	lv.app.logTextView.Highlight()
	// Clear the log view
	lv.app.logTextView.Clear()
	// Get the currently selected game state
	selectedGameState := lv.app.GetSelectedGameState()

	// If there isn't a selected game, then reset the state
	if selectedGameState == nil {
		lv.selectedLogID = nil
		return
	}

	if selectedGameState.SessionDate != nil {
		lv.loadLogsForSession(*selectedGameState.GameID, *selectedGameState.SessionDate)
	} else {
		lv.loadLogsForGame(*selectedGameState.GameID)
	}

	// If there was a selected log, try to highlight and scroll
	if lv.selectedLogID != nil {
		lv.app.logTextView.Highlight(fmt.Sprintf("%d", *lv.selectedLogID))
		lv.app.logTextView.ScrollToHighlight()
	}
}

// loadLogsForGame loads all logs for a game and displays them in the log view
func (lv *LogView) loadLogsForGame(gameID int64) {

	// Load logs from database
	logs, err := lv.app.logService.GetAllForGame(gameID)
	if err != nil {
		lv.app.logTextView.SetText("[red]Error loading logs: " + err.Error())
		return
	}

	if len(logs) == 0 {
		lv.app.logTextView.SetText("[gray]No Logs Yet. Press Ctrl+N to Add")
		return
	}

	lv.displayLogs(logs)
}

// loadLogsForSession loads logs for a specific session and displays them in the log view
func (lv *LogView) loadLogsForSession(gameID int64, sessionDate string) {

	// Load logs for this session from database
	logs, err := lv.app.logService.GetLogsForSession(gameID, sessionDate)
	if err != nil {
		lv.app.logTextView.SetText("[red]Error loading session logs: " + err.Error())
		return
	}

	lv.displayLogs(logs)
}

// displayLogs renders logs in the log view panel
func (lv *LogView) displayLogs(logs []*log.Log) {
	// Display logs grouped by time blocks
	var output string
	var lastBlockLabel string
	var wroteBlockLabel bool

	// If there are no logs to display, make sure to reset the selected log ID
	if len(logs) == 0 {
		lv.selectedLogID = nil
	}

	// Store the IDs of the loaded logs for navigation purposes
	lv.loadedLogIDs = nil

	_, _, w, _ := lv.app.logTextView.GetInnerRect()

	for _, l := range logs {
		lv.loadedLogIDs = append(lv.loadedLogIDs, l.ID)

		blockLabel := " " + l.SessionDate() + " "

		// Print time block header if it changed
		if blockLabel != lastBlockLabel {
			b := len(blockLabel)
			s := w - b
			h := s / 2
			strings.Repeat("─", w)
			output += "[lime::b]" + strings.Repeat("=", h) + blockLabel + strings.Repeat("=", h) + "[-::-]\n\n"
			lastBlockLabel = blockLabel
			wroteBlockLabel = true
		} else {
			wroteBlockLabel = false
		}

		if !wroteBlockLabel {
			output += strings.Repeat("─", w) + "\n\n"
		}

		// Print the log entry
		// Add the region ID so clicking it will open the edit modal
		output += "[\"" + fmt.Sprintf("%d", l.ID) + "\"][::i]"
		// Add the timestamp and log type display name
		output += "[aqua::b]" + l.SessionTime() + " - " + l.LogType.LogTypeDisplayName() + "[-::-][\"\"]\n"
		if len(l.Description) > 0 {
			output += "[::i][yellow::b]Description:[-::-] " + l.Description + "\n"
		}
		if len(l.Result) > 0 {
			output += "[::i][yellow::b]     Result:[-::-] " + l.Result + "\n"
		}
		if len(l.Narrative) > 0 {
			// output += "[::i][yellow::b]  Narrative:[-::-] " + l.Narrative + "[\"\"]\n"
			// output += "\n" + l.Narrative + "[\"\"]\n"
			output += "\n" + l.Narrative + "\n"
		}
		// Close the region and other formatting
		output += "[-::-][\"\"]\n"
	}

	// Write the output to the log
	lv.app.logTextView.SetText(output)
	// Update the title of the text area
	lv.app.logTextView.SetTitle(" [::b]Session Logs (Ctrl+L) ")
}

// HandleSave processes log save operation
func (lv *LogView) HandleSave() {
	logEntity := lv.app.logForm.BuildDomain()

	savedLog, err := lv.app.logService.Save(logEntity)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, lv.app.logForm) {
			return
		}

		// Other errors
		lv.app.notification.ShowError(fmt.Sprintf("Error saving log: %v", err))
		return
	}

	// Set the saved log as the currently selected log
	lv.selectedLogID = &savedLog.ID

	// Dispatch event with saved log
	lv.app.HandleEvent(&LogSavedEvent{
		BaseEvent: BaseEvent{action: LOG_SAVED},
		Log:       *savedLog,
	})
}

// HandleCancel processes log form cancellation
func (lv *LogView) HandleCancel() {
	lv.app.HandleEvent(&LogCancelledEvent{
		BaseEvent: BaseEvent{action: LOG_CANCEL},
	})
}

// HandleDelete processes log deletion with confirmation
func (lv *LogView) HandleDelete() {
	if lv.selectedLogID == nil {
		lv.app.notification.ShowError("Please select a log to delete")
		return
	}

	// Dispatch event to show confirmation
	lv.app.HandleEvent(&LogDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: LOG_DELETE_CONFIRM},
		LogID:     *lv.selectedLogID,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (lv *LogView) ConfirmDelete(logID int64) {
	// Business logic: Delete the log
	err := lv.app.logService.Delete(logID)
	if err != nil {
		// Dispatch failure event with error
		lv.app.HandleEvent(&LogDeleteFailedEvent{
			BaseEvent: BaseEvent{action: LOG_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	lv.selectedLogID = nil

	// Dispatch success event
	lv.app.HandleEvent(&LogDeletedEvent{
		BaseEvent: BaseEvent{action: LOG_DELETED},
	})
}

// ShowModal displays the log form modal for creating a new log entry
func (lv *LogView) ShowModal() {
	selectedGameState := lv.app.GetSelectedGameState()

	// Check if we have a selected game
	if selectedGameState == nil {
		lv.app.notification.ShowError("Please select a game first")
		return
	}

	lv.app.HandleEvent(&LogShowNewEvent{
		BaseEvent: BaseEvent{action: LOG_SHOW_NEW},
	})
}

// ShowEditModal displays the log form modal for editing an existing log entry
func (lv *LogView) ShowEditModal(logID int64) {
	// Load the log entry from the database
	logEntry, err := lv.app.logService.GetByID(logID)
	if err != nil {
		lv.app.notification.ShowError("Error loading log entry: " + err.Error())
		return
	}

	// Set the selected log ID so HandleDelete knows which log to delete
	lv.selectedLogID = &logID

	lv.app.HandleEvent(&LogShowEditEvent{
		BaseEvent: BaseEvent{action: LOG_SHOW_EDIT},
		Log:       logEntry,
	})
}
