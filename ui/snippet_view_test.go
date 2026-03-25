package ui

import (
	"soloterm/domain/snippet"
	testHelper "soloterm/shared/testing"
	"testing"

	// Blank import to trigger snippet migration registration
	_ "soloterm/domain/snippet"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSnippet is a test helper that saves a snippet directly via the service.
func createSnippet(t *testing.T, app *App, name, content string, gameID *int64) *snippet.Snippet {
	t.Helper()
	all, _ := app.snippetView.snippetService.GetAll()
	s := &snippet.Snippet{
		Name:     name,
		Content:  content,
		GameID:   gameID,
		Position: len(all),
	}
	saved, err := app.snippetView.snippetService.Save(s)
	require.NoError(t, err, "failed to create test snippet")
	return saved
}

// openSnippetModal fires the SnippetShow event and asserts the modal is visible.
func openSnippetModal(t *testing.T, app *App) {
	t.Helper()
	app.HandleEvent(&SnippetShowEvent{BaseEvent: BaseEvent{action: SNIPPET_SHOW}})
	require.True(t, app.isPageVisible(SNIPPET_MODAL_ID), "snippet modal should be visible")
}

// TestSnippetView_OpensAndClosesModal verifies open and Esc close.
func TestSnippetView_OpensAndClosesModal(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)

	testHelper.SimulateEscape(app.snippetView.Modal, app.Application)

	assert.False(t, app.isPageVisible(SNIPPET_MODAL_ID))
}

// TestSnippetView_EmptyState_ShowsMessage verifies the empty-state cell is rendered.
func TestSnippetView_EmptyState_ShowsMessage(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)
	assert.Equal(t, 1, app.snippetView.table.GetRowCount(), "should have one placeholder row")
}

// TestSnippetView_CtrlN_OpensFormModal verifies Ctrl+N opens the form modal.
func TestSnippetView_CtrlN_OpensFormModal(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')

	assert.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID), "form modal should open on Ctrl+N")
}

// TestSnippetView_NewSnippet_SavedAndAppears verifies that creating a snippet
// via Ctrl+N + form save closes the form and shows the snippet in the table.
func TestSnippetView_NewSnippet_SavedAndAppears(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))

	app.snippetView.Form.nameField.SetText("Attack")
	app.snippetView.Form.contentField.SetText("1d20+5", false)
	testHelper.SimulateKey(app.snippetView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID), "form modal should close after save")
	assert.True(t, app.isPageVisible(SNIPPET_MODAL_ID), "list modal should remain open")
	assert.Equal(t, 1, app.snippetView.table.GetRowCount(), "saved snippet should appear in table")
}

// TestSnippetView_NewSnippet_ValidationError_KeepsFormOpen verifies that saving
// with an empty name keeps the form open with an error.
func TestSnippetView_NewSnippet_ValidationError_KeepsFormOpen(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))

	app.snippetView.Form.contentField.SetText("1d20", false)
	// Leave name empty
	testHelper.SimulateKey(app.snippetView.Form, app.Application, tcell.KeyCtrlS)

	assert.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID), "form should stay open on validation error")
	assert.True(t, app.snippetView.Form.HasFieldError("name"))
}

// TestSnippetView_NewSnippet_CancelClosesForm verifies Esc on the form modal
// closes the form but leaves the list modal open.
func TestSnippetView_NewSnippet_CancelClosesForm(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))

	testHelper.SimulateEscape(app.snippetView.Form, app.Application)

	assert.False(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID), "form modal should close on Escape")
	assert.True(t, app.isPageVisible(SNIPPET_MODAL_ID), "list modal should remain open")
}

// TestSnippetView_CtrlE_OpensEditForm verifies Ctrl+E opens the form pre-populated.
func TestSnippetView_CtrlE_OpensEditForm(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.table, app.Application, 'e')

	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))
	assert.Equal(t, "Attack", app.snippetView.Form.nameField.GetText(), "form should be pre-populated with name")
	assert.Equal(t, "1d20+5", app.snippetView.Form.contentField.GetText(), "form should be pre-populated with content")
}

// TestSnippetView_EditSnippet_UpdatesName verifies editing and saving persists changes.
func TestSnippetView_EditSnippet_UpdatesName(t *testing.T) {
	app := setupTestApp(t)
	s := createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.table, app.Application, 'e')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))

	app.snippetView.Form.nameField.SetText("Sword Attack")
	testHelper.SimulateKey(app.snippetView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))
	updated, err := app.snippetView.snippetService.GetByID(s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Sword Attack", updated.Name)
}

// TestSnippetView_DeleteSnippet_RemovesFromDB verifies that Ctrl+D + confirm
// deletes the snippet and closes both modals.
func TestSnippetView_DeleteSnippet_RemovesFromDB(t *testing.T) {
	app := setupTestApp(t)
	s := createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.table, app.Application, 'e')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))

	testHelper.SimulateKey(app.snippetView.Form.GetButton(2), app.Application, tcell.KeyEnter)
	require.True(t, app.isPageVisible(CONFIRM_MODAL_ID))

	app.confirmModal.onConfirm()

	assert.False(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID))
	_, err := app.snippetView.snippetService.GetByID(s.ID)
	assert.Error(t, err, "deleted snippet should not be found in DB")
}

// TestSnippetView_DeleteSnippet_CancelKeepsSnippet verifies that cancelling the
// confirm dialog leaves the snippet intact and returns focus to the form.
func TestSnippetView_DeleteSnippet_CancelKeepsSnippet(t *testing.T) {
	app := setupTestApp(t)
	s := createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.table, app.Application, 'e')
	testHelper.SimulateKey(app.snippetView.Form.GetButton(2), app.Application, tcell.KeyEnter)
	require.True(t, app.isPageVisible(CONFIRM_MODAL_ID))

	app.confirmModal.onCancel()

	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID))
	assert.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID), "form should remain open after cancel")
	_, err := app.snippetView.snippetService.GetByID(s.ID)
	assert.NoError(t, err, "snippet should still exist after cancel")
}

// TestSnippetView_UseSnippet_InsertsContentAndCloses verifies that pressing Enter
// on a snippet fires SnippetUseEvent, closes the modal, and inserts content.
func TestSnippetView_UseSnippet_InsertsContentAndCloses(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	testHelper.SimulateEnter(app.snippetView.table, app.Application)

	assert.False(t, app.isPageVisible(SNIPPET_MODAL_ID), "list modal should close after use")
	assert.Equal(t, "1d20+5 ", app.diceView.TextArea.GetText(), "snippet content should be inserted into dice input")
}

// TestSnippetView_UseSnippet_NoneSelected_ShowsWarning verifies that pressing Enter
// on the empty-state row does not close the modal.
func TestSnippetView_UseSnippet_NoneSelected_ShowsWarning(t *testing.T) {
	app := setupTestApp(t)
	openSnippetModal(t, app)
	// Table has the empty-state placeholder row (not selectable, no reference)

	testHelper.SimulateEnter(app.snippetView.table, app.Application)

	assert.True(t, app.isPageVisible(SNIPPET_MODAL_ID), "modal should stay open when nothing is selected")
}

// TestSnippetView_ReorderDown_MovesSnippetDown verifies Ctrl+D reorders within the same scope.
func TestSnippetView_ReorderDown_MovesSnippetDown(t *testing.T) {
	app := setupTestApp(t)
	s1 := createSnippet(t, app, "Alpha", "a", nil)
	s2 := createSnippet(t, app, "Beta", "b", nil)
	openSnippetModal(t, app)
	// s1 is at row 0, selected by default
	testHelper.SimulateRune(app.snippetView.table, app.Application, 'd')

	globals, err := app.snippetView.snippetService.GetGlobal()
	require.NoError(t, err)
	require.Len(t, globals, 2)
	assert.Equal(t, s2.ID, globals[0].ID, "Beta should be first after moving Alpha down")
	assert.Equal(t, s1.ID, globals[1].ID)
}

// TestSnippetView_ReorderUp_MovesSnippetUp verifies Ctrl+U reorders within the same scope.
func TestSnippetView_ReorderUp_MovesSnippetUp(t *testing.T) {
	app := setupTestApp(t)
	s1 := createSnippet(t, app, "Alpha", "a", nil)
	s2 := createSnippet(t, app, "Beta", "b", nil)
	openSnippetModal(t, app)
	// Move selection to row 1 (Beta)
	testHelper.SimulateDownArrow(app.snippetView.table, app.Application)
	testHelper.SimulateRune(app.snippetView.table, app.Application, 'u')

	globals, err := app.snippetView.snippetService.GetGlobal()
	require.NoError(t, err)
	require.Len(t, globals, 2)
	assert.Equal(t, s2.ID, globals[0].ID, "Beta should be first after moving up")
	assert.Equal(t, s1.ID, globals[1].ID)
}

// TestSnippetView_ReorderAtBoundary_DoesNothing verifies that reordering beyond
// the boundary has no effect.
func TestSnippetView_ReorderAtBoundary_DoesNothing(t *testing.T) {
	app := setupTestApp(t)
	s := createSnippet(t, app, "Only", "one", nil)
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.table, app.Application, 'u')

	globals, err := app.snippetView.snippetService.GetGlobal()
	require.NoError(t, err)
	assert.Equal(t, s.ID, globals[0].ID, "single snippet should remain unchanged")
}

// TestSnippetView_GameScopedSnippets_AppearFirst verifies that when a game is active
// its snippets appear before global snippets with a section divider.
func TestSnippetView_GameScopedSnippets_AppearFirst(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	require.NoError(t, app.gameView.SetCurrentGame(g.ID))
	createSnippet(t, app, "Global", "global-content", nil)
	createSnippet(t, app, "GameSnippet", "game-content", &g.ID)
	openSnippetModal(t, app)

	// Expect: row 0 = game snippet, row 1 = divider, row 2 = global snippet
	require.Equal(t, 3, app.snippetView.table.GetRowCount())
	assert.Equal(t, "GameSnippet", app.snippetView.table.GetCell(0, 0).Text)
	assert.Equal(t, "Global", app.snippetView.table.GetCell(2, 0).Text)
	// Middle row is the divider (non-selectable, no reference)
	assert.Nil(t, app.snippetView.table.GetCell(1, 0).GetReference(), "divider row should have no snippet reference")
}

// TestSnippetView_GameSnippetReorder_StaysInGameScope verifies that reordering a
// game-scoped snippet does not affect global snippets.
func TestSnippetView_GameSnippetReorder_StaysInGameScope(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	require.NoError(t, app.gameView.SetCurrentGame(g.ID))
	s1 := createSnippet(t, app, "GameA", "ga", &g.ID)
	s2 := createSnippet(t, app, "GameB", "gb", &g.ID)
	createSnippet(t, app, "Global", "global", nil)
	openSnippetModal(t, app)

	// Row 0 = GameA, Row 1 = GameB, Row 2 = divider, Row 3 = Global
	// Move GameA down
	testHelper.SimulateRune(app.snippetView.table, app.Application, 'd')

	gameSnippets, err := app.snippetView.snippetService.GetByGameID(g.ID)
	require.NoError(t, err)
	assert.Equal(t, s2.ID, gameSnippets[0].ID, "GameB should be first after moving GameA down")
	assert.Equal(t, s1.ID, gameSnippets[1].ID)

	// Global snippet should be unaffected
	globals, err := app.snippetView.snippetService.GetGlobal()
	require.NoError(t, err)
	assert.Len(t, globals, 1)
}

// TestSnippetView_NewSnippet_PreselectsActiveGame verifies that when a game is
// active, the form pre-selects it in the dropdown.
func TestSnippetView_NewSnippet_PreselectsActiveGame(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	require.NoError(t, app.gameView.SetCurrentGame(g.ID))
	openSnippetModal(t, app)

	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')

	_, label := app.snippetView.Form.gameDropdown.GetCurrentOption()
	assert.Equal(t, "Campaign", label, "active game should be pre-selected in the dropdown")
}

// TestSnippetView_Filter_ByName verifies that typing in the filter field narrows
// the table to snippets whose name matches.
func TestSnippetView_Filter_ByName(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	createSnippet(t, app, "Damage", "2d6", nil)
	openSnippetModal(t, app)
	require.Equal(t, 2, app.snippetView.table.GetRowCount())

	app.snippetView.filterField.SetText("att")

	assert.Equal(t, 1, app.snippetView.table.GetRowCount())
	assert.Equal(t, "Attack", app.snippetView.table.GetCell(0, 0).Text)
}

// TestSnippetView_Filter_ByContent verifies that the filter matches against
// snippet content as well as name.
func TestSnippetView_Filter_ByContent(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	createSnippet(t, app, "Damage", "2d6", nil)
	openSnippetModal(t, app)

	app.snippetView.filterField.SetText("2d6")

	assert.Equal(t, 1, app.snippetView.table.GetRowCount())
	assert.Equal(t, "Damage", app.snippetView.table.GetCell(0, 0).Text)
}

// TestSnippetView_Filter_CaseInsensitive verifies that filtering is case-insensitive.
func TestSnippetView_Filter_CaseInsensitive(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	app.snippetView.filterField.SetText("ATTACK")

	assert.Equal(t, 1, app.snippetView.table.GetRowCount())
}

// TestSnippetView_Filter_NoMatch_ShowsEmptyState verifies that a filter with no
// matches shows the empty-state placeholder row.
func TestSnippetView_Filter_NoMatch_ShowsEmptyState(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20+5", nil)
	openSnippetModal(t, app)

	app.snippetView.filterField.SetText("zzz")

	require.Equal(t, 1, app.snippetView.table.GetRowCount())
	assert.Nil(t, app.snippetView.table.GetCell(0, 0).GetReference(), "no-match row should not be selectable")
}

// TestSnippetView_Filter_Cleared_RestoresFullList verifies that clearing the
// filter field restores all snippets including the section divider.
func TestSnippetView_Filter_Cleared_RestoresFullList(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	require.NoError(t, app.gameView.SetCurrentGame(g.ID))
	createSnippet(t, app, "GameSnippet", "game-content", &g.ID)
	createSnippet(t, app, "Global", "global-content", nil)
	openSnippetModal(t, app)

	app.snippetView.filterField.SetText("game")
	require.Equal(t, 1, app.snippetView.table.GetRowCount())

	app.snippetView.filterField.SetText("")

	// game snippet + divider + global snippet
	assert.Equal(t, 3, app.snippetView.table.GetRowCount())
}

// TestSnippetView_Filter_HidesDivider verifies that the section divider is not
// shown when a filter is active.
func TestSnippetView_Filter_HidesDivider(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	require.NoError(t, app.gameView.SetCurrentGame(g.ID))
	createSnippet(t, app, "GameSnippet", "game-content", &g.ID)
	createSnippet(t, app, "Global", "global-content", nil)
	openSnippetModal(t, app)
	require.Equal(t, 3, app.snippetView.table.GetRowCount()) // includes divider

	app.snippetView.filterField.SetText("e") // matches both snippets

	assert.Equal(t, 2, app.snippetView.table.GetRowCount(), "divider should not appear when filter is active")
}

// TestSnippetView_Reorder_NoopWhenFiltered verifies that Ctrl+U and Ctrl+D have
// no effect on snippet positions when a filter is active.
func TestSnippetView_Reorder_NoopWhenFiltered(t *testing.T) {
	app := setupTestApp(t)
	s1 := createSnippet(t, app, "Alpha", "a", nil)
	s2 := createSnippet(t, app, "Beta", "b", nil)
	openSnippetModal(t, app)

	app.snippetView.filterField.SetText("a") // only Alpha visible
	testHelper.SimulateRune(app.snippetView.table, app.Application, 'd')

	globals, err := app.snippetView.snippetService.GetGlobal()
	require.NoError(t, err)
	assert.Equal(t, s1.ID, globals[0].ID, "Alpha should remain first — reorder is a noop when filtered")
	assert.Equal(t, s2.ID, globals[1].ID)
}

// TestSnippetView_EditSnippet_NoDeleteButtonOnNew verifies that Ctrl+N shows
// no delete button, and Ctrl+E does show one.
func TestSnippetView_EditSnippet_DeleteButtonVisibility(t *testing.T) {
	app := setupTestApp(t)
	createSnippet(t, app, "Attack", "1d20", nil)
	openSnippetModal(t, app)

	// New mode: no delete button
	testHelper.SimulateRune(app.snippetView.Modal, app.Application, 'n')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))
	assert.Equal(t, 2, app.snippetView.Form.GetButtonCount(), "new mode should have Save and Cancel only")

	testHelper.SimulateEscape(app.snippetView.Form, app.Application)

	// Edit mode: delete button present
	testHelper.SimulateRune(app.snippetView.table, app.Application, 'e')
	require.True(t, app.isPageVisible(SNIPPET_FORM_MODAL_ID))
	assert.Equal(t, 3, app.snippetView.Form.GetButtonCount(), "edit mode should have Save, Cancel, and Delete")
}
