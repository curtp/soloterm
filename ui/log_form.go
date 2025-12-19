package ui

import (
	"soloterm/domain/log"

	"github.com/rivo/tview"
)

// GameForm represents a form for creating/editing log entries
type LogForm struct {
	*tview.Form
	id               *int64
	logTypeField     *tview.DropDown
	resultField      *tview.InputField
	narrativeField   *tview.TextArea
	descriptionField *tview.InputField
	errorMessage     *tview.TextView
	fieldErrors      map[string]string // Track which fields have errors
	onSave           func(id *int64, logType log.LogType, description string, result string, narrative string)
	onCancel         func()
	onDelete         func(id int64)
}

// NewLogForm creates a new log form
func NewLogForm(onSave func(id *int64, logType log.LogType, description string, result string, narrative string), onCancel func(), onDelete func(id int64)) *LogForm {
	lf := &LogForm{
		Form:        tview.NewForm(),
		onSave:      onSave,
		onCancel:    onCancel,
		onDelete:    onDelete,
		fieldErrors: make(map[string]string),
	}

	lf.errorMessage = tview.NewTextView().
		SetLabel("").
		SetText("")

	// Log Type field
	lf.logTypeField = tview.NewDropDown().
		SetLabel("Log Type").
		SetOptions([]string{string(log.CHARACTER_ACTION), string(log.MECHANICS), string(log.ORACLE_QUESTION)}, nil)

	// Narrative field
	lf.narrativeField = tview.NewTextArea().
		SetLabel("Narrative").
		SetSize(8, 0)

	lf.descriptionField = tview.NewInputField().
		SetLabel("Description").
		SetFieldWidth(0) // 0 means full width

	lf.resultField = tview.NewInputField().
		SetLabel("Result").
		SetFieldWidth(0) // 0 means full width

	lf.setupForm()
	return lf
}

func (lf *LogForm) setupForm() {
	lf.Clear(true)

	// gf.AddFormItem(gf.errorMessage)
	lf.AddFormItem(lf.logTypeField)
	lf.AddFormItem(lf.descriptionField)
	lf.AddFormItem(lf.resultField)
	lf.AddFormItem(lf.narrativeField)

	// Add buttons
	lf.AddButton("Save", func() {
		id := lf.id
		_, selected := lf.logTypeField.GetCurrentOption()
		logType := log.LogTypeFor(selected)
		description := lf.descriptionField.GetText()
		result := lf.resultField.GetText()
		narrative := lf.narrativeField.GetText()

		if lf.onSave != nil {
			lf.onSave(id, logType, description, result, narrative)
		}
	})

	lf.AddButton("Cancel", func() {
		if lf.onCancel != nil {
			lf.onCancel()
		}
	})

	lf.SetBorder(true).
		SetTitle(" New Log ").
		SetTitleAlign(tview.AlignLeft)

	lf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	lf.SetItemPadding(1)
}

// Reset clears all form fields
func (lf *LogForm) Reset() {
	lf.id = nil
	lf.logTypeField.SetCurrentOption(-1)
	lf.narrativeField.SetText("", false)
	lf.resultField.SetText("")
	lf.descriptionField.SetText("")

	// Remove delete button if it exists (going back to create mode)
	if lf.GetButtonCount() == 3 { // Save, Cancel, and Delete
		lf.RemoveButton(2) // Remove the Delete button (index 2)
	}

	lf.ClearFieldErrors()
	lf.SetTitle(" New Log ")
	lf.SetFocus(0)
}

// PopulateForEdit fills the form with existing log data for editing
func (lf *LogForm) PopulateForEdit(logEntry *log.Log) {
	lf.id = &logEntry.ID

	// Set the log type dropdown
	options := []string{string(log.CHARACTER_ACTION), string(log.MECHANICS), string(log.ORACLE_QUESTION)}
	for i, opt := range options {
		if opt == string(logEntry.LogType) {
			lf.logTypeField.SetCurrentOption(i)
			break
		}
	}

	// Set the text fields
	lf.descriptionField.SetText(logEntry.Description)
	lf.resultField.SetText(logEntry.Result)
	lf.narrativeField.SetText(logEntry.Narrative, false)

	// Add delete button for edit mode (insert at the beginning)
	if lf.GetButtonCount() == 2 { // Only Save and Cancel exist
		lf.AddButton("Delete", func() {
			if lf.onDelete != nil && lf.id != nil {
				lf.onDelete(*lf.id)
			}
		})
	}

	lf.ClearFieldErrors()
	lf.SetTitle(" Edit Log ")
	lf.SetFocus(0)
}

// SetFieldError highlights a field and shows an error message
func (lf *LogForm) SetFieldError(field string, message string) {
	// Store the error for this field
	lf.fieldErrors[field] = message
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (lf *LogForm) SetFieldErrors(errors map[string]string) {
	lf.fieldErrors = errors
	lf.UpdateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (lf *LogForm) UpdateFieldLabels() {

	// Update log type field label
	if _, hasError := lf.fieldErrors["log_type"]; hasError {
		lf.logTypeField.SetLabel("[red]Log Type[white]")
	} else {
		lf.logTypeField.SetLabel("Log Type")
	}

	// Update description field label
	if _, hasError := lf.fieldErrors["description"]; hasError {
		lf.descriptionField.SetLabel("[red]Description[white]")
	} else {
		lf.descriptionField.SetLabel("Description")
	}

	// Update result field label
	if _, hasError := lf.fieldErrors["result"]; hasError {
		lf.resultField.SetLabel("[red]Result[white]")
	} else {
		lf.resultField.SetLabel("Result")
	}

	// Update narrative field label
	if _, hasError := lf.fieldErrors["narrative"]; hasError {
		lf.narrativeField.SetLabel("[red]Narrative[white]")
	} else {
		lf.narrativeField.SetLabel("Narrative")
	}

}

// ClearFieldErrors removes all error highlights
func (lf *LogForm) ClearFieldErrors() {
	lf.fieldErrors = make(map[string]string)
	lf.UpdateFieldLabels()
}
