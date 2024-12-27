package main

import (
	"log"
	"net/http"
	"training-tracker/internal/database"
	"training-tracker/internal/handlers"
)

func main() {
	// Initialize database
	db, err := database.InitDB("training.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables
	if err := database.CreateTables(db); err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	mux := http.NewServeMux()
	
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register routes
	handlers.RegisterRoutes(mux, db)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
