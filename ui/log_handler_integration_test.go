package ui

import (
	"testing"

	"soloterm/domain/game"
	"soloterm/domain/log"
	testhelper "soloterm/shared/testing"
)

// TestLogHandler_ShowModal verfies the log modal is displayed
func TestLogHandler_ShowModal(t *testing.T) {
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
		// Need an active game to view the modal
		app.selectedGame = game

		// Open the modal
		app.logHandler.ShowModal()

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// It only has Save and Cancel buttons
		if app.logForm.GetButtonCount() != 2 {
			t.Error("Should only display 2 buttons")
		}

		// Close it and verify it is hidden, main page is showing, and log view has focus
		app.logHandler.HandleCancel()
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}
		inFocus := app.GetFocus()
		if inFocus != app.logView {
			t.Error("log view doesn't have focus")
		}
	})

	t.Run("error message when no game selected", func(t *testing.T) {
		// Open the modal
		app.selectedGame = nil
		app.logHandler.ShowModal()

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

// TestLogHandler_ShowEditModal verfies the log modal is displayed
func TestLogHandler_ShowEditModal(t *testing.T) {
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
		// Need an active game to view the modal
		app.selectedGame = game

		// Open the modal
		app.logHandler.ShowEditModal(1)

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
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
		app.logHandler.HandleCancel()
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}
		inFocus := app.GetFocus()
		if inFocus != app.logView {
			t.Error("log view doesn't have focus")
		}
	})

	t.Run("error message when log doesn't exist", func(t *testing.T) {
		// Open the modal
		app.logHandler.ShowEditModal(2)

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
func TestLogHandler_HandleSave(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.TeardownTestDB(t, db)

	// Create the app
	app := NewApp(db)
	game, _ := game.NewGame("a game")
	app.gameService.Save(game)
	if game.ID != 1 {
		t.Error("game not saved")
	}

	app.selectedGame = game

	//var logEntry log.Log

	t.Run("save a new log", func(t *testing.T) {
		// Open the new modal
		app.logHandler.ShowModal()

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
		app.logHandler.HandleSave()

		// Close it and verify it is hidden, main page is showing, and log view has focus
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}

		inFocus := app.GetFocus()
		if inFocus != app.logView {
			t.Error("log view doesn't have focus")
		}

		// Verify the game has a log
		logs, _ := app.logService.GetAllForGame(game.ID)
		if len(logs) != 1 {
			t.Error("No logs for game")
		}
	})

	t.Run("save an existing log", func(t *testing.T) {
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

		t.Logf("Entry: %+v", logEntry)

		// Verify there are no logs for the game
		logs, _ := app.logService.GetAllForGame(game.ID)
		t.Logf("Log count: %d", len(logs))
		if len(logs) == 0 {
			t.Error("No logs for game")
		}

		// Open the new modal
		app.logHandler.ShowEditModal(logs[0].ID)

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
		app.logHandler.HandleSave()

		// Close it and verify it is hidden, main page is showing, and log view has focus
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != MAIN_PAGE_ID {
			t.Error("Main page doesn't isn't in front")
		}

		inFocus := app.GetFocus()
		if inFocus != app.logView {
			t.Error("log view doesn't have focus")
		}

		// Verify the game has a log
		logs, _ = app.logService.GetAllForGame(game.ID)
		t.Logf("logs: %d", len(logs))
		if len(logs) == 0 {
			t.Error("No logs for game")
		}

		// Make sure the narrative is different
		if logs[0].LogType == logEntry.LogType {
			t.Error("log type didn't change")
		}
		if logs[0].Description == logEntry.Description {
			t.Error("description didn't change")
		}
		if logs[0].Narrative == logEntry.Narrative {
			t.Error("narrative didn't change")
		}
		if logs[0].Result == logEntry.Result {
			t.Error("result didn't change")
		}
	})

	t.Run("save a log with validation errors", func(t *testing.T) {
		// Open the new modal
		app.logHandler.ShowModal()

		// Verify it displays
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Fill in the data
		app.logForm.logTypeField.SetCurrentOption(1)

		// Save it
		app.logHandler.HandleSave()

		// Close it and verify it is hidden, main page is showing, and log view has focus
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log Modal doesn't isn't in front")
		}

		// Verify there are errors on the modal
		if app.logForm.HasFieldErrors() {
			t.Error("No field errors")
		}
	})

}

// TestLogHandler_HandleDelete verfies logs can be deleted with confirmation
func TestLogHandler_HandleDelete(t *testing.T) {
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

		// Open edit modal
		app.logHandler.ShowEditModal(logEntry.ID)

		// Verify modal is showing
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != LOG_MODAL_ID {
			t.Error("Log page doesn't have focus")
		}

		// Call HandleDelete - this sets up the confirmation modal
		app.logHandler.HandleDelete()

		// Verify confirmation modal is showing
		frontPage, _ = app.pages.GetFrontPage()
		if frontPage != CONFIRM_MODAL_ID {
			t.Error("Confirm modal should be showing")
		}

		// Simulate clicking "Delete" by manually calling the onConfirm callback
		if app.confirmModal.onConfirm != nil {
			app.confirmModal.onConfirm()
		}

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

		// Open edit modal
		app.logHandler.ShowEditModal(logEntry.ID)

		// Call HandleDelete
		app.logHandler.HandleDelete()

		// Verify confirmation modal is showing
		frontPage, _ := app.pages.GetFrontPage()
		if frontPage != CONFIRM_MODAL_ID {
			t.Error("Confirm modal should be showing")
		}

		// Simulate clicking "Cancel" by calling onCancel callback
		if app.confirmModal.onCancel != nil {
			app.confirmModal.onCancel()
		}

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
