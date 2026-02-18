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

func TestSessionView_ImportFile(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Import Target")

	// Select the session
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	// Create a temp file to import
	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "import.md")
	err := os.WriteFile(importPath, []byte("# Imported Content\n\nHello from a file!"), 0644)
	require.NoError(t, err)

	// Open import modal via Ctrl+O
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to be visible")

	// Enter the file path and submit
	app.sessionView.FileForm.pathField.SetText(importPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should close after successful import
	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to be hidden after import")

	// Verify the content was imported into the text area
	assert.Equal(t, "# Imported Content\n\nHello from a file!", app.sessionView.TextArea.GetText())
}

func TestSessionView_ImportFileNotFound(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Import Target")

	// Select the session
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	// Open import modal via Ctrl+O
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)

	// Enter a non-existent path and submit
	app.sessionView.FileForm.pathField.SetText("/nonexistent/path/file.md")
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should stay open on error
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to remain visible on error")

	// Error message should be displayed
	assert.Contains(t, app.sessionView.FileForm.errorMessage.GetText(true), "Cannot read file")
}

func TestSessionView_ImportBinaryFileRejected(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Import Target")

	// Select the session
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	// Create a binary file
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "binary.dat")
	err := os.WriteFile(binPath, []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00}, 0644)
	require.NoError(t, err)

	// Open import modal and try to import the binary file
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	app.sessionView.FileForm.pathField.SetText(binPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should stay open
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to remain visible for binary file")
	assert.Contains(t, app.sessionView.FileForm.errorMessage.GetText(true), "binary")
}

func TestSessionView_ImportEmptyPath(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Import Target")

	// Select the session
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	// Open import modal and submit with empty path
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should stay open with error
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to remain visible")
	assert.Contains(t, app.sessionView.FileForm.errorMessage.GetText(true), "File path is required")
}

func TestSessionView_ExportFile(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Export Source")

	// Select the session and add content
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.sessionView.isLoading = true
	app.sessionView.TextArea.SetText("# Exported Content\n\nSession notes here.", false)
	app.sessionView.isLoading = false

	// Open export modal via Ctrl+X
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.md")

	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to be visible")

	// Enter the file path and submit
	app.sessionView.FileForm.pathField.SetText(exportPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should close after successful export
	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to be hidden after export")

	// Verify the file was written with correct content
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Equal(t, "# Exported Content\n\nSession notes here.", string(data))
}

func TestSessionView_ExportToInvalidPath(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Export Source")

	// Select the session
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()

	// Open export modal and try to export to invalid path
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	app.sessionView.FileForm.pathField.SetText("/nonexistent/dir/export.md")
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Modal should stay open on error
	assert.True(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to remain visible on error")
	assert.Contains(t, app.sessionView.FileForm.errorMessage.GetText(true), "Cannot write file")
}

func TestSessionView_ImportCancelDoesNothing(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Cancel Target")

	// Select the session and add content
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.sessionView.isLoading = true
	app.sessionView.TextArea.SetText("Original content", false)
	app.sessionView.isLoading = false

	// Open import modal and cancel
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	assert.True(t, app.isPageVisible(FILE_MODAL_ID))

	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyEscape)
	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to be hidden after cancel")

	// Content should be unchanged
	assert.Equal(t, "Original content", app.sessionView.TextArea.GetText())
}

func TestSessionView_ImportRequiresSession(t *testing.T) {
	app := setupTestApp(t)

	// Without a session selected, Ctrl+O should not open the modal
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to not be visible without a session")
}

func TestSessionView_ExportRequiresSession(t *testing.T) {
	app := setupTestApp(t)

	// Without a session selected, Ctrl+X should not open the modal
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlX)
	assert.False(t, app.isPageVisible(FILE_MODAL_ID), "Expected file modal to not be visible without a session")
}

func TestSessionView_ImportReplacesExistingContent(t *testing.T) {
	app := setupTestApp(t)
	g := createGame(t, app, "Test Game")
	s := createSession(t, app, g.ID, "Replace Target")

	// Select the session and set initial content
	app.sessionView.currentSessionID = &s.ID
	app.sessionView.Refresh()
	app.sessionView.isLoading = true
	app.sessionView.TextArea.SetText("Old content that should be replaced", false)
	app.sessionView.isLoading = false

	// Create a temp file with new content
	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "new.md")
	err := os.WriteFile(importPath, []byte("Brand new content"), 0644)
	require.NoError(t, err)

	// Import the file
	app.SetFocus(app.sessionView.TextArea)
	testHelper.SimulateKey(app.sessionView.TextArea, app.Application, tcell.KeyCtrlO)
	app.sessionView.FileForm.pathField.SetText(importPath)
	testHelper.SimulateKey(app.sessionView.FileForm, app.Application, tcell.KeyCtrlS)

	// Verify content was replaced
	assert.Equal(t, "Brand new content", app.sessionView.TextArea.GetText())
}
