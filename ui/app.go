// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
	"fmt"
	"slices"
	"soloterm/config"
	"soloterm/database"
	"soloterm/domain/character"
	"soloterm/domain/game"
	"soloterm/domain/session"
	"soloterm/domain/tag"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	GAME_MODAL_ID      string = "gameModal"
	TAG_MODAL_ID       string = "tagModal"
	CHARACTER_MODAL_ID string = "characterModal"
	ATTRIBUTE_MODAL_ID string = "attributeModal"
	CONFIRM_MODAL_ID   string = "confirm"
	MAIN_PAGE_ID       string = "main"
	ABOUT_MODAL_ID     string = "about"
	SESSION_MODAL_ID      string = "sessionModal"
	SESSION_HELP_MODAL_ID string = "sessionHelpModal"
)

type GameState struct {
	GameID    *int64
	SessionID *int64
}

type App struct {
	*tview.Application

	// Dependencies
	db             *database.DBStore
	cfg            *config.Config
	gameService    *game.Service
	sessionService *session.Service
	charService    *character.Service
	tagService     *tag.Service

	// View helpers
	gameView      *GameView
	tagView       *TagView
	sessionView   *SessionView
	characterView *CharacterView

	// Application State
	selectedGame *game.Game // Currently selected/active game

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
	aboutModal            *tview.Modal
	confirmModal          *ConfirmationModal
	footer                *tview.TextView
	gameModal             *tview.Flex
	gameForm              *GameForm
	tagModal              *tview.Flex
	tagModalContent       *tview.Flex
	tagList               *tview.List
	tagTable              *tview.Table
	characterModal        *tview.Flex
	characterForm         *CharacterForm
	attributeModal        *tview.Flex
	attributeModalContent *tview.Flex // Container with border that holds form + help
	attributeForm         *AttributeForm
	notification          *Notification
	sessionModal          *tview.Flex
}

func NewApp(db *database.DBStore, cfg *config.Config) *App {
	app := &App{
		Application:    tview.NewApplication(),
		db:             db,
		cfg:            cfg,
		gameService:    game.NewService(game.NewRepository(db)),
		charService:    character.NewService(character.NewRepository(db)),
		tagService:     tag.NewService(session.NewRepository(db)),
		sessionService: session.NewService(session.NewRepository(db)),
	}

	// Initialize forms
	app.gameForm = NewGameForm()
	app.characterForm = NewCharacterForm()
	app.attributeForm = NewAttributeForm()

	// Initialize views
	app.gameView = NewGameView(app)
	app.sessionView = NewSessionView(app, app.sessionService)
	app.tagView = NewTagView(app)
	app.characterView = NewCharacterView(app)

	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// Left pane: Game/Session tree
	a.gameView.Setup()

	// Setup the character view area
	a.characterView.Setup()

	// Setup the tag view
	a.tagView.Setup()

	// Setup the Session view
	a.sessionView.Setup()

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Character pane combining info and attributes
	a.charPane = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.charInfoView, 4, 1, false).  // Proportional height (smaller weight)
		AddItem(a.attributeTable, 0, 2, false) // Proportional height (larger weight - gets 2x space)
	a.charPane.SetBorder(true).
		SetTitle(" [::b]Character Sheet (Ctrl+S) ").
		SetTitleAlign(tview.AlignLeft)

	a.leftSidebar = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.gameTree, 0, 1, false).
		AddItem(a.charTree, 0, 1, false).
		AddItem(a.charPane, 0, 2, false) // Give character pane more space

	// Main layout: horizontal split of tree (left, narrow) and log view (right)
	a.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.leftSidebar, 0, 1, false).        // 1/3 of the width
		AddItem(a.sessionView.TextArea, 0, 2, true) // 2/3 of the width

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
			"Version 1.0.5\n\n" +
			"Lonelog by Loreseed Workshop\n" +
			"https://zeruhur.itch.io/lonelog").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage(ABOUT_MODAL_ID)
		})

	// Create confirmation modal
	a.confirmModal = NewConfirmationModal()

	// Pages for modal overlay (must be created BEFORE notification setup)
	// Note: Pages added later appear on top of earlier pages
	a.pages = tview.NewPages().
		AddPage(MAIN_PAGE_ID, mainContent, true, true).
		AddPage(ABOUT_MODAL_ID, a.aboutModal, true, false).
		AddPage(GAME_MODAL_ID, a.gameModal, true, false).
		AddPage(CHARACTER_MODAL_ID, a.characterModal, true, false).
		AddPage(ATTRIBUTE_MODAL_ID, a.attributeModal, true, false).
		AddPage(SESSION_MODAL_ID, a.sessionView.Modal, true, false).
		AddPage(SESSION_HELP_MODAL_ID, a.sessionView.HelpModal, true, false).
		AddPage(TAG_MODAL_ID, a.tagModal, true, false). // Tag modal on top of forms
		AddPage(CONFIRM_MODAL_ID, a.confirmModal, true, false) // Confirm always on top
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
	a.gameView.Refresh()
	a.characterView.RefreshTree()
}

func (a *App) updateFooterHelp(helpText string) {
	globalHelp := " [yellow]Tab/Shift+Tab[white] Navigate  [yellow]Ctrl+Q[white] Quit  |  "
	a.footer.SetText(globalHelp + helpText)
}

func (a *App) SetModalHelpMessage(form sharedui.DataForm) {
	editing := form.GetButtonCount() == 3
	deleteHelp := "[yellow]Ctrl+D[white] Delete"
	actionMsg := "Add"
	if editing {
		actionMsg = "Edit"
	}
	helpMsg := "[aqua::b]" + actionMsg + "[-::-] :: [yellow]Ctrl+S[white] Save  [yellow]Esc[white] Cancel "
	if form.GetButtonCount() == 3 {
		helpMsg += " " + deleteHelp
	}
	a.updateFooterHelp(helpMsg)
}

func (a *App) setupKeyBindings() {
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switchable := []tview.Primitive{a.gameTree, a.sessionView.TextArea, a.charTree, a.attributeTable}
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.Stop()
			return nil
		case tcell.KeyCtrlG:
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.gameTree)
				return nil
			}
		case tcell.KeyCtrlC:
			// Always consume Ctrl+C to prevent terminal signal handling
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.charTree)
			}
			return nil
		case tcell.KeyCtrlS:
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.attributeTable)
				return nil
			}
		case tcell.KeyCtrlL:
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.sessionView.TextArea)
				return nil
			}
		case tcell.KeyF1:
			a.showAbout()
			return nil
		case tcell.KeyTab:
			// When tabbing on the main view, capture it and set focus properly.
			// When tabbing elsewhere, send the event onward for it to be handled.
			currentFocus := a.GetFocus()
			switch currentFocus {
			case a.gameTree:
				a.SetFocus(a.sessionView.TextArea)
				return nil
			case a.sessionView.TextArea:
				a.SetFocus(a.charTree)
				return nil
			case a.charTree:
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
			case a.sessionView.TextArea:
				a.SetFocus(a.gameTree)
				return nil
			case a.charTree:
				a.SetFocus(a.sessionView.TextArea)
				return nil
			case a.attributeTable:
				a.SetFocus(a.charTree)
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

func (a *App) GetSelectedGameState() *GameState {
	return a.gameView.GetCurrentSelection()
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
	case GAME_SELECTED:
		if e, ok := event.(*GameSelectedEvent); ok {
			a.handleGameSelected(e)
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
	case TAG_SELECTED:
		if e, ok := event.(*TagSelectedEvent); ok {
			a.handleTagSelected(e)
		}
	case TAG_CANCEL:
		if e, ok := event.(*TagCancelledEvent); ok {
			a.handleTagCancelled(e)
		}
	case TAG_SHOW:
		if e, ok := event.(*TagShowEvent); ok {
			a.handleTagShow(e)
		}
	case SESSION_SHOW_NEW:
		if e, ok := event.(*SessionShowNewEvent); ok {
			a.handleSessionShowNew(e)
		}
	case SESSION_CANCEL:
		if e, ok := event.(*SessionCancelledEvent); ok {
			a.handleSessionCancelled(e)
		}
	case SESSION_SELECTED:
		if e, ok := event.(*SessionSelectedEvent); ok {
			a.handleSessionSelected(e)
		}
	case SESSION_SAVED:
		if e, ok := event.(*SessionSavedEvent); ok {
			a.handleSessionSaved(e)
		}
	case SESSION_SHOW_EDIT:
		if e, ok := event.(*SessionShowEditEvent); ok {
			a.handleSessionShowEdit(e)
		}
	case SESSION_DELETE_CONFIRM:
		if e, ok := event.(*SessionDeleteConfirmEvent); ok {
			a.handleSessionDeleteConfirm(e)
		}
	case SESSION_DELETED:
		if e, ok := event.(*SessionDeletedEvent); ok {
			a.handleSessionDeleted(e)
		}
	case SESSION_DELETE_FAILED:
		if e, ok := event.(*SessionDeleteFailedEvent); ok {
			a.handleSessionDeleteFailed(e)
		}
	case SESSION_SHOW_HELP:
		if e, ok := event.(*SessionShowHelpEvent); ok {
			a.handleSessionShowHelp(e)
		}
	case SESSION_CLOSE_HELP:
		if e, ok := event.(*SessionCloseHelpEvent); ok {
			a.handleSessionCloseHelp(e)
		}
	}

}

func (a *App) handleGameSaved(e *GameSavedEvent) {
	a.gameForm.Reset()
	a.pages.HidePage(GAME_MODAL_ID)
	a.gameView.SelectGame(&e.Game.ID)
	a.gameView.Refresh()
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
		"Are you sure you want to delete this game and all associated sessions?\n\nThis action cannot be undone.",
		func() {
			// On confirm, call handler method to perform deletion
			a.gameView.ConfirmDelete(e.GameID)
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

	// Refresh and focus
	a.gameView.Refresh()
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

	a.gameForm.PopulateForEdit(e.Game)
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleGameShowNew(_ *GameShowNewEvent) {
	a.gameForm.Reset()
	a.pages.ShowPage(GAME_MODAL_ID)
	a.SetFocus(a.gameForm)
}

func (a *App) handleGameSelected(_ *GameSelectedEvent) {
}

func (a *App) handleCharacterSaved(e *CharacterSavedEvent) {
	a.characterForm.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.characterView.RefreshTree()
	a.characterView.SelectCharacter(e.Character.ID)
	a.characterView.RefreshDisplay()
	a.SetFocus(a.characterView.ReturnFocus)
	a.notification.ShowSuccess("Character saved successfully")
}

func (a *App) handleCharacterCancel(_ *CharacterCancelledEvent) {
	a.characterForm.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.SetFocus(a.characterView.ReturnFocus)
}

func (a *App) handleCharacterDeleteConfirm(e *CharacterDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this character?\n\nThis will also delete their sheet.",
		func() {
			a.characterView.ConfirmDelete(e.Character.ID)
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
	a.charInfoView.Clear()
	a.attributeTable.Clear()
	a.characterView.RefreshTree()
	a.SetFocus(a.characterView.ReturnFocus)
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
			a.characterView.ConfirmDuplicate(e.Character.ID)
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
	a.characterView.RefreshTree()
	a.characterView.RefreshDisplay()
	a.characterView.SelectCharacter(e.Character.ID)
	a.SetFocus(a.characterView.ReturnFocus)
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
	a.characterForm.PopulateForEdit(e.Character)
	a.pages.ShowPage(CHARACTER_MODAL_ID)
	a.characterForm.SetTitle(" Edit Character ")
	a.SetFocus(a.characterForm)
}

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
	a.attributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.characterView.RefreshDisplay()
	a.characterView.selectAttribute(e.Attribute.ID)
	a.SetFocus(a.attributeTable)
	a.notification.ShowSuccess("Entry saved successfully")
}

func (a *App) handleAttributeCancel(_ *AttributeCancelledEvent) {
	a.attributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeTable)
}

func (a *App) handleAttributeDeleteConfirm(e *AttributeDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this entry?",
		func() {
			a.characterView.ConfirmAttributeDelete(e.Attribute.ID)
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
	a.characterView.RefreshDisplay()
	a.SetFocus(a.attributeTable)
	a.notification.ShowSuccess("Entry deleted successfully")
}

func (a *App) handleAttributeDeleteFailed(e *AttributeDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete entry: " + e.Error.Error())
}

func (a *App) handleAttributeShowNew(e *AttributeShowNewEvent) {
	a.attributeModalContent.SetTitle(" New Entry ")
	a.attributeForm.Reset(e.CharacterID)

	// If there's a selected attribute, use its group and position as defaults
	if e.SelectedAttribute != nil {
		a.attributeForm.groupField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.Group))
		a.attributeForm.positionField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.PositionInGroup+1))
	}

	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeForm)
}

func (a *App) handleAttributeShowEdit(e *AttributeShowEditEvent) {
	a.attributeModalContent.SetTitle(" Edit Entry ")
	a.attributeForm.PopulateForEdit(e.Attribute)
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeForm)
}

func (a *App) handleTagSelected(e *TagSelectedEvent) {
	a.pages.HidePage(TAG_MODAL_ID)
	a.sessionView.InsertAtCursor(e.TagType.Template)
	a.SetFocus(a.sessionView.TextArea)
}

func (a *App) handleTagCancelled(e *TagCancelledEvent) {
	a.pages.HidePage(TAG_MODAL_ID)
}

func (a *App) handleTagShow(e *TagShowEvent) {
	// Store current focus so we can restore it after tag selection
	a.tagView.returnFocus = a.GetFocus()
	a.tagView.Refresh()
	a.pages.ShowPage(TAG_MODAL_ID)
	a.SetFocus(a.tagTable)
}

func (a *App) handleSessionShowNew(e *SessionShowNewEvent) {
	a.sessionView.Modal.SetTitle(" New Session ")
	selectedGameState := a.GetSelectedGameState()
	if selectedGameState == nil {
		a.notification.ShowWarning("Select a game before adding a session.")
		return
	}
	a.sessionView.Form.Reset(*selectedGameState.GameID)
	a.pages.ShowPage(SESSION_MODAL_ID)
	a.SetFocus(a.sessionView.Form)
}

func (a *App) handleSessionCancelled(e *SessionCancelledEvent) {
	a.pages.HidePage(SESSION_MODAL_ID)
}

func (a *App) handleSessionSelected(e *SessionSelectedEvent) {
	a.sessionView.currentSessionID = &e.SessionID
	a.sessionView.gameName = e.GameName
	a.sessionView.Refresh()
}

func (a *App) handleSessionSaved(e *SessionSavedEvent) {
	a.sessionView.Form.ClearFieldErrors()
	a.pages.HidePage(SESSION_MODAL_ID)
	a.sessionView.currentSessionID = &e.Session.ID
	// Load the game to the name can be set
	game := a.gameView.getSelectedGame()
	a.sessionView.gameName = game.Name
	a.sessionView.Refresh()
	a.gameView.Refresh()
	a.gameView.SelectSession(e.Session.ID)
	a.SetFocus(a.sessionView.TextArea)
	a.notification.ShowSuccess("Session saved successfully")
}

func (a *App) handleSessionShowEdit(e *SessionShowEditEvent) {
	if e.Session == nil {
		a.notification.ShowError("Please select a session to edit")
		return
	}

	a.sessionView.Form.PopulateForEdit(e.Session)
	a.pages.ShowPage(SESSION_MODAL_ID)
	a.SetFocus(a.sessionView.Form)
}

func (a *App) handleSessionDeleteConfirm(e *SessionDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this session?",
		func() {
			a.sessionView.ConfirmDelete(e.Session.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleSessionDeleted(_ *SessionDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(SESSION_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.gameView.Refresh()
	a.sessionView.Refresh()
	a.SetFocus(a.gameTree)
	a.notification.ShowSuccess("Session deleted successfully")
}

func (a *App) handleSessionDeleteFailed(e *SessionDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete session: " + e.Error.Error())
}

func (a *App) handleSessionShowHelp(e *SessionShowHelpEvent) {
	a.pages.ShowPage(SESSION_HELP_MODAL_ID)
	a.SetFocus(a.sessionView.HelpModal)
}

func (a *App) handleSessionCloseHelp(e *SessionCloseHelpEvent) {
	a.pages.HidePage(SESSION_HELP_MODAL_ID)
	a.SetFocus(a.sessionView.TextArea)
}
