package ui

import "fmt"

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
	a.attributeView.Form.Reset(e.CharacterID)

	// If there's a selected attribute, use its group and position as defaults
	if e.SelectedAttribute != nil {
		a.attributeView.Form.groupField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.Group))
		a.attributeView.Form.positionField.SetText(fmt.Sprintf("%d", e.SelectedAttribute.PositionInGroup+1))
	}

	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeView.Form)
}

func (a *App) handleAttributeShowEdit(e *AttributeShowEditEvent) {
	a.attributeView.ModalContent.SetTitle(" Edit Entry ")
	a.attributeView.Form.PopulateForEdit(e.Attribute)
	a.pages.ShowPage(ATTRIBUTE_MODAL_ID)
	a.SetFocus(a.attributeView.Form)
}
