package ui

import (
	testHelper "soloterm/shared/testing"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openDiceModal fires the DiceShow event and asserts the modal is visible.
func openDiceModal(t *testing.T, app *App) {
	t.Helper()
	app.HandleEvent(&DiceShowEvent{BaseEvent: BaseEvent{action: DICE_SHOW}})
	require.True(t, app.isPageVisible(DICE_MODAL_ID), "dice modal should be visible")
}

// TestDiceView_CtrlR_OpensModal verifies that Ctrl+R opens the dice modal.
func TestDiceView_CtrlR_OpensModal(t *testing.T) {
	app := setupTestApp(t)
	event := tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModNone)
	if h := app.Application.GetInputCapture(); h != nil {
		h(event)
	}
	assert.True(t, app.isPageVisible(DICE_MODAL_ID))
}

// TestDiceView_EscapeClosesModal verifies that pressing Escape closes the dice modal.
func TestDiceView_EscapeClosesModal(t *testing.T) {
	app := setupTestApp(t)
	openDiceModal(t, app)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyEsc)
	assert.False(t, app.isPageVisible(DICE_MODAL_ID))
}

// TestDiceView_EscapeRestoresFocus verifies that closing the modal restores focus
// to the primitive that had focus before it was opened.
func TestDiceView_EscapeRestoresFocus(t *testing.T) {
	app := setupTestApp(t)
	app.SetFocus(app.sessionView.TextArea)
	openDiceModal(t, app)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyEsc)
	assert.Equal(t, app.sessionView.TextArea, app.GetFocus())
}

// TestDiceView_CtrlR_Roll verifies that Ctrl+R inside the modal executes a roll
// and populates the results view.
func TestDiceView_CtrlR_Roll(t *testing.T) {
	app := setupTestApp(t)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("1d6", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)

	result := app.diceView.resultView.GetText(true)
	assert.NotEmpty(t, result, "results view should be populated after rolling")
}

// TestDiceView_CtrlR_RollMultipleGroups verifies that multiple lines produce
// output for each group.
func TestDiceView_CtrlR_RollMultipleGroups(t *testing.T) {
	app := setupTestApp(t)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("Attack: 1d20\nDamage: 2d6", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)

	result := app.diceView.resultView.GetText(true)
	assert.Contains(t, result, "Attack")
	assert.Contains(t, result, "Damage")
}

// TestDiceView_CtrlO_CanInsert_ReturnsFalseWhenNotFromSession verifies that
// CanInsert is false when the dice modal was opened from outside the session view.
func TestDiceView_CanInsert_FalseWhenNotFromSession(t *testing.T) {
	app := setupTestApp(t)
	app.SetFocus(app.gameView.Tree)
	openDiceModal(t, app)
	assert.False(t, app.diceView.CanInsert())
}

// TestDiceView_CanInsert_TrueWhenFromSession verifies that CanInsert is true when
// the dice modal was opened from the session text area.
func TestDiceView_CanInsert_TrueWhenFromSession(t *testing.T) {
	app := setupTestApp(t)
	app.SetFocus(app.sessionView.TextArea)
	openDiceModal(t, app)
	assert.True(t, app.diceView.CanInsert())
}

// TestDiceView_CtrlO_InsertsResultIntoSession verifies that Ctrl+O inserts the
// roll result into the session text area and closes the modal.
func TestDiceView_CtrlO_InsertsResultIntoSession(t *testing.T) {
	app := setupTestApp(t)
	app.SetFocus(app.sessionView.TextArea)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("1d6", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)
	require.NotEmpty(t, app.diceView.resultView.GetText(true))

	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlO)

	assert.False(t, app.isPageVisible(DICE_MODAL_ID), "modal should close after insert")
	assert.NotEmpty(t, app.sessionView.TextArea.GetText(), "result should be inserted into session")
}

// TestDiceView_CtrlO_DoesNotInsertWhenNotFromSession verifies that Ctrl+O is a
// no-op when the modal was not opened from the session text area.
func TestDiceView_CtrlO_DoesNotInsertWhenNotFromSession(t *testing.T) {
	app := setupTestApp(t)
	app.SetFocus(app.gameView.Tree)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("1d6", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlO)

	// Modal stays open; nothing was inserted
	assert.True(t, app.isPageVisible(DICE_MODAL_ID), "modal should stay open when insert is unavailable")
	assert.Empty(t, app.sessionView.TextArea.GetText())
}

// TestDiceView_Refresh_ClearsState verifies that Refresh wipes the text area,
// results, and hint state.
func TestDiceView_Refresh_ClearsState(t *testing.T) {
	app := setupTestApp(t)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("2d6", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)
	require.NotEmpty(t, app.diceView.resultView.GetText(true))

	app.diceView.Refresh()

	assert.Empty(t, app.diceView.TextArea.GetText())
	assert.Empty(t, app.diceView.resultView.GetText(true))
	assert.False(t, app.diceView.hintsActive)
}

// TestDiceView_TableHints_ShownWhenAtTyped verifies that typing "@" activates the
// table hint panel.
func TestDiceView_TableHints_ShownWhenAtTyped(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin\nOrc")
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("@enc", true)
	// updateTableHints is called via SetChangedFunc; invoke directly for the test
	app.diceView.updateTableHints()

	assert.True(t, app.diceView.hintsActive, "hints should be active when typing @")
	assert.Contains(t, app.diceView.tableHintView.GetText(true), "encounters")
}

// TestDiceView_TableHints_HiddenWhenNoAt verifies that the hint panel is hidden
// when no active "@" prefix exists.
func TestDiceView_TableHints_HiddenWhenNoAt(t *testing.T) {
	app := setupTestApp(t)
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("2d6", true)
	app.diceView.updateTableHints()

	assert.False(t, app.diceView.hintsActive)
}

// TestDiceView_Tab_CompletesFirstHint verifies that pressing Tab when hints are
// active completes the @prefix with the first matching table name.
func TestDiceView_Tab_CompletesFirstHint(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin")
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("@enc", true)
	app.diceView.updateTableHints()
	require.True(t, app.diceView.hintsActive)

	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyTab)

	text := app.diceView.TextArea.GetText()
	assert.Contains(t, text, "Monsters/encounters")
}

// TestDiceView_OracleRoll_PicksEntry verifies that rolling @tablename picks an
// entry from the oracle and displays it in the results view.
func TestDiceView_OracleRoll_PicksEntry(t *testing.T) {
	app := setupTestApp(t)
	createOracle(t, app, "Monsters", "encounters", "Goblin\nOrc\nDragon")
	openDiceModal(t, app)

	app.diceView.TextArea.SetText("@Monsters/encounters", true)
	testHelper.SimulateKey(app.diceView.Modal, app.Application, tcell.KeyCtrlR)

	result := app.diceView.resultView.GetText(true)
	assert.NotEmpty(t, result)
	// Should display "encounters -> <entry>" format (@ prefix stripped)
	assert.Contains(t, result, "encounters")
	assert.Contains(t, result, "->")
}
