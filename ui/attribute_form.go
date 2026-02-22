package ui

import (
	"fmt"
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AttributeForm represents a form for creating/editing attributes
type AttributeForm struct {
	*sharedui.DataForm

	attributeID   *int64
	characterID   int64
	nameField     *tview.InputField
	valueField    *tview.InputField
	groupField    *tview.InputField
	positionField *tview.InputField
}

// NewAttributeForm creates a new attribute form
func NewAttributeForm() *AttributeForm {
	af := &AttributeForm{
		DataForm: sharedui.NewDataForm(),
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

// setupFieldFocusHandlers configures focus handlers for each field to show contextual help
func (af *AttributeForm) setupFieldFocusHandlers() {
	af.nameField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this entry (e.g., Strength, HP, Armor Class)")
	})

	af.valueField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Value:[" + Style.NormalTextColor + "] The current value of this entry (e.g., 18, 50/100, +5)")
	})

	af.groupField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Group:[" + Style.NormalTextColor + "] Organizes related entries together. Entries with the same group number will be displayed in the same section. Use 0 for the first group, 1 for the next, etc.")
	})

	af.positionField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Position In Group:[" + Style.NormalTextColor + "] Controls the order within a group. Position 0 is shown first and will be styled if there are other entries in the same group.")
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

	af.RemoveDeleteButton()

	af.ClearFieldErrors()

	af.SetFocus(0)

	// Trigger initial help text
	af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this attribute (e.g., Strength, HP, Armor Class)")
}

// PopulateForEdit fills the form with existing attribute data for editing
func (af *AttributeForm) PopulateForEdit(attr *character.Attribute) {
	af.attributeID = &attr.ID
	af.characterID = attr.CharacterID
	af.nameField.SetText(attr.Name)
	af.valueField.SetText(attr.Value)
	af.groupField.SetText(fmt.Sprintf("%d", attr.Group))
	af.positionField.SetText(fmt.Sprintf("%d", attr.PositionInGroup))

	af.AddDeleteButton()
	af.ClearFieldErrors()

	af.SetFocus(0)

	// Trigger initial help text
	af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this attribute (e.g., Strength, HP, Armor Class)")
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (af *AttributeForm) SetFieldErrors(errors map[string]string) {
	af.DataForm.SetFieldErrors(errors)
	af.updateFieldLabels()
}

// updateFieldLabels updates field labels to show errors
func (af *AttributeForm) updateFieldLabels() {
	// Update name field label
	if af.HasFieldError("name") {
		af.nameField.SetLabel("[" + Style.ErrorTextColor + "]Name[" + Style.NormalTextColor + "]")
	} else {
		af.nameField.SetLabel("Name")
	}

	// Update value field label
	if af.HasFieldError("value") {
		af.valueField.SetLabel("[" + Style.ErrorTextColor + "]Value[" + Style.NormalTextColor + "]")
	} else {
		af.valueField.SetLabel("Value")
	}

	if af.HasFieldError("group") {
		af.groupField.SetLabel("[" + Style.ErrorTextColor + "]Group[" + Style.NormalTextColor + "]")
	} else {
		af.groupField.SetLabel("Group")
	}

	if af.HasFieldError("position_in_group") {
		af.positionField.SetLabel("[" + Style.ErrorTextColor + "]Position In Group[" + Style.NormalTextColor + "]")
	} else {
		af.positionField.SetLabel("Position In Group")
	}

}

// ClearFieldErrors removes all error highlights
func (af *AttributeForm) ClearFieldErrors() {
	af.DataForm.ClearFieldErrors()
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
