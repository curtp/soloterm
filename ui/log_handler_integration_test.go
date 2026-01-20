package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"soloterm/domain/game"
	"soloterm/domain/log"
	testhelper "soloterm/shared/testing"
)

// TestLogView_ShowModal verfies the log modal is displayed
func TestLogView_ShowModal(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Create the app
	app := NewApp(db)
	game, _ := game.NewGame("a game")
	app.gameService.Save(game)
	if game.ID != 1 {
		t.Error("game not saved")
	}

	t.Run("show log modal requires game", func(t *testing.T) {
		// Refresh to load games into tree
		app.gameView.Refresh()

		// Navigate down in game tree to select the game
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		testhelper.SimulateTab(app.Application)
		if app.GetFocus() != app.logTextView {
			t.Error("Expected to be on the LogTextView, but wasn't")
		}

		// Press Ctrl+N on log view to open new log modal
		testhelper.SimulateCtrlN(app.logTextView, app.Application)

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// It only has Save and Cancel buttons
		if app.logForm.GetButtonCount() != 2 {
			t.Error("Should only display 2 buttons")
		}

		// Press Escape on the form to close the modal
		testhelper.SimulateEscape(app.logForm, app.Application)
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}
		inFocus := app.GetFocus()
		if inFocus != app.logTextView {
			t.Error("log view doesn't have focus")
		}
	})

	t.Run("error message when no game selected", func(t *testing.T) {

		// Clear the selection
		app.gameView.SelectGame(nil)

		if app.gameView.GetCurrentSelection() != nil {
			t.Errorf("There is a game state: %+v", app.GetSelectedGameState())
		}

		// Open the modal
		testhelper.SimulateCtrlN(app.logTextView, app.Application)

		note := app.notification.GetText(false)
		if len(note) == 0 {
			t.Error("Notification wasn't set")
		}

		// Verify the error notification displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't have focus")
		}
	})
}

// TestLog_ShowEditModal verfies the log modal is displayed
func TestLogView_ShowEditModal(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Create the app
	app := NewApp(db)
	game, _ := game.NewGame("a game")
	app.gameService.Save(game)
	if game.ID != 1 {
		t.Error("game not saved")
	}

	logEntry, _ := log.NewLog(game.ID)
	logEntry.LogType = log.CHARACTER_ACTION
	logEntry.Result = "result"
	logEntry.Description = "description"
	logEntry.Narrative = "narrative"
	app.logService.Save(logEntry)
	if logEntry.ID != 1 {
		t.Error("log not saved")
	}

	t.Run("open edit modal", func(t *testing.T) {
		app.gameView.Refresh()

		// Navigate down in game tree to select the game
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		testhelper.SimulateTab(app.Application)
		if app.GetFocus() != app.logTextView {
			t.Error("Expected to be on the LogTextView, but wasn't")
		}

		// Select the log to edit
		testhelper.MoveDown(app.logTextView, app.Application, 1)

		// Hit Ctrl+E
		testhelper.SimulateCtrlE(app.logTextView, app.Application)

		// Verify the log modal displayed
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Errorf("Log page doesn't have focus %s", frontPage)
		}

		// It only has Save, Cancel and Delete buttons
		if app.logForm.GetButtonCount() != 3 {
			t.Error("Should only display 3 buttons")
		}

		// Make sure the pre-filled values match the log
		_, option := app.logForm.logTypeField.GetCurrentOption()
		if option != logEntry.LogType.LogTypeDisplayName() {
			t.Error("LogType not correct")
		}

		if app.logForm.resultField.GetText() != logEntry.Result {
			t.Error("Result not correct")
		}

		if app.logForm.descriptionField.GetText() != logEntry.Description {
			t.Error("Description not correct")
		}

		if app.logForm.narrativeField.GetText() != logEntry.Narrative {
			t.Error("narrative doesn't match")
		}

		if app.logForm.gameID != logEntry.GameID {
			t.Error("Game ID doesn't match")
		}

		if *app.logForm.id != logEntry.ID {
			t.Error("Log ID doesn't match")
		}

		// Close it and verify it is hidden, main page is showing, and log view has focus
		testhelper.SimulateEscape(app.logForm, app.Application)
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}
		inFocus := app.GetFocus()
		if inFocus != app.logTextView {
			t.Error("log view doesn't have focus")
		}
	})

	t.Run("error message when log doesn't exist", func(t *testing.T) {
		app.gameView.Refresh()
		app.gameView.SelectGame(&game.ID)

		// Open the modal
		app.logView.ShowEditModal(2)

		note := app.notification.GetText(false)
		if len(note) == 0 {
			t.Error("Notification wasn't set")
		}

		// Verify the error notification displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't have focus")
		}
	})
}

// TestLogHandler_HandleSave verfies logs are saved (both new and edit)
func TestLogView_HandleSave(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Create the app
	app := NewApp(db)
	game, _ := game.NewGame("a game")
	app.gameService.Save(game)
	if game.ID != 1 {
		t.Error("game not saved")
	}

	t.Run("save a new log", func(t *testing.T) {
		app.gameView.Refresh()

		// Navigate down in game tree to select the game
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		// Tab to the logTextView
		testhelper.SimulateTab(app.Application)

		// Open the new modal
		testhelper.SimulateKey(app.logTextView, app.Application, tcell.KeyCtrlN)

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Fill in the data
		app.logForm.logTypeField.SetCurrentOption(1)
		app.logForm.resultField.SetText("result")
		app.logForm.descriptionField.SetText("descripton")
		app.logForm.narrativeField.SetText("narrative", false)

		// Save it
		testhelper.SimulateKey(app.logForm, app.Application, tcell.KeyCtrlS)

		// Verify the modal closed
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}

		inFocus := app.GetFocus()
		if inFocus != app.logTextView {
			t.Error("log view doesn't have focus")
		}

		// Verify the game has a log
		logs, _ := app.logService.GetAllForGame(game.ID)
		if len(logs) != 1 {
			t.Error("No logs for game")
		}
	})

	t.Run("save an existing log", func(t *testing.T) {
		// start fresh
		app.gameView.SelectGame(nil)
		app.SetFocus(app.gameTree)
		app.gameView.Refresh()

		// Log to update
		logEntry, _ := log.NewLog(game.ID)
		logEntry.LogType = log.CHARACTER_ACTION
		logEntry.Result = "result"
		logEntry.Description = "description"
		logEntry.Narrative = "narrative"
		logEntry, err := app.logService.Save(logEntry)
		if err != nil {
			t.Errorf("Issue saving the log: %e", err)
		}

		// Verify there are no logs for the game
		logs, _ := app.logService.GetAllForGame(game.ID)
		if len(logs) == 0 {
			t.Error("No logs for game")
		}

		// Navigate down in game tree to select the game
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		// Tab to the logTextView
		testhelper.SimulateTab(app.Application)
		// Select the log to edit
		testhelper.MoveDown(app.logTextView, app.Application, 1)
		testhelper.SimulateCtrlE(app.logTextView, app.Application)

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Change everything
		app.logForm.logTypeField.SetCurrentOption(2)
		app.logForm.resultField.SetText("new result")
		app.logForm.descriptionField.SetText("new descripton")
		app.logForm.narrativeField.SetText("new narrative", false)

		// Save it
		testhelper.SimulateKey(app.logForm, app.Application, tcell.KeyCtrlS)

		// Verify the modal closed and the data was saved
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}

		inFocus := app.GetFocus()
		if inFocus != app.logTextView {
			t.Error("log view doesn't have focus")
		}

		// Verify the game has a log
		logs, _ = app.logService.GetAllForGame(game.ID)
		if len(logs) == 0 {
			t.Error("No logs for game")
		}

		// Make sure the narrative is different
		if logs[len(logs)-1].LogType == logEntry.LogType {
			t.Error("log type didn't change")
		}
		if logs[len(logs)-1].Description == logEntry.Description {
			t.Error("description didn't change")
		}
		if logs[len(logs)-1].Narrative == logEntry.Narrative {
			t.Error("narrative didn't change")
		}
		if logs[len(logs)-1].Result == logEntry.Result {
			t.Error("result didn't change")
		}
	})

	t.Run("save a log with validation errors", func(t *testing.T) {
		// Start Fresh
		app.gameView.SelectGame(nil)
		app.gameView.Refresh()

		// Select the game, tab to the log view and open the new modal
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)
		testhelper.SimulateTab(app.Application)
		if app.GetFocus() != app.logTextView {
			t.Error("Log Text View not in focus")
		}
		testhelper.SimulateCtrlN(app.logTextView, app.Application)

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Fill in the data
		app.logForm.logTypeField.SetCurrentOption(1)

		// Save it
		testhelper.SimulateCtrlS(app.logForm, app.Application)

		// Close it and verify it is hidden, main page is showing, and log view has focus
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log Modal doesn't isn't in front")
		}

		// Verify there are errors on the modal
		if !app.logForm.HasFieldErrors() {
			t.Error("No field errors")
		}
	})

}

// TestLogHandler_HandleDelete verfies logs can be deleted with confirmation
func TestLogView_HandleDelete(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Create the app
	app := NewApp(db)
	game, _ := game.NewGame("a game")
	app.gameService.Save(game)
	app.selectedGame = game

	t.Run("delete a log with confirmation", func(t *testing.T) {
		// Create a log to delete
		logEntry, _ := log.NewLog(game.ID)
		logEntry.LogType = log.CHARACTER_ACTION
		logEntry.Result = "result"
		logEntry.Description = "description"
		logEntry.Narrative = "narrative"
		app.logService.Save(logEntry)

		// Verify log exists
		logs, _ := app.logService.GetAllForGame(game.ID)
		initialCount := len(logs)
		if initialCount == 0 {
			t.Error("No logs for game")
		}

		// Reset the game view and select the game
		app.gameView.Refresh()
		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		// Open edit modal
		testhelper.SimulateTab(app.Application)
		if app.GetFocus() != app.logTextView {
			t.Error("Log text view not in focus")
		}
		// Select the log to edit
		testhelper.MoveDown(app.logTextView, app.Application, 1)
		testhelper.SimulateCtrlE(app.logTextView, app.Application)

		// Verify modal is showing
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Delete it - this sets up the confirmation modal
		testhelper.SimulateCtrlD(app.logForm, app.Application)

		// Verify confirmation modal is showing
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != CONFIRM_MODAL_ID {
			t.Error("Confirm modal should be showing")
		}

		// The delete button is in focus, so press enter to confirm
		testhelper.SimulateEnter(app.confirmModal, app.Application)

		// Verify log was deleted from database
		logs, _ = app.logService.GetAllForGame(game.ID)
		if len(logs) >= initialCount {
			t.Error("Log wasn't deleted")
		}

		// Verify we're back on main page
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Should be back on main page after delete")
		}
	})

	t.Run("cancel log deletion", func(t *testing.T) {
		// Create a log
		logEntry, _ := log.NewLog(game.ID)
		logEntry.LogType = log.MECHANICS
		logEntry.Result = "result"
		logEntry.Description = "description"
		logEntry.Narrative = "narrative"
		app.logService.Save(logEntry)

		// Get count before
		logs, _ := app.logService.GetAllForGame(game.ID)
		initialCount := len(logs)

		// Start fresh
		app.gameView.SelectGame(nil)
		app.gameView.Refresh()

		testhelper.SelectTreeEntry(app.gameTree, app.Application, 1)

		// Open edit modal
		testhelper.SimulateTab(app.Application)
		if app.GetFocus() != app.logTextView {
			t.Error("Log text view not in focus")
		}
		// Select the log to edit
		testhelper.MoveDown(app.logTextView, app.Application, 1)
		testhelper.SimulateCtrlE(app.logTextView, app.Application)

		// Delete it - this sets up the confirmation modal
		testhelper.SimulateCtrlD(app.logForm, app.Application)

		// Verify confirmation modal is showing
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != CONFIRM_MODAL_ID {
			t.Error("Confirm modal should be showing")
		}

		// Tab to the cancel button and press enter to cancel the deletion
		testhelper.SimulateKey(app.confirmModal, app.Application, tcell.KeyTab)
		testhelper.SimulateEnter(app.confirmModal, app.Application)

		// Verify log still exists
		logs, _ = app.logService.GetAllForGame(game.ID)
		if len(logs) != initialCount {
			t.Error("Log shouldn't be deleted on cancel")
		}

		// Verify we're back on log modal (not deleted, cancelled)
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Should be back on log modal after cancel")
		}
	})

}
