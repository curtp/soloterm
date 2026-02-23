package ui

import (
	"fmt"
	"soloterm/domain/character"
	testHelper "soloterm/shared/testing"
	"testing"

	// Blank imports to trigger init() migration registration
	_ "soloterm/domain/character"
	_ "soloterm/domain/session"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createCharacter saves a character to the DB, then navigates the character
// tree with simulated keys to select it, and focuses the attribute table.
// Tree structure: Root → System node ("FlexD6", collapsed) → Character node.
func createCharacter(t *testing.T, app *App, name string) *character.Character {
	t.Helper()
	char, err := character.NewCharacter(name, "FlexD6", "Fighter", "Human")
	require.NoError(t, err)
	char, err = app.characterView.charService.Save(char)
	require.NoError(t, err)

	app.characterView.Refresh()
	app.SetFocus(app.characterView.CharTree)

	// Down → system node; Enter → expand it;
	// Down → character node; Enter → select the character.
	testHelper.SimulateDownArrow(app.characterView.CharTree, app.Application)
	testHelper.SimulateEnter(app.characterView.CharTree, app.Application)
	testHelper.SimulateDownArrow(app.characterView.CharTree, app.Application)
	testHelper.SimulateEnter(app.characterView.CharTree, app.Application)

	// Jump to the attribute table (app-level Ctrl+S)
	testHelper.SimulateCtrlS(app.Application)

	return char
}

// createAttr opens the new-attribute modal with Ctrl+N, fills in the fields,
// saves with Ctrl+S on the form, and returns the created attribute from the DB.
func createAttr(t *testing.T, app *App, charID int64, group, pos int, name string) *character.Attribute {
	t.Helper()

	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ATTRIBUTE_MODAL_ID), "Expected attribute modal to be visible")

	app.attributeView.Form.nameField.SetText(name)
	app.attributeView.Form.valueField.SetText("val")
	app.attributeView.Form.groupField.SetText(fmt.Sprintf("%d", group))
	app.attributeView.Form.positionField.SetText(fmt.Sprintf("%d", pos))

	testHelper.SimulateKey(app.attributeView.Form, app.Application, tcell.KeyCtrlS)
	require.False(t, app.isPageVisible(ATTRIBUTE_MODAL_ID), "Expected attribute modal to close after save")

	// Locate the saved attribute in the DB by its name, group, and position.
	attrs, err := app.attributeView.attrService.GetForCharacter(charID)
	require.NoError(t, err)
	for _, a := range attrs {
		if a.Name == name && a.Group == group && a.PositionInGroup == pos {
			return a
		}
	}
	t.Fatalf("createAttr: could not find saved attribute %q (group=%d pos=%d)", name, group, pos)
	return nil
}

// reloadAttrs returns the current sorted attribute list from the DB.
func reloadAttrs(t *testing.T, app *App, charID int64) []*character.Attribute {
	t.Helper()
	attrs, err := app.attributeView.attrService.GetForCharacter(charID)
	require.NoError(t, err)
	return attrs
}

// --- Ctrl+D: individual move down ---

func TestAttributeView_CtrlDown_SwapsAdjacentItems(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Alpha")
	b := createAttr(t, app, char.ID, 0, 1, "Beta")

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Beta should be first after Ctrl+D on Alpha")
	assert.Equal(t, a.ID, attrs[1].ID, "Alpha should be second after Ctrl+D")
}

func TestAttributeView_CtrlDown_AtBottom_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, 0, 0, "Alpha")
	b := createAttr(t, app, char.ID, 0, 1, "Beta")

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[1].ID, "Beta should remain last after no-op")
}

// --- Ctrl+U: individual move up ---

func TestAttributeView_CtrlUp_SwapsAdjacentItems(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Alpha")
	b := createAttr(t, app, char.ID, 0, 1, "Beta")

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Beta should be first after Ctrl+U")
	assert.Equal(t, a.ID, attrs[1].ID, "Alpha should be second after Ctrl+U on Beta")
}

func TestAttributeView_CtrlUp_AtTop_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Alpha")
	createAttr(t, app, char.ID, 0, 1, "Beta")

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, a.ID, attrs[0].ID, "Alpha should remain first after no-op")
}

// --- Ctrl+D crossing a group boundary ---

func TestAttributeView_CtrlDown_CrossGroupBoundary(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Alpha")
	b := createAttr(t, app, char.ID, 1, 0, "Beta")

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Beta should be first (group 0) after boundary cross")
	assert.Equal(t, a.ID, attrs[1].ID, "Alpha should be second (group 1) after boundary cross")
}

// --- D (Shift+D): group move down ---

func TestAttributeView_GroupDown_SwapsGroups(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Stats")
	b := createAttr(t, app, char.ID, 1, 0, "Skills")

	app.attributeView.Select(a.ID)
	testHelper.SimulateRune(app.attributeView.Table, app.Application, 'D')

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Skills should be first after D on Stats group")
	assert.Equal(t, a.ID, attrs[1].ID, "Stats should be second after D")
}

func TestAttributeView_GroupDown_MovesEntireGroup(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	header := createAttr(t, app, char.ID, 0, 0, "Stats")
	child := createAttr(t, app, char.ID, 0, 1, "HP")
	other := createAttr(t, app, char.ID, 1, 0, "Skills")

	// Select the child (not the header) to verify group move works from any item.
	app.attributeView.Select(child.ID)
	testHelper.SimulateRune(app.attributeView.Table, app.Application, 'D')

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 3)
	assert.Equal(t, other.ID, attrs[0].ID, "Skills should be first")
	assert.Equal(t, header.ID, attrs[1].ID, "Stats header should be second")
	assert.Equal(t, child.ID, attrs[2].ID, "HP child should be third")
}

func TestAttributeView_GroupDown_AtBottom_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, 0, 0, "Stats")
	b := createAttr(t, app, char.ID, 1, 0, "Skills")

	app.attributeView.Select(b.ID)
	testHelper.SimulateRune(app.attributeView.Table, app.Application, 'D')

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[1].ID, "Skills should remain last after no-op")
}

// --- U (Shift+U): group move up ---

func TestAttributeView_GroupUp_SwapsGroups(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Stats")
	b := createAttr(t, app, char.ID, 1, 0, "Skills")

	app.attributeView.Select(b.ID)
	testHelper.SimulateRune(app.attributeView.Table, app.Application, 'U')

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Skills should be first after U")
	assert.Equal(t, a.ID, attrs[1].ID, "Stats should be second after U on Skills")
}

func TestAttributeView_GroupUp_AtTop_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Stats")
	createAttr(t, app, char.ID, 1, 0, "Skills")

	app.attributeView.Select(a.ID)
	testHelper.SimulateRune(app.attributeView.Table, app.Application, 'U')

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, a.ID, attrs[0].ID, "Stats should remain first after no-op")
}

// --- Focus is restored to the table after reorder ---

func TestAttributeView_ReorderRestoresFocus(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, 0, 0, "Alpha")
	createAttr(t, app, char.ID, 0, 1, "Beta")

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	assert.Equal(t, app.attributeView.Table, app.GetFocus(), "Table should retain focus after reorder")
}

// --- No character selected: event is swallowed without panic ---

func TestAttributeView_NoCharacterSelected_ReorderIsIgnored(t *testing.T) {
	app := setupTestApp(t)
	// No character selected; the input capture shows a warning and swallows the event.
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyDown)
}
