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

const version = "1.1.2"

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

	// Resolve database path: DB_PATH env → config database_dir → default data dir
	dbPath := database.ResolveDBPath(loadedCfg.DatabaseDir, dataDir)

	// Setup database (connect + migrate)
	db, err := database.Setup(dbPath)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Database setup failed: ", err)
	}
	log.Printf("Database is stored in: %s", dbPath)
	defer db.Connection.Close()

	log.Print("Starting...")
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	info := ui.AppInfo{
		Version:      version,
		ConfigFile:   loadedCfg.FullFilePath,
		LogFile:      logPath,
		DatabasePath: dbPath,
	}

	// Create and run the TUI application
	app := ui.NewApp(db, loadedCfg, info)
	if err := app.EnableMouse(false).Run(); err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Application error:", err)
	}
}
