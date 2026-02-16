package ui

func (a *App) handleCharacterSaved(e *CharacterSavedEvent) {
	a.characterView.Form.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.characterView.RefreshTree()
	a.characterView.SelectCharacter(e.Character.ID)
	a.characterView.RefreshDisplay()
	a.SetFocus(a.characterView.ReturnFocus)
	a.notification.ShowSuccess("Character saved successfully")
}

func (a *App) handleCharacterCancel(_ *CharacterCancelledEvent) {
	a.characterView.Form.ClearFieldErrors()
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.SetFocus(a.characterView.ReturnFocus)
}

func (a *App) handleCharacterDeleteConfirm(e *CharacterDeleteConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to delete this character?\n\nThis will also delete their sheet.",
		func() {
			a.characterView.ConfirmDelete(e.Character.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleCharacterDeleted(_ *CharacterDeletedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.pages.HidePage(CHARACTER_MODAL_ID)
	a.pages.SwitchToPage(MAIN_PAGE_ID)
	a.characterView.InfoView.Clear()
	a.attributeView.Table.Clear()
	a.characterView.RefreshTree()
	a.SetFocus(a.characterView.ReturnFocus)
	a.notification.ShowSuccess("Character deleted successfully")
}

func (a *App) handleCharacterDeleteFailed(e *CharacterDeleteFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to delete character: " + e.Error.Error())
}

func (a *App) handleCharacterDuplicateConfirm(e *CharacterDuplicateConfirmEvent) {
	returnFocus := a.GetFocus()

	a.confirmModal.Configure(
		"Are you sure you want to duplicate this character and their sheet?",
		func() {
			a.characterView.ConfirmDuplicate(e.Character.ID)
		},
		func() {
			a.pages.HidePage(CONFIRM_MODAL_ID)
			a.SetFocus(returnFocus)
		},
		"Duplicate",
	)

	a.pages.ShowPage(CONFIRM_MODAL_ID)
}

func (a *App) handleCharacterDuplicated(e *CharacterDuplicatedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.characterView.RefreshTree()
	a.characterView.SelectCharacter(e.Character.ID)
	a.characterView.RefreshDisplay()
	a.SetFocus(a.characterView.ReturnFocus)
	a.notification.ShowSuccess("Character duplicated successfully")
}

func (a *App) handleCharacterDuplicateFailed(e *CharacterDuplicateFailedEvent) {
	a.pages.HidePage(CONFIRM_MODAL_ID)
	a.notification.ShowError("Failed to duplicate the character: " + e.Error.Error())
}

func (a *App) handleCharacterShowNew(_ *CharacterShowNewEvent) {
	a.characterView.Form.Reset()
	a.pages.ShowPage(CHARACTER_MODAL_ID)
	a.characterView.Form.SetTitle(" New Character ")
	a.SetFocus(a.characterView.Form)
}

func (a *App) handleCharacterShowEdit(e *CharacterShowEditEvent) {
	a.characterView.Form.PopulateForEdit(e.Character)
	a.pages.ShowPage(CHARACTER_MODAL_ID)
	a.characterView.Form.SetTitle(" Edit Character ")
	a.SetFocus(a.characterView.Form)
}
