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

// openCharacterModal focuses the character tree and opens the New Character modal via n.
func openCharacterModal(t *testing.T, app *App) {
	t.Helper()
	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'n')
}

func TestCharacterView_OpenAndClose(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)
	assert.True(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be visible")

	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyEscape)
	assert.False(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be hidden after Escape")
}

func TestCharacterView_AddCharacter(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)

	app.characterView.Form.nameField.SetText("Aria")
	app.characterView.Form.systemField.SetText("D&D 5e")
	app.characterView.Form.roleField.SetText("Wizard")
	app.characterView.Form.speciesField.SetText("Elf")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyCtrlS)

	assert.False(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be hidden after save")

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	require.Len(t, chars, 1)
	assert.Equal(t, "Aria", chars[0].Name)
	assert.Equal(t, "D&D 5e", chars[0].System)
	assert.Equal(t, "Wizard", chars[0].Role)
	assert.Equal(t, "Elf", chars[0].Species)
}

func TestCharacterView_AddCharacterValidationError(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)

	// Leave name empty and attempt save
	app.characterView.Form.nameField.SetText("")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyCtrlS)

	assert.True(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to remain visible on validation error")

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, chars, "Expected no character to be saved")
	assert.True(t, app.characterView.Form.HasFieldError("name"), "Expected field error on 'name'")
}

func TestCharacterView_CancelDoesNotSave(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)

	app.characterView.Form.nameField.SetText("Ghost")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyEscape)

	assert.False(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be hidden after cancel")

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, chars, "Expected no character to be saved after cancel")
}

func TestCharacterView_FocusReturnedToTreeAfterCancel(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)

	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyEscape)

	assert.Equal(t, app.characterView.CharTree, app.GetFocus(), "Expected focus to return to character tree after cancel")
}

func TestCharacterView_FocusReturnedToTreeAfterSave(t *testing.T) {
	app := setupTestApp(t)
	openCharacterModal(t, app)

	app.characterView.Form.nameField.SetText("Hero")
	app.characterView.Form.systemField.SetText("FlexD6")
	app.characterView.Form.roleField.SetText("Fighter")
	app.characterView.Form.speciesField.SetText("Human")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyCtrlS)

	assert.Equal(t, app.characterView.CharTree, app.GetFocus(), "Expected focus to return to character tree after save")
}

func TestCharacterView_EditCharacter(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Old Name")

	// Move focus back to the tree and open edit modal
	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'e')
	assert.True(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be visible for edit")

	// The form should be pre-populated
	assert.Equal(t, "Old Name", app.characterView.Form.nameField.GetText())

	// Update the name and save
	app.characterView.Form.nameField.SetText("New Name")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyCtrlS)
	assert.False(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be hidden after save")

	updated, err := app.characterView.charService.GetByID(char.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestCharacterView_EditCharacterValidationError(t *testing.T) {
	app := setupTestApp(t)
	createCharacter(t, app, "Valid Name")

	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'e')

	// Clear the name and try to save
	app.characterView.Form.nameField.SetText("")
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyCtrlS)

	assert.True(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to remain visible on validation error")
	assert.True(t, app.characterView.Form.HasFieldError("name"), "Expected field error on 'name'")
}

func TestCharacterView_FormResetOnNew(t *testing.T) {
	app := setupTestApp(t)
	createCharacter(t, app, "Existing")

	// Open edit modal to populate the form
	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'e')
	assert.Equal(t, "Existing", app.characterView.Form.nameField.GetText(), "Expected edit form to be pre-populated")

	// Cancel, then open new character form
	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyEscape)
	openCharacterModal(t, app)

	// Fields should be cleared for a new entry
	assert.Empty(t, app.characterView.Form.nameField.GetText(), "Expected empty name field after reset")
	assert.Empty(t, app.characterView.Form.systemField.GetText(), "Expected empty system field after reset")
	assert.Empty(t, app.characterView.Form.roleField.GetText(), "Expected empty role field after reset")
	assert.Empty(t, app.characterView.Form.speciesField.GetText(), "Expected empty species field after reset")

	testHelper.SimulateKey(app.characterView.Form, app.Application, tcell.KeyEscape)
}

func TestCharacterView_DeleteCharacter(t *testing.T) {
	app := setupTestApp(t)
	char := createCharacter(t, app, "Doomed")

	// Open edit modal, then trigger delete
	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'e')
	assert.True(t, app.isPageVisible(CHARACTER_MODAL_ID))

	testHelper.SimulateKey(app.characterView.Form.GetButton(2), app.Application, tcell.KeyEnter)
	assert.True(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be visible")

	// Confirm deletion
	app.characterView.ConfirmDelete(char.ID)
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be hidden")
	assert.False(t, app.isPageVisible(CHARACTER_MODAL_ID), "Expected character modal to be hidden")

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	assert.Empty(t, chars, "Expected character to be deleted")
}

func TestCharacterView_DeleteLeavesOtherCharacters(t *testing.T) {
	app := setupTestApp(t)
	char1 := createCharacter(t, app, "First")
	char2, err := character.NewCharacter("Second", "FlexD6", "Rogue", "Halfling")
	require.NoError(t, err)
	char2, err = app.characterView.charService.Save(char2)
	require.NoError(t, err)

	app.characterView.ConfirmDelete(char1.ID)

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	require.Len(t, chars, 1, "Expected only the second character to remain")
	assert.Equal(t, char2.ID, chars[0].ID)
}

func TestCharacterView_DuplicateCharacter(t *testing.T) {
	app := setupTestApp(t)
	createCharacter(t, app, "Original")

	charID := *app.characterView.GetSelectedCharacterID()

	// d on the tree triggers a duplicate confirmation
	app.SetFocus(app.characterView.CharTree)
	testHelper.SimulateRune(app.characterView.CharTree, app.Application, 'd')
	assert.True(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal for duplicate")

	app.characterView.ConfirmDuplicate(charID)
	assert.False(t, app.isPageVisible(CONFIRM_MODAL_ID), "Expected confirmation modal to be hidden after duplicate")

	chars, err := app.characterView.charService.GetAll()
	require.NoError(t, err)
	assert.Len(t, chars, 2, "Expected original and duplicate character")
}

func TestCharacterView_EmptyTreeShowsPlaceholder(t *testing.T) {
	app := setupTestApp(t)

	root := app.characterView.CharTree.GetRoot()
	require.NotNil(t, root)
	children := root.GetChildren()
	require.Len(t, children, 1)
	assert.Equal(t, "(No Characters Yet - Press n to Add)", children[0].GetText())
}

func TestCharacterView_TreeGroupsCharactersBySystem(t *testing.T) {
	app := setupTestApp(t)

	char1, err := character.NewCharacter("Aria", "D&D 5e", "Wizard", "Elf")
	require.NoError(t, err)
	_, err = app.characterView.charService.Save(char1)
	require.NoError(t, err)

	char2, err := character.NewCharacter("Ragnar", "Pathfinder", "Barbarian", "Dwarf")
	require.NoError(t, err)
	_, err = app.characterView.charService.Save(char2)
	require.NoError(t, err)

	app.characterView.RefreshTree()

	// Root should have two system nodes, sorted alphabetically
	root := app.characterView.CharTree.GetRoot()
	children := root.GetChildren()
	require.Len(t, children, 2, "Expected two system nodes")
	assert.Equal(t, "D&D 5e", children[0].GetText())
	assert.Equal(t, "Pathfinder", children[1].GetText())

	// Each system node should contain its character
	require.Len(t, children[0].GetChildren(), 1)
	assert.Equal(t, "Aria", children[0].GetChildren()[0].GetText())

	require.Len(t, children[1].GetChildren(), 1)
	assert.Equal(t, "Ragnar", children[1].GetChildren()[0].GetText())
}

func TestCharacterView_TreeGroupsMultipleCharactersUnderSameSystem(t *testing.T) {
	app := setupTestApp(t)

	char1, err := character.NewCharacter("Aria", "FlexD6", "Wizard", "Elf")
	require.NoError(t, err)
	_, err = app.characterView.charService.Save(char1)
	require.NoError(t, err)

	char2, err := character.NewCharacter("Brom", "FlexD6", "Fighter", "Human")
	require.NoError(t, err)
	_, err = app.characterView.charService.Save(char2)
	require.NoError(t, err)

	app.characterView.RefreshTree()

	root := app.characterView.CharTree.GetRoot()
	children := root.GetChildren()
	require.Len(t, children, 1, "Expected one system node for FlexD6")
	assert.Equal(t, "FlexD6", children[0].GetText())
	assert.Len(t, children[0].GetChildren(), 2, "Expected two characters under FlexD6")
}

func TestCharacterView_CharacterInfoDisplaysAfterSelection(t *testing.T) {
	app := setupTestApp(t)
	createCharacter(t, app, "Aria")

	// createCharacter selects the character via the tree, which triggers RefreshDisplay
	infoText := app.characterView.InfoView.GetText(true)
	assert.Contains(t, infoText, "Aria", "Expected character name in info view")
	assert.Contains(t, infoText, "FlexD6", "Expected system in info view")
	assert.Contains(t, infoText, "Fighter", "Expected role in info view")
	assert.Contains(t, infoText, "Human", "Expected species in info view")
}

func TestCharacterView_SelectingCharacterFocusesAttributeTable(t *testing.T) {
	app := setupTestApp(t)
	createCharacter(t, app, "Hero")

	// createCharacter ends with SimulateCtrlS which switches focus to the attribute table
	assert.Equal(t, app.attributeView.Table, app.GetFocus(), "Expected attribute table to have focus after character selection")
}
