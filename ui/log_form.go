package ui

import (
	"soloterm/domain/log"
	sharedui "soloterm/shared/ui"

	"github.com/rivo/tview"
)

// LogForm represents a form for creating/editing log entries
type LogForm struct {
	*sharedui.DataForm
	id               *int64
	gameID           int64
	logTypeField     *tview.DropDown
	resultField      *tview.InputField
	narrativeField   *tview.TextArea
	descriptionField *tview.InputField
	errorMessage     *tview.TextView
}

// NewLogForm creates a new log form
func NewLogForm() *LogForm {
	lf := &LogForm{
		DataForm: sharedui.NewDataForm(),
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
	lf.setupFieldFocusHandlers()
	return lf
}

func (lf *LogForm) setupForm() {
	lf.Clear(true)

	lf.AddFormItem(lf.logTypeField)
	lf.AddFormItem(lf.descriptionField)
	lf.AddFormItem(lf.resultField)
	lf.AddFormItem(lf.narrativeField)

	// Buttons will be set up when handlers are attached
	// Note: Border and title managed by modal container
	lf.SetBorder(false)

	lf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	lf.SetItemPadding(1)
}

// setupFieldFocusHandlers configures focus handlers for each field to show contextual help
func (lf *LogForm) setupFieldFocusHandlers() {
	lf.logTypeField.SetFocusFunc(func() {
		lf.NotifyHelpTextChange("[yellow]Log Type:[white] Choose the type of log entry - Character Action (requires narrative), Oracle Question, Mechanics, or Story (narrative only)")
	})

	lf.descriptionField.SetFocusFunc(func() {
		lf.NotifyHelpTextChange("[yellow]Description:[white] A brief summary of what happened or what question was asked (required for non-Story entries)")
	})

	lf.resultField.SetFocusFunc(func() {
		lf.NotifyHelpTextChange("[yellow]Result:[white] The outcome or answer - dice rolls, oracle results, or resolution (required for non-Story entries)")
	})

	lf.narrativeField.SetFocusFunc(func() {
		lf.NotifyHelpTextChange("[yellow]Narrative:[white] The story description of events as they unfolded (required for Character Action and Story entries)")
	})
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
	lf.SetFocus(0)

	// Trigger initial help text
	lf.NotifyHelpTextChange("[yellow]Log Type:[white] Choose the type of log entry - Character Action (requires narrative), Oracle Question, Mechanics, or Story (narrative only)")
}

// PopulateForEdit fills the form with existing log data for editing
func (lf *LogForm) PopulateForEdit(logEntry *log.Log) {
	lf.id = &logEntry.ID
	lf.gameID = logEntry.GameID

	// Set the log type dropdown by matching LogType to index
	for i, lt := range log.LogTypes() {
		if lt == logEntry.LogType {
			lf.logTypeField.SetCurrentOption(i)
			break
		}
	}

	// Set the text fields
	lf.descriptionField.SetText(logEntry.Description)
	lf.resultField.SetText(logEntry.Result)
	lf.narrativeField.SetText(logEntry.Narrative, false)

	lf.AddDeleteButton()

	lf.ClearFieldErrors()
	lf.SetFocus(0)

	// Trigger initial help text
	lf.NotifyHelpTextChange("[yellow]Log Type:[white] Choose the type of log entry - Character Action (requires narrative), Oracle Question, Mechanics, or Story (narrative only)")
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (lf *LogForm) SetFieldErrors(errors map[string]string) {
	lf.DataForm.SetFieldErrors(errors)
	lf.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (lf *LogForm) updateFieldLabels() {

	// Update log type field label
	if lf.HasFieldError("log_type") {
		lf.logTypeField.SetLabel("[red]Log Type[white]")
	} else {
		lf.logTypeField.SetLabel("Log Type")
	}

	// Update description field label
	if lf.HasFieldError("description") {
		lf.descriptionField.SetLabel("[red]Description[white]")
	} else {
		lf.descriptionField.SetLabel("Description")
	}

	// Update result field label
	if lf.HasFieldError("result") {
		lf.resultField.SetLabel("[red]Result[white]")
	} else {
		lf.resultField.SetLabel("Result")
	}

	// Update narrative field label
	if lf.HasFieldError("narrative") {
		lf.narrativeField.SetLabel("[red]Narrative[white]")
	} else {
		lf.narrativeField.SetLabel("Narrative")
	}

}

// ClearFieldErrors removes all error highlights
func (lf *LogForm) ClearFieldErrors() {
	lf.DataForm.ClearFieldErrors()
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
