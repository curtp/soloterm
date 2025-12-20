package ui

import (
	"soloterm/domain/log"

	"github.com/rivo/tview"
)

// LogForm represents a form for creating/editing log entries
type LogForm struct {
	*tview.Form
	id               *int64
	gameID           int64
	logTypeField     *tview.DropDown
	resultField      *tview.InputField
	narrativeField   *tview.TextArea
	descriptionField *tview.InputField
	errorMessage     *tview.TextView
	fieldErrors      map[string]string // Track which fields have errors
	onSave           func()
	onCancel         func()
	onDelete         func()
}

// NewLogForm creates a new log form
func NewLogForm() *LogForm {
	lf := &LogForm{
		Form:        tview.NewForm(),
		fieldErrors: make(map[string]string),
	}

	lf.errorMessage = tview.NewTextView().
		SetLabel("").
		SetText("")

	// Log Type field (using display names)
	lf.logTypeField = tview.NewDropDown().
		SetLabel("Log Type").
		SetOptions([]string{
			log.CHARACTER_ACTION.LogTypeDisplayName(),
			log.MECHANICS.LogTypeDisplayName(),
			log.ORACLE_QUESTION.LogTypeDisplayName(),
			log.STORY.LogTypeDisplayName(),
		}, nil)

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

	lf.AddFormItem(lf.logTypeField)
	lf.AddFormItem(lf.descriptionField)
	lf.AddFormItem(lf.resultField)
	lf.AddFormItem(lf.narrativeField)

	// Buttons will be set up when handlers are attached
	lf.SetBorder(true).
		SetTitle(" New Log ").
		SetTitleAlign(tview.AlignLeft)

	lf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	lf.SetItemPadding(1)
}

// Reset clears all form fields and sets the game ID for new log entry
func (lf *LogForm) Reset(gameID int64) {
	lf.id = nil
	lf.gameID = gameID
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
	lf.gameID = logEntry.GameID

	// Set the log type dropdown by matching LogType to index
	logTypes := log.LogTypes()
	for i, lt := range logTypes {
		if lt == logEntry.LogType {
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
			if lf.onDelete != nil {
				lf.onDelete()
			}
		})
	}

	lf.ClearFieldErrors()
	lf.SetTitle(" Edit Log ")
	lf.SetFocus(0)
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (lf *LogForm) SetFieldErrors(errors map[string]string) {
	lf.fieldErrors = errors
	lf.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (lf *LogForm) updateFieldLabels() {

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
	lf.updateFieldLabels()
}

// BuildDomain constructs a Log entity from the form data
func (lf *LogForm) BuildDomain() *log.Log {
	index, _ := lf.logTypeField.GetCurrentOption()

	// Map dropdown index to LogType
	logTypes := log.LogTypes()

	var logType log.LogType
	if index >= 0 && index < len(logTypes) {
		logType = logTypes[index]
	}

	l := &log.Log{
		GameID:      lf.gameID,
		LogType:     logType,
		Description: lf.descriptionField.GetText(),
		Result:      lf.resultField.GetText(),
		Narrative:   lf.narrativeField.GetText(),
	}

	// If editing an existing log, set the ID
	if lf.id != nil {
		l.ID = *lf.id
	}

	return l
}

// SetupHandlers configures all form button handlers
func (lf *LogForm) SetupHandlers(onSave, onCancel, onDelete func()) {
	lf.onSave = onSave
	lf.onCancel = onCancel
	lf.onDelete = onDelete

	// Clear and re-add buttons
	lf.ClearButtons()

	lf.AddButton("Save", func() {
		if lf.onSave != nil {
			lf.onSave()
		}
	})

	lf.AddButton("Cancel", func() {
		if lf.onCancel != nil {
			lf.onCancel()
		}
	})
}
