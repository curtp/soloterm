// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
	"soloterm/database"
	"soloterm/domain/game"
	"soloterm/domain/log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	GAME_MODAL_ID    string = "gameModal"
	LOG_MODAL_ID     string = "logModal"
	CONFIRM_MODAL_ID string = "confirm"
	MAIN_PAGE_ID     string = "main"
	HELP_MODAL_ID    string = "help"
)

// UserAction represents user-triggered application events
type UserAction string

const (
	GAME_SAVED     UserAction = "game_saved"
	GAME_DELETED   UserAction = "game_deleted"
	GAME_CANCEL    UserAction = "game_cancel"
	GAME_SHOW_NEW  UserAction = "game_show_new"
	GAME_SHOW_EDIT UserAction = "game_show_edit"
	LOG_SAVED      UserAction = "log_saved"
	LOG_DELETED    UserAction = "log_deleted"
	LOG_CANCEL     UserAction = "log_cancel"
	LOG_SHOW_NEW   UserAction = "log_show_new"
	LOG_SHOW_EDIT  UserAction = "log_show_edit"
	CONFIRM_SHOW   UserAction = "confirm_show"
	CONFIRM_CANCEL UserAction = "confirm_cancel"
)

type App struct {
	*tview.Application

	// Dependencies
	db          *database.DBStore
	gameService *game.Service
	logService  *log.Service

	// Handlers
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

func NewApp(db *database.DBStore) *App {
	app := &App{
		Application: tview.NewApplication(),
		db:          db,
		gameService: game.NewService(game.NewRepository(db)),
		logService:  log.NewService(log.NewRepository(db)),
	}

	// Initialize handlers (after app is created so they can reference it)
	app.gameHandler = NewGameHandler(app)
	app.logHandler = NewLogHandler(app)

	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// Left pane: Game/Session tree
	a.gameTree = tview.NewTreeView()
	a.gameTree.SetBorder(true).
		SetTitle(" Games & Sessions (Ctrl+E Edit) ").
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
		SetTitle(" Select Game To View ").
		SetTitleAlign(tview.AlignLeft)

	// Handle region clicks to edit log entries
	a.logView.SetHighlightedFunc(func(added, removed, remaining []string) {
		if len(added) > 0 {
			// Convert the region ID (log ID string) back to int64
			if logID, err := strconv.ParseInt(added[0], 10, 64); err == nil {
				a.logHandler.ShowEditModal(logID)
			}
		}
	})

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]F1[white] Help  [yellow]Tab[white] Switch Pane  [yellow]Ctrl+G[white] New Game  [yellow]Ctrl+L[white] New Log  [yellow]Ctrl+B[white] Toggle Sidebar  [yellow]Ctrl+C[white] Quit")

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
		case tcell.KeyTab:
			// Only handle Tab on main view, not in modals
			currentFocus := a.GetFocus()
			switch currentFocus {
			case a.gameTree:
				a.SetFocus(a.logView)
				return nil
			case a.logView:
				a.SetFocus(a.gameTree)
				return nil
			}
			// If focus is not on main views (ie. in a modal), let Tab work normally
			return event
		case tcell.KeyCtrlG:
			a.gameHandler.ShowModal()
			return nil
		case tcell.KeyCtrlB:
			a.togglePane(a.gameTree)
			a.SetFocus(a.logView)
			return nil
		case tcell.KeyCtrlL:
			a.logHandler.ShowModal()
			return nil
		case tcell.KeyCtrlE:
			a.gameHandler.ShowEditModal()
		}

		return event
	})
}

func (a *App) setupGameModal() {
	// Create the form
	a.gameForm = NewGameForm()

	// Set up handlers
	a.gameForm.SetupHandlers(
		a.gameHandler.HandleSave,
		a.gameHandler.HandleCancel,
		a.gameHandler.HandleDelete,
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
	a.logForm = NewLogForm()

	// Set up handlers
	a.logForm.SetupHandlers(
		a.logHandler.HandleSave,
		a.logHandler.HandleCancel,
		a.logHandler.HandleDelete,
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

// UpdateView orchestrates UI updates based on application events
func (a *App) UpdateView(event UserAction) {
	switch event {
	case GAME_DELETED:
		// Close modals
		a.pages.HidePage(CONFIRM_MODAL_ID)
		a.pages.HidePage(GAME_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)

		// Reset log view to default state
		a.logView.Clear()
		a.logView.SetTitle(" Select Game To View ")
		a.selectedLog = nil

		// Refresh and focus
		a.refreshGameTree()
		a.SetFocus(a.gameTree)

		// Show success notification
		a.notification.ShowSuccess("Game deleted successfully")

	case GAME_SAVED:
		// Clear form errors
		a.gameForm.ClearFieldErrors()

		// Close modal
		a.pages.HidePage(GAME_MODAL_ID)

		// Refresh and focus
		a.refreshGameTree()
		a.SetFocus(a.gameTree)

		// Show success notification
		a.notification.ShowSuccess("Game saved successfully")

	case GAME_CANCEL:
		// Clear form errors
		a.gameForm.ClearFieldErrors()

		// Close modal and focus tree
		a.pages.HidePage(GAME_MODAL_ID)
		a.SetFocus(a.gameTree)

	case LOG_SAVED:
		// Clear form errors
		a.logForm.ClearFieldErrors()

		// Close modal
		a.pages.HidePage(LOG_MODAL_ID)

		// Clear highlights and refresh
		a.logView.Highlight()
		a.loadLogsForSelectedGameEntry()
		a.refreshGameTree()
		a.SetFocus(a.logView)

		// Show success notification
		a.notification.ShowSuccess("Log saved successfully")

	case LOG_DELETED:
		// Close modals
		a.pages.HidePage(CONFIRM_MODAL_ID)
		a.pages.HidePage(LOG_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)

		// Refresh and focus
		a.loadLogsForSelectedGameEntry()
		a.refreshGameTree()
		a.SetFocus(a.logView)

		// Show success notification
		a.notification.ShowSuccess("Log entry deleted successfully")

	case LOG_CANCEL:
		// Clear form errors
		a.logForm.ClearFieldErrors()

		// Close modal, clear highlights and focus log view
		a.pages.HidePage(LOG_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)
		a.logView.Highlight()
		a.SetFocus(a.logView)

	case GAME_SHOW_NEW:
		// Show modal for creating new game
		a.gameForm.Reset()
		a.pages.ShowPage(GAME_MODAL_ID)
		a.SetFocus(a.gameForm)

	case GAME_SHOW_EDIT:
		// Show modal for editing existing game
		if a.selectedGame != nil {
			a.gameForm.PopulateForEdit(a.selectedGame)
			a.pages.ShowPage(GAME_MODAL_ID)
			a.SetFocus(a.gameForm)
		}

	case LOG_SHOW_NEW:
		// Show modal for creating new log
		if a.selectedGame != nil {
			a.logForm.Reset(a.selectedGame.ID)
			a.pages.ShowPage(LOG_MODAL_ID)
			a.SetFocus(a.logForm)
		}

	case LOG_SHOW_EDIT:
		// Show modal for editing existing log
		if a.selectedLog != nil {
			a.logForm.PopulateForEdit(a.selectedLog)
			a.pages.ShowPage(LOG_MODAL_ID)
			a.SetFocus(a.logForm)
		}

	case CONFIRM_SHOW:
		// Show confirmation modal
		a.pages.ShowPage(CONFIRM_MODAL_ID)

	case CONFIRM_CANCEL:
		// Reusable across all confirmation cancellations
		a.pages.HidePage(CONFIRM_MODAL_ID)
	}
}

func (a *App) loadLogsForSelectedGameEntry() {
	// Get the reference from the node
	if ref := a.gameTree.GetCurrentNode().GetReference(); ref != nil {
		// Check if it's a game
		if g, ok := ref.(*game.Game); ok {
			// Update selected game state
			a.selectedGame = g

			// Load all logs for this game
			a.logHandler.LoadLogsForGame(g.ID)
		} else if s, ok := ref.(*log.Session); ok {
			// It's a session - load the game from service using session's GameID
			g, err := a.gameService.GetByID(s.GameID)
			if err == nil {
				a.selectedGame = g
			}

			// Load logs for this specific session date
			a.logHandler.LoadLogsForSession(s.GameID, s.Date)
		}
	}
}

func (a *App) refreshGameTree() {

	// Save the currently selected reference (game or session) before clearing
	var selectedGameID *int64
	var selectedSessionDate *string
	var selectedSessionGameID *int64

	if currentNode := a.gameTree.GetCurrentNode(); currentNode != nil {
		if ref := currentNode.GetReference(); ref != nil {
			if g, ok := ref.(*game.Game); ok {
				selectedGameID = &g.ID
			} else if s, ok := ref.(*log.Session); ok {
				selectedSessionGameID = &s.GameID
				selectedSessionDate = &s.Date
			}
		}
	}

	root := a.gameTree.GetRoot()
	root.ClearChildren()

	// Load all games from database
	games, err := a.gameService.GetAll()
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
		if selectedGameID != nil && g.ID == *selectedGameID {
			nodeToSelect = gameNode
		}

		// Load sessions for this game
		sessions, err := a.logService.GetSessionsForGame(g.ID)
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
