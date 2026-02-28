package ui

import (
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runSearch is a test helper that sets the search term and fires the Done
// callback (same as pressing Enter in the input field).
func runSearch(app *App, term string) {
	app.searchView.searchTermInput.SetText(term)
	app.searchView.performSearch(tcell.KeyEnter)
}

// openSearchFromNotes navigates to the Notes node for the first game and
// opens the search modal via F5.
func openSearchFromNotes(t *testing.T, app *App) {
	t.Helper()
	app.gameView.Refresh()
	selectNotes(t, app)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyF5)
}

func TestSearch_FindsMatchInNotes(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	err := app.sessionView.gameService.SaveNotes(g.ID, "Malichi is a hostile mage")
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "hostile")

	require.Len(t, app.searchView.matches, 1)
	assert.True(t, app.searchView.matches[0].isNotes)
	assert.Equal(t, "Notes", app.searchView.matches[0].sessionName)
}

func TestSearch_FindsMatchesInBothSessionsAndNotes(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")

	err := app.sessionView.gameService.SaveNotes(g.ID, "dragon sighted in the north")
	require.NoError(t, err)

	s := createSession(t, app, g.ID, "Session One")
	s.Content = "the dragon attacked the village"
	_, err = app.sessionView.sessionService.Save(s)
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "dragon")

	require.Len(t, app.searchView.matches, 2)

	sessionMatches := 0
	notesMatches := 0
	for _, m := range app.searchView.matches {
		if m.isNotes {
			notesMatches++
		} else {
			sessionMatches++
		}
	}
	assert.Equal(t, 1, sessionMatches, "expected one session match")
	assert.Equal(t, 1, notesMatches, "expected one notes match")
}

func TestSearch_NoNotesMatchWhenTermAbsent(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	err := app.sessionView.gameService.SaveNotes(g.ID, "Malichi is a hostile mage")
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "dragon")

	assert.Empty(t, app.searchView.matches)
}

func TestSearch_SelectNotesResult_LoadsNotesPane(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	err := app.sessionView.gameService.SaveNotes(g.ID, "Malichi is a hostile mage")
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "hostile")

	require.Len(t, app.searchView.matches, 1)
	require.True(t, app.searchView.matches[0].isNotes)

	app.HandleEvent(&SearchSelectResultEvent{
		BaseEvent: BaseEvent{action: SEARCH_SELECT_RESULT},
	})

	assert.True(t, app.sessionView.IsNotesMode(), "Expected notes mode after selecting a notes result")
	assert.Equal(t, g.ID, app.sessionView.currentGame.ID)
	assert.False(t, app.isPageVisible(SEARCH_MODAL_ID), "Expected search modal to be hidden")
}

func TestSearch_SelectNotesResult_SelectsNotesNodeInTree(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	err := app.sessionView.gameService.SaveNotes(g.ID, "Malichi is a hostile mage")
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "hostile")
	require.Len(t, app.searchView.matches, 1)

	app.HandleEvent(&SearchSelectResultEvent{
		BaseEvent: BaseEvent{action: SEARCH_SELECT_RESULT},
	})

	// The tree's current node should be the Notes node for this game.
	state := app.gameView.GetCurrentSelection()
	require.NotNil(t, state)
	assert.True(t, state.IsNotes)
	assert.Equal(t, g.ID, *state.GameID)
}

func TestSearch_NotesOnlyResult_NoSessionsNeeded(t *testing.T) {
	// Search works and returns notes results even when the game has no sessions.
	app := setupTestApp(t)
	g := createGame(t, app, "Campaign")
	err := app.sessionView.gameService.SaveNotes(g.ID, "tower of the archmage")
	require.NoError(t, err)

	openSearchFromNotes(t, app)
	runSearch(app, "archmage")

	require.Len(t, app.searchView.matches, 1)
	assert.True(t, app.searchView.matches[0].isNotes)
}
