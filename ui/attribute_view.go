package ui

import (
	"slices"
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AttributeView provides attribute UI operations
type AttributeView struct {
	app         *App
	attrService *character.AttributeService
	attrOrder   []int64
	Table       *tview.Table
	Form        *AttributeForm
	Modal       *tview.Flex
	ModalContent *tview.Flex
}

// NewAttributeView creates a new attribute view
func NewAttributeView(app *App, attrService *character.AttributeService) *AttributeView {
	av := &AttributeView{
		app:         app,
		attrService: attrService,
	}
	av.Setup()
	return av
}

// Setup initializes all attribute-related UI components
func (av *AttributeView) Setup() {
	av.setupForm()
	av.setupTable()
	av.setupModal()
}

func (av *AttributeView) setupForm() {
	av.Form = NewAttributeForm()
}

// setupTable configures the attribute table
func (av *AttributeView) setupTable() {
	av.Table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false). // Make rows selectable
		SetFixed(1, 0)              // Fix the header and divider rows

	av.Table.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack))

	av.Table.SetBorder(false)

	// Set up input capture for attribute table
	av.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if av.app.GetSelectedCharacterID() == nil {
			av.app.notification.ShowWarning("Select a character before editing the sheet")
			return nil
		}

		// Handle Ctrl+N or Insert key to add new attribute
		if event.Key() == tcell.KeyCtrlN {
			av.ShowNewModal()
			return nil
		}

		// Handle Ctrl+E to edit selected attribute
		if event.Key() == tcell.KeyCtrlE {
			attr := av.GetSelected()
			if attr != nil {
				av.ShowEditModal(attr)
			}
			return nil
		}

		return event
	})
}

// setupModal configures the attribute form modal
func (av *AttributeView) setupModal() {
	// Create help text view at modal level (no border, will be inside form container)
	helpTextView := tview.NewTextView()
	helpTextView.SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true).
		SetBorder(false)

	// Create container with border that holds both form and help
	av.ModalContent = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(av.Form, 0, 1, true).
		AddItem(helpTextView, 3, 0, false) // Fixed 3-line height for help
	av.ModalContent.SetBorder(true).
		SetTitleAlign(tview.AlignLeft)

	// Subscribe to form's help text changes
	av.Form.SetHelpTextChangeHandler(func(text string) {
		helpTextView.SetText(text)
	})

	// Set up handlers
	av.Form.SetupHandlers(
		av.HandleSave,
		av.HandleCancel,
		av.HandleDelete,
	)

	// Center the modal on screen
	av.Modal = tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(av.ModalContent, 16, 1, true). // Increased height for help text
				AddItem(nil, 0, 1, false),
			60, 1, true, // Fixed width
		).
		AddItem(nil, 0, 1, false)

	av.Form.SetFocusFunc(func() {
		av.app.SetModalHelpMessage(*av.Form.DataForm)
		av.ModalContent.SetBorderColor(Style.BorderFocusColor)
	})

	av.Form.SetBlurFunc(func() {
		av.ModalContent.SetBorderColor(Style.BorderColor)
	})
}

// LoadAndDisplay loads and displays attributes for a character
func (av *AttributeView) LoadAndDisplay(characterID int64) {
	// Load attributes for this character
	attrs, err := av.attrService.GetForCharacter(characterID)
	if err != nil {
		attrs = []*character.Attribute{}
	}

	// Clear and repopulate attribute table
	av.Table.Clear()
	// Clear the array of attribute IDs
	av.attrOrder = nil

	// Add header row
	av.Table.SetCell(0, 0, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	av.Table.SetCell(0, 1, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	// Add attribute rows (starting from row 1)
	for i, attr := range attrs {
		// Put the attribute ID in the attrOrder array to track what attribute is where.
		av.attrOrder = append(av.attrOrder, attr.ID)

		name := tview.Escape(attr.Name)

		if attr.PositionInGroup > 0 && attr.GroupCount > 1 && attr.GroupCountAfterZero > 0 {
			name = " " + name
		} else {
			if attr.PositionInGroup == 0 && attr.GroupCount > 1 && attr.GroupCountAfterZero > 0 {
				name = "[::u]" + name + "[::-]"
			}
		}

		row := i + 1
		av.Table.SetCell(row, 0, tview.NewTableCell(name).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		av.Table.SetCell(row, 1, tview.NewTableCell(tview.Escape(attr.Value)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
	}

	// Show message if no attributes
	if len(attrs) == 0 {
		av.Table.SetCell(2, 0, tview.NewTableCell("(No Entries - Ctrl+N to Add)").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignCenter).
			SetExpansion(2))
	}
}

// Select selects an attribute by ID in the table
func (av *AttributeView) Select(attributeID int64) {
	// Get the index of the attribute ID to select
	index := slices.Index(av.attrOrder, attributeID)

	if index != -1 {
		// Add 1 to the index to account for the header row
		av.Table.Select(index+1, 0)
	}
}

// GetSelected returns the currently selected attribute
func (av *AttributeView) GetSelected() *character.Attribute {
	if av.app.GetSelectedCharacterID() == nil {
		return nil
	}

	// Load the attribute which is currently selected
	row, _ := av.Table.GetSelection()
	attrs, _ := av.attrService.GetForCharacter(*av.app.GetSelectedCharacterID())
	attrIndex := row - 1
	if attrIndex >= 0 && attrIndex < len(attrs) {
		return attrs[attrIndex]
	}

	return nil
}

// HandleSave saves the attribute from the form
func (av *AttributeView) HandleSave() {
	attr := av.Form.BuildDomain()

	// Validate and save
	savedAttr, err := av.attrService.Save(attr)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, av.Form) {
			return
		}
		av.app.notification.ShowError("Failed to save entry: " + err.Error())
		return
	}

	// Dispatch event with saved attribute
	av.app.HandleEvent(&AttributeSavedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SAVED},
		Attribute: savedAttr,
	})
}

// HandleCancel cancels attribute editing
func (av *AttributeView) HandleCancel() {
	av.app.HandleEvent(&AttributeCancelledEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
	})
}

// HandleDelete deletes the current attribute
func (av *AttributeView) HandleDelete() {
	attr := av.Form.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if attr.ID == 0 {
		av.app.HandleEvent(&AttributeCancelledEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
		})
		return
	}

	// Dispatch event to show confirmation
	av.app.HandleEvent(&AttributeDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_CONFIRM},
		Attribute: attr,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (av *AttributeView) ConfirmDelete(attributeID int64) {
	// Business logic: Delete the attribute
	err := av.attrService.Delete(attributeID)
	if err != nil {
		// Dispatch failure event with error
		av.app.HandleEvent(&AttributeDeleteFailedEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	av.app.HandleEvent(&AttributeDeletedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETED},
	})
}

// ShowEditModal displays the attribute form modal for editing an existing attribute
func (av *AttributeView) ShowEditModal(attr *character.Attribute) {
	if attr == nil {
		return
	}

	av.app.HandleEvent(&AttributeShowEditEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SHOW_EDIT},
		Attribute: attr,
	})
}

// ShowNewModal displays the attribute form modal for creating a new attribute
func (av *AttributeView) ShowNewModal() {
	if av.app.GetSelectedCharacterID() == nil {
		return
	}

	// Get the currently selected attribute to use its group as default
	attr := av.GetSelected()

	av.app.HandleEvent(&AttributeShowNewEvent{
		BaseEvent:         BaseEvent{action: ATTRIBUTE_SHOW_NEW},
		CharacterID:       *av.app.GetSelectedCharacterID(),
		SelectedAttribute: attr, // Pass selected attribute for default values
	})
}

// ShowHelpModal displays the character sheet help modal
func (av *AttributeView) ShowHelpModal() {
	av.app.HandleEvent(&ShowHelpEvent{
		BaseEvent:   BaseEvent{action: SHOW_HELP},
		Title:       "Character Sheet Help",
		ReturnFocus: av.Table,
		Text: strings.NewReplacer(
			"[yellow]", "["+Style.HelpKeyTextColor+"]",
			"[white]", "["+Style.NormalTextColor+"]",
			"[green]", "["+Style.HelpSectionColor+"]",
		).Replace(`Scroll Down To View All Help Options

[green]What to Track?[white]

It's recommended to only track the bare essentials like health, or key pieces of information that may change during the adventure. Short names and values are recommended.

[green]Sheet Entries[white]

[yellow]Name[white]: Name of the entry (HP, XP, Level). May be the name of a section header (Skills, Gear, Stats)
[yellow]Value[white]: Value to assign to the entry (10/10, 2, +2). It may be blank.
[yellow]Group[white]: Organizes related entries together. Entries with the same group number will be displayed in the same section.
[yellow]Position[white]: Controls the order within a group. Position 0 is shown first and will be styled if there are other entries in the same group.
`),
	})
}
