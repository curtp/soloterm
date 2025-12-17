package main

import (
	"log"

	"soloterm/database"
)

func main() {
	// Setup database (connect + migrate)
	db, err := database.Setup()
	if err != nil {
		log.Fatal("Database setup failed:", err)
	}
	defer db.Close()

	log.Println("Database connected and migrated!")

	log.Println("Hello World")
}
