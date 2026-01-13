package ui

import (
	syslog "log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DataForm contains common functionality for forms that manage data entities
type DataForm struct {
	*tview.Form
	fieldErrors      map[string]string
	onSave           func()
	onCancel         func()
	onDelete         func()
	onHelpTextChange func(string)
}

// NewDataForm creates a new DataForm with initialized fields
func NewDataForm() *DataForm {
	return &DataForm{
		Form:        tview.NewForm(),
		fieldErrors: make(map[string]string),
	}
}

// SetFieldErrors sets multiple field errors at once
// This implements the ValidatableForm interface
func (df *DataForm) SetFieldErrors(errors map[string]string) {
	syslog.Print("in DataForm.SetFieldErrors")
	df.fieldErrors = errors
}

// ClearFieldErrors removes all error highlights
func (df *DataForm) ClearFieldErrors() {
	df.fieldErrors = make(map[string]string)
}

// HasFieldError checks if a specific field has an error
func (df *DataForm) HasFieldError(fieldName string) bool {
	_, exists := df.fieldErrors[fieldName]
	return exists
}

// HasFieldErrors checks to see if there are any errors on the form
func (df *DataForm) HasFieldErrors() bool {
	return len(df.fieldErrors) > 0
}

// RemoveDeleteButton removes the delete button if it exists (3 buttons = Save, Cancel, Delete)
func (df *DataForm) RemoveDeleteButton() {
	if df.GetButtonCount() == 3 {
		df.RemoveButton(2) // Remove the Delete button (index 2)
	}
}

func (df *DataForm) AddDeleteButton() {
	// Add delete button for edit mode (insert at the beginning)
	if df.GetButtonCount() == 2 { // Only Save and Cancel exist
		df.AddButton("Delete (Ctrl+D)", func() {
			if df.onDelete != nil {
				df.onDelete()
			}
		})
	}
}

// SetHelpTextChangeHandler sets a callback for when help text changes
func (df *DataForm) SetHelpTextChangeHandler(handler func(string)) {
	df.onHelpTextChange = handler
}

// NotifyHelpTextChange notifies the handler when help text changes
// Make it public so forms can call it
func (df *DataForm) NotifyHelpTextChange(text string) {
	if df.onHelpTextChange != nil {
		df.onHelpTextChange(text)
	}
}

// SetupHandlers stores the handler functions
// Note: This doesn't add buttons - that's left to the specific form implementation
func (df *DataForm) SetupHandlers(onSave, onCancel, onDelete func()) {
	df.onSave = onSave
	df.onCancel = onCancel
	df.onDelete = onDelete

	// Clear and re-add buttons
	df.ClearButtons()

	df.AddButton("Save (Ctrl+S)", func() {
		if df.onSave != nil {
			df.onSave()
		}
	})

	df.AddButton("Cancel (Esc)", func() {
		if df.onCancel != nil {
			df.onCancel()
		}
	})

	// Add key capture for buttons
	df.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if df.onCancel != nil {
				df.onCancel()
			}
			return nil
		case tcell.KeyCtrlS:
			if df.onSave != nil {
				df.onSave()
			}
			return nil
		case tcell.KeyCtrlD:
			if df.GetButtonCount() == 3 && df.onDelete != nil {
				df.onDelete()
			}
		}
		return event
	})

}
