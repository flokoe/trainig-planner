package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
			SELECT id, name, start_date, end_date, description 
			FROM training_plans 
			WHERE start_date <= DATE('now') 
			AND end_date >= DATE('now') 
			ORDER BY start_date DESC LIMIT 1
		`).Scan(&activePlan.ID, &activePlan.Name, &activePlan.StartDate, &activePlan.EndDate, &activePlan.Description)

		data := struct {
			ActivePlan *models.TrainingPlan
			CurrentDay int
			TotalDays int
		}{nil, 0, 0}

		if err == nil {
			data.ActivePlan = &activePlan
			data.CurrentDay = int(time.Since(activePlan.StartDate).Hours()/24) + 1
			data.TotalDays = int(activePlan.EndDate.Sub(activePlan.StartDate).Hours()/24) + 1
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, data)
	}))

	// Handle plans listing
	http.HandleFunc("/plans", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, start_date, end_date, description FROM training_plans ORDER BY start_date DESC")
		if err != nil {
			http.Error(w, "Failed to fetch plans", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var plans []models.TrainingPlan
		for rows.Next() {
			var plan models.TrainingPlan
			err := rows.Scan(&plan.ID, &plan.Name, &plan.StartDate, &plan.EndDate, &plan.Description)
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
		startDateStr := r.FormValue("start_date")
		endDateStr := r.FormValue("end_date")
		description := r.FormValue("description")

		// Validate required fields
		if name == "" || startDateStr == "" || endDateStr == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}

		// Create training plan
		plan := &models.TrainingPlan{
			Name:        name,
			StartDate:   startDate,
			EndDate:     endDate,
			Description: description,
		}

		// Insert into database
		result, err := db.Exec(
			"INSERT INTO training_plans (name, start_date, end_date, description) VALUES (?, ?, ?, ?)",
			plan.Name,
			plan.StartDate,
			plan.EndDate,
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

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}))

	// Handle session form
	http.HandleFunc("/plans/", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[len("/plans/"):] == "new" {
			http.ServeFile(w, r, "templates/plan_form.html")
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) == 4 && pathParts[3] == "new" {
			planID := pathParts[2]
			tmpl := template.Must(template.ParseFiles("templates/session_form.html"))
			tmpl.Execute(w, struct{ PlanID string }{planID})
			return
		}
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
		sessionType := r.FormValue("type")
		description := r.FormValue("description")
		intensityStr := r.FormValue("intensity")

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

		intensity, err := strconv.Atoi(intensityStr)
		if err != nil {
			http.Error(w, "Invalid intensity value", http.StatusBadRequest)
			return
		}

		// Insert into database
		_, err = db.Exec(
			"INSERT INTO training_sessions (plan_id, scheduled_date, type, description, intensity) VALUES (?, ?, ?, ?, ?)",
			planIDInt,
			date,
			sessionType,
			description,
			intensity,
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
