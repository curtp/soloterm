package ui

import (
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

// createAttr opens the new-attribute modal with Ctrl+N, fills in name and value,
// selects the group via the dropdown (0 = "- New -", 1+ = existing group by index),
// saves with Ctrl+S, and returns the created attribute from the DB.
func createAttr(t *testing.T, app *App, charID int64, name string, groupDropdownIdx int) *character.Attribute {
	t.Helper()

	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ATTRIBUTE_MODAL_ID), "Expected attribute modal to be visible")

	app.attributeView.Form.nameField.SetText(name)
	app.attributeView.Form.valueField.SetText("val")
	app.attributeView.Form.groupDropDown.SetCurrentOption(groupDropdownIdx)

	testHelper.SimulateKey(app.attributeView.Form, app.Application, tcell.KeyCtrlS)
	require.False(t, app.isPageVisible(ATTRIBUTE_MODAL_ID), "Expected attribute modal to close after save")

	attrs, err := app.attributeView.attrService.GetForCharacter(charID)
	require.NoError(t, err)
	for _, a := range attrs {
		if a.Name == name {
			return a
		}
	}
	t.Fatalf("createAttr: could not find saved attribute %q", name)
	return nil
}

// reloadAttrs returns the current sorted attribute list from the DB.
func reloadAttrs(t *testing.T, app *App, charID int64) []*character.Attribute {
	t.Helper()
	attrs, err := app.attributeView.attrService.GetForCharacter(charID)
	require.NoError(t, err)
	return attrs
}

// --- Ctrl+D on a child: move down within group ---

func TestAttributeView_CtrlDown_ChildSwapsWithNextItem(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Alpha", 0) // header, pos 0
	b := createAttr(t, app, char.ID, "Beta", 1)  // child, pos 1
	c := createAttr(t, app, char.ID, "Gamma", 1) // child, pos 2

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 3)
	assert.Equal(t, c.ID, attrs[1].ID, "Gamma should be second after Ctrl+D on Beta")
	assert.Equal(t, b.ID, attrs[2].ID, "Beta should be third after Ctrl+D")
}

func TestAttributeView_CtrlDown_ChildAtGroupBoundary_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Alpha", 0)
	b := createAttr(t, app, char.ID, "Beta", 1)  // child, last in group 0
	createAttr(t, app, char.ID, "Gamma", 0)       // standalone in group 1

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 3)
	assert.Equal(t, b.ID, attrs[1].ID, "Beta should remain at position 1 — children cannot cross group boundaries")
}

// --- Ctrl+U on a child: move up within group ---

func TestAttributeView_CtrlUp_ChildSwapsWithSiblingAbove(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Alpha", 0) // header, pos 0 — stays put
	b := createAttr(t, app, char.ID, "Beta", 1)  // child, pos 1
	c := createAttr(t, app, char.ID, "Gamma", 1) // child, pos 2

	app.attributeView.Select(c.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 3)
	assert.Equal(t, c.ID, attrs[1].ID, "Gamma should be at pos 1 after Ctrl+U")
	assert.Equal(t, b.ID, attrs[2].ID, "Beta should be at pos 2 after Ctrl+U on Gamma")
}

func TestAttributeView_CtrlUp_ChildAtHeaderBoundary_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, "Alpha", 0) // header
	b := createAttr(t, app, char.ID, "Beta", 1)  // child directly below header

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, a.ID, attrs[0].ID, "Alpha should remain the header")
	assert.Equal(t, b.ID, attrs[1].ID, "Beta should remain the child")
}

// --- Ctrl+D/U on a header or standalone: move entire group ---

func TestAttributeView_CtrlDown_StandaloneMovesGroup(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, "Alpha", 0) // standalone group 0
	b := createAttr(t, app, char.ID, "Beta", 0)  // standalone group 1

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Beta should be first after Ctrl+D on standalone Alpha")
	assert.Equal(t, a.ID, attrs[1].ID, "Alpha should be second")
}

func TestAttributeView_CtrlDown_HeaderMovesEntireGroup(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	header := createAttr(t, app, char.ID, "Stats", 0)  // group 0, pos 0
	child := createAttr(t, app, char.ID, "HP", 1)      // group 0, pos 1
	other := createAttr(t, app, char.ID, "Skills", 0)  // group 1

	app.attributeView.Select(header.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 3)
	assert.Equal(t, other.ID, attrs[0].ID, "Skills should be first")
	assert.Equal(t, header.ID, attrs[1].ID, "Stats header should be second")
	assert.Equal(t, child.ID, attrs[2].ID, "HP child should be third")
}

func TestAttributeView_CtrlDown_HeaderAtBottom_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Stats", 0)
	b := createAttr(t, app, char.ID, "Skills", 0)

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlD)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[1].ID, "Skills should remain last after no-op")
}

func TestAttributeView_CtrlUp_StandaloneMovesGroup(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, "Stats", 0)
	b := createAttr(t, app, char.ID, "Skills", 0)

	app.attributeView.Select(b.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, b.ID, attrs[0].ID, "Skills should be first after Ctrl+U")
	assert.Equal(t, a.ID, attrs[1].ID, "Stats should be second")
}

func TestAttributeView_CtrlUp_HeaderAtTop_IsNoop(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, "Stats", 0)
	createAttr(t, app, char.ID, "Skills", 0)

	app.attributeView.Select(a.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlU)

	attrs := reloadAttrs(t, app, char.ID)
	require.Len(t, attrs, 2)
	assert.Equal(t, a.ID, attrs[0].ID, "Stats should remain first after no-op")
}

// --- Focus is restored to the table after reorder ---

func TestAttributeView_ReorderRestoresFocus(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	a := createAttr(t, app, char.ID, "Alpha", 0) // standalone group 0
	createAttr(t, app, char.ID, "Beta", 0)        // standalone group 1

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

// --- Group dropdown: new group assignment ---

func TestAttributeView_NewAttr_NewGroup_AssignsCorrectGroupAndPosition(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	// First "- New -" → group 0, position 0
	alpha := createAttr(t, app, char.ID, "Alpha", 0)
	assert.Equal(t, 0, alpha.Group, "First new group should be group 0")
	assert.Equal(t, 0, alpha.PositionInGroup, "First item in new group should be position 0")

	// Second "- New -" → group 1, position 0
	beta := createAttr(t, app, char.ID, "Beta", 0)
	assert.Equal(t, 1, beta.Group, "Second new group should be group 1")
	assert.Equal(t, 0, beta.PositionInGroup, "First item in second new group should be position 0")
}

// --- Group dropdown: append to existing group ---

func TestAttributeView_NewAttr_ExistingGroup_AppendsToEnd(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	alpha := createAttr(t, app, char.ID, "Alpha", 0) // group 0, pos 0
	beta := createAttr(t, app, char.ID, "Beta", 1)   // append to group 0 → pos 1
	gamma := createAttr(t, app, char.ID, "Gamma", 1) // append again → pos 2

	assert.Equal(t, alpha.Group, beta.Group, "Beta should be in the same group as Alpha")
	assert.Equal(t, 1, beta.PositionInGroup, "Beta should be at position 1")
	assert.Equal(t, alpha.Group, gamma.Group, "Gamma should be in the same group as Alpha")
	assert.Equal(t, 2, gamma.PositionInGroup, "Gamma should be at position 2")
}

// --- Group dropdown: correct headers are shown ---

func TestAttributeView_NewAttr_DropdownShowsGroupHeaders(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Stats", 0)  // group 0
	createAttr(t, app, char.ID, "Skills", 0) // group 1

	// Open the new entry modal — handler fetches attrs and rebuilds groupHeaders.
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlN)
	require.True(t, app.isPageVisible(ATTRIBUTE_MODAL_ID))

	headers := app.attributeView.Form.groupHeaders
	require.Len(t, headers, 2, "Dropdown should list 2 existing groups")
	assert.Equal(t, "Stats", headers[0].Name, "First group header should be Stats")
	assert.Equal(t, "Skills", headers[1].Name, "Second group header should be Skills")

	testHelper.SimulateKey(app.attributeView.Form, app.Application, tcell.KeyEscape)
}

// --- Edit mode: group and position are preserved ---

func TestAttributeView_EditAttr_PreservesGroupAndPosition(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Alpha", 0)        // group 0, pos 0
	beta := createAttr(t, app, char.ID, "Beta", 1) // group 0, pos 1

	app.attributeView.Select(beta.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlE)
	require.True(t, app.isPageVisible(ATTRIBUTE_MODAL_ID))

	// Only change the name; leave the group dropdown on its pre-selected value.
	app.attributeView.Form.nameField.SetText("Beta Updated")
	testHelper.SimulateKey(app.attributeView.Form, app.Application, tcell.KeyCtrlS)
	require.False(t, app.isPageVisible(ATTRIBUTE_MODAL_ID))

	attrs := reloadAttrs(t, app, char.ID)
	var updated *character.Attribute
	for _, a := range attrs {
		if a.ID == beta.ID {
			updated = a
			break
		}
	}
	require.NotNil(t, updated)
	assert.Equal(t, "Beta Updated", updated.Name)
	assert.Equal(t, beta.Group, updated.Group, "Group should be preserved after edit")
	assert.Equal(t, beta.PositionInGroup, updated.PositionInGroup, "Position should be preserved after edit")
}

func TestAttributeView_EditAttr_ChangingGroupAppendsToNewGroup(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Hero")

	createAttr(t, app, char.ID, "Alpha", 0) // group 0, pos 0
	beta := createAttr(t, app, char.ID, "Beta", 0)  // group 1, pos 0
	gamma := createAttr(t, app, char.ID, "Gamma", 1) // group 0, pos 1 (appended to Alpha's group)

	// Edit Gamma and move it to Beta's group (dropdown index 2).
	app.attributeView.Select(gamma.ID)
	testHelper.SimulateKey(app.attributeView.Table, app.Application, tcell.KeyCtrlE)
	require.True(t, app.isPageVisible(ATTRIBUTE_MODAL_ID))

	app.attributeView.Form.groupDropDown.SetCurrentOption(2) // dropdown: 0="-New-", 1="Alpha", 2="Beta"
	testHelper.SimulateKey(app.attributeView.Form, app.Application, tcell.KeyCtrlS)
	require.False(t, app.isPageVisible(ATTRIBUTE_MODAL_ID))

	attrs := reloadAttrs(t, app, char.ID)
	var gammaAfter *character.Attribute
	for _, a := range attrs {
		if a.ID == gamma.ID {
			gammaAfter = a
			break
		}
	}
	require.NotNil(t, gammaAfter)
	assert.Equal(t, beta.Group, gammaAfter.Group, "Gamma should now be in Beta's group")
	assert.Equal(t, 1, gammaAfter.PositionInGroup, "Gamma should be appended after Beta (pos 1)")
}
