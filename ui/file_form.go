package ui

import (
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FileForm represents a form for importing/exporting files
type FileForm struct {
	*sharedui.DataForm
	pathField    *tview.InputField
	errorMessage *tview.TextView
}

// NewFileForm creates a new file form
func NewFileForm() *FileForm {
	ff := &FileForm{
		DataForm: sharedui.NewDataForm(),
	}

	ff.pathField = tview.NewInputField().
		SetLabel("File Path").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	ff.errorMessage = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)

	ff.setupForm()
	return ff
}

func (ff *FileForm) setupForm() {
	ff.Clear(true)

	ff.AddFormItem(ff.pathField)

	ff.SetBorder(true).
		SetTitle(" Import File ").
		SetTitleAlign(tview.AlignLeft)
	ff.SetButtonsAlign(tview.AlignCenter)
	ff.SetItemPadding(1)
}

// Reset clears the form fields and error
func (ff *FileForm) Reset() {
	ff.pathField.SetText("")
	ff.ClearError()
	ff.SetFocus(0)
}

// GetPath returns the file path entered by the user
func (ff *FileForm) GetPath() string {
	return ff.pathField.GetText()
}

// ShowError displays an error message on the form
func (ff *FileForm) ShowError(msg string) {
	ff.errorMessage.SetText("[red]" + msg)
}

// ClearError clears the error message
func (ff *FileForm) ClearError() {
	ff.errorMessage.SetText("")
}
