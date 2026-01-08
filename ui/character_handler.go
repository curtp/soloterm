package ui

import (
	syslog "log"
	"soloterm/domain/character"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CharacterHandler coordinates character-related UI operations
type CharacterHandler struct {
	app *App
}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler(app *App) *CharacterHandler {
	return &CharacterHandler{
		app: app,
	}
}

// LoadCharacters loads all characters
func (h *CharacterHandler) LoadCharacters() (map[string][]*character.Character, error) {
	// Initialize the map
	charsBySystem := make(map[string][]*character.Character)

	// Load characters from database
	chars, err := h.app.charService.GetAll()
	if err != nil {
		return nil, err
	}

	// If there aren't any, return an empty map
	if len(chars) == 0 {
		return charsBySystem, nil
	}

	// Group them by system
	for _, c := range chars {
		charsBySystem[c.System] = append(charsBySystem[c.System], c)
	}

	return charsBySystem, nil
}

// HandleSave saves the character from the form
func (h *CharacterHandler) HandleSave() {
	char := h.app.characterForm.BuildDomain()

	// Validate and save - get the saved character back from the service
	savedChar, err := h.app.charService.Save(char)
	if err != nil {
		// Check if it's a validation error
		if handleValidationError(err, h.app.characterForm) {
			return
		}
		h.app.notification.ShowError("Failed to save character: " + err.Error())
		return
	}

	// Update the app's selected character with the saved version from database
	h.app.selectedCharacter = savedChar

	// Orchestrate UI updates
	h.app.UpdateView(CHARACTER_SAVED)
}

// HandleCancel cancels character editing
func (h *CharacterHandler) HandleCancel() {
	h.app.UpdateView(CHARACTER_CANCEL)
}

func (h *CharacterHandler) HandleDuplicate() {
	// Show confirmation
	h.app.confirmModal.Show(
		"Are you sure you want to duplicate this character and their sheet?",
		func() {
			// User confirmed - duplicate the character
			char, err := h.app.charService.Duplicate(h.app.selectedCharacter.ID)
			if err != nil {
				h.app.UpdateView(CONFIRM_CANCEL)
				h.app.notification.ShowError("Failed to duplicate the character: " + err.Error())
				return
			}

			// Select the duplicated character
			h.app.selectedCharacter = char

			// Orchestrate UI updates
			h.app.UpdateView(CHARACTER_DUPLICATED)
		},
		func() {
			// User cancelled - just close confirmation modal
			h.app.UpdateView(CONFIRM_CANCEL)
		},
		"Duplicate", // Custom confirm button label
	)
	h.app.UpdateView(CONFIRM_SHOW)

}

// HandleDelete deletes the current character
func (h *CharacterHandler) HandleDelete() {
	char := h.app.characterForm.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if char.ID == 0 {
		h.app.UpdateView(CHARACTER_CANCEL)
		return
	}

	// Show confirmation
	h.app.confirmModal.Show(
		"Are you sure you want to delete this character?\n\nThis will also delete all associated attributes.",
		func() {
			// User confirmed - delete the character
			err := h.app.charService.Delete(char.ID)
			if err != nil {
				h.app.UpdateView(CONFIRM_CANCEL)
				h.app.notification.ShowError("Failed to delete character: " + err.Error())
				return
			}

			// Orchestrate UI updates
			h.app.UpdateView(CHARACTER_DELETED)
		},
		func() {
			// User cancelled - just close confirmation modal
			h.app.UpdateView(CONFIRM_CANCEL)
		},
	)
	h.app.UpdateView(CONFIRM_SHOW)
}

// ShowModal displays the character form modal for creating a new character
func (h *CharacterHandler) ShowModal() {
	h.app.UpdateView(CHARACTER_SHOW_NEW)
}

// ShowEditCharacterModal displays the character form modal for editing the selected character
func (h *CharacterHandler) ShowEditCharacterModal() {
	if h.app.selectedCharacter != nil {
		h.app.UpdateView(CHARACTER_SHOW_EDIT)
	}
}

// HandleAttributeSave saves the attribute from the form
func (h *CharacterHandler) HandleAttributeSave() {
	attr := h.app.attributeForm.BuildDomain()

	// Validate and save
	_, err := h.app.charService.SaveAttribute(attr)
	if err != nil {
		syslog.Printf("Problem saving attribute: %v", err)
		// Check if it's a validation error
		if handleValidationError(err, h.app.attributeForm) {
			return
		}
		h.app.notification.ShowError("Failed to save attribute: " + err.Error())
		return
	}

	// Reload attributes for the current character
	if h.app.selectedCharacter != nil {
		h.LoadAndDisplayAttributes(h.app.selectedCharacter.ID)
	}

	// Orchestrate UI updates
	h.app.UpdateView(ATTRIBUTE_SAVED)
}

// HandleAttributeCancel cancels attribute editing
func (h *CharacterHandler) HandleAttributeCancel() {
	h.app.UpdateView(ATTRIBUTE_CANCEL)
}

// HandleAttributeDelete deletes the current attribute
func (h *CharacterHandler) HandleAttributeDelete() {
	attr := h.app.attributeForm.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if attr.ID == 0 {
		h.app.UpdateView(ATTRIBUTE_CANCEL)
		return
	}

	// Show confirmation
	h.app.confirmModal.Show(
		"Are you sure you want to delete this attribute?",
		func() {
			// User confirmed - delete the attribute
			err := h.app.charService.DeleteAttribute(attr.ID)
			if err != nil {
				h.app.UpdateView(CONFIRM_CANCEL)
				h.app.notification.ShowError("Failed to delete attribute: " + err.Error())
				return
			}

			// Reload attributes for the current character
			if h.app.selectedCharacter != nil {
				h.LoadAndDisplayAttributes(h.app.selectedCharacter.ID)
			}

			// Orchestrate UI updates
			h.app.UpdateView(ATTRIBUTE_DELETED)
		},
		func() {
			// User cancelled - just close confirmation modal
			h.app.UpdateView(CONFIRM_CANCEL)
		},
	)
	h.app.UpdateView(CONFIRM_SHOW)
}

// ShowEditAttributeModal displays the attribute form modal for editing an existing attribute
func (h *CharacterHandler) ShowEditAttributeModal(attr *character.Attribute) {
	h.app.attributeForm.PopulateForEdit(attr)
	h.app.UpdateView(ATTRIBUTE_SHOW_EDIT)
}

// ShowNewAttributeModal displays the attribute form modal for creating a new attribute
func (h *CharacterHandler) ShowNewAttributeModal() {
	if h.app.selectedCharacter != nil {
		h.app.attributeForm.Reset(h.app.selectedCharacter.ID)
		h.app.UpdateView(ATTRIBUTE_SHOW_NEW)
	}
}

// DisplayCharacterInfo displays character information in the character info view
func (h *CharacterHandler) DisplayCharacterInfo(char *character.Character) {
	charInfo := "[aqua::b]" + char.Name + "[-::-]\n"
	charInfo += "[yellow::b]      System:[white::-] " + char.System + "\n"
	charInfo += "[yellow::b]  Role/Class:[white::-] " + char.Role + "\n"
	charInfo += "[yellow::b]Species/Race:[white::-] " + char.Species
	h.app.charInfoView.SetText(charInfo)
}

// LoadAndDisplayAttributes loads and displays attributes for a character
func (h *CharacterHandler) LoadAndDisplayAttributes(characterID int64) {
	// Load attributes for this character
	attrs, err := h.app.charService.GetAttributesForCharacter(characterID)
	if err != nil {
		attrs = []*character.Attribute{}
	}

	// Clear and repopulate attribute table
	h.app.attributeTable.Clear()

	// Add header row
	h.app.attributeTable.SetCell(0, 0, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))
	h.app.attributeTable.SetCell(0, 1, tview.NewTableCell("").
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignLeft).
		SetSelectable(false))

	// Add attribute rows (starting from row 1)
	for i, attr := range attrs {
		row := i + 1
		h.app.attributeTable.SetCell(row, 0, tview.NewTableCell(tview.Escape(attr.Name)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(0))
		h.app.attributeTable.SetCell(row, 1, tview.NewTableCell(tview.Escape(attr.Value)).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1))
	}

	// Show message if no attributes
	if len(attrs) == 0 {
		h.app.attributeTable.SetCell(2, 0, tview.NewTableCell("(No attributes - press 'a' to add)").
			SetTextColor(tcell.ColorGray).
			SetAlign(tview.AlignCenter).
			SetExpansion(2))
	}
}
