package main

import (
	"log"
	"os"
	"path/filepath"
	"soloterm/config"
	"soloterm/database"
	"soloterm/shared/dirs"
	"soloterm/ui"
)

func main() {
	log.SetOutput(os.Stdout)

	// Resolve directories
	configDir, err := dirs.ConfigDir()
	if err != nil {
		log.Fatal("Failed to resolve config directory: ", err)
	}
	dataDir, err := dirs.DataDir()
	if err != nil {
		log.Fatal("Failed to resolve data directory: ", err)
	}

	// Load configuration
	var cfg config.Config
	loadedCfg, err := cfg.Load(configDir)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Failed to load config: ", err)
	}
	log.Printf("Using configuration file: %s", cfg.FullFilePath)

	// Setup logging to file
	logPath := filepath.Join(dataDir, "soloterm.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer logFile.Close()
	log.Printf("Logs are written to: %s", logFile.Name())

	// Setup database (connect + migrate)
	db, err := database.Setup(nil)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Database setup failed: ", err)
	}
	log.Printf("Database is stored in: %s", dataDir)
	defer db.Connection.Close()

	log.Print("Starting...")
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create and run the TUI application
	app := ui.NewApp(db, loadedCfg)
	if err := app.EnableMouse(false).Run(); err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Application error:", err)
	}
}
