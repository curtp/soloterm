package ui

import (
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"mellium.im/filechooser"
)

// ImportPosition controls where imported content is placed in the text area.
type ImportPosition int

const (
	ImportReplace  ImportPosition = iota // Replace all current content (default)
	ImportBefore                         // Insert before current content
	ImportAfter                          // Append after current content
	ImportAtCursor                       // Insert at the cursor position
)

// FileForm represents a form for importing/exporting files
type FileForm struct {
	*sharedui.DataForm
	pathField     *filechooser.PathInputField
	positionField *tview.DropDown
	lastErrorText string
}

// NewFileForm creates a new file form
func NewFileForm() *FileForm {
	ff := &FileForm{
		DataForm: sharedui.NewDataForm(),
	}

	ff.pathField = filechooser.NewPathInputField()
	ff.pathField.SetLabel("File Path").
		SetPlaceholder("Type / to browse files").
		SetPlaceholderStyle(ff.pathField.GetFieldStyle().Foreground(tcell.ColorDarkGrey)).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)
	ff.pathField.SetAutocompleteStyles(
		tcell.ColorGray,
		tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGray),
		tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorAqua),
	)

	ff.positionField = tview.NewDropDown().
		SetLabel("Insert at").
		SetOptions([]string{
			"Replace (default)",
			"Before current",
			"After current",
			"At cursor",
		}, nil).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetCurrentOption(0)

	ff.setupForm()
	return ff
}

func (ff *FileForm) setupForm() {
	ff.Clear(true)

	ff.AddFormItem(ff.pathField)

	ff.SetBorder(false)
	ff.SetButtonsAlign(tview.AlignCenter)
	ff.SetItemPadding(1)
}

// SetImportMode rebuilds the form items for import (with position selector) or export (path only).
func (ff *FileForm) SetImportMode(isImport bool) {
	ff.Clear(false) // clear items, keep buttons
	ff.AddFormItem(ff.pathField)
	if isImport {
		ff.positionField.SetCurrentOption(0)
		ff.AddFormItem(ff.positionField)
	}
}

// GetImportPosition returns the currently selected import position.
func (ff *FileForm) GetImportPosition() ImportPosition {
	idx, _ := ff.positionField.GetCurrentOption()
	return ImportPosition(idx)
}

// Reset clears the form fields and error, setting the path to defaultPath
func (ff *FileForm) Reset(defaultPath string) {
	ff.pathField.SetText(defaultPath)
	ff.pathField.Autocomplete() // dismiss any lingering autocomplete list
	ff.ClearError()
	ff.SetFocus(0)
}

// GetPath returns the file path entered by the user
func (ff *FileForm) GetPath() string {
	return ff.pathField.GetText()
}

// ShowError displays an error message on the form
func (ff *FileForm) ShowError(msg string) {
	ff.lastErrorText = "[red]" + msg
	ff.NotifyHelpTextChange(ff.lastErrorText)
}

// ClearError clears the error message
func (ff *FileForm) ClearError() {
	ff.lastErrorText = ""
	ff.NotifyHelpTextChange("")
}

// GetErrorText returns the current error text for testing
func (ff *FileForm) GetErrorText() string {
	return ff.lastErrorText
}
