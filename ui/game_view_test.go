package ui

import (
	"soloterm/config"
	"soloterm/domain/game"
	"soloterm/domain/session"
	"soloterm/domain/tag"
	testHelper "soloterm/shared/testing"
	"testing"

	// Blank imports to trigger init() migration registration
	_ "soloterm/domain/character"
	_ "soloterm/domain/session"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestApp creates a fully wired App backed by an in-memory database.
func setupTestApp(t *testing.T) *App {
	t.Helper()

	db := testHelper.SetupTestDB(t)
	t.Cleanup(func() { testHelper.TeardownTestDB(t, db) })

	cfg := &config.Config{
		TagTypes:        tag.DefaultTagTypes(),
		TagExcludeWords: []string{"closed", "abandoned"},
	}

	app := NewApp(db, cfg)
	return app
}

// createGame is a test helper that creates a game, refreshes the tree, and selects it.
func createGame(t *testing.T, app *App, name string) *game.Game {
	t.Helper()
	g := &game.Game{Name: name}
	g, err := app.gameView.gameService.Save(g)
	require.NoError(t, err, "Failed to create test game")
	app.gameView.Refresh()
	app.gameView.SelectGame(&g.ID)
	return g
}

func createSession(t *testing.T, app *App, gameID int64, name string) *session.Session {
	t.Helper()
	s := &session.Session{GameID: gameID, Name: name}
	s, err := app.sessionView.sessionService.Save(s)
	require.NoError(t, err, "Failed to create test service")
	return s
}

func TestGameView_AddGame(t *testing.T) {
	app := setupTestApp(t)

	// Open the new modal via Ctrl+N on the tree
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlN)
	assert.True(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be visible")

	// Fill in the form and save via Ctrl+S
	app.gameView.Form.nameField.SetText("My Test Game")
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlS)
	assert.False(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be hidden after save")

	// Verify the game was saved to the database
	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Equal(t, "My Test Game", games[0].Name)
	assert.Equal(t, app.gameView.Tree, app.GetFocus(), "Expected Tree to be in focus")
}

func TestGameView_AddGameWithDescription(t *testing.T) {
	app := setupTestApp(t)

	// Open the new modal via Ctrl+N
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlN)
	assert.True(t, app.isPageVisible(GAME_MODAL_ID))

	// Fill in both fields and save via Ctrl+S
	app.gameView.Form.nameField.SetText("RPG Campaign")
	app.gameView.Form.descriptionField.SetText("A solo adventure", false)
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlS)

	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Equal(t, "RPG Campaign", games[0].Name)
	require.NotNil(t, games[0].Description)
	assert.Equal(t, "A solo adventure", *games[0].Description)
}

func TestGameView_AddGameValidationError(t *testing.T) {
	app := setupTestApp(t)

	// Open the new modal via Ctrl+N
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlN)

	// Leave name empty and attempt save via Ctrl+S
	app.gameView.Form.nameField.SetText("")
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlS)

	// Modal should stay open on validation failure
	assert.True(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to remain visible on validation error")

	// Verify no game was saved
	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, games)

	// Verify the form has a field error
	assert.True(t, app.gameView.Form.HasFieldError("name"), "Expected field error on 'name'")
}

func TestGameView_EditGame(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Original Name")

	// Open edit modal via Ctrl+E on the tree
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.True(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be visible")

	// Change the name and save via Ctrl+S
	app.gameView.Form.nameField.SetText("Updated Name")
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlS)
	assert.False(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be hidden after save")

	// Verify the game was updated
	updated, err := app.gameView.gameService.GetByID(g.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, app.gameView.Tree, app.GetFocus(), "Expected Tree to be in focus")
}

func TestGameView_EditGameAddDescription(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Game")

	// Open edit modal via Ctrl+E
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)

	// Add a description and save via Ctrl+S
	app.gameView.Form.descriptionField.SetText("Now with a description", false)
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlS)

	updated, err := app.gameView.gameService.GetByID(g.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "Now with a description", *updated.Description)
}

func TestGameView_DeleteGame(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Game To Delete")

	// Open edit modal via Ctrl+E, then delete via Ctrl+D
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.True(t, app.isPageVisible(GAME_MODAL_ID))

	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyCtrlD)
	assert.True(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be visible")

	// Confirm the deletion
	app.gameView.ConfirmDelete(g.ID)
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be hidden")
	assert.False(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be hidden")

	// Verify the game was deleted
	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, games)
}

func TestGameView_DeleteGameLeavesOtherGames(t *testing.T) {
	app := setupTestApp(t)
	g1 := createGame(t, app, "Game One")
	createGame(t, app, "Game Two")

	// Select and delete game one
	app.gameView.SelectGame(&g1.ID)
	app.gameView.ConfirmDelete(g1.ID)

	// Verify only game two remains
	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Equal(t, "Game Two", games[0].Name)
}

func TestGameView_TreeShowsGames(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "Alpha Game")
	createGame(t, app, "Beta Game")

	root := app.gameView.Tree.GetRoot()
	children := root.GetChildren()
	require.Len(t, children, 2)
	assert.Equal(t, "Alpha Game", children[0].GetText())
	assert.Equal(t, "Beta Game", children[1].GetText())
}

func TestGameView_EmptyTreeShowsPlaceholder(t *testing.T) {
	app := setupTestApp(t)

	root := app.gameView.Tree.GetRoot()
	children := root.GetChildren()
	require.Len(t, children, 1)
	assert.Equal(t, "(No Games Yet - Press Ctrl+N to Add)", children[0].GetText())
}

func TestGameView_CancelDoesNotSave(t *testing.T) {
	app := setupTestApp(t)

	// Open new game modal via Ctrl+N and fill in data
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlN)
	app.gameView.Form.nameField.SetText("Should Not Save")

	// Cancel via Escape
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyEscape)
	assert.False(t, app.isPageVisible(GAME_MODAL_ID), "Expected game modal to be hidden after cancel")

	// Verify nothing was saved
	games, err := app.gameView.gameService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, games)
	assert.Equal(t, app.gameView.Tree, app.GetFocus(), "Expected Tree to be in focus")
}

func TestGameView_FormResetOnNew(t *testing.T) {
	app := setupTestApp(t)
	createGame(t, app, "Existing Game")

	// Open edit modal to populate the form
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlE)
	assert.Equal(t, "Existing Game", app.gameView.Form.nameField.GetText())

	// Cancel then open new game form
	testHelper.SimulateKey(app.gameView.Form, app.Application, tcell.KeyEscape)
	testHelper.SimulateKey(app.gameView.Tree, app.Application, tcell.KeyCtrlN)

	// Verify the form fields are cleared
	assert.Empty(t, app.gameView.Form.nameField.GetText(), "Expected empty name field after reset")
	assert.Empty(t, app.gameView.Form.descriptionField.GetText(), "Expected empty description field after reset")
}

func TestGameView_SelectGame(t *testing.T) {
	app := setupTestApp(t)
	game := createGame(t, app, "Existing Game")
	assert.NotNil(t, game)

	// Refresh the tree view so the game is visible
	app.gameView.Refresh()
	root := app.gameView.Tree.GetRoot()
	children := root.GetChildren()
	assert.Len(t, children, 1)

	// Select the game using the keyboard
	testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
	testHelper.SimulateEnter(app.gameView.Tree, app.Application)
	assert.Equal(t, game.ID, *app.gameView.GetCurrentSelection().GameID)

	// The child node letting them know there are no sessions yet
	assert.Equal(t, "(No sessions yet)", app.gameView.Tree.GetCurrentNode().GetChildren()[0].GetText())

	t.Run("select a session", func(t *testing.T) {
		session := createSession(t, app, game.ID, "New Session")
		app.gameView.Refresh()

		// Scroll down and select the session
		testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
		testHelper.SimulateDownArrow(app.gameView.Tree, app.Application)
		testHelper.SimulateEnter(app.gameView.Tree, app.Application)
		assert.Equal(t, session.ID, *app.sessionView.currentSessionID)

		sessionLabel := "New Session (" + session.CreatedAt.Format("2006-01-02") + ")"
		assert.Equal(t, sessionLabel, app.gameView.Tree.GetCurrentNode().GetText())

		// Refreshing the view will re-select the session
		app.gameView.Refresh()
		assert.Equal(t, session.ID, *app.sessionView.currentSessionID)
		assert.Equal(t, sessionLabel, app.gameView.Tree.GetCurrentNode().GetText())
	})

}
