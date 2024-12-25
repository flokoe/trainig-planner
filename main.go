package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mattn/go-sqlite3"
	"training-tracker/internal/database"
	"training-tracker/internal/middleware"
	"training-tracker/internal/models"
)

func main() {
	// Initialize database
	db, err := database.InitDB("training.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle main page
	http.HandleFunc("/", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	}))

	// Handle training plan form
	http.HandleFunc("/plans/new", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/plan_form.html"))
		tmpl.Execute(w, nil)
	}))

	// Handle training plan creation
	http.HandleFunc("/plans/create", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// TODO: Implement plan creation logic with database
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
