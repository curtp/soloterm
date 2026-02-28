package ui

import (
	"soloterm/domain/game"
	"soloterm/domain/tag"
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

// openTagModalFromNotes is a test helper that creates a game with notes content,
// selects the Notes node, and opens the tag modal via Ctrl+T.
func openTagModalFromNotes(t *testing.T, app *App, notesContent string) *game.Game {
	t.Helper()
	g := createGame(t, app, "Test Game")
	err := app.sessionView.gameService.SaveNotes(g.ID, notesContent)
	require.NoError(t, err)
	app.gameView.Refresh()
	selectNotes(t, app)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)
	return g
}

// findTagInTable scans the table and returns the row index where column 0 matches label, or -1.
func findTagInTable(table *tview.Table, label string) int {
	for row := 1; row < table.GetRowCount(); row++ {
		if cell := table.GetCell(row, 0); cell != nil && cell.Text == label {
			return row
		}
	}
	return -1
}

func TestTagView_ShowsNotesTags(t *testing.T) {
	app := setupTestApp(t)
	openTagModalFromNotes(t, app, "[N:Malichi | Hostile mage]\n[L:Sunken Tower | Flooded]")

	assert.NotEqual(t, -1, findTagInTable(app.tagView.TagTable, "N:Malichi"), "Expected 'N:Malichi' in notes tags")
	assert.NotEqual(t, -1, findTagInTable(app.tagView.TagTable, "L:Sunken Tower"), "Expected 'L:Sunken Tower' in notes tags")

	// The Notes Tags section header must be present
	found := false
	for row := 1; row < app.tagView.TagTable.GetRowCount(); row++ {
		if cell := app.tagView.TagTable.GetCell(row, 0); cell != nil && cell.Text == "─── Notes Tags ───" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected Notes Tags section header")
}

func TestTagView_HidesNotesSectionWhenEmpty(t *testing.T) {
	app := setupTestApp(t)
	openTagModalFromNotes(t, app, "") // empty notes

	for row := 1; row < app.tagView.TagTable.GetRowCount(); row++ {
		if cell := app.tagView.TagTable.GetCell(row, 0); cell != nil && cell.Text == "─── Notes Tags ───" {
			t.Fatal("Expected no Notes Tags section header when notes are empty")
		}
	}
}

func TestTagView_NoteTagsIndependentOfSessionClose(t *testing.T) {
	// Closing a tag in a session removes it from Active Tags but must NOT remove it from Notes Tags.
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")

	// Notes contain Malichi as an open tag.
	err := app.sessionView.gameService.SaveNotes(g.ID, "[N:Malichi | Hostile mage]")
	require.NoError(t, err)

	// A session closes Malichi.
	s := createSession(t, app, g.ID, "Session One")
	s.Content = "[N:Malichi | Hostile mage; Closed]"
	_, err = app.sessionView.sessionService.Save(s)
	require.NoError(t, err)

	app.gameView.Refresh()
	selectNotes(t, app)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)

	// Malichi must NOT appear in Active Tags (closed in session).
	// Malichi MUST appear in Notes Tags (still open in notes).
	activeSection := false
	notesSection := false
	activeHasMalichi := false
	notesHasMalichi := false

	for row := 1; row < app.tagView.TagTable.GetRowCount(); row++ {
		cell := app.tagView.TagTable.GetCell(row, 0)
		if cell == nil {
			continue
		}
		switch cell.Text {
		case "─── Active Tags ───":
			activeSection = true
			notesSection = false
		case "─── Notes Tags ───":
			notesSection = true
			activeSection = false
		case "N:Malichi":
			if activeSection {
				activeHasMalichi = true
			}
			if notesSection {
				notesHasMalichi = true
			}
		}
	}

	assert.False(t, activeHasMalichi, "Malichi must not appear in Active Tags (closed in session)")
	assert.True(t, notesHasMalichi, "Malichi must appear in Notes Tags (open in notes)")
}

func TestTagView_ShowsBothActiveAndNotesTags(t *testing.T) {
	// Both Active Tags (from sessions) and Notes Tags must appear in the modal simultaneously.
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")

	err := app.sessionView.gameService.SaveNotes(g.ID, "[N:Malichi | Hostile mage]")
	require.NoError(t, err)

	s := createSession(t, app, g.ID, "Session One")
	s.Content = "[L:Tavern | Cozy]"
	_, err = app.sessionView.sessionService.Save(s)
	require.NoError(t, err)

	app.gameView.Refresh()
	selectNotes(t, app)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlT)

	assert.NotEqual(t, -1, findTagInTable(app.tagView.TagTable, "L:Tavern"), "Expected session tag 'L:Tavern' in Active Tags")
	assert.NotEqual(t, -1, findTagInTable(app.tagView.TagTable, "N:Malichi"), "Expected notes tag 'N:Malichi' in Notes Tags")

	activeFound, notesFound := false, false
	for row := 1; row < app.tagView.TagTable.GetRowCount(); row++ {
		if cell := app.tagView.TagTable.GetCell(row, 0); cell != nil {
			switch cell.Text {
			case "─── Active Tags ───":
				activeFound = true
			case "─── Notes Tags ───":
				notesFound = true
			}
		}
	}
	assert.True(t, activeFound, "Expected Active Tags section header")
	assert.True(t, notesFound, "Expected Notes Tags section header")
}

func TestTagView_NotesTagsExcludesClosed(t *testing.T) {
	// A tag closed in the notes document must not appear in Notes Tags.
	app := setupTestApp(t)
	openTagModalFromNotes(t, app, "[N:Malichi | Closed]\n[L:Dungeon | Active]")

	assert.Equal(t, -1, findTagInTable(app.tagView.TagTable, "N:Malichi"), "Malichi is closed in notes — must not appear in Notes Tags")
	assert.NotEqual(t, -1, findTagInTable(app.tagView.TagTable, "L:Dungeon"), "Dungeon is open in notes — must appear in Notes Tags")
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
