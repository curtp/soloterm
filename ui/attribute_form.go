package ui

import (
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AttributeForm represents a form for creating/editing attributes
type AttributeForm struct {
	*sharedui.DataForm

	attributeID      *int64
	characterID      int64
	nameField        *tview.InputField
	valueField       *tview.InputField
	groupDropDown    *tview.DropDown
	groupHeaders     []*character.Attribute // first attr per group, in display order
	maxGroupNum      int                    // max group number across current attrs; -1 if none
	originalGroup    int                    // preserved in edit mode
	originalPosition int                    // preserved in edit mode
}

// NewAttributeForm creates a new attribute form
func NewAttributeForm() *AttributeForm {
	af := &AttributeForm{
		DataForm: sharedui.NewDataForm(),
	}

	af.nameField = tview.NewInputField().
		SetLabel("Name").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	af.valueField = tview.NewInputField().
		SetLabel("Value").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldWidth(0)

	af.groupDropDown = tview.NewDropDown().
		SetLabel("Section").
		SetFieldBackgroundColor(tcell.ColorDefault)

	af.setupForm()
	af.setupFieldFocusHandlers()
	return af
}

func (af *AttributeForm) setupForm() {
	af.Clear(true)
	af.SetBorder(false)
	af.SetButtonsAlign(tview.AlignCenter)
	af.SetItemPadding(1)
}

// rebuildItems clears form items (keeping buttons) and re-adds the given items.
// tview.Form has no show/hide for individual items, so this is the only way
// to switch between new mode (with dropdown) and edit mode (name+value only).
func (af *AttributeForm) rebuildItems(items []tview.FormItem) {
	af.Clear(false)
	for _, item := range items {
		af.AddFormItem(item)
	}
}

// buildGroupHeaders populates groupHeaders, maxGroupNum, and the dropdown labels
// from the current character's attribute list.
func (af *AttributeForm) buildGroupHeaders(attrs []*character.Attribute) {
	af.groupHeaders = nil
	af.maxGroupNum = -1

	seen := map[int]bool{}
	for _, a := range attrs {
		if !seen[a.Group] {
			seen[a.Group] = true
			af.groupHeaders = append(af.groupHeaders, a)
		}
		if a.Group > af.maxGroupNum {
			af.maxGroupNum = a.Group
		}
	}

	labels := []string{"- New -"}
	for _, h := range af.groupHeaders {
		labels = append(labels, h.Name)
	}
	af.groupDropDown.SetOptions(labels, nil)
	af.groupDropDown.SetCurrentOption(0)
}

// setupFieldFocusHandlers configures focus handlers for each field to show contextual help
func (af *AttributeForm) setupFieldFocusHandlers() {
	af.nameField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this entry (e.g., Strength, HP, Armor Class)")
	})

	af.valueField.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Value:[" + Style.NormalTextColor + "] The current value of this entry (e.g., 18, 50/100, +5)")
	})

	af.groupDropDown.SetFocusFunc(func() {
		af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Section:[" + Style.NormalTextColor + "] Choose an existing section to append to, or \"- New -\" to create a standalone section at the bottom of the sheet.")
	})
}

// Reset clears the form and sets it up for creating a new attribute.
// attrs is the current list for the character, used to populate the group dropdown.
func (af *AttributeForm) Reset(characterID int64, attrs []*character.Attribute) {
	af.attributeID = nil
	af.characterID = characterID
	af.nameField.SetText("")
	af.valueField.SetText("")
	af.buildGroupHeaders(attrs)
	af.rebuildItems([]tview.FormItem{af.nameField, af.valueField, af.groupDropDown})
	af.RemoveDeleteButton()
	af.ClearFieldErrors()
	af.SetFocus(0)
	af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this entry (e.g., Strength, HP, Armor Class)")
}

// SelectGroup pre-selects the dropdown to the option matching groupNum.
// Called after Reset when there is a currently selected attribute.
func (af *AttributeForm) SelectGroup(groupNum int) {
	for i, h := range af.groupHeaders {
		if h.Group == groupNum {
			af.groupDropDown.SetCurrentOption(i + 1) // +1 for "- New -" at index 0
			return
		}
	}
}

// PopulateForEdit fills the form with existing attribute data for editing.
// The group dropdown is shown so the user can move the entry to a different group.
func (af *AttributeForm) PopulateForEdit(attr *character.Attribute, attrs []*character.Attribute) {
	af.attributeID = &attr.ID
	af.characterID = attr.CharacterID
	af.originalGroup = attr.Group
	af.originalPosition = attr.PositionInGroup
	af.nameField.SetText(attr.Name)
	af.valueField.SetText(attr.Value)
	af.buildGroupHeaders(attrs)
	af.rebuildItems([]tview.FormItem{af.nameField, af.valueField, af.groupDropDown})
	af.SelectGroup(attr.Group)
	af.AddDeleteButton()
	af.ClearFieldErrors()
	af.SetFocus(0)
	af.NotifyHelpTextChange("[" + Style.HelpKeyTextColor + "]Name:[" + Style.NormalTextColor + "] The display name for this entry (e.g., Strength, HP, Armor Class)")
}

// SetFieldErrors sets multiple field errors at once and updates labels
func (af *AttributeForm) SetFieldErrors(errors map[string]string) {
	af.DataForm.SetFieldErrors(errors)
	af.updateFieldLabels()
}

// updateFieldLabels updates field labels to show validation errors
func (af *AttributeForm) updateFieldLabels() {
	if af.HasFieldError("name") {
		af.nameField.SetLabel("[" + Style.ErrorTextColor + "]Name[" + Style.NormalTextColor + "]")
	} else {
		af.nameField.SetLabel("Name")
	}

	if af.HasFieldError("value") {
		af.valueField.SetLabel("[" + Style.ErrorTextColor + "]Value[" + Style.NormalTextColor + "]")
	} else {
		af.valueField.SetLabel("Value")
	}
}

// ClearFieldErrors removes all error highlights
func (af *AttributeForm) ClearFieldErrors() {
	af.DataForm.ClearFieldErrors()
	af.updateFieldLabels()
}

// BuildDomain constructs an Attribute entity from the form data
func (af *AttributeForm) BuildDomain() *character.Attribute {
	var group, position int

	idx, _ := af.groupDropDown.GetCurrentOption()
	if idx == 0 || len(af.groupHeaders) == 0 {
		// "- New -": next available group at position 0
		group = af.maxGroupNum + 1
		position = 0
	} else {
		header := af.groupHeaders[idx-1]
		if af.attributeID != nil && header.Group == af.originalGroup {
			// Edit mode, same group selected: preserve original position
			group = af.originalGroup
			position = af.originalPosition
		} else {
			// New mode, or edit mode changing to a different group: append to end
			group = header.Group
			position = header.GroupCount
		}
	}

	attr := &character.Attribute{
		CharacterID:     af.characterID,
		Name:            af.nameField.GetText(),
		Value:           af.valueField.GetText(),
		Group:           group,
		PositionInGroup: position,
	}

	if af.attributeID != nil {
		attr.ID = *af.attributeID
	}

	return attr
}
