package ui

import (
	"soloterm/domain/tag"
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openTagModal is a test helper that creates a game with a session,
// selects the session, and opens the tag modal via Ctrl+T.
func openTagModal(t *testing.T, app *App) {
	t.Helper()
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Test Session")
	app.gameView.Refresh()
	app.gameView.SelectSession(s.ID)
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)
}

func TestTagView_OpenAndClose(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)
	assert.True(t, app.isPageVisible(TAG_MODAL_ID), "Expected tag modal to be visible")

	// Close via Escape
	testHelper.SimulateKey(app.tagView.TagTable, app.Application, tcell.KeyEsc)
	assert.False(t, app.isPageVisible(TAG_MODAL_ID), "Expected tag modal to be hidden after Escape")
}

func TestTagView_ShowsConfiguredTags(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)

	// The tag table should have a header row plus one row per configured tag type
	expectedCount := len(tag.DefaultTagTypes())
	// Row 0 is the header, rows 1..N are the tag types
	actualRows := app.tagView.TagTable.GetRowCount() - 1 // subtract header
	assert.Equal(t, expectedCount, actualRows, "Expected one row per configured tag type")

	// Verify the first configured tag appears (sorted alphabetically)
	firstCell := app.tagView.TagTable.GetCell(1, 0)
	require.NotNil(t, firstCell)
	assert.Equal(t, "Clock", firstCell.Text, "Expected first tag to be 'Clock' (alphabetically sorted)")
}

func TestTagView_SelectTagInsertsTemplate(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)

	// Select the first tag row (row 1, since row 0 is header)
	app.tagView.TagTable.Select(1, 0)
	testHelper.SimulateKey(app.tagView.TagTable, app.Application, tcell.KeyEnter)

	// Tag modal should close
	assert.False(t, app.isPageVisible(TAG_MODAL_ID), "Expected tag modal to be hidden after selection")

	// The template should be inserted into the session text area
	text := app.sessionView.TextArea.GetText()
	assert.NotEmpty(t, text, "Expected tag template to be inserted into text area")

	// Focus should return to the text area
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus(), "Expected TextArea to be in focus after tag selection")
}

func TestTagView_SelectSpecificTag(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)

	// Find the "NPC" tag row by scanning the table
	npcRow := -1
	for row := 1; row < app.tagView.TagTable.GetRowCount(); row++ {
		cell := app.tagView.TagTable.GetCell(row, 0)
		if cell != nil && cell.Text == "NPC" {
			npcRow = row
			break
		}
	}
	require.NotEqual(t, -1, npcRow, "Expected to find NPC tag in the table")

	// Select the NPC tag
	app.tagView.TagTable.Select(npcRow, 0)
	testHelper.SimulateKey(app.tagView.TagTable, app.Application, tcell.KeyEnter)

	// Verify the NPC template was inserted
	text := app.sessionView.TextArea.GetText()
	assert.Contains(t, text, "[N:", "Expected NPC template to be inserted")
}

func TestTagView_ShowsRecentTagsFromSession(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Test Session")

	// Add content with tags to the session
	s.Content = "[L:Tavern | A cozy place]\n[N:Bartender | Friendly]\n[N:Bartender | Friendly; Knowledgeable]"
	app.sessionView.sessionService.Save(s)

	// Select the session and open the tag modal
	app.gameView.Refresh()
	app.gameView.SelectSession(s.ID)
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)

	// Count rows: header + configured tags + separator + recent tags
	configTagCount := len(tag.DefaultTagTypes())
	totalRows := app.tagView.TagTable.GetRowCount()

	// Should have more rows than just header + config tags (separator + recent tags)
	assert.Greater(t, totalRows, configTagCount+1, "Expected recent tags to appear in the table")

	// Look for our recent tags in the table
	foundTavern := false
	foundBartender := false
	for row := 1; row < totalRows; row++ {
		cell := app.tagView.TagTable.GetCell(row, 0)
		if cell != nil {
			if cell.Text == "L:Tavern" {
				foundTavern = true
			}
			// Bartender should be the most recent tag
			if cell.Text == "N:Bartender" {
				if app.tagView.TagTable.GetCell(row, 1).Text == "[N:Bartender | Friendly; Knowledgeable]" {
					foundBartender = true
				}
			}
		}
	}
	assert.True(t, foundTavern, "Expected 'L:Tavern' in recent tags")
	assert.True(t, foundBartender, "Expected 'N:Bartender' in recent tags")
}

func TestTagView_ExcludesClosedTags(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Test Session")

	// Add content with a closed tag (exclude word from default config: "closed")
	s.Content = "[L:Dungeon | Explored; Closed]\n[N:Guard | Alert]"
	app.sessionView.sessionService.Save(s)

	// Select and open tag modal
	app.gameView.Refresh()
	app.gameView.SelectSession(s.ID)
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)

	// Look for tags - Dungeon should be excluded, Guard should be present
	totalRows := app.tagView.TagTable.GetRowCount()
	foundDungeon := false
	foundGuard := false
	for row := 1; row < totalRows; row++ {
		cell := app.tagView.TagTable.GetCell(row, 0)
		if cell != nil {
			if cell.Text == "L:Dungeon" {
				foundDungeon = true
			}
			if cell.Text == "N:Guard" {
				foundGuard = true
			}
		}
	}
	assert.False(t, foundDungeon, "Expected 'L:Dungeon' to be excluded (closed)")
	assert.True(t, foundGuard, "Expected 'N:Guard' in recent tags")
}

func TestTagView_ShowHelpModal(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)

	// Open help from the tag modal via F12
	testHelper.SimulateKey(app.tagView.TagTable, app.Application, tcell.KeyF12)
	assert.True(t, app.isPageVisible(HELP_MODAL_ID), "Expected help modal to be visible")

	testHelper.SimulateEscape(app.helpModal, app.Application)
	assert.False(t, app.isPageVisible(HELP_MODAL_ID), "Expected help modal to be hidden")
}

func TestTagView_NavigateWithArrowKeys(t *testing.T) {
	app := setupTestApp(t)
	openTagModal(t, app)

	// Start at the first selectable row
	app.tagView.TagTable.Select(1, 0)
	initialRow, _ := app.tagView.TagTable.GetSelection()

	// Move down
	testHelper.SimulateDownArrow(app.tagView.TagTable, app.Application)
	newRow, _ := app.tagView.TagTable.GetSelection()
	assert.Equal(t, initialRow+1, newRow, "Expected selection to move down")

	// Move back up
	testHelper.SimulateUpArrow(app.tagView.TagTable, app.Application)
	newRow, _ = app.tagView.TagTable.GetSelection()
	assert.Equal(t, initialRow, newRow, "Expected selection to move back up")
}
