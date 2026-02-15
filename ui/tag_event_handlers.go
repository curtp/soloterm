package ui

func (a *App) handleTagSelected(e *TagSelectedEvent) {
	a.pages.HidePage(TAG_MODAL_ID)
	a.sessionView.InsertAtCursor(e.TagType.Template)
	a.SetFocus(a.sessionView.TextArea)
}

func (a *App) handleTagCancelled(e *TagCancelledEvent) {
	a.pages.HidePage(TAG_MODAL_ID)
}

func (a *App) handleTagShow(e *TagShowEvent) {
	// Store current focus so we can restore it after tag selection
	a.tagView.returnFocus = a.GetFocus()
	a.tagView.Refresh()
	a.pages.ShowPage(TAG_MODAL_ID)
	a.SetFocus(a.tagView.TagTable)
}
