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
	TextArea      *tview.TextArea
	textAreaFrame *tview.Frame
	Form          *SessionForm
	Modal         *tview.Flex
	formModal     *sharedui.FormModal
	app           *App
	sessionService   *session.Service
	currentSessionID *int64
	currentSession   *session.Session
	isNotes          bool
	isLoading        bool
	isDirty          bool
	autosaveTicker   *time.Ticker
	autosaveStop     chan struct{}
}

// IsNotesMode reports whether the pane is displaying game notes rather than a session.
func (sv *SessionView) IsNotesMode() bool {
	return sv.isNotes
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
	sv.setupKeyBindings()
	sv.setupFocusHandlers()
}

// setupTextArea configures the text area for displaying the session
func (sv *SessionView) setupTextArea() {
	sv.TextArea = tview.NewTextArea()
	sv.TextArea.SetPlaceholder("Select a session to view from the Games view or select a Game and press Ctrl+N here to create a new session.")
	sv.TextArea.SetPlaceholderStyle(tcell.StyleDefault.
		Background(Style.PrimitiveBackgroundColor).
		Foreground(Style.EmptyStateMessageColor))
	sv.TextArea.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorYellow).
		Foreground(Style.ContrastSecondaryTextColor))

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

	sv.formModal = sharedui.NewFormModal(sv.Form, 7)
	sv.Modal = sv.formModal.Modal

	sv.Form.SetFocusFunc(func() {
		sv.app.SetModalHelpMessage(*sv.Form.DataForm)
		sv.formModal.SetBorderColor(Style.BorderFocusColor)
	})

	sv.Form.SetBlurFunc(func() {
		sv.formModal.SetBorderColor(Style.BorderColor)
	})
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
				sv.InsertAtCursor(sv.app.cfg.CoreTags.Action.Template)
			}
			return nil
		case tcell.KeyF3:
			if sv.currentSessionID != nil {
				sv.InsertAtCursor(sv.app.cfg.CoreTags.Oracle.Template)
			}
			return nil
		case tcell.KeyF4:
			if sv.currentSessionID != nil {
				sv.InsertAtCursor(sv.app.cfg.CoreTags.Dice.Template)
			}
			return nil
		case tcell.KeyF5:
			if sv.currentSessionID != nil || sv.IsNotesMode() {
				sv.app.Autosave()
				sv.app.HandleEvent(&SearchShowEvent{
					BaseEvent: BaseEvent{action: SEARCH_SHOW},
				})
			}
		case tcell.KeyCtrlT:
			if sv.currentSessionID != nil || sv.IsNotesMode() {
				sv.app.Autosave()
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
	sv.TextArea.SetFocusFunc(func() {
		if sv.currentSessionID != nil {
			sv.app.updateFooterHelp(helpBar("Session", []helpEntry{
				{"PgUp/PgDn/↑/↓", "Scroll"},
				{"F12", "Help"},
				{"Ctrl+N", "New"},
				{"Ctrl+T", "Tag"},
				{"F2", "Action"},
				{"F3", "Oracle"},
				{"F4", "Dice"},
				{"F5", "Search"},
			}))
		} else if sv.IsNotesMode() {
			sv.app.updateFooterHelp(helpBar("Notes", []helpEntry{
				{"PgUp/PgDn/↑/↓", "Scroll"},
				{"F12", "Help"},
				{"Ctrl+N", "New Session"},
				{"Ctrl+T", "Tag"},
				{"F5", "Search"},
			}))
		} else {
			sv.app.updateFooterHelp(helpBar("Session", []helpEntry{
				{"Ctrl+N", "New"},
			}))
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
	sv.isNotes = false
	sv.isLoading = false
	sv.isDirty = false
	sv.stopAutosave()
}

func (sv *SessionView) SetText(text string, cursorAtEnd bool) {
	sv.isLoading = true
	sv.TextArea.SetText(text, cursorAtEnd)
	sv.isLoading = false
}

// Refresh reloads the session tree from the database and restores selection
func (sv *SessionView) Refresh() {
	sv.app.Autosave()

	if sv.IsNotesMode() {
		g := sv.app.CurrentGame()
		if g == nil {
			return
		}
		if g.Notes != sv.TextArea.GetText() {
			sv.SetText(g.Notes, false)
		}
		sv.updateTitle()
		sv.TextArea.SetDisabled(false)
		return
	}

	if sv.currentSessionID == nil {
		sv.textAreaFrame.SetTitle(DEFAULT_SECTION_TITLE)
		sv.SetText("", true)
		sv.currentSession = nil
		return
	}

	// Load the session and set the content
	loadedSession, err := sv.sessionService.GetByID(*sv.currentSessionID)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
	}

	sv.currentSession = loadedSession

	// Skip SetText if content is unchanged (e.g. rename) to preserve cursor and scroll
	if loadedSession.Content != sv.TextArea.GetText() {
		sv.SetText(loadedSession.Content, false)
	}
	sv.updateTitle()
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
	var title string
	if sv.IsNotesMode() {
		title = "Notes Help"
	} else {
		title = "Session Help"
	}

	sv.app.HandleEvent(&ShowHelpEvent{
		BaseEvent:   BaseEvent{action: SHOW_HELP},
		Title:       title,
		ReturnFocus: sv.TextArea,
		Text:        sv.buildHelpText(),
	})
}

func (sv *SessionView) buildHelpText() string {
	isNotes := sv.IsNotesMode()
	var b strings.Builder

	b.WriteString("Scroll Down To View All Help Options\n\n")

	if isNotes {
		b.WriteString("[green]Notes[white]\n\n")
		b.WriteString("Notes are where you can track things for the entire game. For example, key NPCs, locations, or other details that cross multiple sessions. Notes are also searchable and tags added here will appear in the list of available tags.\n\n")
	} else {
		b.WriteString("[green]Session Management[white]\n\n")
		b.WriteString("Select the session in the game view to edit the name or delete the session.\n\n")
	}

	b.WriteString("[yellow]Note:[white] Do not paste large amounts of text into the session log or notes. It is slow. Instead, use Import.\n\n")
	b.WriteString("[yellow]Ctrl+N[white]: Add a new session.\n")
	b.WriteString("[yellow]Ctrl-O[white]: Open a text file to import.\n")
	b.WriteString("[yellow]Ctrl-X[white]: Export to a text file.\n")
	b.WriteString("[yellow]F5[white]: Search the notes and sessions.\n")

	if !isNotes {
		b.WriteString("\n[green][:::https://zeruhur.itch.io/lonelog]Lonelog[:::-] https://zeruhur.itch.io/lonelog\n\n")
		b.WriteString("[yellow]F2[white]: Insert the Character Action template.\n")
		b.WriteString("[yellow]F3[white]: Insert the Oracle template.\n")
		b.WriteString("[yellow]F4[white]: Insert the Dice template.\n")
	}

	b.WriteString("[yellow]Ctrl+T[white]: Select a template (NPC, Event, Location, etc.) to insert.\n")

	b.WriteString(`
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
[yellow]Ctrl-Z[white]: Undo.
[yellow]Ctrl-Y[white]: Redo.
`)

	return strings.NewReplacer(
		"[yellow]", "["+Style.HelpKeyTextColor+"]",
		"[white]", "["+Style.NormalTextColor+"]",
		"[green]", "["+Style.HelpSectionColor+"]",
	).Replace(b.String())
}

// ShowEditModal displays the session form modal for editing an existing session
func (sv *SessionView) ShowEditModal(sessionID int64) {
	sv.app.Autosave()
	session, err := sv.sessionService.GetByID(sessionID)
	if err != nil {
		sv.app.notification.ShowError(fmt.Sprintf("Error loading session: %v", err))
		return
	}

	sv.app.HandleEvent(&SessionShowEditEvent{
		BaseEvent: BaseEvent{action: SESSION_SHOW_EDIT},
		Session:   session,
	})
}

func (sv *SessionView) updateTitle() {
	keyHelp := " ([" + Style.HelpKeyTextColor + "]Ctrl+L[" + Style.NormalTextColor + "]) "
	counts := sv.wordCharCount()
	prefix := ""
	body := ""
	if sv.isDirty {
		prefix = "[" + Style.ErrorTextColor + "]●[-] "
	}

	if sv.IsNotesMode() {
		g := sv.app.CurrentGame()
		if g == nil {
			return
		}
		body = tview.Escape(g.Name) + ": Notes"
	} else {
		if sv.currentSession == nil {
			return
		}
		body = tview.Escape(sv.currentSession.GameName) + ": " + tview.Escape(sv.currentSession.Name)
	}

	sv.textAreaFrame.SetTitle(" " + prefix + "[::b]" + body + keyHelp + counts)
}

func (sv *SessionView) wordCharCount() string {
	text := sv.TextArea.GetText()
	words := len(strings.Fields(text))
	chars := len([]rune(text))
	return fmt.Sprintf("[::d] %d words · %d chars ", words, chars)
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
					sv.app.Autosave()
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

// SelectSession switches to session mode and loads the given session into the editor.
func (sv *SessionView) SelectSession(sessionID int64) {
	sv.isNotes = false
	sv.currentSession = nil
	sv.currentSessionID = &sessionID
	sv.Refresh()
}

// SelectNotes switches to notes mode and loads the active game's notes into the editor.
func (sv *SessionView) SelectNotes() {
	sv.isNotes = true
	sv.currentSession = nil
	sv.currentSessionID = nil
	sv.Refresh()
}

func (sv *SessionView) InsertAtCursor(template string) {
	_, start, _ := sv.TextArea.GetSelection()
	// When inserting the string would push the cursor below the
	// viewable area, scroll up just enough so the cursor remains
	// visible after inserting the content.
	sv.TextArea.SetMovedFunc(func() {
		sv.TextArea.SetMovedFunc(nil)
		row, _, _, _ := sv.TextArea.GetCursor()
		offsetRow, _ := sv.TextArea.GetOffset()
		_, _, _, height := sv.TextArea.GetInnerRect()
		if row >= offsetRow+height {
			sv.TextArea.SetOffset(row-height+1, 0)
		}
	})
	sv.TextArea.Replace(start, start, template)
}

// ====== FileTarget implementation ======

func (sv *SessionView) GetFileContent() string {
	return sv.TextArea.GetText()
}

func (sv *SessionView) SetFileContent(data string, position ImportPosition) {
	switch position {
	case ImportBefore:
		sv.SetText(data+sv.TextArea.GetText(), false)
	case ImportAfter:
		sv.SetText(sv.TextArea.GetText()+data, true)
	case ImportAtCursor:
		sv.InsertAtCursor(data)
	default: // ImportReplace
		sv.SetText(data, false)
	}
	sv.isDirty = true
	sv.updateTitle()
}

func (sv *SessionView) UsePositionField() bool { return true }

func (sv *SessionView) FileDir() string { return "" } // unused; FileView uses dirs.ExportDir()

func (sv *SessionView) OnFileDone() {
	sv.app.Autosave()
}
