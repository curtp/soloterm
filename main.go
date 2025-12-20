package main

import (
	"log"
	"os"
	"soloterm/database"
	"soloterm/ui"

	"github.com/rivo/tview"
)

type AppUI struct {
	*tview.Application

	// Layout containers
	mainFlex  *tview.Flex
	rightFlex *tview.Flex
	pages     *tview.Pages

	// UI Components
	gameTree  *tview.TreeView
	logView   *tview.TextView
	inputArea *tview.TextArea
	helpModal *tview.Modal
	footer    *tview.TextView
}

func main() {
	// Setup logging to file
	logFile, err := os.OpenFile("soloterm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("=== SoloTerm Starting ===")

	// Setup database (connect + migrate)
	db, err := database.Setup()
	if err != nil {
		log.Fatal("Database setup failed:", err)
	}
	defer db.Close()

	log.Println("Database setup complete")

	// Create and run the TUI application
	app := ui.NewApp(db)
	log.Println("App created, starting...")
	if err := app.EnableMouse(true).Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}
