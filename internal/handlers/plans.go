package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"training-tracker/internal/models"
)

func handleCreatePlan(db *sql.DB) http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("internal/templates/create_plan.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Get workout types for the dropdown
			rows, err := db.Query("SELECT id, name FROM workout_types")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var workoutTypes []models.WorkoutType
			for rows.Next() {
				var wt models.WorkoutType
				if err := rows.Scan(&wt.ID, &wt.Name); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				workoutTypes = append(workoutTypes, wt)
			}

			data := struct {
				WorkoutTypes []models.WorkoutType
			}{
				WorkoutTypes: workoutTypes,
			}

			tmpl.Execute(w, data)
			return
		}

		if r.Method == "POST" {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			name := r.FormValue("name")
			workoutTypeID := r.FormValue("workout_type_id")

			// Insert new plan into database
			result, err := db.Exec(`
				INSERT INTO training_plans (name, workout_type_id, created_at)
				VALUES (?, ?, ?)
			`, name, workoutTypeID, time.Now())

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get the ID of the newly inserted plan
			id, err := result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect to plan view
			http.Redirect(w, r, fmt.Sprintf("/plans/%d", id), http.StatusSeeOther)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleListPlans(db *sql.DB) http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("internal/templates/list_plans.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get all plans
		rows, err := db.Query(`
			SELECT id, name, workout_type_id, created_at 
			FROM training_plans 
			ORDER BY created_at DESC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var plans []models.TrainingPlan
		for rows.Next() {
			var plan models.TrainingPlan
			if err := rows.Scan(&plan.ID, &plan.Name, &plan.WorkoutTypeID, &plan.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			plans = append(plans, plan)
		}

		// Get all workout types to create a map of ID to name
		workoutTypeRows, err := db.Query("SELECT id, name FROM workout_types")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer workoutTypeRows.Close()

		workoutTypeNames := make(map[int64]string)
		for workoutTypeRows.Next() {
			var id int64
			var name string
			if err := workoutTypeRows.Scan(&id, &name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			workoutTypeNames[id] = name
		}

		data := struct {
			Plans            []models.TrainingPlan
			WorkoutTypeNames map[int64]string
		}{
			Plans:            plans,
			WorkoutTypeNames: workoutTypeNames,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleViewPlan(db *sql.DB) http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("internal/templates/view_plan.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract plan ID from URL path
		planID := r.URL.Path[len("/plans/"):]
		if planID == "" {
			http.Error(w, "Plan ID is required", http.StatusBadRequest)
			return
		}

		// Get plan details
		var plan models.TrainingPlan
		err := db.QueryRow(`
			SELECT id, name, workout_type_id, created_at 
			FROM training_plans 
			WHERE id = ?`, planID).Scan(&plan.ID, &plan.Name, &plan.WorkoutTypeID, &plan.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Plan not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get workout type name
		var workoutTypeName string
		err = db.QueryRow(`
			SELECT name 
			FROM workout_types 
			WHERE id = ?`, plan.WorkoutTypeID).Scan(&workoutTypeName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get training sessions
		rows, err := db.Query(`
			SELECT id, session_order, description, date 
			FROM training_sessions 
			WHERE plan_id = ? 
			ORDER BY session_order`, plan.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sessions []models.TrainingSession
		for rows.Next() {
			var session models.TrainingSession
			if err := rows.Scan(&session.ID, &session.SessionOrder, &session.Description, &session.Date); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sessions = append(sessions, session)
		}

		data := struct {
			Plan           models.TrainingPlan
			WorkoutTypeName string
			Sessions       []models.TrainingSession
		}{
			Plan:           plan,
			WorkoutTypeName: workoutTypeName,
			Sessions:       sessions,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
