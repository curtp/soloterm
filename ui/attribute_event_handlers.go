package ui

import "fmt"

func (a *App) handleAttributeSaved(e *AttributeSavedEvent) {
	a.characterView.AttributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.characterView.RefreshDisplay()
	a.characterView.selectAttribute(e.Attribute.ID)
	a.SetFocus(a.characterView.AttributeTable)
	a.notification.ShowSuccess("Entry saved successfully")
}

func (a *App) handleAttributeCancel(_ *AttributeCancelledEvent) {
	a.characterView.AttributeForm.ClearFieldErrors()
	a.pages.HidePage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.characterView.AttributeTable)
}

func (a *App) handleAttributeDeleteConfirm(e *AttributeDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this entry?",
		func() {
			a.characterView.ConfirmAttributeDelete(e.Attribute.ID)
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
	a.SetFocus(a.characterView.AttributeTable)
	a.notification.ShowSuccess("Entry deleted successfully")
}

func (a *App) handleAttributeDeleteFailed(e *AttributeDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete entry: " + e.Error.Error())
}

func (a *App) handleAttributeShowNew(e *AttributeShowNewEvent) {
	a.characterView.AttributeModalContent.SetTitle(" New Entry ")
	a.characterView.AttributeForm.Reset(e.CharacterID)

	// If there's a selected attribute, use its group and position as defaults
	if e.SelectedAttribute != nil {
		a.characterView.AttributeForm.groupField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.Group))
		a.characterView.AttributeForm.positionField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.PositionInGroup+1))
	}

	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.characterView.AttributeForm)
}

func (a *App) handleAttributeShowEdit(e *AttributeShowEditEvent) {
	a.characterView.AttributeModalContent.SetTitle(" Edit Entry ")
	a.characterView.AttributeForm.PopulateForEdit(e.Attribute)
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.characterView.AttributeForm)
}
