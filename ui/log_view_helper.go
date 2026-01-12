package ui

import (
	"fmt"
	"soloterm/domain/game"
	"soloterm/domain/log"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LogViewHelper provides log-specific UI operations
type LogViewHelper struct {
	app *App
}

// NewLogViewHelper creates a new log view
func NewLogViewHelper(app *App) *LogViewHelper {
	return &LogViewHelper{app: app}
}

// Setup initializes all log UI components
func (lv *LogViewHelper) Setup() {
	lv.setupTextView()
	lv.setupModal()
	lv.setupKeyBindings()
	lv.setupFocusHandlers()
}

// setupTextView configures the log text view
func (lv *LogViewHelper) setupTextView() {
	lv.app.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetChangedFunc(func() {
			lv.app.Draw()
		})

	lv.app.logView.SetBorder(true).
		SetTitle(" [::b]Select Game To View ").
		SetTitleAlign(tview.AlignLeft)
}

// setupModal configures the log form modal
func (lv *LogViewHelper) setupModal() {
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
		lv.app.logHandler.HandleSave,
		lv.app.logHandler.HandleCancel,
		lv.app.logHandler.HandleDelete,
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
}

// setupKeyBindings configures keyboard shortcuts for the log view
func (lv *LogViewHelper) setupKeyBindings() {
	lv.app.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			lv.highlightPreviousLog()
			return nil
		case tcell.KeyDown:
			lv.highlightNextLog()
			return nil
		case tcell.KeyCtrlE:
			// Get currently highlighted region and open edit modal
			regions := lv.app.logView.GetHighlights()
			if len(regions) > 0 {
				if logID, err := strconv.ParseInt(regions[0], 10, 64); err == nil {
					lv.app.logHandler.ShowEditModal(logID)
				}
			}
			return nil
		case tcell.KeyCtrlN:
			lv.app.logHandler.ShowModal()
			return nil
		}
		return event
	})
}

// setupFocusHandlers configures focus event handlers
func (lv *LogViewHelper) setupFocusHandlers() {
	lv.app.logView.SetFocusFunc(func() {
		lv.app.updateFooterHelp("[aqua::b]Session Logs[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New")
	})
}

// highlightPreviousLog navigates to the previous log entry
func (lv *LogViewHelper) highlightPreviousLog() {
	if len(lv.app.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := lv.app.logView.GetHighlights()

	// If nothing is highlighted, highlight the last region
	if len(currentHighlights) == 0 {
		lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[len(lv.app.loadedLogIDs)-1]))
		lv.app.logView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]
	for i, logID := range lv.app.loadedLogIDs {
		// Convert the log ID to a string for all the following logic
		regionID := fmt.Sprintf("%d", logID)
		if regionID == currentRegion {
			// Move to previous region (wrap around to end if at beginning)
			if i > 0 {
				lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[i-1]))
			} else {
				lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[len(lv.app.loadedLogIDs)-1]))
			}
			lv.app.logView.ScrollToHighlight()
			return
		}
	}
}

// highlightNextLog navigates to the next log entry
func (lv *LogViewHelper) highlightNextLog() {
	if len(lv.app.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := lv.app.logView.GetHighlights()

	// If nothing is highlighted, highlight the first region
	if len(currentHighlights) == 0 {
		lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[0]))
		lv.app.logView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]
	for i, logID := range lv.app.loadedLogIDs {
		// Convert the log ID to a string for all the following logic
		regionID := fmt.Sprintf("%d", logID)
		if regionID == currentRegion {
			// Move to next region (wrap around to beginning if at end)
			if i < len(lv.app.loadedLogIDs)-1 {
				lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[i+1]))
			} else {
				lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.loadedLogIDs[0]))
			}
			lv.app.logView.ScrollToHighlight()
			return
		}
	}
}

// Refresh loads logs for the currently selected game or session
func (lv *LogViewHelper) Refresh() {
	// Clear the highlight
	lv.app.logView.Highlight()

	// Get the reference from the game tree node
	if ref := lv.app.gameTree.GetCurrentNode().GetReference(); ref != nil {
		// Check if it's a game
		if g, ok := ref.(*game.Game); ok {
			// Update selected game state
			lv.app.selectedGame = g

			// Load all logs for this game
			lv.loadLogsForGame(g.ID)
		} else if s, ok := ref.(*log.Session); ok {
			// It's a session - load the game from service using session's GameID
			g, err := lv.app.gameService.GetByID(s.GameID)
			if err == nil {
				lv.app.selectedGame = g
			}

			// Load logs for this specific session date
			lv.loadLogsForSession(s.GameID, s.Date)
		}
	}

	// If there's a selected log, highlight it
	if lv.app.selectedLog != nil {
		lv.app.logView.Highlight(fmt.Sprintf("%d", lv.app.selectedLog.ID))
	}
}

// loadLogsForGame loads all logs for a game and displays them in the log view
func (lv *LogViewHelper) loadLogsForGame(gameID int64) {
	// Clear the log view
	lv.app.logView.Clear()

	// Load logs from database
	logs, err := lv.app.logService.GetAllForGame(gameID)
	if err != nil {
		lv.app.logView.SetText("[red]Error loading logs: " + err.Error())
		return
	}

	if len(logs) == 0 {
		lv.app.logView.SetText("[gray]No Logs Yet. Press Ctrl+N to Add")
		return
	}

	lv.displayLogs(logs)
}

// loadLogsForSession loads logs for a specific session and displays them in the log view
func (lv *LogViewHelper) loadLogsForSession(gameID int64, sessionDate string) {
	// Clear the log view
	lv.app.logView.Clear()

	// Load logs for this session from database
	logs, err := lv.app.logService.GetLogsForSession(gameID, sessionDate)
	if err != nil {
		lv.app.logView.SetText("[red]Error loading session logs: " + err.Error())
		return
	}

	lv.displayLogs(logs)
}

// displayLogs renders logs in the log view panel
func (lv *LogViewHelper) displayLogs(logs []*log.Log) {
	// Display logs grouped by time blocks
	var output string
	var lastBlockLabel string
	var wroteBlockLabel bool

	// Store the IDs of the loaded logs for navigation purposes
	lv.app.loadedLogIDs = nil

	_, _, w, _ := lv.app.logView.GetInnerRect()

	for _, l := range logs {
		lv.app.loadedLogIDs = append(lv.app.loadedLogIDs, l.ID)

		// Format timestamp in local timezone
		localTime := l.CreatedAt.Local()

		// Create time block label (date + hour)
		blockLabel := localTime.Format(" 2006-01-02 ")

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
		timestamp := localTime.Format("03:04 PM")
		// Add the region ID so clicking it will open the edit modal
		output += "[\"" + fmt.Sprintf("%d", l.ID) + "\"][::i]"
		// Add the timestamp and log type display name
		output += "[aqua::b]" + timestamp + " - " + l.LogType.LogTypeDisplayName() + "[-::-][\"\"]\n"
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
	lv.app.logView.SetText(output)
	// Update the title of the text area
	lv.app.logView.SetTitle(" [::b]Session Logs ")
}
