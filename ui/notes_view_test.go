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

// selectNotes navigates to the Notes node for the current (top) game and presses Enter.
// Requires the game node to already be current and expanded in the tree
// (call app.gameView.Refresh() after createGame before using this helper).
func selectNotes(t *testing.T, app *App) {
	t.Helper()
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application) // game → Notes
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)     // fires GameNotesSelectedEvent
}

// TestNotes_SelectingLoadsContent verifies that selecting the Notes node
// loads pre-existing notes content into the text area.
func TestNotes_SelectingLoadsContent(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")

	err := app.sessionView.gameService.SaveNotes(g.ID, "[N:Malichi | Hostile mage]")
	require.NoError(t, err)

	app.gameView.Refresh()
	selectNotes(t, app)

	assert.True(t, app.sessionView.IsNotesMode())
	assert.Equal(t, "[N:Malichi | Hostile mage]", app.sessionView.TextArea.GetText())
	assert.False(t, app.sessionView.TextArea.GetDisabled())
	assert.Nil(t, app.sessionView.currentSessionID)
	assert.NotNil(t, app.sessionView.currentGame)
	assert.Equal(t, g.ID, app.sessionView.currentGame.ID)
}

// TestNotes_EmptyNotesEnablesTextArea verifies that selecting Notes on a game
// with no notes still enables the text area (empty is valid).
func TestNotes_EmptyNotesEnablesTextArea(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)

	assert.True(t, app.sessionView.IsNotesMode())
	assert.Equal(t, "", app.sessionView.TextArea.GetText())
	assert.False(t, app.sessionView.TextArea.GetDisabled())
}

// TestNotes_TitleShowsGameName verifies the pane title includes the game name
// and the "Notes" suffix when notes are loaded.
func TestNotes_TitleShowsGameName(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)

	title := app.sessionView.textAreaFrame.GetTitle()
	assert.Contains(t, title, "My Campaign")
	assert.Contains(t, title, "Notes")
}

// TestNotes_SelectingDoesNotMarkDirty verifies that loading notes does not
// mark the view as dirty (the ChangedFunc isLoading guard must hold).
func TestNotes_SelectingDoesNotMarkDirty(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)

	assert.True(t, app.sessionView.IsNotesMode())
	assert.False(t, app.sessionView.isDirty, "loading notes must not mark dirty")
}

// TestNotes_AutosavePersistsContent verifies that calling Autosave() while in
// notes mode writes the text area content to the database.
func TestNotes_AutosavePersistsContent(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	// Simulate the user typing (SetText triggers ChangedFunc → isDirty = true).
	app.sessionView.TextArea.SetText("[N:Malichi | Hostile mage]", false)
	require.True(t, app.sessionView.isDirty)

	app.sessionView.Autosave()

	saved, err := app.sessionView.gameService.GetByID(g.ID)
	require.NoError(t, err)
	assert.Equal(t, "[N:Malichi | Hostile mage]", saved.Notes)
	assert.False(t, app.sessionView.isDirty, "isDirty must clear after autosave")
}

// TestNotes_TitleShowsDirtyIndicator verifies that the dirty dot appears in the
// title when notes have unsaved changes, and disappears after autosave.
func TestNotes_TitleShowsDirtyIndicator(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	// Type something — isDirty becomes true, title gets the dirty indicator.
	app.sessionView.TextArea.SetText("draft", false)
	require.True(t, app.sessionView.isDirty)
	assert.Contains(t, app.sessionView.textAreaFrame.GetTitle(), "●")

	app.sessionView.Autosave()
	assert.NotContains(t, app.sessionView.textAreaFrame.GetTitle(), "●")
}

// TestNotes_SwitchToSessionAutosavesNotes verifies that switching from notes
// mode to a session automatically saves any dirty notes content.
func TestNotes_SwitchToSessionAutosavesNotes(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")
	s := createSession(t, app, g.ID, "Session One")

	// Select Notes and type some content.
	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())
	app.sessionView.TextArea.SetText("NPC notes here", false)
	require.True(t, app.sessionView.isDirty)

	// Switch to the session — handleSessionSelected calls Autosave first.
	app.HandleEvent(&SessionSelectedEvent{
		BaseEvent: BaseEvent{action: SESSION_SELECTED},
		SessionID: s.ID,
	})

	// Notes were saved.
	saved, err := app.sessionView.gameService.GetByID(g.ID)
	require.NoError(t, err)
	assert.Equal(t, "NPC notes here", saved.Notes)

	// Session is now loaded.
	assert.False(t, app.sessionView.IsNotesMode())
	require.NotNil(t, app.sessionView.currentSessionID)
	assert.Equal(t, s.ID, *app.sessionView.currentSessionID)
	assert.Nil(t, app.sessionView.currentGame)
}

// TestNotes_SwitchToNotesAutosavesSession verifies that switching from an open
// session to notes automatically saves any dirty session content.
func TestNotes_SwitchToNotesAutosavesSession(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")
	s := createSession(t, app, g.ID, "Session One")

	// Load the session.
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	require.False(t, app.sessionView.IsNotesMode())

	// Type content into the session — isDirty becomes true.
	app.sessionView.TextArea.SetText("Session content here", false)
	require.True(t, app.sessionView.isDirty)

	// Switch to notes — handleGameNotesSelected calls Autosave first.
	app.HandleEvent(&GameNotesSelectedEvent{
		BaseEvent: BaseEvent{action: GAME_NOTES_SELECTED},
		GameID:    g.ID,
	})

	// Session content was saved.
	saved, err := app.sessionView.sessionService.GetByID(s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Session content here", saved.Content)

	// Notes mode is now active.
	assert.True(t, app.sessionView.IsNotesMode())
	assert.Nil(t, app.sessionView.currentSessionID)
}

// TestNotes_CtrlNOpensSessionModal verifies that pressing Ctrl+N while in notes
// mode opens the new session form scoped to the correct game.
func TestNotes_CtrlNOpensSessionModal(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID), "Ctrl+N must open the session modal")

	// Saving the form creates a session for the correct game.
	app.sessionView.Form.nameField.SetText("New Session")
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlS)

	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, g.ID, sessions[0].GameID)
}

// TestNotes_CtrlTOpensTagModal verifies that pressing Ctrl+T while in notes
// mode opens the tag selection modal.
func TestNotes_CtrlTOpensTagModal(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)
	assert.True(t, app.isPageVisible(TAG_MODAL_ID), "Ctrl+T must open the tag modal in notes mode")
}

// TestNotes_CtrlO_OpensImportModal verifies that pressing Ctrl+O while in
// notes mode opens the import file modal.
func TestNotes_CtrlO_OpensImportModal(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Ctrl+O must open the import modal in notes mode")
}

// TestNotes_CtrlX_OpensExportModal verifies that pressing Ctrl+X while in
// notes mode opens the export file modal.
func TestNotes_CtrlX_OpensExportModal(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Ctrl+X must open the export modal in notes mode")
}

// TestNotes_ImportFile verifies that importing a file while in notes mode
// replaces the notes content and persists it to the database.
func TestNotes_ImportFile(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "notes.md")
	err := os.WriteFile(importPath, []byte("Imported notes content"), 0644)
	require.NoError(t, err)

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	require.True(t, app.isPageVisible(FILE_MODAL_ID))

	app.sessionView.FileForm.pathField.SetText(importPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to close after import")
	assert.Equal(t, "Imported notes content", app.sessionView.TextArea.GetText())

	saved, err := app.sessionView.gameService.GetByID(g.ID)
	require.NoError(t, err)
	assert.Equal(t, "Imported notes content", saved.Notes)
}

// TestNotes_ExportFile verifies that exporting while in notes mode writes the
// current notes content to the specified file.
func TestNotes_ExportFile(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Campaign")

	err := app.sessionView.gameService.SaveNotes(g.ID, "Notes to export")
	require.NoError(t, err)

	app.gameView.Refresh()
	selectNotes(t, app)
	require.True(t, app.sessionView.IsNotesMode())

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "notes_export.md")

	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	require.True(t, app.isPageVisible(FILE_MODAL_ID))

	app.sessionView.FileForm.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to close after export")

	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Equal(t, "Notes to export", string(data))
}
