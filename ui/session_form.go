package ui

import (
	"soloterm/domain/session"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SesisonForm represents a form for creating/editing sessions
type SessionForm struct {
	*sharedui.DataForm
	sessionID    *int64
	gameID       *int64
	content      string
	nameField    *tview.InputField
	errorMessage *tview.TextView
}

// NewSessionForm creates a new session form
func NewSessionForm() *SessionForm {
	sf := &SessionForm{
		DataForm: sharedui.NewDataForm(),
	}

	sf.errorMessage = tview.NewTextView().
		SetLabel("").
		SetText("")

	// Name field
	sf.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0) // 0 means full width

	sf.setupForm()
	return sf
}

// Fill the fields with the data from the session passed in
func (sf *SessionForm) PopulateForEdit(session *session.Session) {
	sf.gameID = &session.GameID
	sf.sessionID = &session.ID
	sf.content = session.Content

	sf.nameField.SetText(session.Name)

	sf.AddDeleteButton()

	sf.SetFocus(0)
	sf.SetTitle(" Edit Session ")
}

func (sf *SessionForm) setupForm() {
	sf.Clear(true)

	sf.AddFormItem(sf.nameField)

	// Buttons will be set up when handlers are attached
	sf.SetBorder(true).
		SetTitle(" New Session ").
		SetTitleAlign(tview.AlignLeft)

	sf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	sf.SetItemPadding(1)
}

// Reset clears all form fields
func (sf *SessionForm) Reset(gameID int64) {
	sf.gameID = &gameID
	sf.content = ""
	sf.nameField.SetText("")
	sf.ClearFieldErrors()
	sf.SetTitle(" New Session ")
	sf.RemoveDeleteButton()
	sf.SetFocus(0)
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (sf *SessionForm) SetFieldErrors(errors map[string]string) {
	sf.DataForm.SetFieldErrors(errors)
	sf.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (sf *SessionForm) updateFieldLabels() {

	// Update name field label
	if sf.HasFieldError("name") {
		sf.nameField.SetLabel("[red]Name[white]")
	} else {
		sf.nameField.SetLabel("Name")
	}

}

// ClearFieldErrors removes all error highlights
func (sf *SessionForm) ClearFieldErrors() {
	sf.DataForm.ClearFieldErrors()
	sf.updateFieldLabels()
}

// BuildDomain constructs a Session entity from the form data
func (sf *SessionForm) BuildDomain() *session.Session {
	s := &session.Session{
		Name:    sf.nameField.GetText(),
		Content: sf.content,
	}

	if sf.sessionID != nil {
		s.ID = *sf.sessionID
	}

	// If editing an existing session, set the ID
	if sf.gameID != nil {
		s.GameID = *sf.gameID
	}

	return s
}
