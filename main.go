package main

import (
	"log"
	"os"
	"soloterm/database"
	"soloterm/ui"
)

func main() {
	// Setup logging to file
	logFile, err := os.OpenFile("soloterm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Failed to open log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// log.SetFlags(log.Ldate | log.Ltime)

	// Setup database (connect + migrate)
	db, err := database.Setup()
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Database setup failed: ", err)
	}
	defer db.Connection.Close()

	log.Printf("App starting with database: %s", *db.Path)

	// Create and run the TUI application
	app := ui.NewApp(db)
	if err := app.EnableMouse(true).Run(); err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Application error:", err)
	}
}
