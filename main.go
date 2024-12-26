package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"

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

		// Get the current active plan (most recent one that's ongoing)
		var activePlan models.TrainingPlan
		err := db.QueryRow(`
			SELECT id, name, description 
			FROM training_plans 
			ORDER BY id DESC LIMIT 1
		`).Scan(&activePlan.ID, &activePlan.Name, &activePlan.Description)

		data := struct {
			ActivePlan *models.TrainingPlan
		}{nil}

		if err == nil {
			data.ActivePlan = &activePlan
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, data)
	}))

	// Handle plans listing
	http.HandleFunc("/plans", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, description FROM training_plans ORDER BY id DESC")
		if err != nil {
			http.Error(w, "Failed to fetch plans", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var plans []models.TrainingPlan
		for rows.Next() {
			var plan models.TrainingPlan
			err := rows.Scan(&plan.ID, &plan.Name, &plan.Description)
			if err != nil {
				log.Printf("Error scanning plan row: %v", err)
				continue
			}
			plans = append(plans, plan)
		}

		tmpl := template.Must(template.ParseFiles("templates/plans.html"))
		tmpl.Execute(w, struct{ Plans []models.TrainingPlan }{plans})
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

		// Parse form values
		name := r.FormValue("name")
		description := r.FormValue("description")

		// Validate required fields
		if name == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create training plan
		plan := &models.TrainingPlan{
			Name:        name,
			Description: description,
		}

		// Insert into database
		result, err := db.Exec(
			"INSERT INTO training_plans (name, description) VALUES (?, ?)",
			plan.Name,
			plan.Description,
		)
		if err != nil {
			log.Printf("Error creating training plan: %v", err)
			http.Error(w, "Failed to create training plan", http.StatusInternalServerError)
			return
		}

		// Get the ID of the newly created plan
		planID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last insert ID: %v", err)
		} else {
			log.Printf("Created training plan with ID: %d", planID)
		}

		http.Redirect(w, r, "/plans/"+strconv.FormatInt(planID, 10), http.StatusSeeOther)
	}))

	// Handle plan views and session form
	http.HandleFunc("/plans/", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// if r.URL.Path[len("/plans/"):] == "new" {
		// 	http.ServeFile(w, r, "templates/plan_form.html")
		// 	return
		// }

		pathParts := strings.Split(r.URL.Path, "/")
		fmt.Println(pathParts)
		// if len(pathParts) < 3 {
		// 	http.NotFound(w, r)
		// 	return
		// }

		// Handle session form
		if len(pathParts) == 5 && pathParts[3] == "sessions" {
			if pathParts[4] == "new" {
				planID := pathParts[2]
				tmpl := template.Must(template.ParseFiles("templates/session_form.html"))
				tmpl.Execute(w, struct{ PlanID string }{planID})
				return
			}
		}

		// Handle single plan view
		planID, err := strconv.ParseInt(pathParts[2], 10, 64)
		if err != nil {
			http.Error(w, "Invalid plan ID", http.StatusBadRequest)
			return
		}

		// Get plan details
		var plan models.TrainingPlan
		err = db.QueryRow(`
			SELECT id, name, description 
			FROM training_plans 
			WHERE id = ?`, planID).Scan(&plan.ID, &plan.Name, &plan.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Database error", http.StatusInternalServerError)
			}
			return
		}

		// Get associated sessions
		rows, err := db.Query(`
			SELECT id, scheduled_date, workout, description, intensity 
			FROM training_sessions 
			WHERE plan_id = ? 
			ORDER BY scheduled_date ASC`, planID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sessions []models.TrainingSession
		for rows.Next() {
			var session models.TrainingSession
			err := rows.Scan(&session.ID, &session.ScheduledDate, &session.Workout, &session.Description, &session.Intensity)
			if err != nil {
				log.Printf("Error scanning session row: %v", err)
				continue
			}
			sessions = append(sessions, session)
		}

		data := struct {
			Plan     models.TrainingPlan
			Sessions []models.TrainingSession
		}{
			Plan:     plan,
			Sessions: sessions,
		}

		tmpl := template.Must(template.ParseFiles("templates/plan.html"))
		tmpl.Execute(w, data)
	}))

	// Handle session creation
	http.HandleFunc("/sessions/create", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		planID := r.FormValue("plan_id")
		dateStr := r.FormValue("date")
		workout := r.FormValue("workout")
		description := r.FormValue("description")

		// Parse values
		planIDInt, err := strconv.ParseInt(planID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid plan ID", http.StatusBadRequest)
			return
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		// Insert into database
		_, err = db.Exec(
			"INSERT INTO training_sessions (plan_id, scheduled_date, workout, description, intensity) VALUES (?, ?, ?, ?, ?)",
			planIDInt,
			date,
			workout,
			description,
			0, // Default intensity to 0 since we're not using it
		)
		if err != nil {
			log.Printf("Error creating training session: %v", err)
			http.Error(w, "Failed to create training session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/plans", http.StatusSeeOther)
	}))

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
