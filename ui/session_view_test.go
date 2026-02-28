package ui

import (
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionView_AddSession(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")

	// Focus the text area and open the new session modal via Ctrl+N
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be visible")

	// Fill in the form and save via Ctrl+S
	app.sessionView.Form.nameField.SetText("Session One")
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlS)
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be hidden after save")

	// Verify the session was saved to the database
	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "Session One", sessions[0].Name)
	assert.Equal(t, g.ID, sessions[0].GameID)

	// Verify the session is now selected and the text area is focused
	require.NotNil(t, app.sessionView.currentSessionID)
	assert.Equal(t, sessions[0].ID, *app.sessionView.currentSessionID)
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus(), "Expected TextArea to be in focus")

}

func TestSessionView_AddSessionValidationError(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")

	// Open the new session modal via Ctrl+N
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)

	// Leave name empty and attempt save via Ctrl+S
	app.sessionView.Form.nameField.SetText("")
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlS)

	// Modal should stay open on validation failure
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to remain visible on validation error")

	// Verify no session was saved
	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Verify the form has a field error
	assert.True(t, app.sessionView.Form.HasFieldError("name"), "Expected field error on 'name'")
}

func TestSessionView_AddSessionRequiresGame(t *testing.T) {
	app := setupTestApp(t)

	// Try to open new session modal without selecting a game
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)

	// Session modal should not be shown since no game is selected
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to not be visible without a game selected")
}

func TestSessionView_EditSession(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Original Session")

	// Navigate to and select the session via the game tree, then open edit modal
	app.gameView.Refresh()
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be visible")

	// Change the name and save via Ctrl+S
	app.sessionView.Form.nameField.SetText("Renamed Session")
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlS)
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be hidden after save")

	// Verify the session was updated
	updated, err := app.sessionView.sessionService.GetByID(s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed Session", updated.Name)
}

func TestSessionView_DeleteSession(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	addNotesToGame(t, app, g.ID, "these are some notes")
	s := createSession(t, app, g.ID, "Session To Delete")

	// Navigate to and select the session via the game tree, then open edit modal
	app.gameView.Refresh()
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID))

	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlD)
	assert.True(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be visible")

	// Confirm the deletion
	app.sessionView.ConfirmDelete(s.ID)
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be hidden")
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be hidden")

	// Verify the session was deleted
	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Verify the game still has notes
	g, err = app.gameView.gameService.GetByID(g.ID)
	assert.Nil(t, err, "Expected err to be nil")
	assert.NotEqual(t, "", g.Notes, "Expected notes to not be empty")

	// Verify the current session is cleared
	assert.Nil(t, app.sessionView.currentSessionID, "Expected current session to be nil after deletion")
}

func TestSessionView_DeleteSessionLeavesOtherSessions(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s1 := createSession(t, app, g.ID, "Session One")
	createSession(t, app, g.ID, "Session Two")

	// Select and delete session one
	app.sessionView.currentSessionID = &s1.ID
	app.sessionView.ConfirmDelete(s1.ID)

	// Verify only session two remains
	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "Session Two", sessions[0].Name)
}

func TestSessionView_CancelDoesNotSave(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")

	// Open new session modal and fill in data
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)
	app.sessionView.Form.nameField.SetText("Should Not Save")

	// Cancel via Escape
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyEscape)
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session modal to be hidden after cancel")

	// Verify nothing was saved
	sessions, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionView_SessionContentLoads(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "My Session")

	// Add content to the session directly
	s.Content = "Some session notes here"
	_, err := app.sessionView.sessionService.Save(s)
	require.NoError(t, err)

	// Select the session and verify content loads into the text area
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	assert.Equal(t, "Some session notes here", app.sessionView.TextArea.GetText())
	assert.False(t, app.sessionView.TextArea.GetDisabled(), "Expected TextArea to be enabled after loading session")
}

func TestSessionView_TextAreaDisabledWithoutSession(t *testing.T) {
	app := setupTestApp(t)

	// Without selecting a session, the text area should be disabled
	assert.True(t, app.sessionView.TextArea.GetDisabled(), "Expected TextArea to be disabled without a session")
}

func TestSessionView_FormResetOnNew(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	createSession(t, app, g.ID, "Existing Session")

	// Navigate to and select the session via the game tree, then open edit modal
	app.gameView.Refresh()
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.Equal(t, "Existing Session", app.sessionView.Form.nameField.GetText())

	// Cancel then open new session form
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyEscape)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)

	// Verify the form field is cleared
	assert.Empty(t, app.sessionView.Form.nameField.GetText(), "Expected empty name field after reset")
}

func TestSessionView_SessionAppearsInGameTree(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "New Session")
	app.gameView.Refresh()

	// The game node should have Notes first, then the session
	root := app.gameView.Tree.GetRoot()
	gameNode := root.GetChildren()[0]
	children := gameNode.GetChildren()

	require.Len(t, children, 2)
	assert.Equal(t, "Notes", children[0].GetText())
	expectedLabel := "New Session (" + s.CreatedAt.Format("2006-01-02") + ")"
	assert.Equal(t, expectedLabel, children[1].GetText())
}

func TestSessionView_ShowHelpModal(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "New Session")
	app.gameView.Refresh()

	// Select the session
	app.gameView.SelectSession(s.ID)

	// Show the help modal
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyF12)
	assert.True(t, app.isPageVisible(HELP_MODAL_ID), "Expected help modal to be visible")
}

func TestSessionView_InsertMainTemplates(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	createSession(t, app, g.ID, "New Session")
	app.gameView.Refresh()

	// Select the session
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)

	// Insert the templates using the function keys
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyF2)
	assert.Equal(t, "@ \nd: ->\n=> ", app.sessionView.TextArea.GetText())

	// Reset the text, then insert the next template
	app.sessionView.SetText("", true)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyF3)
	assert.Equal(t, "? \nd: ->\n=> ", app.sessionView.TextArea.GetText())

	// Reset the text, then insert the next template
	app.sessionView.SetText("", true)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyF4)
	assert.Equal(t, "d: ->\n=> ", app.sessionView.TextArea.GetText())

}

func TestSessionView_ShowTagModal(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	createSession(t, app, g.ID, "New Session")
	app.gameView.Refresh()

	// Select the session
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)

	// Insert the templates using the function keys
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)
	assert.True(t, app.isPageVisible(TAG_MODAL_ID), "Expected tag modal to be visible")
}

func TestSessionView_EditAndDeleteSessionFromGameTree(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	sA := createSession(t, app, g.ID, "Session A")

	// Setup: navigate to and select Session A via the game tree
	app.gameView.Refresh()
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	require.Equal(t, sA.ID, *app.sessionView.currentSessionID, "Expected Session A to be selected")

	// Create Session B via the UI
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(SESSION_MODAL_ID))
	app.sessionView.Form.nameField.SetText("Session B")
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlS)
	require.False(t, app.isPageVisible(SESSION_MODAL_ID))

	// After saving, Session B is current — capture its ID before navigating away
	sBID := *app.sessionView.currentSessionID

	// Navigate UP to Session A and select it (tree cursor is on Session B after save)
	testHelper.SimulateUpArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	require.Equal(t, sA.ID, *app.sessionView.currentSessionID, "Expected Session A to be selected again")

	// Navigate DOWN to Session B without selecting it
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	assert.Equal(t, sA.ID, *app.sessionView.currentSessionID, "Session A should still be selected (Session B not entered)")

	// Ctrl+E on the game tree edits the highlighted Session B (not the selected one)
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.True(t, app.isPageVisible(SESSION_MODAL_ID), "Expected session edit modal to open for Session B")
	assert.Equal(t, "Session B", app.sessionView.Form.nameField.GetText(), "Expected form to show Session B's name")

	// Delete Session B via Ctrl+D, then confirm
	testHelper.SimulateKey(app.sessionView.Form, app.Application, tcell.KeyCtrlD)
	assert.True(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal")
	app.sessionView.ConfirmDelete(sBID)
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID))
	assert.False(t, app.isPageVisible(SESSION_MODAL_ID))

	// Result: only Session A remains
	remaining, err := app.sessionView.sessionService.GetAllForGame(g.ID)
	require.NoError(t, err)
	require.Len(t, remaining, 1, "Expected only Session A to remain")
	assert.Equal(t, "Session A", remaining[0].Name)
}
