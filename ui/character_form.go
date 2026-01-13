package ui

import (
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CharacterForm represents a form for creating/editing characters
type CharacterForm struct {
	*sharedui.DataForm
	characterID  *int64
	nameField    *tview.InputField
	systemField  *tview.InputField
	roleField    *tview.InputField
	speciesField *tview.InputField
}

// NewCharacterForm creates a new character form
func NewCharacterForm() *CharacterForm {
	cf := &CharacterForm{
		DataForm: sharedui.NewDataForm(),
	}

	// Name field
	cf.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0) // 0 means full width

	// System field
	cf.systemField = tview.NewInputField().
		SetLabel("System").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	// Role field
	cf.roleField = tview.NewInputField().
		SetLabel("Role/Class").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	// Species field
	cf.speciesField = tview.NewInputField().
		SetLabel("Species/Race").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	cf.setupForm()
	return cf
}

// Fill the fields with the data from the character passed in
func (cf *CharacterForm) PopulateForEdit(char *character.Character) {
	cf.characterID = &char.ID

	cf.nameField.SetText(char.Name)
	cf.systemField.SetText(char.System)
	cf.roleField.SetText(char.Role)
	cf.speciesField.SetText(char.Species)

	cf.AddDeleteButton()

	cf.SetFocus(0)
	cf.SetTitle(" Edit Character ")
}

func (cf *CharacterForm) setupForm() {
	cf.Clear(true)

	cf.AddFormItem(cf.nameField)
	cf.AddFormItem(cf.systemField)
	cf.AddFormItem(cf.roleField)
	cf.AddFormItem(cf.speciesField)

	// Buttons will be set up when handlers are attached
	cf.SetBorder(true).
		SetTitle(" New Character ").
		SetTitleAlign(tview.AlignLeft)

	cf.SetButtonsAlign(tview.AlignCenter)

	// Add spacing between form items (1 line vertical space)
	cf.SetItemPadding(1)

}

// Reset clears all form fields
func (cf *CharacterForm) Reset() {
	cf.characterID = nil
	cf.nameField.SetText("")
	cf.systemField.SetText("")
	cf.roleField.SetText("")
	cf.speciesField.SetText("")
	cf.ClearFieldErrors()

	cf.RemoveDeleteButton()

	cf.SetFocus(0)
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (cf *CharacterForm) SetFieldErrors(errors map[string]string) {
	cf.DataForm.SetFieldErrors(errors)
	cf.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (cf *CharacterForm) updateFieldLabels() {

	// Update name field label
	if cf.HasFieldError("name") {
		cf.nameField.SetLabel("[red]Name[white]")
	} else {
		cf.nameField.SetLabel("Name")
	}

	// Update system field label
	if cf.HasFieldError("system") {
		cf.systemField.SetLabel("[red]System[white]")
	} else {
		cf.systemField.SetLabel("System")
	}

	// Update role field label
	if cf.HasFieldError("role") {
		cf.roleField.SetLabel("[red]Role/Class[white]")
	} else {
		cf.roleField.SetLabel("Role/Class")
	}

	// Update species field label
	if cf.HasFieldError("species") {
		cf.speciesField.SetLabel("[red]Species/Race[white]")
	} else {
		cf.speciesField.SetLabel("Species/Race")
	}

}

// ClearFieldErrors removes all error highlights
func (cf *CharacterForm) ClearFieldErrors() {
	cf.DataForm.ClearFieldErrors()
	cf.updateFieldLabels()
}

// BuildDomain constructs a Character entity from the form data
func (cf *CharacterForm) BuildDomain() *character.Character {

	c := &character.Character{
		Name:    cf.nameField.GetText(),
		System:  cf.systemField.GetText(),
		Role:    cf.roleField.GetText(),
		Species: cf.speciesField.GetText(),
	}

	// If editing an existing character, set the ID
	if cf.characterID != nil {
		c.ID = *cf.characterID
	}

	return c
}
