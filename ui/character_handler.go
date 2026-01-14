package ui

import (
	"soloterm/domain/character"
	sharedui "soloterm/shared/ui"
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
		if sharedui.HandleValidationError(err, h.app.characterForm) {
			return
		}
		h.app.notification.ShowError("Failed to save character: " + err.Error())
		return
	}

	// Dispatch event with saved character
	h.app.HandleEvent(&CharacterSavedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SAVED},
		Character: savedChar,
	})
}

// HandleCancel cancels character editing
func (h *CharacterHandler) HandleCancel() {
	h.app.HandleEvent(&CharacterCancelledEvent{
		BaseEvent: BaseEvent{action: CHARACTER_CANCEL},
	})
}

func (h *CharacterHandler) HandleDuplicate() {
	if h.app.selectedCharacter == nil {
		h.app.notification.ShowWarning("Select a character to duplicate")
		return
	}

	// Dispatch event to show confirmation
	h.app.HandleEvent(&CharacterDuplicateConfirmEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DUPLICATE_CONFIRM},
		Character: h.app.selectedCharacter,
	})
}

// ConfirmDuplicate executes the actual duplication after user confirmation
func (h *CharacterHandler) ConfirmDuplicate(characterID int64) {
	// Business logic: Duplicate the character
	char, err := h.app.charService.Duplicate(characterID)
	if err != nil {
		// Dispatch failure event with error
		h.app.HandleEvent(&CharacterDuplicateFailedEvent{
			BaseEvent: BaseEvent{action: CHARACTER_DUPLICATE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	h.app.HandleEvent(&CharacterDuplicatedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DUPLICATED},
		Character: char,
	})
}

// HandleDelete deletes the current character
func (h *CharacterHandler) HandleDelete() {
	char := h.app.characterForm.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if char.ID == 0 {
		h.app.HandleEvent(&CharacterCancelledEvent{
			BaseEvent: BaseEvent{action: CHARACTER_CANCEL},
		})
		return
	}

	// Dispatch event to show confirmation
	h.app.HandleEvent(&CharacterDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DELETE_CONFIRM},
		Character: char,
	})
}

// ConfirmDelete executes the actual deletion after user confirmation
func (h *CharacterHandler) ConfirmDelete(characterID int64) {
	// Business logic: Delete the character
	err := h.app.charService.Delete(characterID)
	if err != nil {
		// Dispatch failure event with error
		h.app.HandleEvent(&CharacterDeleteFailedEvent{
			BaseEvent: BaseEvent{action: CHARACTER_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	h.app.HandleEvent(&CharacterDeletedEvent{
		BaseEvent: BaseEvent{action: CHARACTER_DELETED},
	})
}

// ShowModal displays the character form modal for creating a new character
func (h *CharacterHandler) ShowModal() {
	h.app.HandleEvent(&CharacterShowNewEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SHOW_NEW},
	})
}

// ShowEditCharacterModal displays the character form modal for editing the selected character
func (h *CharacterHandler) ShowEditCharacterModal() {
	if h.app.selectedCharacter == nil {
		return
	}

	h.app.HandleEvent(&CharacterShowEditEvent{
		BaseEvent: BaseEvent{action: CHARACTER_SHOW_EDIT},
		Character: h.app.selectedCharacter,
	})
}

// HandleAttributeSave saves the attribute from the form
func (h *CharacterHandler) HandleAttributeSave() {
	attr := h.app.attributeForm.BuildDomain()

	// Validate and save
	savedAttr, err := h.app.charService.SaveAttribute(attr)
	if err != nil {
		// Check if it's a validation error
		if sharedui.HandleValidationError(err, h.app.attributeForm) {
			return
		}
		h.app.notification.ShowError("Failed to save attribute: " + err.Error())
		return
	}

	// Dispatch event with saved attribute
	h.app.HandleEvent(&AttributeSavedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SAVED},
		Attribute: savedAttr,
	})
}

// HandleAttributeCancel cancels attribute editing
func (h *CharacterHandler) HandleAttributeCancel() {
	h.app.HandleEvent(&AttributeCancelledEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
	})
}

// HandleAttributeDelete deletes the current attribute
func (h *CharacterHandler) HandleAttributeDelete() {
	attr := h.app.attributeForm.BuildDomain()

	// Only delete if it has an ID (exists in database)
	if attr.ID == 0 {
		h.app.HandleEvent(&AttributeCancelledEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_CANCEL},
		})
		return
	}

	// Dispatch event to show confirmation
	h.app.HandleEvent(&AttributeDeleteConfirmEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_CONFIRM},
		Attribute: attr,
	})
}

// ConfirmAttributeDelete executes the actual deletion after user confirmation
func (h *CharacterHandler) ConfirmAttributeDelete(attributeID int64) {
	// Business logic: Delete the attribute
	err := h.app.charService.DeleteAttribute(attributeID)
	if err != nil {
		// Dispatch failure event with error
		h.app.HandleEvent(&AttributeDeleteFailedEvent{
			BaseEvent: BaseEvent{action: ATTRIBUTE_DELETE_FAILED},
			Error:     err,
		})
		return
	}

	// Dispatch success event
	h.app.HandleEvent(&AttributeDeletedEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_DELETED},
	})
}

// ShowEditAttributeModal displays the attribute form modal for editing an existing attribute
func (h *CharacterHandler) ShowEditAttributeModal(attr *character.Attribute) {
	if attr == nil {
		return
	}

	h.app.HandleEvent(&AttributeShowEditEvent{
		BaseEvent: BaseEvent{action: ATTRIBUTE_SHOW_EDIT},
		Attribute: attr,
	})
}

// ShowNewAttributeModal displays the attribute form modal for creating a new attribute
func (h *CharacterHandler) ShowNewAttributeModal() {
	if h.app.selectedCharacter == nil {
		return
	}

	// Get the currently selected attribute to use its group as default
	attr := h.GetSelectedAttribute()

	h.app.HandleEvent(&AttributeShowNewEvent{
		BaseEvent:         BaseEvent{action: ATTRIBUTE_SHOW_NEW},
		SelectedAttribute: attr, // Pass selected attribute for default values
	})
}

func (h *CharacterHandler) GetSelectedAttribute() *character.Attribute {
	if h.app.selectedCharacter == nil {
		return nil
	}

	// Load the attribute which is currently selected
	row, _ := h.app.attributeTable.GetSelection()
	attrs, _ := h.app.charService.GetAttributesForCharacter(h.app.selectedCharacter.ID)
	attrIndex := row - 1
	if attrIndex >= 0 && attrIndex < len(attrs) {
		return attrs[attrIndex]
	}

	return nil
}
