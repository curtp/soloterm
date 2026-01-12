package ui

import (
	"fmt"
	"soloterm/domain/character"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AttributeForm represents a form for creating/editing attributes
type AttributeForm struct {
	*tview.Form
	attributeID      *int64
	characterID      int64
	nameField        *tview.InputField
	valueField       *tview.InputField
	groupField       *tview.InputField
	positionField    *tview.InputField
	fieldErrors      map[string]string
	onSave           func()
	onCancel         func()
	onDelete         func()
	onHelpTextChange func(string)
}

// NewAttributeForm creates a new attribute form
func NewAttributeForm() *AttributeForm {
	af := &AttributeForm{
		Form:        tview.NewForm(),
		fieldErrors: make(map[string]string),
	}

	// Name field
	af.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	// Value field
	af.valueField = tview.NewInputField().
		SetLabel("Value").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	af.groupField = tview.NewInputField().
		SetLabel("Group").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	af.positionField = tview.NewInputField().
		SetLabel("Position In Group").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	af.setupForm()
	af.setupFieldFocusHandlers()
	return af
}

func (af *AttributeForm) setupForm() {
	af.Clear(true)

	af.AddFormItem(af.nameField)
	af.AddFormItem(af.valueField)
	af.AddFormItem(af.groupField)
	af.AddFormItem(af.positionField)

	// Buttons will be set up when handlers are attached
	// Note: Border and title managed by modal container
	af.SetBorder(false)

	af.SetButtonsAlign(tview.AlignCenter)
	af.SetItemPadding(1)
}

// SetHelpTextChangeHandler sets a callback for when help text changes
func (af *AttributeForm) SetHelpTextChangeHandler(handler func(string)) {
	af.onHelpTextChange = handler
}

// notifyHelpTextChange notifies the handler when help text changes
func (af *AttributeForm) notifyHelpTextChange(text string) {
	if af.onHelpTextChange != nil {
		af.onHelpTextChange(text)
	}
}

// setupFieldFocusHandlers configures focus handlers for each field to show contextual help
func (af *AttributeForm) setupFieldFocusHandlers() {
	af.nameField.SetFocusFunc(func() {
		af.notifyHelpTextChange("[yellow]Name:[white] The display name for this attribute (e.g., Strength, HP, Armor Class)")
	})

	af.valueField.SetFocusFunc(func() {
		af.notifyHelpTextChange("[yellow]Value:[white] The current value of this attribute (e.g., 18, 50/100, +5)")
	})

	af.groupField.SetFocusFunc(func() {
		af.notifyHelpTextChange("[yellow]Group:[white] Organizes related attributes together. Attributes with the same group number will be displayed in the same section. Use 0 for the first group, 1 for the next, etc.")
	})

	af.positionField.SetFocusFunc(func() {
		af.notifyHelpTextChange("[yellow]Position In Group:[white] Controls the order within a group. Position 0 is shown first and will be styled if there are other attributes in the same group.")
	})
}

// Reset clears all form fields and sets the character ID for new attribute
func (af *AttributeForm) Reset(characterID int64) {
	af.attributeID = nil
	af.characterID = characterID
	af.nameField.SetText("")
	af.valueField.SetText("")
	af.groupField.SetText("0")
	af.positionField.SetText("0")

	// Remove delete button if it exists (going back to create mode)
	if af.GetButtonCount() == 3 { // Save, Cancel, and Delete
		af.RemoveButton(2) // Remove the Delete button (index 2)
	}

	af.ClearFieldErrors()

	af.SetFocus(0)

	// Trigger initial help text
	af.notifyHelpTextChange("[yellow]Name:[white] The display name for this attribute (e.g., Strength, HP, Armor Class)")
}

// PopulateForEdit fills the form with existing attribute data for editing
func (af *AttributeForm) PopulateForEdit(attr *character.Attribute) {
	af.attributeID = &attr.ID
	af.characterID = attr.CharacterID
	af.nameField.SetText(attr.Name)
	af.valueField.SetText(attr.Value)
	af.groupField.SetText(fmt.Sprintf("%d", attr.Group))
	af.positionField.SetText(fmt.Sprintf("%d", attr.PositionInGroup))

	// Add delete button for edit mode
	if af.GetButtonCount() == 2 { // Only Save and Cancel exist
		af.AddButton("Delete", func() {
			if af.onDelete != nil {
				af.onDelete()
			}
		})
	}

	af.ClearFieldErrors()

	af.SetFocus(0)

	// Trigger initial help text
	af.notifyHelpTextChange("[yellow]Name:[white] The display name for this attribute (e.g., Strength, HP, Armor Class)")
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (af *AttributeForm) SetFieldErrors(errors map[string]string) {
	af.fieldErrors = errors
	af.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (af *AttributeForm) updateFieldLabels() {
	// Update name field label
	if _, hasError := af.fieldErrors["name"]; hasError {
		af.nameField.SetLabel("[red]Name[white]")
	} else {
		af.nameField.SetLabel("Name")
	}

	// Update value field label
	if _, hasError := af.fieldErrors["value"]; hasError {
		af.valueField.SetLabel("[red]Value[white]")
	} else {
		af.valueField.SetLabel("Value")
	}

	if _, hasError := af.fieldErrors["group"]; hasError {
		af.groupField.SetLabel("[red]Group[white]")
	} else {
		af.groupField.SetLabel("Group")
	}

	if _, hasError := af.fieldErrors["position_in_group"]; hasError {
		af.positionField.SetLabel("[red]Position In Group[white]")
	} else {
		af.positionField.SetLabel("Position In Group")
	}

}

// ClearFieldErrors removes all error highlights
func (af *AttributeForm) ClearFieldErrors() {
	af.fieldErrors = make(map[string]string)
	af.updateFieldLabels()
}

// BuildDomain constructs an Attribute entity from the form data
func (af *AttributeForm) BuildDomain() *character.Attribute {
	group, err := strconv.Atoi(af.groupField.GetText())
	if err != nil {
		group = 0
	}

	position, err := strconv.Atoi(af.positionField.GetText())
	if err != nil {
		position = 0
	}

	attr := &character.Attribute{
		CharacterID:     af.characterID,
		Name:            af.nameField.GetText(),
		Value:           af.valueField.GetText(),
		Group:           group,
		PositionInGroup: position,
	}

	// If editing an existing attribute, set the ID
	if af.attributeID != nil {
		attr.ID = *af.attributeID
	}

	return attr
}

// SetupHandlers configures all form button handlers
func (af *AttributeForm) SetupHandlers(onSave, onCancel, onDelete func()) {
	af.onSave = onSave
	af.onCancel = onCancel
	af.onDelete = onDelete

	// Clear and re-add buttons
	af.ClearButtons()

	af.AddButton("Save", func() {
		if af.onSave != nil {
			af.onSave()
		}
	})

	af.AddButton("Cancel", func() {
		if af.onCancel != nil {
			af.onCancel()
		}
	})
}
