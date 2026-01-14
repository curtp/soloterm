// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
	"fmt"
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
	gameTree              *tview.TreeView
	charTree              *tview.TreeView
	charInfoView          *tview.TextView
	attributeTable        *tview.Table
	logView               *tview.TextView
	aboutModal            *tview.Modal
	confirmModal          *ConfirmationModal
	footer                *tview.TextView
	gameModal             *tview.Flex
	gameForm              *GameForm
	logModal              *tview.Flex
	logModalContent       *tview.Flex // Container with border that holds form + help
	logForm               *LogForm
	characterModal        *tview.Flex
	characterForm         *CharacterForm
	attributeModal        *tview.Flex
	attributeModalContent *tview.Flex // Container with border that holds form + help
	attributeForm         *AttributeForm
	notification          *Notification
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
			"Version 0.13").
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
			a.showAbout()
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

func (a *App) showAbout() {
	a.pages.ShowPage(ABOUT_MODAL_ID)
}

func (a *App) HandleEvent(event Event) {
	switch event.Action() {
	case GAME_SAVED:
		if e, ok := event.(*GameSavedEvent); ok {
			a.handleGameSaved(e)
		}
	case GAME_CANCEL:
		if e, ok := event.(*GameCancelledEvent); ok {
			a.handleGameCancel(e)
		}
	case GAME_DELETE_CONFIRM:
		if e, ok := event.(*GameDeleteConfirmEvent); ok {
			a.handleGameDeleteConfirm(e)
		}
	case GAME_DELETED:
		if e, ok := event.(*GameDeletedEvent); ok {
			a.handleGameDeleted(e)
		}
	case GAME_DELETE_FAILED:
		if e, ok := event.(*GameDeleteFailedEvent); ok {
			a.handleGameDeleteFailed(e)
		}
	case GAME_SHOW_EDIT:
		if e, ok := event.(*GameShowEditEvent); ok {
			a.handleGameShowEdit(e)
		}
	case GAME_SHOW_NEW:
		if e, ok := event.(*GameShowNewEvent); ok {
			a.handleGameShowNew(e)
		}
	case LOG_SAVED:
		if e, ok := event.(*LogSavedEvent); ok {
			a.handleLogSaved(e)
		}
	case LOG_CANCEL:
		if e, ok := event.(*LogCancelledEvent); ok {
			a.handleLogCancel(e)
		}
	case LOG_DELETE_CONFIRM:
		if e, ok := event.(*LogDeleteConfirmEvent); ok {
			a.handleLogDeleteConfirm(e)
		}
	case LOG_DELETED:
		if e, ok := event.(*LogDeletedEvent); ok {
			a.handleLogDeleted(e)
		}
	case LOG_DELETE_FAILED:
		if e, ok := event.(*LogDeleteFailedEvent); ok {
			a.handleLogDeleteFailed(e)
		}
	case LOG_SHOW_NEW:
		if e, ok := event.(*LogShowNewEvent); ok {
			a.handleLogShowNew(e)
		}
	case LOG_SHOW_EDIT:
		if e, ok := event.(*LogShowEditEvent); ok {
			a.handleLogShowEdit(e)
		}
	case CHARACTER_SAVED:
		if e, ok := event.(*CharacterSavedEvent); ok {
			a.handleCharacterSaved(e)
		}
	case CHARACTER_CANCEL:
		if e, ok := event.(*CharacterCancelledEvent); ok {
			a.handleCharacterCancel(e)
		}
	case CHARACTER_DELETE_CONFIRM:
		if e, ok := event.(*CharacterDeleteConfirmEvent); ok {
			a.handleCharacterDeleteConfirm(e)
		}
	case CHARACTER_DELETED:
		if e, ok := event.(*CharacterDeletedEvent); ok {
			a.handleCharacterDeleted(e)
		}
	case CHARACTER_DELETE_FAILED:
		if e, ok := event.(*CharacterDeleteFailedEvent); ok {
			a.handleCharacterDeleteFailed(e)
		}
	case CHARACTER_DUPLICATE_CONFIRM:
		if e, ok := event.(*CharacterDuplicateConfirmEvent); ok {
			a.handleCharacterDuplicateConfirm(e)
		}
	case CHARACTER_DUPLICATED:
		if e, ok := event.(*CharacterDuplicatedEvent); ok {
			a.handleCharacterDuplicated(e)
		}
	case CHARACTER_DUPLICATE_FAILED:
		if e, ok := event.(*CharacterDuplicateFailedEvent); ok {
			a.handleCharacterDuplicateFailed(e)
		}
	case CHARACTER_SHOW_NEW:
		if e, ok := event.(*CharacterShowNewEvent); ok {
			a.handleCharacterShowNew(e)
		}
	case CHARACTER_SHOW_EDIT:
		if e, ok := event.(*CharacterShowEditEvent); ok {
			a.handleCharacterShowEdit(e)
		}
	case ATTRIBUTE_SAVED:
		if e, ok := event.(*AttributeSavedEvent); ok {
			a.handleAttributeSaved(e)
		}
	case ATTRIBUTE_CANCEL:
		if e, ok := event.(*AttributeCancelledEvent); ok {
			a.handleAttributeCancel(e)
		}
	case ATTRIBUTE_DELETE_CONFIRM:
		if e, ok := event.(*AttributeDeleteConfirmEvent); ok {
			a.handleAttributeDeleteConfirm(e)
		}
	case ATTRIBUTE_DELETED:
		if e, ok := event.(*AttributeDeletedEvent); ok {
			a.handleAttributeDeleted(e)
		}
	case ATTRIBUTE_DELETE_FAILED:
		if e, ok := event.(*AttributeDeleteFailedEvent); ok {
			a.handleAttributeDeleteFailed(e)
		}
	case ATTRIBUTE_SHOW_NEW:
		if e, ok := event.(*AttributeShowNewEvent); ok {
			a.handleAttributeShowNew(e)
		}
	case ATTRIBUTE_SHOW_EDIT:
		if e, ok := event.(*AttributeShowEditEvent); ok {
			a.handleAttributeShowEdit(e)
		}
	}
}

func (a *App) handleGameSaved(e *GameSavedEvent) {
	a.gameForm.Reset()
	a.pages.HidePage(GAME_MODAL_ID)
	a.selectedGame = e.Game
	a.gameViewHelper.Refresh()
	a.SetFocus(a.gameTree)
	a.notification.ShowSuccess("Game saved successfully")
}

func (a *App) handleGameCancel(_ *GameCancelledEvent) {
	a.pages.HidePage(GAME_MODAL_ID)
	a.SetFocus(a.gameTree)
}

func (a *App) handleGameDeleteConfirm(e *GameDeleteConfirmEvent) {
	// Capture focus for restoration on cancel
	returnFocus := a.GetFocus()

	// Configure confirmation modal
	a.confirmModal.Configure(
		"Are you sure you want to delete this game and all associated log entries?\n\nThis action cannot be undone.",
		func() {
			// On confirm, call handler method to perform deletion
			a.gameHandler.ConfirmDelete(e.Game.ID)
		},
		func() {
			// On cancel, restore focus
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	// Show the confirmation modal
	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleGameDeleted(_ *GameDeletedEvent) {
	// Close modals
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(GAME_MODAL_ID)

	// Reset log view to default state
	a.logView.Clear()
	a.logView.SetTitle(" Select Game To View ")

	// Update state
	a.selectedGame = nil
	a.selectedLog = nil

	// Refresh and focus
	a.gameViewHelper.Refresh()
	a.SetFocus(a.gameTree)

	// Show success notification
	a.notification.ShowSuccess("Game deleted successfully")
}

func (a *App) handleGameDeleteFailed(e *GameDeleteFailedEvent) {
	// Close confirmation modal
	a.pages.HidePage(CONFIRM_MODAL_ID)

	// Show error notification
	a.notification.ShowError("Error deleting game: " + e.Error.Error())
}

func (a *App) handleGameShowEdit(e *GameShowEditEvent) {
	if e.Game == nil {
		a.notification.ShowError("Please select a game to edit")
		return
	}

	a.selectedGame = e.Game
	a.logViewHelper.Refresh()
	a.gameForm.PopulateForEdit(a.selectedGame)
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleGameShowNew(_ *GameShowNewEvent) {
	a.gameForm.Reset()
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleLogSaved(e *LogSavedEvent) {
	a.logForm.ClearFieldErrors()
	a.pages.HidePage(LOG_MODAL_ID)
	a.selectedLog = e.Log
	a.logViewHelper.Refresh()
	a.gameViewHelper.Refresh()
	a.SetFocus(a.logView)
	a.logView.ScrollToHighlight()
	a.notification.ShowSuccess("Log saved successfully")
}

func (a *App) handleLogCancel(_ *LogCancelledEvent) {
	a.logForm.ClearFieldErrors()
	a.pages.HidePage(LOG_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.SetFocus(a.logView)
	a.logView.ScrollToHighlight()
}

func (a *App) handleLogDeleteConfirm(e *LogDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this log entry?\n\nThis action cannot be undone.",
		func() {
			a.logHandler.ConfirmDelete(e.Log.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleLogDeleted(_ *LogDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(LOG_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.logViewHelper.Refresh()
	a.gameViewHelper.Refresh()
	a.SetFocus(a.logView)
	a.notification.ShowSuccess("Log entry deleted successfully")
}

func (a *App) handleLogDeleteFailed(e *LogDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Error deleting log entry: " + e.Error.Error())
}

func (a *App) handleLogShowNew(_ *LogShowNewEvent) {
	a.logModalContent.SetTitle(" New Log ")
	a.logForm.Reset(a.selectedGame.ID)
	a.pages.ShowPage(LOG_MODAL_ID)
	a.SetFocus(a.logForm)
}

func (a *App) handleLogShowEdit(e *LogShowEditEvent) {
	a.selectedLog = e.Log
	a.logModalContent.SetTitle(" Edit Log ")
	a.logForm.PopulateForEdit(e.Log)
	a.pages.ShowPage(LOG_MODAL_ID)
	a.SetFocus(a.logForm)
}

func (a *App) handleCharacterSaved(e *CharacterSavedEvent) {
	a.characterForm.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.selectedCharacter = e.Character
	a.characterViewHelper.RefreshTree()
	a.characterViewHelper.RefreshDisplay()
	a.SetFocus(a.characterViewHelper.ReturnFocus)
	a.notification.ShowSuccess("Character saved successfully")
}

func (a *App) handleCharacterCancel(_ *CharacterCancelledEvent) {
	a.characterForm.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.SetFocus(a.characterViewHelper.ReturnFocus)
}

func (a *App) handleCharacterDeleteConfirm(e *CharacterDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this character?\n\nThis will also delete all associated attributes.",
		func() {
			a.charHandler.ConfirmDelete(e.Character.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleCharacterDeleted(_ *CharacterDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.selectedCharacter = nil
	a.charInfoView.Clear()
	a.attributeTable.Clear()
	a.characterViewHelper.RefreshTree()
	a.SetFocus(a.characterViewHelper.ReturnFocus)
	a.notification.ShowSuccess("Character deleted successfully")
}

func (a *App) handleCharacterDeleteFailed(e *CharacterDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete character: " + e.Error.Error())
}

func (a *App) handleCharacterDuplicateConfirm(e *CharacterDuplicateConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to duplicate this character and their sheet?",
		func() {
			a.charHandler.ConfirmDuplicate(e.Character.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
		"Duplicate",
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleCharacterDuplicated(e *CharacterDuplicatedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.selectedCharacter = e.Character
	a.characterViewHelper.RefreshTree()
	a.characterViewHelper.RefreshDisplay()
	a.SetFocus(a.characterViewHelper.ReturnFocus)
	a.notification.ShowSuccess("Character duplicated successfully")
}

func (a *App) handleCharacterDuplicateFailed(e *CharacterDuplicateFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to duplicate the character: " + e.Error.Error())
}

func (a *App) handleCharacterShowNew(_ *CharacterShowNewEvent) {
	a.characterForm.Reset()
	a.pages.ShowPage(CHARACTER_MODAL_ID)
	a.characterForm.SetTitle(" New Character ")
	a.SetFocus(a.characterForm)
}

func (a *App) handleCharacterShowEdit(e *CharacterShowEditEvent) {
	a.selectedCharacter = e.Character
	a.characterForm.PopulateForEdit(e.Character)
	a.pages.ShowPage(CHARACTER_MODAL_ID)
	a.characterForm.SetTitle(" Edit Character ")
	a.SetFocus(a.characterForm)
}

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
	a.attributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.characterViewHelper.RefreshDisplay()
	a.characterViewHelper.selectAttribute(e.Attribute.ID)
	a.SetFocus(a.attributeTable)
	a.notification.ShowSuccess("Attribute saved successfully")
}

func (a *App) handleAttributeCancel(_ *AttributeCancelledEvent) {
	a.attributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeTable)
}

func (a *App) handleAttributeDeleteConfirm(e *AttributeDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this attribute?",
		func() {
			a.charHandler.ConfirmAttributeDelete(e.Attribute.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleAttributeDeleted(_ *AttributeDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.characterViewHelper.RefreshDisplay()
	a.SetFocus(a.attributeTable)
	a.notification.ShowSuccess("Attribute deleted successfully")
}

func (a *App) handleAttributeDeleteFailed(e *AttributeDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete attribute: " + e.Error.Error())
}

func (a *App) handleAttributeShowNew(e *AttributeShowNewEvent) {
	a.attributeModalContent.SetTitle(" New Attribute ")
	a.attributeForm.Reset(a.selectedCharacter.ID)

	// If there's a selected attribute, use its group and position as defaults
	if e.SelectedAttribute != nil {
		a.attributeForm.groupField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.Group))
		a.attributeForm.positionField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.PositionInGroup+1))
	}

	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeForm)
}

func (a *App) handleAttributeShowEdit(e *AttributeShowEditEvent) {
	a.attributeModalContent.SetTitle(" Edit Attribute ")
	a.attributeForm.PopulateForEdit(e.Attribute)
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeForm)
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
		a.SetFocus(a.logView)
		a.logView.ScrollToHighlight()

	case GAME_SHOW_NEW:
		// Show modal for creating new game
		a.gameForm.Reset()
		a.pages.ShowPage(GAME_MODAL_ID)
		a.SetFocus(a.gameForm)

	case GAME_SHOW_EDIT:
		a.logViewHelper.Refresh()

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
		//a.SetFocus(a.charInfoView)
		// Set the focus back to where the user was when they started
		a.SetFocus(a.characterViewHelper.ReturnFocus)

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
		// Set the focus back to where the user was when they started
		a.SetFocus(a.characterViewHelper.ReturnFocus)
		//		a.SetFocus(a.charTree)

		// Show success notification
		a.notification.ShowSuccess("Character deleted successfully")

	case CHARACTER_CANCEL:
		// Clear form errors
		a.characterForm.ClearFieldErrors()

		// Close modal and focus tree
		a.pages.HidePage(CHARACTER_MODAL_ID)
		// Set the focus back to where the user was when they started
		a.SetFocus(a.characterViewHelper.ReturnFocus)
		//		a.SetFocus(a.charTree)

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

	}
}
