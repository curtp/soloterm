package ui

import (
	"os"
	"path/filepath"
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openOracleModal fires the OracleShow event and asserts the modal is visible.
func openOracleModal(t *testing.T, app *App) {
	t.Helper()
	app.HandleEvent(&OracleShowEvent{BaseEvent: BaseEvent{action: ORACLE_SHOW}})
	require.True(t, app.isPageVisible(ORACLE_MODAL_ID), "oracle modal should be visible")
}

// TestOracleView_CtrlP_OpensModal verifies that Ctrl+P opens the oracle modal.
func TestOracleView_CtrlP_OpensModal(t *testing.T) {
	app := setupTestApp(t)
	event := tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone)
	if h := app.Application.GetInputCapture(); h != nil {
		h(event)
	}
	assert.True(t, app.isPageVisible(ORACLE_MODAL_ID))
}

// TestOracleView_EscapeClosesModal verifies that pressing Escape closes the oracle modal.
func TestOracleView_EscapeClosesModal(t *testing.T) {
	app := setupTestApp(t)
	openOracleModal(t, app)
	testHelper.SimulateEscape(app.oracleView.Modal, app.Application)
	assert.False(t, app.isPageVisible(ORACLE_MODAL_ID))
}

// TestOracleView_TreeLoadedOnce verifies that the tree is only loaded on the first
// open and not reloaded on subsequent opens.
func TestOracleView_TreeLoadedOnce(t *testing.T) {
	app := setupTestApp(t)
	assert.False(t, app.oracleView.treeLoaded, "tree should not be loaded before first open")
	openOracleModal(t, app)
	assert.True(t, app.oracleView.treeLoaded, "tree should be loaded after first open")

	// Close and reopen — treeLoaded should remain true (no second Refresh)
	app.HandleEvent(&OracleCancelEvent{BaseEvent: BaseEvent{action: ORACLE_CANCEL}})
	require.False(t, app.isPageVisible(ORACLE_MODAL_ID))
	openOracleModal(t, app)
	assert.True(t, app.oracleView.treeLoaded)
}

// TestOracleView_EmptyState_ContentAreaDisabled verifies that the content area
// is disabled when no oracles exist.
func TestOracleView_EmptyState_ContentAreaDisabled(t *testing.T) {
	app := setupTestApp(t)
	openOracleModal(t, app)
	assert.True(t, app.oracleView.ContentArea.GetDisabled(), "content area should be disabled with no tables")
}

// TestOracleView_SelectingOracleLoadsContent verifies that opening the modal
// with an existing oracle loads its content into the ContentArea.
func TestOracleView_SelectingOracleLoadsContent(t *testing.T) {
	app := setupTestApp(t)
	o := createOracle(t, app, "Monsters", "encounters", "Goblin\nOrc\nDragon")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)
	assert.Equal(t, o.ID, app.oracleView.currentOracle.ID)
	assert.Equal(t, "Goblin\nOrc\nDragon", app.oracleView.ContentArea.GetText())
	assert.False(t, app.oracleView.ContentArea.GetDisabled(), "content area should be enabled")
}

// TestOracleView_SelectingOracleDoesNotMarkDirty verifies that loading content
// on selection does not set isDirty.
func TestOracleView_SelectingOracleDoesNotMarkDirty(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin")
	openOracleModal(t, app)
	assert.False(t, app.oracleView.isDirty, "loading content must not mark dirty")
}

// TestOracleView_NewOracle_SavedAndSelectedInTree verifies that Ctrl+N + form
// save creates a new oracle, closes the form, and selects the new oracle.
func TestOracleView_NewOracle_SavedAndSelectedInTree(t *testing.T) {
	app := setupTestApp(t)
	openOracleModal(t, app)
	app.SetFocus(app.oracleView.OracleTree)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID), "form modal should open on Ctrl+N")

	app.oracleView.Form.categoryField.SetText("Monsters")
	app.oracleView.Form.nameField.SetText("encounters")
	testHelper.SimulateKey(app.oracleView.Form, app.Application, tcell.KeyCtrlS)

	require.False(t, app.isPageVisible(ORACLE_FORM_MODAL_ID), "form modal should close after save")
	require.NotNil(t, app.oracleView.currentOracle, "a table should be selected after save")
	assert.Equal(t, "encounters", app.oracleView.currentOracle.Name)
	assert.Equal(t, "Monsters", app.oracleView.currentOracle.Category)
}

// TestOracleView_NewOracle_ValidationError_EmptyName verifies that saving with
// an empty name shows a validation error and keeps the form open.
func TestOracleView_NewOracle_ValidationError_EmptyName(t *testing.T) {
	app := setupTestApp(t)
	openOracleModal(t, app)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))

	app.oracleView.Form.categoryField.SetText("Monsters")
	// Leave name empty
	testHelper.SimulateKey(app.oracleView.Form, app.Application, tcell.KeyCtrlS)

	assert.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID), "form should stay open on validation error")
	assert.True(t, app.oracleView.Form.HasFieldError("name"), "name field should have an error")
}

// TestOracleView_NewOracle_CancelClosesForm verifies that pressing Escape on
// the form closes the form modal and returns focus to the oracle tree.
func TestOracleView_NewOracle_CancelClosesForm(t *testing.T) {
	app := setupTestApp(t)
	openOracleModal(t, app)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))

	testHelper.SimulateEscape(app.oracleView.Form, app.Application)
	assert.False(t, app.isPageVisible(ORACLE_FORM_MODAL_ID), "form modal should close on Escape")
	assert.True(t, app.isPageVisible(ORACLE_MODAL_ID), "oracle modal should remain open")
}

// TestOracleView_EditOracle_UpdatesName verifies that Ctrl+E opens the edit
// form pre-populated, and saving persists the updated name.
func TestOracleView_EditOracle_UpdatesName(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlE)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))
	assert.Equal(t, "encounters", app.oracleView.Form.nameField.GetText(), "form should be pre-populated")

	app.oracleView.Form.nameField.SetText("random-encounters")
	testHelper.SimulateKey(app.oracleView.Form, app.Application, tcell.KeyCtrlS)

	require.False(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))
	require.NotNil(t, app.oracleView.currentOracle)
	assert.Equal(t, "random-encounters", app.oracleView.currentOracle.Name)
}

// TestOracleView_DeleteOracle_RemovesFromDB verifies that deleting an oracle
// removes it from the database and clears the current oracle.
func TestOracleView_DeleteOracle_RemovesFromDB(t *testing.T) {
	app := setupTestApp(t)
	o := createOracle(t, app, "Monsters", "encounters", "")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlE)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))
	testHelper.SimulateKey(app.oracleView.Form, app.Application, tcell.KeyCtrlD)

	require.True(t, app.isPageVisible(CONFIRM_MODAL_ID))
	app.confirmModal.onConfirm()

	assert.False(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))
	assert.Nil(t, app.oracleView.currentOracle)
	_, err := app.oracleView.oracleService.GetByID(o.ID)
	assert.Error(t, err, "deleted oracle should not be found in DB")
}

// TestOracleView_DeleteOracle_StaysInSameCategory verifies that after deleting
// an oracle, the tree selects another oracle from the same category rather than
// jumping to the first oracle overall.
func TestOracleView_DeleteOracle_StaysInSameCategory(t *testing.T) {
	app := setupTestApp(t)
	o1 := createOracle(t, app, "Monsters", "encounters", "")
	createOracle(t, app, "Monsters", "bosses", "")
	createOracle(t, app, "Locations", "cities", "")
	openOracleModal(t, app)

	// Select and delete the first oracle in Monsters
	app.oracleView.SelectOracle(o1.ID)
	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlE)
	require.True(t, app.isPageVisible(ORACLE_FORM_MODAL_ID))
	testHelper.SimulateKey(app.oracleView.Form, app.Application, tcell.KeyCtrlD)
	require.True(t, app.isPageVisible(CONFIRM_MODAL_ID))
	app.confirmModal.onConfirm()

	require.NotNil(t, app.oracleView.currentOracle, "a table should still be selected after delete")
	assert.Equal(t, "Monsters", app.oracleView.currentOracle.Category, "should stay in Monsters category")
}

// TestOracleView_EditingContentMarksDirty verifies that typing in the content
// area sets the dirty flag.
func TestOracleView_EditingContentMarksDirty(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	app.oracleView.ContentArea.SetText("Goblin\nOrc", false)
	assert.True(t, app.oracleView.isDirty, "editing content should set dirty")
}

// TestOracleView_AutosavePersistsContent verifies that AutosaveContent writes
// dirty content to the database and clears the dirty flag.
func TestOracleView_AutosavePersistsContent(t *testing.T) {
	app := setupTestApp(t)
	o := createOracle(t, app, "Monsters", "encounters", "Goblin")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	app.oracleView.ContentArea.SetText("Goblin\nOrc\nDragon", false)
	require.True(t, app.oracleView.isDirty)

	app.oracleView.AutosaveContent()

	saved, err := app.oracleView.oracleService.GetByID(o.ID)
	require.NoError(t, err)
	assert.Equal(t, "Goblin\nOrc\nDragon", saved.Content)
	assert.False(t, app.oracleView.isDirty, "isDirty should clear after autosave")
}

// TestOracleView_DirtyIndicatorAppearsAndClears verifies that the dirty dot
// appears in the content title when dirty and disappears after autosave.
func TestOracleView_DirtyIndicatorAppearsAndClears(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	app.oracleView.ContentArea.SetText("Goblin\nOrc", false)
	require.True(t, app.oracleView.isDirty)
	assert.Contains(t, app.oracleView.contentFrame.GetTitle(), "●")

	app.oracleView.AutosaveContent()
	assert.NotContains(t, app.oracleView.contentFrame.GetTitle(), "●")
}

// TestOracleView_ReorderDown_MovesOracleDown verifies that Ctrl+D moves the
// selected oracle down within its category.
func TestOracleView_ReorderDown_MovesOracleDown(t *testing.T) {
	app := setupTestApp(t)
	o1 := createOracle(t, app, "Monsters", "encounters", "")
	o2 := createOracle(t, app, "Monsters", "bosses", "")
	openOracleModal(t, app)

	app.oracleView.SelectOracle(o1.ID)
	app.SetFocus(app.oracleView.OracleTree)
	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlD)

	oracles, err := app.oracleView.oracleService.GetAll()
	require.NoError(t, err)
	require.Len(t, oracles, 2)
	assert.Equal(t, o2.ID, oracles[0].ID, "bosses should be first after moving encounters down")
	assert.Equal(t, o1.ID, oracles[1].ID, "encounters should be second")
}

// TestOracleView_ReorderUp_MovesOracleUp verifies that Ctrl+U moves the
// selected oracle up within its category.
func TestOracleView_ReorderUp_MovesOracleUp(t *testing.T) {
	app := setupTestApp(t)
	o1 := createOracle(t, app, "Monsters", "encounters", "")
	o2 := createOracle(t, app, "Monsters", "bosses", "")
	openOracleModal(t, app)

	app.oracleView.SelectOracle(o2.ID)
	app.SetFocus(app.oracleView.OracleTree)
	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlU)

	oracles, err := app.oracleView.oracleService.GetAll()
	require.NoError(t, err)
	require.Len(t, oracles, 2)
	assert.Equal(t, o2.ID, oracles[0].ID, "bosses should be first after moving up")
	assert.Equal(t, o1.ID, oracles[1].ID, "encounters should be second")
}

// TestOracleView_ReorderDoesNotFireFromContentArea verifies that Ctrl+U/D
// are ignored when the content area has focus (reserved for text editing).
func TestOracleView_ReorderDoesNotFireFromContentArea(t *testing.T) {
	app := setupTestApp(t)
	o1 := createOracle(t, app, "Monsters", "encounters", "")
	createOracle(t, app, "Monsters", "bosses", "")
	openOracleModal(t, app)

	app.oracleView.SelectOracle(o1.ID)
	app.SetFocus(app.oracleView.ContentArea) // focus is on content, not tree
	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlD)

	oracles, err := app.oracleView.oracleService.GetAll()
	require.NoError(t, err)
	assert.Equal(t, o1.ID, oracles[0].ID, "order should be unchanged when reorder fires from content area")
}

// TestOracleView_ImportFile_ReplacesContent verifies that importing a file
// replaces the oracle's content and persists it.
func TestOracleView_ImportFile_ReplacesContent(t *testing.T) {
	app := setupTestApp(t)
	o := createOracle(t, app, "Monsters", "encounters", "Old content")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "monsters.txt")
	err := os.WriteFile(importPath, []byte("Goblin\nOrc\nDragon"), 0644)
	require.NoError(t, err)

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlO)
	require.True(t, app.isPageVisible(FILE_MODAL_ID))

	app.fileView.Form.pathField.SetText(importPath)
	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "file modal should close after import")
	assert.Equal(t, "Goblin\nOrc\nDragon", app.oracleView.ContentArea.GetText())

	saved, err := app.oracleView.oracleService.GetByID(o.ID)
	require.NoError(t, err)
	assert.Equal(t, "Goblin\nOrc\nDragon", saved.Content)
}

// TestOracleView_ExportFile_WritesContent verifies that exporting writes the
// oracle's current content to the specified file.
func TestOracleView_ExportFile_WritesContent(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin\nOrc\nDragon")
	openOracleModal(t, app)
	require.NotNil(t, app.oracleView.currentOracle)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "encounters.txt")

	testHelper.SimulateKey(app.oracleView.Modal, app.Application, tcell.KeyCtrlX)
	require.True(t, app.isPageVisible(FILE_MODAL_ID))

	app.fileView.Form.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID))
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Equal(t, "Goblin\nOrc\nDragon", string(data))
}
