// Package ui provides the terminal user interface for soloterm.
// It implements a TUI using tview and tcell for managing games and log entries.
package ui

import (
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
	FILE_MODAL_ID      string = "fileModal"
	CONFIRM_MODAL_ID   string = "confirm"
	MAIN_PAGE_ID       string = "main"
	ABOUT_MODAL_ID     string = "about"
	SESSION_MODAL_ID   string = "sessionModal"
	HELP_MODAL_ID      string = "helpModal"
)

type App struct {
	*tview.Application

	// View helpers
	gameView      *GameView
	tagView       *TagView
	sessionView   *SessionView
	characterView *CharacterView
	attributeView *AttributeView

	// Layout containers
	mainFlex         *tview.Flex
	pages            *tview.Pages
	rootFlex         *tview.Flex // Root container with notification
	leftSidebar      *tview.Flex // Left sidebar
	notificationFlex *tview.Flex // Container for notification banner

	// UI Components
	aboutModal   *tview.Modal
	helpModal    *HelpModal
	confirmModal *ConfirmationModal
	footer       *tview.TextView
	notification *Notification
}

func NewApp(db *database.DBStore, cfg *config.Config) *App {
	gameService := game.NewService(game.NewRepository(db))
	charRepo := character.NewRepository(db)
	attrRepo := character.NewAttributeRepository(db)
	attrService := character.NewAttributeService(attrRepo)
	charService := character.NewService(charRepo, attrService)
	sessionRepo := session.NewRepository(db)
	tagService := tag.NewService(sessionRepo)
	sessionService := session.NewService(sessionRepo)

	app := &App{
		Application: tview.NewApplication(),
	}

	// Initialize views
	app.gameView = NewGameView(app, gameService, sessionService)
	app.sessionView = NewSessionView(app, sessionService)
	app.tagView = NewTagView(app, cfg, tagService)
	app.attributeView = NewAttributeView(app, attrService)
	app.characterView = NewCharacterView(app, charService)

	app.setupUI()
	return app
}

func (a *App) setupUI() {

	// Footer with help text
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	a.leftSidebar = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.gameView.Tree, 0, 1, false).
		AddItem(a.characterView.CharTree, 0, 1, false).
		AddItem(a.characterView.CharPane, 0, 2, false) // Give character pane more space

	// Main layout: horizontal split of tree (left, narrow) and log view (right)
	a.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.leftSidebar, 0, 1, false).             // 1/3 of the width
		AddItem(a.sessionView.textAreaFrame, 0, 2, true) // 2/3 of the width

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
			"Version 1.0.6\n\n" +
			"Lonelog by Loreseed Workshop\n" +
			"https://zeruhur.itch.io/lonelog").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.HidePage(ABOUT_MODAL_ID)
		})

	// Create app-level modals
	a.helpModal = NewHelpModal(a)
	a.confirmModal = NewConfirmationModal()

	// Pages for modal overlay (must be created BEFORE notification setup)
	// Note: Pages added later appear on top of earlier pages
	a.pages = tview.NewPages().
		AddPage(MAIN_PAGE_ID, mainContent, true, true).
		AddPage(ABOUT_MODAL_ID, a.aboutModal, true, false).
		AddPage(GAME_MODAL_ID, a.gameView.Modal, true, false).
		AddPage(CHARACTER_MODAL_ID, a.characterView.Modal, true, false).
		AddPage(ATTRIBUTE_MODAL_ID, a.attributeView.Modal, true, false).
		AddPage(SESSION_MODAL_ID, a.sessionView.Modal, true, false).
		AddPage(TAG_MODAL_ID, a.tagView.Modal, true, false).
		AddPage(FILE_MODAL_ID, a.sessionView.FileModal, true, false).
		AddPage(HELP_MODAL_ID, a.helpModal, true, false).
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

	a.SetRoot(a.rootFlex, true).SetFocus(a.gameView.Tree)
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
		switchable := []tview.Primitive{a.gameView.Tree, a.sessionView.TextArea, a.characterView.CharTree, a.attributeView.Table}
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.sessionView.Autosave()
			a.Stop()
			return nil
		case tcell.KeyCtrlG:
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.gameView.Tree)
				return nil
			}
		case tcell.KeyCtrlC:
			// Always consume Ctrl+C to prevent terminal signal handling
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.characterView.CharTree)
			}
			return nil
		case tcell.KeyCtrlS:
			if slices.Contains(switchable, a.GetFocus()) {
				a.SetFocus(a.attributeView.Table)
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
			case a.gameView.Tree:
				a.SetFocus(a.sessionView.TextArea)
				return nil
			case a.sessionView.TextArea:
				a.SetFocus(a.characterView.CharTree)
				return nil
			case a.characterView.CharTree:
				a.SetFocus(a.attributeView.Table)
				return nil
			case a.attributeView.Table:
				a.SetFocus(a.gameView.Tree)
				return nil
			}
			// If focus is not on main views (ie. in a modal), let Tab work normally
			return event
		case tcell.KeyBacktab:
			currentFocus := a.GetFocus()
			switch currentFocus {
			case a.gameView.Tree:
				a.SetFocus(a.attributeView.Table)
				return nil
			case a.sessionView.TextArea:
				a.SetFocus(a.gameView.Tree)
				return nil
			case a.characterView.CharTree:
				a.SetFocus(a.sessionView.TextArea)
				return nil
			case a.attributeView.Table:
				a.SetFocus(a.characterView.CharTree)
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

func (a *App) handleShowHelp(e *ShowHelpEvent) {
	a.helpModal.Show(e.Title, e.Text, e.ReturnFocus)
	a.pages.ShowPage(HELP_MODAL_ID)
	a.SetFocus(a.helpModal)
}

func (a *App) handleCloseHelp(e *CloseHelpEvent) {
	a.pages.HidePage(HELP_MODAL_ID)
	a.SetFocus(a.helpModal.returnFocus)
}

func (a *App) isPageVisible(pageID string) bool {
	return slices.Contains(a.pages.GetPageNames(true), pageID)
}

func (a *App) GetSelectedGameState() *GameState {
	return a.gameView.GetCurrentSelection()
}

func (a *App) GetSelectedCharacterID() *int64 {
	return a.characterView.GetSelectedCharacterID()
}

func (a *App) HandleEvent(event Event) {
	switch event.Action() {
	case GAME_SAVED:
		dispatch(event, a.handleGameSaved)
	case GAME_CANCEL:
		dispatch(event, a.handleGameCancel)
	case GAME_DELETE_CONFIRM:
		dispatch(event, a.handleGameDeleteConfirm)
	case GAME_DELETED:
		dispatch(event, a.handleGameDeleted)
	case GAME_DELETE_FAILED:
		dispatch(event, a.handleGameDeleteFailed)
	case GAME_SHOW_EDIT:
		dispatch(event, a.handleGameShowEdit)
	case GAME_SHOW_NEW:
		dispatch(event, a.handleGameShowNew)
	case GAME_SELECTED:
		dispatch(event, a.handleGameSelected)
	case CHARACTER_SAVED:
		dispatch(event, a.handleCharacterSaved)
	case CHARACTER_CANCEL:
		dispatch(event, a.handleCharacterCancel)
	case CHARACTER_DELETE_CONFIRM:
		dispatch(event, a.handleCharacterDeleteConfirm)
	case CHARACTER_DELETED:
		dispatch(event, a.handleCharacterDeleted)
	case CHARACTER_DELETE_FAILED:
		dispatch(event, a.handleCharacterDeleteFailed)
	case CHARACTER_DUPLICATE_CONFIRM:
		dispatch(event, a.handleCharacterDuplicateConfirm)
	case CHARACTER_DUPLICATED:
		dispatch(event, a.handleCharacterDuplicated)
	case CHARACTER_DUPLICATE_FAILED:
		dispatch(event, a.handleCharacterDuplicateFailed)
	case CHARACTER_SHOW_NEW:
		dispatch(event, a.handleCharacterShowNew)
	case CHARACTER_SHOW_EDIT:
		dispatch(event, a.handleCharacterShowEdit)
	case ATTRIBUTE_SAVED:
		dispatch(event, a.handleAttributeSaved)
	case ATTRIBUTE_CANCEL:
		dispatch(event, a.handleAttributeCancel)
	case ATTRIBUTE_DELETE_CONFIRM:
		dispatch(event, a.handleAttributeDeleteConfirm)
	case ATTRIBUTE_DELETED:
		dispatch(event, a.handleAttributeDeleted)
	case ATTRIBUTE_DELETE_FAILED:
		dispatch(event, a.handleAttributeDeleteFailed)
	case ATTRIBUTE_SHOW_NEW:
		dispatch(event, a.handleAttributeShowNew)
	case ATTRIBUTE_SHOW_EDIT:
		dispatch(event, a.handleAttributeShowEdit)
	case TAG_SELECTED:
		dispatch(event, a.handleTagSelected)
	case TAG_CANCEL:
		dispatch(event, a.handleTagCancelled)
	case TAG_SHOW:
		dispatch(event, a.handleTagShow)
	case SHOW_HELP:
		dispatch(event, a.handleShowHelp)
	case CLOSE_HELP:
		dispatch(event, a.handleCloseHelp)
	case SESSION_SHOW_NEW:
		dispatch(event, a.handleSessionShowNew)
	case SESSION_CANCEL:
		dispatch(event, a.handleSessionCancelled)
	case SESSION_SELECTED:
		dispatch(event, a.handleSessionSelected)
	case SESSION_SAVED:
		dispatch(event, a.handleSessionSaved)
	case SESSION_SHOW_EDIT:
		dispatch(event, a.handleSessionShowEdit)
	case SESSION_DELETE_CONFIRM:
		dispatch(event, a.handleSessionDeleteConfirm)
	case SESSION_DELETED:
		dispatch(event, a.handleSessionDeleted)
	case SESSION_DELETE_FAILED:
		dispatch(event, a.handleSessionDeleteFailed)
	case SESSION_SHOW_IMPORT:
		dispatch(event, a.handleSessionShowImport)
	case SESSION_SHOW_EXPORT:
		dispatch(event, a.handleSessionShowExport)
	case SESSION_IMPORT:
		dispatch(event, a.handleSessionImport)
	case SESSION_EXPORT:
		dispatch(event, a.handleSessionExport)
	case SESSION_IMPORT_DONE:
		dispatch(event, a.handleSessionImportDone)
	case SESSION_EXPORT_DONE:
		dispatch(event, a.handleSessionExportDone)
	case FILE_FORM_CANCEL:
		dispatch(event, a.handleFileFormCancelled)
	}
}
