package ui

import (
	"testing"

	// Blank imports to trigger init() migration registration
	_ "soloterm/domain/character"
	_ "soloterm/domain/session"
	testHelper "soloterm/shared/testing"

	"github.com/stretchr/testify/assert"
)

func TestGameView_Navigate(t *testing.T) {
	app := setupTestApp(t)

	// Make sure it starts out on the Game Tree
	assert.Equal(t, app.gameView.Tree, app.GetFocus())
	// Tab goes to the Session Log
	testHelper.SimulateTab(app.Application)
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus())
	// Tab goes to the Character Tee
	testHelper.SimulateTab(app.Application)
	assert.Equal(t, app.characterView.CharTree, app.GetFocus())
	// Tab goes to the Character Sheet
	testHelper.SimulateTab(app.Application)
	assert.Equal(t, app.attributeView.Table, app.GetFocus())
	// Wraps around to the Game Tree
	testHelper.SimulateTab(app.Application)
	assert.Equal(t, app.gameView.Tree, app.GetFocus())

	// Go back to attribute table
	testHelper.SimulateBacktab(app.Application)

	// Backtab flows backward through the elements
	testHelper.SimulateBacktab(app.Application)
	assert.Equal(t, app.characterView.CharTree, app.GetFocus())
	testHelper.SimulateBacktab(app.Application)
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus())
	testHelper.SimulateBacktab(app.Application)
	assert.Equal(t, app.gameView.Tree, app.GetFocus())
	// Wraps around to the Attribute table
	testHelper.SimulateBacktab(app.Application)
	assert.Equal(t, app.attributeView.Table, app.GetFocus())

	// Jump to the Game Tree
	testHelper.SimulateCtrlG(app.Application)
	assert.Equal(t, app.gameView.Tree, app.GetFocus())

	// Jump to the session view
	testHelper.SimulateCtrlL(app.Application)
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus())

	// Jump to the character tree
	testHelper.SimulateCtrlC(app.Application)
	assert.Equal(t, app.characterView.CharTree, app.GetFocus())

	// Jump to the character sheet
	testHelper.SimulateCtrlS(app.Application)
	assert.Equal(t, app.attributeView.Table, app.GetFocus())
}

func TestGameView_ShowAbout(t *testing.T) {
	app := setupTestApp(t)

	testHelper.SimulateF1(app.Application)
	assert.True(t, app.isPageVisible(ABOUT_MODAL_ID), "Expected about modal to be visible")

	testHelper.SimulateEscape(app.aboutModal, app.Application)
	assert.False(t, app.isPageVisible(ABOUT_MODAL_ID), "Expected about modal to be hidden")
}
