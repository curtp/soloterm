package ui

import (
	"os"
	"path/filepath"
	"soloterm/domain/oracle"
	testHelper "soloterm/shared/testing"
	"testing"

	// Blank import to trigger oracle migration registration
	_ "soloterm/domain/oracle"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createOracle(t *testing.T, app *App, category, name, content string) *oracle.Oracle {
	t.Helper()
	// Compute correct sort positions so reorder tests work correctly.
	categories, err := app.oracleView.oracleService.GetCategoryInfo()
	require.NoError(t, err)

	var catPos, posInCat int
	maxCatPos := -1
	found := false
	for _, ci := range categories {
		if ci.CategoryPosition > maxCatPos {
			maxCatPos = ci.CategoryPosition
		}
		if ci.Name == category {
			found = true
			catPos = ci.CategoryPosition
			posInCat = ci.MaxPositionInCategory + 1
		}
	}
	if !found {
		catPos = maxCatPos + 1
		posInCat = 0
	}

	o := &oracle.Oracle{
		Category:           category,
		Name:               name,
		Content:            content,
		CategoryPosition:   catPos,
		PositionInCategory: posInCat,
	}
	saved, err := app.oracleView.oracleService.Save(o)
	require.NoError(t, err, "Failed to create test oracle")
	return saved
}

func TestFileView_SessionExport_WritesContentToFile(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Game")
	s := createSession(t, app, g.ID, "Session One")
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.sessionView.SetText("Session content to export.", false)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "out.md")

	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	require.True(t, app.isPageVisible(FILE_MODAL_ID), "export modal should open")

	app.fileView.Form.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "modal should close after export")
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Equal(t, "Session content to export.", string(data))
}

func TestFileView_SessionExport_RestoresFocusToTextArea(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Game")
	s := createSession(t, app, g.ID, "Session One")
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "out.md")

	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	app.fileView.Form.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyCtrlS)

	assert.Equal(t, app.sessionView.TextArea, app.GetFocus(), "focus should return to TextArea after export")
}

func TestFileView_SessionExport_CancelRestoresFocus(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "My Game")
	s := createSession(t, app, g.ID, "Session One")
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	require.True(t, app.isPageVisible(FILE_MODAL_ID))

	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyEscape)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "modal should close after cancel")
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus(), "focus should return to TextArea after cancel")
}

func TestFileView_OracleExport_WritesOracleContent(t *testing.T) {
	// Proves FileView routes to the correct target: exporting from the oracle
	// modal writes oracle content, not session content.
	app := setupTestApp(t)
	o := createOracle(t, app, "Monsters", "encounters", "Goblin\nOrc\nDragon")
	app.oracleView.Refresh()
	require.NotNil(t, app.oracleView.currentOracle, "oracle should be selected after Refresh")

	app.HandleEvent(&OracleShowExportEvent{BaseEvent: BaseEvent{action: ORACLE_SHOW_EXPORT}})
	require.True(t, app.isPageVisible(FILE_MODAL_ID), "export modal should open")

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "oracle.txt")
	app.fileView.Form.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.fileView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "modal should close after export")
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Equal(t, o.Content, string(data))
}
