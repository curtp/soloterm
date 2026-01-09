// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
	"soloterm/database"
	"soloterm/domain/character"
	"soloterm/domain/game"
	"soloterm/domain/log"

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
	ABOUT_MODAL_ID     string = "about"
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

	// View helpers
	gameViewHelper      *GameViewHelper
	logViewHelper       *LogViewHelper
	characterViewHelper *CharacterViewHelper

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
	aboutModal     *tview.Modal
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

	// Initialize forms
	app.gameForm = NewGameForm()
	app.logForm = NewLogForm()
	app.characterForm = NewCharacterForm()
	app.attributeForm = NewAttributeForm()

	// Initialize handlers (after app is created so they can reference it)
	app.gameHandler = NewGameHandler(app)
	app.logHandler = NewLogHandler(app)
	app.charHandler = NewCharacterHandler(app)

	// Initialize view helpers
	app.gameViewHelper = NewGameViewHelper(app)
	app.logViewHelper = NewLogViewHelper(app)
	app.characterViewHelper = NewCharacterViewHelper(app)

	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// Left pane: Game/Session tree
	a.gameViewHelper.Setup()

	// Setup the character view area
	a.characterViewHelper.Setup()

	// Setup the log view area of the app and everything it needs to work
	a.logViewHelper.Setup()

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
	a.aboutModal = tview.NewModal().
		SetText("SoloTerm - Solo RPG Session Logger\n\n" +
			"By Squidhead Games\n" +
			"https://squidhead-games.itch.io\n\n" +
			"Version 0.0.1").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage(ABOUT_MODAL_ID)
		})

	// Create confirmation modal
	a.confirmModal = NewConfirmationModal()

	// Pages for modal overlay (must be created BEFORE notification setup)
	a.pages = tview.NewPages().
		AddPage(MAIN_PAGE_ID, mainContent, true, true).
		AddPage(ABOUT_MODAL_ID, a.aboutModal, true, false).
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
	a.gameViewHelper.Refresh()
	a.characterViewHelper.RefreshTree()
}

func (a *App) updateFooterHelp(helpText string) {
	globalHelp := " [yellow]Tab/Shift+Tab[white] Navigate  [yellow]Ctrl+C[white] Quit  |  "
	a.footer.SetText(globalHelp + helpText)
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

func (a *App) showHelp() {
	a.pages.ShowPage(ABOUT_MODAL_ID)
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
		a.gameViewHelper.Refresh()
		a.SetFocus(a.gameTree)

		// Show success notification
		a.notification.ShowSuccess("Game deleted successfully")

	case GAME_SAVED:
		// Clear form errors
		a.gameForm.ClearFieldErrors()

		// Close modal
		a.pages.HidePage(GAME_MODAL_ID)

		// Refresh and focus
		a.gameViewHelper.Refresh()
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
		// a.logView.Highlight()
		a.logViewHelper.Refresh()
		a.gameViewHelper.Refresh()
		a.SetFocus(a.logView)
		a.logView.ScrollToHighlight()

		// Show success notification
		a.notification.ShowSuccess("Log saved successfully")

	case LOG_DELETED:
		// Close modals
		a.pages.HidePage(CONFIRM_MODAL_ID)
		a.pages.HidePage(LOG_MODAL_ID)
		a.pages.SwitchToPage(MAIN_PAGE_ID)

		// Refresh and focus
		a.logViewHelper.Refresh()
		a.gameViewHelper.Refresh()
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
		a.characterViewHelper.RefreshTree()

		// Focus back on character info view
		a.SetFocus(a.charInfoView)

		// Display them
		a.characterViewHelper.RefreshDisplay()

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
		a.characterViewHelper.RefreshTree()
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

		// Restore focus to where user was before modal
		if a.confirmModal.ReturnFocus != nil {
			a.SetFocus(a.confirmModal.ReturnFocus)
			a.confirmModal.ReturnFocus = nil // Clear for next use
		}
	}
}
