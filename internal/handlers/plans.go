package handlers

import (
	"database/sql"
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
			_, err := db.Exec(`
				INSERT INTO training_plans (name, workout_type_id, created_at)
				VALUES (?, ?, ?)
			`, name, workoutTypeID, time.Now())

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect to plans list
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
