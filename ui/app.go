package ui

import (
	"soloterm/domain/game"
	"soloterm/domain/log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rivo/tview"
)

const (
	GAME_MODAL_ID    string = "gameModal"
	LOG_MODAL_ID     string = "logModal"
	CONFIRM_MODAL_ID string = "confirm"
	MAIN_PAGE_ID     string = "main"
	HELP_MODAL_ID    string = "help"
)

type App struct {
	*tview.Application

	// Dependencies
	db          *sqlx.DB
	gameHandler *GameHandler
	logHandler  *LogHandler

	// Application State
	selectedGame *game.Game // Currently selected/active game
	selectedLog  *log.Log   // Currently selected log (for editing)

	// Layout containers
	mainFlex         *tview.Flex
	pages            *tview.Pages
	rootFlex         *tview.Flex // Root container with notification
	notificationFlex *tview.Flex // Container for notification banner

	// UI Components
	gameTree     *tview.TreeView
	logView      *tview.TextView
	helpModal    *tview.Modal
	confirmModal *ConfirmationModal
	footer       *tview.TextView
	gameModal    *tview.Flex
	gameForm     *GameForm
	logModal     *tview.Flex
	logForm      *LogForm
	notification *Notification
}

func NewApp(db *sqlx.DB) *App {
	app := &App{
		Application: tview.NewApplication(),
		db:          db,
		gameHandler: NewGameHandler(db),
		logHandler:  NewLogHandler(db),
	}
	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// Left pane: Game/Session tree
	a.gameTree = tview.NewTreeView()
	a.gameTree.SetBorder(true).
		SetTitle(" Games & Sessions (Ctrl+E Edit)").
		SetTitleAlign(tview.AlignLeft)

	// Placeholder root node
	root := tview.NewTreeNode("Games").SetColor(tcell.ColorYellow)
	a.gameTree.SetRoot(root).SetCurrentNode(root)

	// Set up selection handler for the tree
	a.gameTree.SetSelectedFunc(func(node *tview.TreeNode) {
		a.loadLogsForSelectedGameEntry()
		a.logView.ScrollToEnd()
	})

	// Right top pane: Log display
	a.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		// SetToggleHighlights(true).
		SetChangedFunc(func() {
			a.Draw()
		})
	a.logView.SetBorder(true).
		SetTitle(" Session Log (click log to edit) ").
		SetTitleAlign(tview.AlignLeft)

	// Handle region clicks to edit log entries
	a.logView.SetHighlightedFunc(func(added, removed, remaining []string) {
		if len(added) > 0 {
			// Convert the region ID (log ID string) back to int64
			if logID, err := strconv.ParseInt(added[0], 10, 64); err == nil {
				a.handleLogEdit(logID)
			}
		}
	})

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F1[white] Help  [yellow]Ctrl+G[white] New Game  [yellow]Ctrl+L[white] New Log  [yellow]Ctrl+B[white] Toggle Sidebar  [yellow]Ctrl+C[white] Quit")

	// Main layout: horizontal split of tree (left, narrow) and log view (right)
	a.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.gameTree, 0, 1, false). // 1/3 of the width
		AddItem(a.logView, 0, 2, true)    // 2/3 of the width

	// Main content with footer
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.mainFlex, 0, 1, true).
		AddItem(a.footer, 1, 0, false)

	// Help modal
	a.helpModal = tview.NewModal().
		SetText("SoloTerm - Solo RPG Session Logger\n\n" +
			"Keyboard Shortcuts:\n\n" +
			"F1          - Show this help\n" +
			"Ctrl+G      - Create new game\n" +
			"Ctrl+L		 - Create new log entry\n" +
			"Ctrl+B      - Toggle sidebar\n" +
			"Ctrl+S      - Save current entry\n" +
			"Tab         - Switch between panes\n" +
			"Ctrl+C      - Quit application\n\n" +
			"Navigation:\n\n" +
			"Arrow Keys  - Navigate tree/scroll log\n" +
			"Enter       - Select game/session\n" +
			"Space       - Expand/collapse tree nodes").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage(HELP_MODAL_ID)
		})

	// New game modal
	a.setupGameModal()

	// New Log Modal
	a.setupLogModal()

	// Create confirmation modal
	a.confirmModal = NewConfirmationModal()

	// Pages for modal overlay (must be created BEFORE notification setup)
	a.pages = tview.NewPages().
		AddPage(MAIN_PAGE_ID, mainContent, true, true).
		AddPage(HELP_MODAL_ID, a.helpModal, true, false).
		AddPage(GAME_MODAL_ID, a.gameModal, true, false).
		AddPage(LOG_MODAL_ID, a.logModal, true, false).
		AddPage(CONFIRM_MODAL_ID, a.confirmModal, true, false)
	a.pages.SetBackgroundColor(tcell.ColorDefault)

	// Create notification flex (initially hidden - just shows pages)
	a.notificationFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.pages, 0, 1, true)

	// Create notification system (after notificationFlex and pages are created)
	a.notification = NewNotification(a.notificationFlex, a.pages, a.Application)

	// Root flex that can show/hide notification
	a.rootFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.notificationFlex, 0, 1, true)

	a.SetRoot(a.rootFlex, true).SetFocus(a.gameTree)
	a.setupKeyBindings()

	// Load initial game data
	a.refreshGameTree()
}

func (a *App) setupKeyBindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			a.showHelp()
			return nil
		case tcell.KeyCtrlG:
			a.showGameModal()
			return nil
		case tcell.KeyCtrlB:
			a.togglePane(a.gameTree)
			return nil
		case tcell.KeyCtrlL:
			if a.selectedGame == nil {
				a.notification.ShowError("Please select a game first")
				return nil
			}
			a.showLogModal()
			return nil
		case tcell.KeyCtrlE:
			if a.selectedGame == nil {
				a.notification.ShowError("Please select a game first")
				return nil
			}
			a.showEditGameModal(a.selectedGame)
		}

		return event
	})
}

func (a *App) setupGameModal() {
	// Create the form
	a.gameForm = NewGameForm(
		a.handleGameSave,
		a.handleGameCancel,
		a.handleGameDelete,
	)

	// Center the modal on screen
	a.gameModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(a.gameForm, 12, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
	a.gameModal.SetBackgroundColor(tcell.ColorBlack)
}

func (a *App) setupLogModal() {
	// Create the form
	a.logForm = NewLogForm(
		a.handleLogSave,
		a.handleLogCancel,
		a.handleLogDelete,
	)

	// Center the modal on screen
	a.logModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(a.logForm, 20, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
}

func (a *App) showHelp() {
	a.pages.ShowPage(HELP_MODAL_ID)
}

func (a *App) showGameModal() {
	a.gameForm.Reset()
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) showLogModal() {
	a.logForm.Reset()
	a.pages.ShowPage(LOG_MODAL_ID)
	a.SetFocus(a.logForm)
}

func (a *App) showEditGameModal(game *game.Game) {
	a.gameForm.PopulateForEdit(game)
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleGameSave(id *int64, name string, description string) {
	// Create the game
	game, err := a.gameHandler.SaveGame(id, name, description)
	if err != nil {
		// Check if it's a validation error
		if valErr, ok := err.(*ValidationError); ok {
			// Build field errors map from validator
			fieldErrors := make(map[string]string)
			for _, fieldErr := range valErr.Validator.Errors {
				fieldErrors[fieldErr.Field] = fieldErr.FormattedErrorMessage()
			}

			// Set all errors at once and update labels
			a.gameForm.SetFieldErrors(fieldErrors)
			return
		}
		// Other error - could show a generic error modal
		return
	}

	// Success! Update state and refresh the tree
	a.selectedGame = game
	a.pages.HidePage(GAME_MODAL_ID)
	a.refreshGameTree()
	a.SetFocus(a.gameTree)

	// Clear logs view for the new game
	a.loadLogsForGame(game.Id)

	// Set the success message based on update vs. create
	message := "Game created successfuly"
	if id == nil {
		message = "Game saved successfuly"
	}
	// Show success notification
	a.notification.ShowSuccess(message)
}

func (a *App) handleGameCancel() {
	a.pages.HidePage(GAME_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.SetFocus(a.gameTree)
}

func (a *App) handleLogSave(id *int64, logType log.LogType, description string, result string, narrative string) {

	// Check if we have a selected game
	if a.selectedGame == nil {
		// Show error notification
		a.notification.ShowError("Please select a game first")
		return
	}

	// Save the log
	logEntry, err := a.logHandler.SaveLog(id, a.selectedGame.Id, logType, description, result, narrative)
	if err != nil {
		// Check if it's a validation error
		if valErr, ok := err.(*ValidationError); ok {
			// Build field errors map from validator
			fieldErrors := make(map[string]string)
			for _, fieldErr := range valErr.Validator.Errors {
				fieldErrors[fieldErr.Field] = fieldErr.FormattedErrorMessage()
			}

			// Set all errors at once and update labels
			a.logForm.SetFieldErrors(fieldErrors)
			return
		}
		// Other error - could show a generic error modal
		return
	}

	// Refresh the game tree
	a.refreshGameTree()

	// Reload logs for the selected game
	a.loadLogsForSelectedGameEntry()

	// Scroll based on whether this was a new log or an update
	if id == nil {
		// New log - scroll to the end
		a.logView.ScrollToEnd()
	} else {
		// Updated log - scroll to and highlight the specific entry
		a.logView.Highlight(strconv.FormatInt(logEntry.ID, 10))
		a.logView.ScrollToHighlight()
		a.logView.Highlight() // Clear highlight so next click will always trigger
	}

	// Close the modal and refresh
	a.pages.HidePage(GAME_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.SetFocus(a.logView)

	a.notification.ShowSuccess("Log entry saved successfully")
}

func (a *App) loadLogsForSelectedGameEntry() {
	// Get the reference from the node
	if ref := a.gameTree.GetCurrentNode().GetReference(); ref != nil {
		// Check if it's a game
		if g, ok := ref.(*game.Game); ok {
			// Update selected game state
			a.selectedGame = g

			// Load all logs for this game
			a.loadLogsForGame(g.Id)
		} else if s, ok := ref.(*log.Session); ok {
			// It's a session - load the game from repository using session's GameID
			g, err := a.gameHandler.gameRepo.GetByID(s.GameID)
			if err == nil {
				a.selectedGame = g
			}

			// Load logs for this specific session date
			a.loadLogsForSession(s.GameID, s.Date)
		}
	}
}

func (a *App) handleLogCancel() {
	a.pages.HidePage(LOG_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.SetFocus(a.logView)
	a.logView.Highlight() // Clear highlight so next click will always trigger
}

func (a *App) handleLogEdit(logID int64) {
	// Load the log entry from the database
	logEntry, err := a.logHandler.GetByID(logID)
	if err != nil {
		a.notification.ShowError("Error loading log entry: " + err.Error())
		return
	}

	// Populate the form with the existing log data
	a.logForm.PopulateForEdit(logEntry)

	// Show the modal
	a.pages.ShowPage(LOG_MODAL_ID)
	a.SetFocus(a.logForm)
}

func (a *App) handleLogDelete(logID int64) {
	// Show confirmation modal
	a.confirmModal.Show(
		"Are you sure you want to delete this log entry?\n\nThis action cannot be undone.",
		func() {
			// Delete the log entry
			err := a.logHandler.Delete(logID)
			if err != nil {
				a.pages.HidePage(CONFIRM_MODAL_ID)
				a.notification.ShowError("Error deleting log entry: " + err.Error())
				return
			}

			// Close modals
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.pages.HidePage(LOG_MODAL_ID)
			a.pages.SwitchToPage(MAIN_PAGE_ID)

			// Refresh the UI
			a.refreshGameTree()
			a.loadLogsForSelectedGameEntry()
			a.SetFocus(a.logView)

			// Show success notification
			a.notification.ShowSuccess("Log entry deleted successfully")
		},
		func() {
			// Cancel - just hide the confirmation modal
			a.pages.HidePage(CONFIRM_MODAL_ID)
		},
	)
	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleGameEdit(gameID int64) {
	// Load the game from the database
	game, err := a.gameHandler.gameRepo.GetByID(gameID)
	if err != nil {
		a.notification.ShowError("Error loading game: " + err.Error())
		return
	}

	// Populate the form with the existing game data
	a.gameForm.PopulateForEdit(game)

	// Show the modal
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleGameDelete(gameID int64) {
	// Show confirmation modal
	a.confirmModal.Show(
		"Are you sure you want to delete this game and all associated log entries?\n\nThis action cannot be undone.",
		func() {
			// Delete the game
			err := a.gameHandler.Delete(gameID)
			if err != nil {
				a.pages.HidePage(CONFIRM_MODAL_ID)
				a.notification.ShowError("Error deleting game: " + err.Error())
				return
			}

			// Close modals
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.pages.HidePage(GAME_MODAL_ID)
			a.pages.SwitchToPage(MAIN_PAGE_ID)

			// Refresh the UI
			a.logView.Clear()
			a.refreshGameTree()
			a.SetFocus(a.gameTree)

			// Show success notification
			a.notification.ShowSuccess("Game deleted successfully")
		},
		func() {
			// Cancel - just hide the confirmation modal
			a.pages.HidePage(CONFIRM_MODAL_ID)
		},
	)
	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) loadLogsForGame(gameID int64) {

	// Clear the log view
	a.logView.Clear()

	// Load logs from database
	logs, err := a.logHandler.GetAllForGame(gameID)
	if err != nil {
		a.logView.SetText("[red]Error loading logs: " + err.Error())
		return
	}

	if len(logs) == 0 {
		a.logView.SetText("[gray]No logs yet for this game. Press Ctrl+L to create one.")
		return
	}

	a.displayLogs(logs)
}

func (a *App) loadLogsForSession(gameID int64, sessionDate string) {
	// Clear the log view
	a.logView.Clear()

	// Load logs for this session from database
	logs, err := a.logHandler.GetLogsForSession(gameID, sessionDate)
	if err != nil {
		a.logView.SetText("[red]Error loading session logs: " + err.Error())
		return
	}

	a.displayLogs(logs)
}

func (a *App) displayLogs(logs []*log.Log) {
	// Display logs grouped by time blocks
	var output string
	var lastBlockLabel string

	for _, l := range logs {
		// Format timestamp in local timezone
		localTime := l.CreatedAt.In(time.Local)

		// Create time block label (date + hour)
		blockLabel := localTime.Format("2006-01-02 3PM")

		// Print time block header if it changed
		if blockLabel != lastBlockLabel {
			if lastBlockLabel != "" {
				output += "\n" // Add spacing between blocks
			}
			output += "[lime::b]" + blockLabel + "[-::-]\n"
			output += "[gray]─────[-]\n"
			lastBlockLabel = blockLabel
		}

		// Print the log entry
		timestamp := localTime.Format("15:04")
		output += "[\"" + strconv.FormatInt(l.ID, 10) + "\"][::i][aqua::b]" + timestamp + "[-::-] "
		output += "[::i][yellow::b]" + string(l.LogType) + "[-::-]\n"
		output += "[::i][yellow::b]Description:[-::-] " + l.Description + "\n"
		output += "[::i][yellow::b]Result:[-::-] " + l.Result + "\n"
		output += "[::i][yellow::b]Narrative:[-::-] " + l.Narrative + "[\"\"]\n\n"
	}

	a.logView.SetText(output)
}

func (a *App) refreshGameTree() {

	// Save the currently selected reference (game or session) before clearing
	var selectedGameID *int64
	var selectedSessionDate *string
	var selectedSessionGameID *int64

	if currentNode := a.gameTree.GetCurrentNode(); currentNode != nil {
		if ref := currentNode.GetReference(); ref != nil {
			if g, ok := ref.(*game.Game); ok {
				selectedGameID = &g.Id
			} else if s, ok := ref.(*log.Session); ok {
				selectedSessionGameID = &s.GameID
				selectedSessionDate = &s.Date
			}
		}
	}

	root := a.gameTree.GetRoot()
	root.ClearChildren()

	// Load all games from database
	games, err := a.gameHandler.gameRepo.GetAll()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading games: " + err.Error()).
			SetColor(tcell.ColorRed)
		root.AddChild(errorNode)
		return
	}

	if len(games) == 0 {
		// No games yet
		placeholder := tview.NewTreeNode("(No games yet - press Ctrl+G to create one)").
			SetColor(tcell.ColorGray)
		root.AddChild(placeholder)
		return
	}

	var nodeToSelect *tview.TreeNode

	// Add each game to the tree
	for _, g := range games {
		gameNode := tview.NewTreeNode(g.Name).
			SetReference(g).
			SetColor(tcell.ColorLime).
			SetSelectable(true)
		root.AddChild(gameNode)

		// Check if this game was previously selected
		if selectedGameID != nil && g.Id == *selectedGameID {
			nodeToSelect = gameNode
		}

		// Load sessions for this game
		sessions, err := a.logHandler.GetSessionsForGame(g.Id)
		if err != nil || len(sessions) == 0 {
			sessionPlaceholder := tview.NewTreeNode("(No sessions yet)").
				SetColor(tcell.ColorGray).
				SetSelectable(false)
			gameNode.AddChild(sessionPlaceholder)
		} else {
			// Add session nodes
			for _, s := range sessions {
				// Parse the date to format it nicely
				sessionLabel := s.Date + " (" + strconv.Itoa(s.LogCount) + " entries)"
				sessionNode := tview.NewTreeNode(sessionLabel).
					SetReference(s).
					SetColor(tcell.ColorAqua).
					SetSelectable(true)
				gameNode.AddChild(sessionNode)

				// Check if this session was previously selected
				if selectedSessionGameID != nil && selectedSessionDate != nil &&
					s.GameID == *selectedSessionGameID && s.Date == *selectedSessionDate {
					nodeToSelect = sessionNode
					// Expand the parent game node so the session is visible
					gameNode.SetExpanded(true)
				}
			}
		}
	}

	root.SetExpanded(true)

	// Restore the selection if we found a matching node
	if nodeToSelect != nil {
		a.gameTree.SetCurrentNode(nodeToSelect)
	}
}

func (a *App) togglePane(pane tview.Primitive) {
	// Check if pane is in main flex
	itemCount := a.mainFlex.GetItemCount()
	for i := 0; i < itemCount; i++ {
		item := a.mainFlex.GetItem(i)
		if item == pane {
			// Found in main flex, remove it
			a.mainFlex.RemoveItem(pane)
			return
		}
	}

	// Not found, so add it back (only handles gameTree)
	if pane == a.gameTree {
		// Re-add to main flex at the beginning
		a.mainFlex.Clear()
		a.mainFlex.AddItem(a.gameTree, 0, 1, false).
			AddItem(a.logView, 0, 2, true)
	}
}
