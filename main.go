package main

import (
	"log"
	"os"
	"soloterm/database"
	"soloterm/ui"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)

	setupEnvironment()

	log.Print("Starting...")
	time.Sleep(3 * time.Second)

	// Setup logging to file
	logFile, err := os.OpenFile(getWorkingDirectory()+"/soloterm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Setup database (connect + migrate)
	db, err := database.Setup(nil)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Database setup failed: ", err)
	}
	defer db.Connection.Close()

	// Create and run the TUI application
	app := ui.NewApp(db)
	if err := app.EnableMouse(false).Run(); err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal("Application error:", err)
	}
}

func setupEnvironment() {
	if workDir := os.Getenv("SOLOTERM_WORK_DIR"); workDir == "" {
		workDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Unable to locate the users home directory. Set the SOLOTERM_WORK_DIR environment variable to a directory to store the data and log.")
		}
		workDir += "/soloterm"
		os.Setenv("SOLOTERM_WORK_DIR", workDir)
		log.Printf("SOLOTERM_WORK_DIR environment variable not found")
		log.Printf("Setting to: %s", workDir)
		log.Printf("To change the work dir, set the environment variable and restart the app.")
		log.Printf("The database files and application log will be located in this directory.\n\n")
	}
}

func getWorkingDirectory() string {
	return os.Getenv("SOLOTERM_WORK_DIR")
}
