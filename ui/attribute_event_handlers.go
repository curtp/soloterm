package ui

import (
	syslog "log"
)

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
	a.attributeView.Form.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.characterView.RefreshDisplay()
	a.attributeView.Select(e.Attribute.ID)
	a.SetFocus(a.attributeView.Table)
	a.notification.ShowSuccess("Entry saved successfully")
}

func (a *App) handleAttributeCancel(_ *AttributeCancelledEvent) {
	a.attributeView.Form.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeView.Table)
}

func (a *App) handleAttributeDeleteConfirm(e *AttributeDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this entry?",
		func() {
			a.attributeView.ConfirmDelete(e.Attribute.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleAttributeDeleted(_ *AttributeDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.characterView.RefreshDisplay()
	a.SetFocus(a.attributeView.Table)
	a.notification.ShowSuccess("Entry deleted successfully")
}

func (a *App) handleAttributeDeleteFailed(e *AttributeDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete entry: " + e.Error.Error())
}

func (a *App) handleAttributeShowNew(e *AttributeShowNewEvent) {
	a.attributeView.ModalContent.SetTitle(" New Entry ")
	attrs, err := a.attributeView.attrService.GetForCharacter(e.CharacterID)
	if err != nil {
		syslog.Printf("Failed to open the New Entry modal: %s", err)
		a.notification.ShowError("Failed to open new entry form: " + err.Error())
		return
	}
	a.attributeView.Form.Reset(e.CharacterID, attrs)
	if e.SelectedAttribute != nil {
		a.attributeView.Form.SelectGroup(e.SelectedAttribute.Group)
	}
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeView.Form)
}

func (a *App) handleAttributeShowEdit(e *AttributeShowEditEvent) {
	a.attributeView.ModalContent.SetTitle(" Edit Entry ")
	attrs, err := a.attributeView.attrService.GetForCharacter(e.Attribute.CharacterID)
	if err != nil {
		syslog.Printf("Failed to open the edit entry modal: %s", err)
		a.notification.ShowError("Failed to open edit entry modal: " + err.Error())
		return
	}
	a.attributeView.Form.PopulateForEdit(e.Attribute, attrs)
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeView.Form)
}

func (a *App) handleAttributeReorder(e *AttributeReorderEvent) {
	movedID, err := a.attributeView.attrService.Reorder(e.CharacterID, e.AttributeID, e.Direction)
	if err != nil {
		syslog.Printf("Failed to reorder the entry: %s", err)
		a.notification.ShowError("Failed to reorder: " + err.Error())
		return
	}
	a.characterView.RefreshDisplay()
	if movedID != 0 {
		a.attributeView.Select(movedID)
	}
	a.SetFocus(a.attributeView.Table)
}
