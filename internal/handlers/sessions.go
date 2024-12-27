package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func handleCreateSession(db *sql.DB) http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("internal/templates/create_session.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		// Extract plan ID from URL
		planID := strings.TrimPrefix(r.URL.Path, "/sessions/create/")
		if planID == "" {
			http.Error(w, "Plan ID is required", http.StatusBadRequest)
			return
		}

		// Get workout type for the plan
		var workoutType string
		err := db.QueryRow(`
			SELECT wt.name 
			FROM workout_types wt 
			JOIN training_plans tp ON tp.workout_type_id = wt.id 
			WHERE tp.id = ?`, planID).Scan(&workoutType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Method == "GET" {
			data := struct {
				PlanID      string
				WorkoutType string
			}{
				PlanID:      planID,
				WorkoutType: workoutType,
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		if r.Method == "POST" {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Begin transaction
			tx, err := db.Begin()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			// Parse and validate date
			date, err := time.Parse("2006-01-02T15:04", r.FormValue("date"))
			if err != nil {
				http.Error(w, "Invalid date format", http.StatusBadRequest)
				return
			}

			// Insert base session
			result, err := tx.Exec(`
				INSERT INTO training_sessions (plan_id, session_order, description, date)
				VALUES (?, NULLIF(?, ''), ?, ?)`,
				planID,
				r.FormValue("session_order"),
				r.FormValue("description"),
				date)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			sessionID, err := result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Handle workout-type specific data
			switch workoutType {
			case "cycling":
				hfMax, _ := strconv.Atoi(r.FormValue("hf_max"))
				_, err = tx.Exec(`
					INSERT INTO cycling_sessions (session_id, hf_max)
					VALUES (?, ?)`,
					sessionID, hfMax)
			case "mobility":
				_, err = tx.Exec(`
					INSERT INTO mobility_sessions (session_id)
					VALUES (?)`,
					sessionID)
			case "sandbag":
				_, err = tx.Exec(`
					INSERT INTO sandbag_sessions (session_id)
					VALUES (?)`,
					sessionID)
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect back to plan view
			http.Redirect(w, r, "/plans/"+planID, http.StatusSeeOther)
		}
	}
}
