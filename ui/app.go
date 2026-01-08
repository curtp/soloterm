// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
	"fmt"
	"soloterm/database"
	"soloterm/domain/character"
	"soloterm/domain/game"
	"soloterm/domain/log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	GAME_MODAL_ID      string = "gameModal"
	LOG_MODAL_ID       string = "logModal"
	CHARACTER_MODAL_ID string = "characterModal"
	ATTRIBUTE_MODAL_ID string = "attributeModal"
	CONFIRM_MODAL_ID   string = "confirm"
	MAIN_PAGE_ID       string = "main"
	HELP_MODAL_ID      string = "help"
)

// UserAction represents user-triggered application events
type UserAction string

const (
	GAME_SAVED           UserAction = "game_saved"
	GAME_DELETED         UserAction = "game_deleted"
	GAME_CANCEL          UserAction = "game_cancel"
	GAME_SHOW_NEW        UserAction = "game_show_new"
	GAME_SHOW_EDIT       UserAction = "game_show_edit"
	LOG_SAVED            UserAction = "log_saved"
	LOG_DELETED          UserAction = "log_deleted"
	LOG_CANCEL           UserAction = "log_cancel"
	LOG_SHOW_NEW         UserAction = "log_show_new"
	LOG_SHOW_EDIT        UserAction = "log_show_edit"
	CHARACTER_SAVED      UserAction = "character_saved"
	CHARACTER_DELETED    UserAction = "character_deleted"
	CHARACTER_DUPLICATED UserAction = "character_duplicated"
	CHARACTER_CANCEL     UserAction = "character_cancel"
	CHARACTER_SHOW_NEW   UserAction = "character_show_new"
	CHARACTER_SHOW_EDIT  UserAction = "character_show_edit"
	ATTRIBUTE_SAVED      UserAction = "attribute_saved"
	ATTRIBUTE_DELETED    UserAction = "attribute_deleted"
	ATTRIBUTE_CANCEL     UserAction = "attribute_cancel"
	ATTRIBUTE_SHOW_NEW   UserAction = "attribute_show_new"
	ATTRIBUTE_SHOW_EDIT  UserAction = "attribute_show_edit"
	CONFIRM_SHOW         UserAction = "confirm_show"
	CONFIRM_CANCEL       UserAction = "confirm_cancel"
)

type App struct {
	*tview.Application

	// Dependencies
	db          *database.DBStore
	gameService *game.Service
	logService  *log.Service
	charService *character.Service

	// Handlers
	gameHandler *GameHandler
	logHandler  *LogHandler
	charHandler *CharacterHandler

	// Application State
	selectedGame      *game.Game           // Currently selected/active game
	selectedLog       *log.Log             // Currently selected log (for editing)
	selectedCharacter *character.Character // Currently selected character (for display)
	loadedLogIDs      []int64

	// Layout containers
	mainFlex         *tview.Flex
	pages            *tview.Pages
	rootFlex         *tview.Flex // Root container with notification
	leftSidebar      *tview.Flex // Left sidebar
	notificationFlex *tview.Flex // Container for notification banner
	charPane         *tview.Flex // Character section

	// UI Components
	gameTree       *tview.TreeView
	charTree       *tview.TreeView
	charInfoView   *tview.TextView
	attributeTable *tview.Table
	logView        *tview.TextView
	helpModal      *tview.Modal
	confirmModal   *ConfirmationModal
	footer         *tview.TextView
	gameModal      *tview.Flex
	gameForm       *GameForm
	logModal       *tview.Flex
	logForm        *LogForm
	characterModal *tview.Flex
	characterForm  *CharacterForm
	attributeModal *tview.Flex
	attributeForm  *AttributeForm
	notification   *Notification
}

func NewApp(db *database.DBStore) *App {
	app := &App{
		Application: tview.NewApplication(),
		db:          db,
		gameService: game.NewService(game.NewRepository(db)),
		logService:  log.NewService(log.NewRepository(db)),
		charService: character.NewService(character.NewRepository(db)),
	}

	// Initialize handlers (after app is created so they can reference it)
	app.gameHandler = NewGameHandler(app)
	app.logHandler = NewLogHandler(app)
	app.charHandler = NewCharacterHandler(app)

	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// Left pane: Game/Session tree
	a.gameTree = tview.NewTreeView()
	a.gameTree.SetBorder(true).
		SetTitle(" Games & Sessions ").
		SetTitleAlign(tview.AlignLeft)

	// Placeholder root node
	root := tview.NewTreeNode("Games").SetColor(tcell.ColorYellow).SetSelectable(false)
	a.gameTree.SetRoot(root).SetCurrentNode(root)

	// Set up selection handler for the tree (triggered by Space)
	a.gameTree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return
		}

		// If node has children (it's a game), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
		}

		// Always load logs for the selected game or session
		a.loadLogsForSelectedGameEntry()
		a.logView.ScrollToBeginning()
	})

	// Set up input capture for game tree - Ctrl+E to edit game, Ctrl+N to add game
	a.gameTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			node := a.gameTree.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()
				// Only open edit modal if it's a game (not a session)
				if _, ok := reference.(*game.Game); ok {
					a.loadLogsForSelectedGameEntry()
					a.gameHandler.ShowEditModal()
					return nil
				}
			}
		case tcell.KeyCtrlN:
			a.gameHandler.ShowModal()
			return nil
		}
		return event
	})

	// Setup the char tree
	a.charTree = tview.NewTreeView()
	a.charTree.SetBorder(true).
		SetTitle(" Characters ").
		SetTitleAlign(tview.AlignLeft)

	// Set up selection handler for the tree
	a.charTree.SetSelectedFunc(func(node *tview.TreeNode) {
		// If node has children (it's a system), expand/collapse it
		if len(node.GetChildren()) > 0 {
			node.SetExpanded(!node.IsExpanded())
			return
		}

		a.loadSelectedCharacter()
	})

	// Set up input capture for character tree - Ctrl+N to add character
	a.charTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			a.charHandler.ShowModal()
			return nil
		}
		return event
	})

	// Setup character info display
	a.charInfoView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetText("")
	a.charInfoView.SetBorder(true).
		SetTitle(" Character Info ").
		SetTitleAlign(tview.AlignLeft)

	// Set up input capture for character info - Ctrl+E to edit, Ctrl+D to duplicate character
	a.charInfoView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			a.charHandler.ShowEditCharacterModal()
			return nil
		case tcell.KeyCtrlD:
			a.charHandler.HandleDuplicate()
			return nil
		}
		return event
	})

	// Setup attribute table
	a.attributeTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows
	a.attributeTable.SetBorder(true).
		SetTitle(" Sheet ").
		SetTitleAlign(tview.AlignLeft)

	// Set up input capture for attribute table
	a.attributeTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := a.attributeTable.GetSelection()

		// Handle Ctrl+N or Insert key to add new attribute
		if event.Key() == tcell.KeyCtrlN || event.Key() == tcell.KeyInsert {
			a.charHandler.ShowNewAttributeModal()
			return nil
		}

		// Handle Ctrl+E to edit selected attribute
		if event.Key() == tcell.KeyCtrlE {
			if a.selectedCharacter != nil {
				attrs, _ := a.charService.GetAttributesForCharacter(a.selectedCharacter.ID)
				attrIndex := row - 1
				if attrIndex >= 0 && attrIndex < len(attrs) {
					a.charHandler.ShowEditAttributeModal(attrs[attrIndex])
				}
			}
			return nil
		}

		return event
	})

	// Right top pane: Log display
	a.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetChangedFunc(func() {
			a.Draw()
		})
	a.logView.SetBorder(true).
		SetTitle(" Select Game To View ").
		SetTitleAlign(tview.AlignLeft)

	// Set up input capture for log view navigation
	a.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			a.highlightPreviousLog()
			return nil
		case tcell.KeyDown:
			a.highlightNextLog()
			return nil
		case tcell.KeyCtrlE:
			// Get currently highlighted region and open edit modal
			regions := a.logView.GetHighlights()
			if len(regions) > 0 {
				if logID, err := strconv.ParseInt(regions[0], 10, 64); err == nil {
					a.logHandler.ShowEditModal(logID)
				}
			}
			return nil
		case tcell.KeyCtrlN:
			a.logHandler.ShowModal()
			return nil
		}
		return event
	})

	// Set up focus handlers for context-sensitive help
	a.gameTree.SetFocusFunc(func() {
		a.updateFooterHelp("gameTree")
	})
	a.logView.SetFocusFunc(func() {
		a.updateFooterHelp("logView")
	})
	a.charTree.SetFocusFunc(func() {
		a.updateFooterHelp("charTree")
	})
	a.charInfoView.SetFocusFunc(func() {
		a.updateFooterHelp("charInfo")
	})
	a.attributeTable.SetFocusFunc(func() {
		a.updateFooterHelp("attributeTable")
	})

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Character pane combining info and attributes
	a.charPane = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.charInfoView, 6, 1, false).  // Proportional height (smaller weight)
		AddItem(a.attributeTable, 0, 2, false) // Proportional height (larger weight - gets 2x space)

	a.leftSidebar = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.gameTree, 0, 1, false).
		AddItem(a.charTree, 0, 1, false).
		AddItem(a.charPane, 0, 2, false) // Give character pane more space

	// Main layout: horizontal split of tree (left, narrow) and log view (right)
	a.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.leftSidebar, 0, 1, false). // 1/3 of the width
		AddItem(a.logView, 0, 2, true)       // 2/3 of the width

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
			"Enter       - Select game/session/character\n" +
			"Space       - Expand/collapse tree nodes").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage(HELP_MODAL_ID)
		})

	// New game modal
	a.setupGameModal()

	// New Log Modal
	a.setupLogModal()

	// New Character Modal
	a.setupCharacterModal()

	// New Attribute Modal
	a.setupAttributeModal()

	// Create confirmation modal
	a.confirmModal = NewConfirmationModal()

	// Pages for modal overlay (must be created BEFORE notification setup)
	a.pages = tview.NewPages().
		AddPage(MAIN_PAGE_ID, mainContent, true, true).
		AddPage(HELP_MODAL_ID, a.helpModal, true, false).
		AddPage(GAME_MODAL_ID, a.gameModal, true, false).
		AddPage(LOG_MODAL_ID, a.logModal, true, false).
		AddPage(CHARACTER_MODAL_ID, a.characterModal, true, false).
		AddPage(ATTRIBUTE_MODAL_ID, a.attributeModal, true, false).
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
	a.refreshCharacterTree()
}

func (a *App) updateFooterHelp(context string) {
	var helpText string
	globalHelp := " [yellow]Tab/Shift-Tab[white] Navigate  [yellow]Ctrl+C[white] Quit  |  "
	switch context {
	case "logView":
		helpText = globalHelp + "[aqua::b]Session Logs[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New"
	case "gameTree":
		helpText = globalHelp + "[aqua::b]Games[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space[white] Select/Expand  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New"
	case "charTree":
		helpText = globalHelp + "[aqua::b]Characters[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Space/Enter[white] Select/Expand  [yellow]Ctrl+N[white] New"
	case "charInfo":
		helpText = globalHelp + "[aqua::b]Character Info[-::-] :: [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+D[white] Duplicate"
	case "attributeTable":
		helpText = globalHelp + "[aqua::b]Sheet[-::-] :: [yellow]↑/↓[white] Navigate  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New"
	}

	a.footer.SetText(helpText)

}

func (a *App) setupKeyBindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			a.Stop()
			return nil
		case tcell.KeyF1:
			a.showHelp()
			return nil
		case tcell.KeyTab:
			// When tabbing on the main view, capture it and set focus properly.
			// When tabbing elsewhere, send the event onward for it to be handled.
			currentFocus := a.GetFocus()
			switch currentFocus {
			case a.gameTree:
				a.SetFocus(a.logView)
				return nil
			case a.logView:
				a.SetFocus(a.charTree)
				return nil
			case a.charTree:
				a.SetFocus(a.charInfoView)
				return nil
			case a.charInfoView:
				a.SetFocus(a.attributeTable)
				return nil
			case a.attributeTable:
				a.SetFocus(a.gameTree)
				return nil
			}
			// If focus is not on main views (ie. in a modal), let Tab work normally
			return event
		case tcell.KeyBacktab:
			currentFocus := a.GetFocus()
			switch currentFocus {
			case a.gameTree:
				a.SetFocus(a.attributeTable)
				return nil
			case a.logView:
				a.SetFocus(a.gameTree)
				return nil
			case a.charTree:
				a.SetFocus(a.logView)
				return nil
			case a.charInfoView:
				a.SetFocus(a.charTree)
				return nil
			case a.attributeTable:
				a.SetFocus(a.charInfoView)
				return nil
			}
			// If focus is not on main views (ie. in a modal), let Tab work normally
			return event
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
				AddItem(a.logForm, 20, 2, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			100, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
}

func (a *App) setupCharacterModal() {
	// Create the form
	a.characterForm = NewCharacterForm()

	// Set up handlers
	a.characterForm.SetupHandlers(
		a.charHandler.HandleSave,
		a.charHandler.HandleCancel,
		a.charHandler.HandleDelete,
	)

	// Center the modal on screen
	a.characterModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(a.characterForm, 13, 1, true). // Fixed height for form
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)

	// Prevent mouse clicks outside modal
	//	a.characterModal.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	//		// Only pass through mouse events if they're within the modal form
	//		return action, nil
	//	})
}

func (a *App) setupAttributeModal() {
	// Create the form
	a.attributeForm = NewAttributeForm()

	// Set up handlers
	a.attributeForm.SetupHandlers(
		a.charHandler.HandleAttributeSave,
		a.charHandler.HandleAttributeCancel,
		a.charHandler.HandleAttributeDelete,
	)

	// Center the modal on screen
	a.attributeModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(a.attributeForm, 13, 1, true). // Fixed height for simple form
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
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
		a.logView.ScrollToEnd()

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

	case CHARACTER_SAVED, CHARACTER_DUPLICATED:
		// Clear form errors
		a.characterForm.ClearFieldErrors()

		if event == CHARACTER_SAVED {
			a.pages.HidePage(CHARACTER_MODAL_ID)
		}

		if event == CHARACTER_DUPLICATED {
			a.pages.HidePage(CONFIRM_MODAL_ID)
		}

		// Refresh character tree
		a.refreshCharacterTree()

		// Focus back on character info view
		a.SetFocus(a.charInfoView)

		// Display them
		a.displaySelectedCharacter()

		// Show success notification
		a.notification.ShowSuccess("Character saved successfully")

	case CHARACTER_DELETED:
		// Close modals
		a.pages.HidePage(CONFIRM_MODAL_ID)
		a.pages.HidePage(CHARACTER_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)

		// Clear character display
		a.selectedCharacter = nil
		a.charInfoView.Clear()
		a.attributeTable.Clear()

		// Refresh and focus
		a.refreshCharacterTree()
		a.SetFocus(a.charTree)

		// Show success notification
		a.notification.ShowSuccess("Character deleted successfully")

	case CHARACTER_CANCEL:
		// Clear form errors
		a.characterForm.ClearFieldErrors()

		// Close modal and focus tree
		a.pages.HidePage(CHARACTER_MODAL_ID)
		a.SetFocus(a.charTree)

	case CHARACTER_SHOW_NEW:
		// Show modal for creating new character
		a.characterForm.Reset()
		a.pages.ShowPage(CHARACTER_MODAL_ID)
		a.characterForm.SetTitle(" New Character ")
		a.SetFocus(a.characterForm)

	case CHARACTER_SHOW_EDIT:
		// Show modal for editing existing character
		if a.selectedCharacter != nil {
			a.characterForm.PopulateForEdit(a.selectedCharacter)
			a.pages.ShowPage(CHARACTER_MODAL_ID)
			a.characterForm.SetTitle(" Edit Character ")
			a.SetFocus(a.characterForm)
		}

	case ATTRIBUTE_SAVED:
		// Clear form errors
		a.attributeForm.ClearFieldErrors()

		// Close modal
		a.pages.HidePage(ATTRIBUTE_MODAL_ID)

		// Focus back on attribute table
		a.SetFocus(a.attributeTable)

		// Show success notification
		a.notification.ShowSuccess("Attribute saved successfully")

	case ATTRIBUTE_DELETED:
		// Close modals
		a.pages.HidePage(CONFIRM_MODAL_ID)
		a.pages.HidePage(ATTRIBUTE_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)

		// Focus back on attribute table
		a.SetFocus(a.attributeTable)

		// Show success notification
		a.notification.ShowSuccess("Attribute deleted successfully")

	case ATTRIBUTE_CANCEL:
		// Clear form errors
		a.attributeForm.ClearFieldErrors()

		// Close modal and focus attribute table
		a.pages.HidePage(ATTRIBUTE_MODAL_ID)
		a.SetFocus(a.attributeTable)

	case ATTRIBUTE_SHOW_NEW:
		// Show modal for creating new attribute
		a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
		a.SetFocus(a.attributeForm)

	case ATTRIBUTE_SHOW_EDIT:
		// Show modal for editing existing attribute
		a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
		a.SetFocus(a.attributeForm)

	case CONFIRM_SHOW:
		// Show confirmation modal
		a.pages.ShowPage(CONFIRM_MODAL_ID)

	case CONFIRM_CANCEL:
		// Reusable across all confirmation cancellations
		a.pages.HidePage(CONFIRM_MODAL_ID)
	}
}

func (a *App) loadLogsForSelectedGameEntry() {
	// Clear the highlight
	a.logView.Highlight()

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

func (a *App) refreshCharacterTree() {

	root := a.charTree.GetRoot()
	if root == nil {
		root = tview.NewTreeNode("Systems").SetColor(tcell.ColorYellow).SetSelectable(false)
		a.charTree.SetRoot(root).SetCurrentNode(root)
	}
	root.ClearChildren()

	// Load all characters from database
	charsBySystem, err := a.charHandler.LoadCharacters()
	if err != nil {
		// Show error in tree
		errorNode := tview.NewTreeNode("Error loading characters: " + err.Error()).
			SetColor(tcell.ColorRed)
		root.AddChild(errorNode)
		return
	}

	if len(charsBySystem) == 0 {
		// No characters yet
		placeholder := tview.NewTreeNode("(No characters yet - press Ctrl+H to create one)").
			SetColor(tcell.ColorGray)
		root.AddChild(placeholder)
		return
	}

	// Add each system to the tree, then the characters for that system
	for system, chars := range charsBySystem {
		// Create system node
		systemNode := tview.NewTreeNode(system).
			SetColor(tcell.ColorLime).
			SetSelectable(true)
		root.AddChild(systemNode)

		// Add character nodes under the system
		for _, c := range chars {
			charNode := tview.NewTreeNode(c.Name).
				SetReference(c).
				SetColor(tcell.ColorAqua).
				SetSelectable(true)
			systemNode.AddChild(charNode)

			// If this character is currently selected, then select them in the tree
			if a.selectedCharacter != nil && c.ID == a.selectedCharacter.ID {
				a.charTree.SetCurrentNode(charNode)
			}
		}
	}

	root.SetExpanded(true)
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

func (a *App) displaySelectedCharacter() {
	// Display character info
	a.charHandler.DisplayCharacterInfo(a.selectedCharacter)

	// Resize the pane to fit the character info provided
	a.charInfoView.ScrollToBeginning()

	// Load and display attributes
	a.charHandler.LoadAndDisplayAttributes(a.selectedCharacter.ID)
}

func (a *App) loadSelectedCharacter() {
	// Get the reference from the node
	if ref := a.charTree.GetCurrentNode().GetReference(); ref != nil {
		// Check if it's a character (not a system node)
		if char, ok := ref.(*character.Character); ok {

			// Update selected character
			a.selectedCharacter = char

			// Display them
			a.displaySelectedCharacter()
		}
	}
}

// highlightPreviousLog navigates to the previous log entry in the log view
func (a *App) highlightPreviousLog() {
	// Log view uses regions which are just the log.ID as a string.

	if len(a.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := a.logView.GetHighlights()

	// If nothing is highlighted, highlight the last region
	if len(currentHighlights) == 0 {
		a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[len(a.loadedLogIDs)-1]))
		a.logView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]
	for i, logID := range a.loadedLogIDs {
		// Convert the log ID to a string for all the following logic
		regionID := fmt.Sprintf("%d", logID)
		if regionID == currentRegion {
			// Move to previous region (wrap around to end if at beginning)
			if i > 0 {
				a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[i-1]))
			} else {
				a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[len(a.loadedLogIDs)-1]))
			}
			a.logView.ScrollToHighlight()
			return
		}
	}
}

// highlightNextLog navigates to the next log entry in the log view
func (a *App) highlightNextLog() {
	// Log view uses regions which are just the log.ID as a string.

	if len(a.loadedLogIDs) == 0 {
		return
	}

	currentHighlights := a.logView.GetHighlights()

	// If nothing is highlighted, highlight the first region
	if len(currentHighlights) == 0 {
		a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[0]))
		a.logView.ScrollToHighlight()
		return
	}

	// Find the current region index
	currentRegion := currentHighlights[0]
	for i, logID := range a.loadedLogIDs {
		// Convert the log ID to a string for all the following logic
		regionID := fmt.Sprintf("%d", logID)
		if regionID == currentRegion {
			// Move to previous region (wrap around to end if at beginning)
			if i < len(a.loadedLogIDs)-1 {
				a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[i+1]))
			} else {
				a.logView.Highlight(fmt.Sprintf("%d", a.loadedLogIDs[0]))
			}
			a.logView.ScrollToHighlight()
			return
		}
	}

}
