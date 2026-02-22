package ui

import (
	"fmt"
	"soloterm/domain/session"
	sharedui "soloterm/shared/ui"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SessionView provides session-specific UI operations
type SessionView struct {
	TextArea          *tview.TextArea
	textAreaFrame     *tview.Frame
	Form              *SessionForm
	Modal             *tview.Flex
	FileForm          *FileForm
	FileModal         *tview.Flex
	fileFormContainer *tview.Flex
	app               *App
	sessionService    *session.Service
	currentSessionID  *int64
	currentSession    *session.Session
	isLoading         bool
	isDirty           bool
	isImporting       bool
	autosaveTicker    *time.Ticker
	autosaveStop      chan struct{}
}

const (
	DEFAULT_SECTION_TITLE = " [::b]Select/Add Session To View (Ctrl+L) "
)

// NewSessionView creates a new session view helper
func NewSessionView(app *App, service *session.Service) *SessionView {
	sessionView := &SessionView{
		app:            app,
		sessionService: service,
		isDirty:        false,
	}

	sessionView.Setup()

	return sessionView
}

// Setup initializes all session UI components
func (sv *SessionView) Setup() {
	sv.setupTextArea()
	sv.setupModal()
	sv.setupFileModal()
	sv.setupKeyBindings()
	sv.setupFocusHandlers()
}

// setupTextArea configures the text area for displaying the session
func (sv *SessionView) setupTextArea() {
	sv.TextArea = tview.NewTextArea()
	sv.TextArea.SetDisabled(true)
	sv.TextArea.SetChangedFunc(func() {
		if sv.isLoading {
			return
		}
		sv.isDirty = true
		sv.updateTitle()
		sv.startAutosave()
	})

	sv.textAreaFrame = tview.NewFrame(sv.TextArea).
		SetBorders(1, 1, 0, 0, 1, 1)
	sv.textAreaFrame.SetTitle(DEFAULT_SECTION_TITLE).
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)
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
				AddItem(sv.Form, 7, 1, true). // Dynamic height: expands to fit content
				AddItem(nil, 0, 1, false),
			60, 1, true, // Dynamic width: expands to fit content (up to screen width)
		).
		AddItem(nil, 0, 1, false)
	// sv.Modal.SetBackgroundColor(tcell.ColorBlack)

	sv.Form.SetFocusFunc(func() {
		sv.app.SetModalHelpMessage(*sv.Form.DataForm)
		sv.Form.SetBorderColor(Style.BorderFocusColor)
	})

	sv.Form.SetBlurFunc(func() {
		sv.Form.SetBorderColor(Style.BorderColor)
	})
}

// setupFileModal configures the file import/export form modal
func (sv *SessionView) setupFileModal() {
	sv.FileForm = NewFileForm()

	sv.FileForm.SetupHandlers(
		func() {
			if sv.isImporting {
				sv.app.HandleEvent(&SessionImportEvent{
					BaseEvent: BaseEvent{action: SESSION_IMPORT},
				})
			} else {
				sv.app.HandleEvent(&SessionExportEvent{
					BaseEvent: BaseEvent{action: SESSION_EXPORT},
				})
			}
		},
		func() {
			sv.app.HandleEvent(&FileFormCancelledEvent{
				BaseEvent: BaseEvent{action: FILE_FORM_CANCEL},
			})
		},
		nil,
	)

	helpTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)

	sv.FileForm.SetHelpTextChangeHandler(func(text string) {
		helpTextView.SetText(text)
	})

	sv.fileFormContainer = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sv.FileForm, 0, 1, true).
		AddItem(helpTextView, 3, 0, false)
	sv.fileFormContainer.SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	sv.FileForm.SetFocusFunc(func() {
		sv.fileFormContainer.SetBorderColor(Style.BorderFocusColor)
	})

	sv.FileForm.SetBlurFunc(func() {
		sv.fileFormContainer.SetBorderColor(Style.BorderColor)
	})

	sv.FileModal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(sv.fileFormContainer, 12, 1, true).
				AddItem(nil, 0, 1, false),
			60, 1, true,
		).
		AddItem(nil, 0, 1, false)
	// sv.FileModal.SetBackgroundColor(tcell.ColorBlack)
}

// setupKeyBindings configures keyboard shortcuts for the session tree
func (sv *SessionView) setupKeyBindings() {
	sv.TextArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF12:
			sv.ShowHelpModal()
			return nil
		case tcell.KeyCtrlN:
			sv.ShowNewModal()
			return nil
		case tcell.KeyF2:
			if sv.currentSessionID != nil {
				sv.InsertAtCursor("@ \nd: ->\n=> ")
			}
			return nil
		case tcell.KeyF3:
			if sv.currentSessionID != nil {
				sv.InsertAtCursor("? \nd: ->\n=> ")
			}
			return nil
		case tcell.KeyF4:
			if sv.currentSessionID != nil {
				sv.InsertAtCursor("d: ->\n=> ")
			}
			return nil
		case tcell.KeyCtrlT:
			if sv.currentSessionID != nil {
				sv.Autosave()
				sv.app.HandleEvent(&TagShowEvent{
					BaseEvent: BaseEvent{action: TAG_SHOW},
				})
			}
			return nil
		case tcell.KeyCtrlO:
			sv.app.HandleEvent(&SessionShowImportEvent{
				BaseEvent: BaseEvent{action: SESSION_SHOW_IMPORT},
			})
			return nil
		case tcell.KeyCtrlX:
			sv.app.HandleEvent(&SessionShowExportEvent{
				BaseEvent: BaseEvent{action: SESSION_SHOW_EXPORT},
			})
			return nil
		}

		return event
	})
}

// setupFocusHandlers configures focus event handlers
func (sv *SessionView) setupFocusHandlers() {
	editHelp := "[" + Style.HelpKeyTextColor + "]PgUp/PgDn/↑/↓[" + Style.NormalTextColor + "] Scroll  [" + Style.HelpKeyTextColor + "]F12[" + Style.NormalTextColor + "] Help  [" + Style.HelpKeyTextColor + "]Ctrl+N[" + Style.NormalTextColor + "] New  [" + Style.HelpKeyTextColor + "]Ctrl+T[" + Style.NormalTextColor + "] Tag  [" + Style.HelpKeyTextColor + "]F2[" + Style.NormalTextColor + "] Action  [" + Style.HelpKeyTextColor + "]F3[" + Style.NormalTextColor + "] Oracle"
	newHelp := "[" + Style.HelpKeyTextColor + "]Ctrl+N[" + Style.NormalTextColor + "] New"
	baseHelp := "[" + Style.ContextLabelTextColor + "::b]Session[-::-] :: "
	sv.TextArea.SetFocusFunc(func() {
		if sv.currentSessionID != nil {
			sv.app.updateFooterHelp(baseHelp + editHelp)
		} else {
			sv.app.updateFooterHelp(baseHelp + newHelp)
		}
		sv.textAreaFrame.SetBorderColor(Style.BorderFocusColor)
	})

	sv.TextArea.SetBlurFunc(func() {
		sv.textAreaFrame.SetBorderColor(Style.BorderColor)
	})
}

// Reset removes the state of the view
func (sv *SessionView) Reset() {
	sv.currentSessionID = nil
	sv.currentSession = nil
	sv.isLoading = false
	sv.isDirty = false
	sv.isImporting = false

}

// Refresh reloads the session tree from the database and restores selection
func (sv *SessionView) Refresh() {
	sv.Autosave()

	if sv.currentSessionID == nil {
		sv.textAreaFrame.SetTitle(DEFAULT_SECTION_TITLE)
		sv.TextArea.SetText("", true)
		sv.currentSession = nil
		return
	}

	// Load the session and set the content
	loadedSession, err := sv.sessionService.GetByID(*sv.currentSessionID)
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

	session, err := sv.sessionService.GetByID(*sv.currentSessionID)
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
	err := sv.sessionService.Delete(sessionID)
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
	sv.app.HandleEvent(&ShowHelpEvent{
		BaseEvent:   BaseEvent{action: SHOW_HELP},
		Title:       "Session Help",
		ReturnFocus: sv.TextArea,
		Text:        sv.buildHelpText(),
	})
}

func (sv *SessionView) buildHelpText() string {
	return strings.NewReplacer(
		"[yellow]", "["+Style.HelpKeyTextColor+"]",
		"[white]", "["+Style.NormalTextColor+"]",
		"[green]", "["+Style.HelpSectionColor+"]",
	).Replace(`Scroll Down To View All Help Options

[green]Session Management[white]

Select the session in the game view to edit the name or delete the session.

[yellow]Note:[white] Do not paste large amounts of text into the session log. It is slow. Instead, use Import (see the help below)

[yellow]Ctrl+N[white]: Add a new session.
[yellow]Ctrl+O[white]: Import content from a file.
[yellow]Ctrl+X[white]: Export content to a file.

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
[yellow]Ctrl-E, End[white]: Move to the end of the current line.
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
[yellow]Ctrl-Y[white]: Redo.

[green]Import/Export

[yellow]Ctrl-O[white]: Open a text file to import. You can choose where the imported text is inserted into the log.
[yellow]Ctrl-X[white]: Export the current session to a text file.
`)
}

// ShowEditModal displays the session form modal for editing an existing session
func (sv *SessionView) ShowEditModal(sessionID int64) {
	sv.Autosave()
	session, err := sv.sessionService.GetByID(sessionID)
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
	body := tview.Escape(sv.currentSession.GameName) + ": " + tview.Escape(sv.currentSession.Name)
	prefix := ""
	if sv.isDirty {
		prefix = "[" + Style.ErrorTextColor + "]●[-] "
	}
	sv.textAreaFrame.SetTitle(" " + prefix + "[::b]" + body + " (Ctrl+L) ")
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
	_, start, _ := sv.TextArea.GetSelection()
	sv.TextArea.Replace(start, start, template)
}
