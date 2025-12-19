package ui

import (
	stdlog "log"
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
}

// NewLogForm creates a new log form
func NewLogForm(onSave func(id *int64, logType log.LogType, description string, result string, narrative string), onCancel func()) *LogForm {
	lf := &LogForm{
		Form:        tview.NewForm(),
		onSave:      onSave,
		onCancel:    onCancel,
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
		stdlog.Println("Log form Save button clicked")
		id := lf.id
		_, selected := lf.logTypeField.GetCurrentOption()
		logType := log.LogTypeFor(selected)
		description := lf.descriptionField.GetText()
		result := lf.resultField.GetText()
		narrative := lf.narrativeField.GetText()
		stdlog.Printf("Form data: logType=%s, desc=%s, result=%s, narrative=%s", logType, description, result, narrative)
		if lf.onSave != nil {
			stdlog.Println("Calling onSave callback")
			lf.onSave(id, logType, description, result, narrative)
		} else {
			stdlog.Println("WARNING: onSave callback is nil!")
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
	lf.ClearFieldErrors()
	lf.SetTitle(" New Log ")
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

	lf.ClearFieldErrors()
	lf.SetTitle(" Edit Log ")
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
