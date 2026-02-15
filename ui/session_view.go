package ui

import (
	"fmt"
	"soloterm/domain/session"
	sharedui "soloterm/shared/ui"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SessionView provides session-specific UI operations
type SessionView struct {
	TextArea         *tview.TextArea
	Form             *SessionForm
	Modal            *tview.Flex
	HelpModal        *tview.Flex
	app              *App
	sessionService   *session.Service
	helper           *SessionViewHelper
	currentSessionID *int64
	currentSession   *session.Session
	isLoading        bool
	isDirty          bool
	autosaveTicker   *time.Ticker
	autosaveStop     chan struct{}
}

const (
	DEFAULT_SECTION_TITLE = " [::b]Select/Add Session To View (Ctrl+L)"
)

// NewSessionView creates a new session view helper
func NewSessionView(app *App, service *session.Service) *SessionView {
	sessionView := &SessionView{
		app:            app,
		sessionService: service,
		isDirty:        false,
		helper:         NewSessionViewHelper(service),
	}

	sessionView.Setup()

	return sessionView
}

// Setup initializes all session UI components
func (sv *SessionView) Setup() {
	sv.setupTextArea()
	sv.setupModal()
	sv.setupHelpModal()
	sv.setupKeyBindings()
	sv.setupFocusHandlers()
}

// setupTextArea configures the text area for displaying the session
func (sv *SessionView) setupTextArea() {
	sv.TextArea = tview.NewTextArea()
	sv.TextArea.SetDisabled(true)
	sv.TextArea.SetTitle(DEFAULT_SECTION_TITLE).
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)
	sv.TextArea.SetChangedFunc(func() {
		if sv.isLoading {
			return
		}
		sv.isDirty = true
		sv.updateTitle()
		sv.startAutosave()
	})
}

func (sv *SessionView) setupHelpModal() {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`Scroll Down To View All Help Options

[green]Session Management

[yellow]Ctrl+E[white]: Edit the session name or Delete the session.
[yellow]Ctrl+N[white]: Add a new session.

[green][:::https://zeruhur.itch.io/lonelog]Lonelog[:::-] https://zeruhur.itch.io/lonelog

[yellow]F2[white]: Insert the Character Action template.
[yellow]F3[white]: Insert the Oracle template.
[yellow]F4[white]: Insert the Dice template.
[yellow]Ctrl+T[white]: Select a template (NPC, Event, Location, etc.) to insert.
		
[green]Navigation

[yellow]Left arrow[white]: Move left.
[yellow]Right arrow[white]: Move right.
[yellow]Down arrow[white]: Move down.
[yellow]Up arrow[white]: Move up.
[yellow]Ctrl-A, Home[white]: Move to the beginning of the current line.
[yellow]End[white]: Move to the end of the current line.
[yellow]Ctrl-F, page down[white]: Move down by one page.
[yellow]Ctrl-B, page up[white]: Move up by one page.
[yellow]Alt-Up arrow[white]: Scroll the page up.
[yellow]Alt-Down arrow[white]: Scroll the page down.
[yellow]Alt-Left arrow[white]: Scroll the page to the left.
[yellow]Alt-Right arrow[white]: Scroll the page to the right.
[yellow]Alt-B, Ctrl-Left arrow[white]: Move back by one word.
[yellow]Alt-F, Ctrl-Right arrow[white]: Move forward by one word.

[green]Editing[white]

Type to enter text.
[yellow]Backspace[white]: Delete the left character.
[yellow]Delete[white]: Delete the right character.
[yellow]Ctrl-K[white]: Delete until the end of the line.
[yellow]Ctrl-W[white]: Delete the rest of the word.
[yellow]Ctrl-U[white]: Delete the current line.

[green]Undo

[yellow]Ctrl-Z[white]: Undo.
[yellow]Ctrl-Y[white]: Redo.`)

	helpFrame := tview.NewFrame(helpText).
		SetBorders(1, 1, 0, 0, 2, 2)
	helpFrame.SetBorder(true).
		SetTitle(" [::b]Help (Esc = Close) ").
		SetTitleAlign(tview.AlignLeft).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				sv.app.HandleEvent(&SessionCloseHelpEvent{
					BaseEvent: BaseEvent{action: SESSION_CLOSE_HELP},
				})
				return nil
			}
			return event
		})

	// Scroll to the beginning when the help text receives focus
	helpText.SetFocusFunc(func() {
		helpText.ScrollToBeginning()
	})

	// Center the modal on screen
	sv.HelpModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(helpFrame, 22, 1, true).
				AddItem(nil, 0, 1, false),
			64, 1, true,
		).
		AddItem(nil, 0, 1, false)
}

// setupModal configures the session form modal
func (sv *SessionView) setupModal() {

	sv.Form = NewSessionForm()

	// Set up handlers
	sv.Form.SetupHandlers(
		sv.HandleSave,
		sv.HandleCancel,
		sv.HandleDelete,
	)

	// Center the modal on screen
	sv.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(sv.Form, 8, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
	sv.Modal.SetBackgroundColor(tcell.ColorBlack)

	sv.Modal.SetFocusFunc(func() {
		sv.app.SetModalHelpMessage(*sv.Form.DataForm)
	})

}

// setupKeyBindings configures keyboard shortcuts for the session tree
func (sv *SessionView) setupKeyBindings() {
	sv.TextArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			if sv.currentSessionID != nil {
				sv.ShowEditModal(*sv.currentSessionID)
				return nil
			}
		case tcell.KeyCtrlH:
			sv.ShowHelpModal()
			return nil
		case tcell.KeyCtrlN:
			sv.ShowNewModal()
			return nil
		case tcell.KeyF2:
			sv.InsertAtCursor("@ \nd: ->\n=> ")
			return nil
		case tcell.KeyF3:
			sv.InsertAtCursor("? \nd: ->\n=> ")
			return nil
		case tcell.KeyF4:
			sv.InsertAtCursor("d: ->\n=> ")
			return nil
		case tcell.KeyCtrlT:
			// Dispatch event with saved log
			sv.app.HandleEvent(&TagShowEvent{
				BaseEvent: BaseEvent{action: TAG_SHOW},
			})
			return nil
		}

		return event
	})
}

// setupFocusHandlers configures focus event handlers
func (sv *SessionView) setupFocusHandlers() {
	sv.TextArea.SetFocusFunc(func() {
		sv.app.updateFooterHelp("[aqua::b]Session[-::-] :: [yellow]PgUp/PgDn/↑/↓[white] Scroll  [yellow]Ctrl+E[white] Edit  [yellow]Ctrl+N[white] New  [yellow]Ctrl+T[white] Tag  [yellow]F2[white] Action  [yellow]F3[white] Oracle  [yellow]F4[white] Dice  [yellow]Ctrl+H[white] Help")
	})
}

// Refresh reloads the session tree from the database and restores selection
func (sv *SessionView) Refresh() {
	sv.Autosave()

	if sv.currentSessionID == nil {
		sv.TextArea.SetTitle(DEFAULT_SECTION_TITLE)
		sv.TextArea.SetText("", true)
		sv.currentSession = nil
		return
	}

	// Load the session and set the content
	loadedSession, err := sv.helper.LoadSession(*sv.currentSessionID)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
	}

	sv.currentSession = loadedSession
	sv.updateTitle()

	// Skip SetText if content is unchanged (e.g. rename) to preserve cursor and scroll
	if loadedSession.Content != sv.TextArea.GetText() {
		sv.isLoading = true
		sv.TextArea.SetText(loadedSession.Content, false)
		sv.isLoading = false
	}
	sv.TextArea.SetDisabled(false)
}

// HandleSave processes session save operation
func (sv *SessionView) HandleSave() {
	session := sv.Form.BuildDomain()

	session, err := sv.sessionService.Save(session)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, sv.Form) {
			return
		}

		// Other errors
		sv.app.notification.ShowError(fmt.Sprintf("Error saving session: %v", err))
		return
	}

	sv.currentSessionID = &session.ID

	sv.app.HandleEvent(&SessionSavedEvent{
		BaseEvent: BaseEvent{action: SESSION_SAVED},
		Session:   *session,
	})

}

// HandleCancel processes session form cancellation
func (sv *SessionView) HandleCancel() {
	sv.app.HandleEvent(&SessionCancelledEvent{
		BaseEvent: BaseEvent{action: SESSION_CANCEL},
	})
}

// HandleDelete processes session deletion with confirmation
func (sv *SessionView) HandleDelete() {

	if sv.currentSessionID == nil {
		sv.app.notification.ShowError("Please select a session to delete")
		return
	}

	session, err := sv.helper.LoadSession(*sv.currentSessionID)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
	}

	// Dispatch event to show confirmation
	sv.app.HandleEvent(&SessionDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: SESSION_DELETE_CONFIRM},
		Session:   session,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (sv *SessionView) ConfirmDelete(sessionID int64) {
	// Business logic: Delete the session
	err := sv.helper.Delete(sessionID)
	if err != nil {
		// Dispatch failure event with error
		sv.app.HandleEvent(&SessionDeleteFailedEvent{
			BaseEvent: BaseEvent{action: SESSION_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	sv.currentSessionID = nil

	// Dispatch success event
	sv.app.HandleEvent(&SessionDeletedEvent{
		BaseEvent: BaseEvent{action: SESSION_DELETED},
	})
}

// ShowNewModal displays the session form modal for creating a new session
func (sv *SessionView) ShowNewModal() {
	sv.app.HandleEvent(&SessionShowNewEvent{
		BaseEvent: BaseEvent{action: SESSION_SHOW_NEW},
	})
}

func (sv *SessionView) ShowHelpModal() {
	sv.app.HandleEvent(&SessionShowHelpEvent{
		BaseEvent: BaseEvent{action: SESSION_SHOW_HELP},
	})
}

// ShowEditModal displays the session form modal for editing an existing session
func (sv *SessionView) ShowEditModal(sessionID int64) {
	sv.Autosave()
	session, err := sv.helper.LoadSession(sessionID)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
	}

	sv.app.HandleEvent(&SessionShowEditEvent{
		BaseEvent: BaseEvent{action: SESSION_SHOW_EDIT},
		Session:   session,
	})
}

func (sv *SessionView) updateTitle() {
	if sv.currentSession == nil {
		return
	}
	title := " [::b]" + sv.currentSession.GameName + ": " + sv.currentSession.Name + " (Ctrl+L) "
	if sv.isDirty {
		title = " [red]●[-] [::b]" + sv.currentSession.GameName + ": " + sv.currentSession.Name + " (Ctrl+L) "
	}
	sv.TextArea.SetTitle(title)
}

// Autosave persists the current TextArea content if dirty
func (sv *SessionView) Autosave() {
	if !sv.isDirty || sv.currentSession == nil {
		return
	}
	sv.currentSession.Content = sv.TextArea.GetText()
	_, err := sv.sessionService.Save(sv.currentSession)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Autosave failed: %v", err))
		return
	}
	sv.isDirty = false
	sv.updateTitle()
	sv.stopAutosave()
}

func (sv *SessionView) startAutosave() {
	if sv.autosaveTicker != nil {
		return
	}
	sv.autosaveTicker = time.NewTicker(3 * time.Second)
	sv.autosaveStop = make(chan struct{})
	ticker := sv.autosaveTicker
	stop := sv.autosaveStop
	go func() {
		for {
			select {
			case <-ticker.C:
				sv.app.QueueUpdateDraw(func() {
					sv.Autosave()
				})
			case <-stop:
				return
			}
		}
	}()
}

func (sv *SessionView) stopAutosave() {
	if sv.autosaveTicker != nil {
		sv.autosaveTicker.Stop()
		sv.autosaveTicker = nil
	}
	if sv.autosaveStop != nil {
		close(sv.autosaveStop)
		sv.autosaveStop = nil
	}
}

func (sv *SessionView) InsertAtCursor(template string) {
	row, col, _, _ := sv.TextArea.GetCursor()
	content := sv.TextArea.GetText()

	// Convert row:col to byte offset
	currentRow := 0
	currentCol := 0

	for i, r := range content {
		if currentRow == row && currentCol == col {
			sv.TextArea.Replace(i, i, template)
			return
		}
		if r == '\n' {
			currentRow++
			currentCol = 0
		} else {
			currentCol++
		}
	}

	// Cursor is at the very end of text
	sv.TextArea.Replace(len(content), len(content), template)
}
