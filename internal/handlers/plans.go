package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
	"training-tracker/internal/models"
)

type SessionsYAML struct {
	Sessions []struct {
		Order       int       `yaml:"order"`
		Description string    `yaml:"description"`
		Date        time.Time `yaml:"date"`
		// Type-specific fields
		HFMax       string    `yaml:"hfmax,omitempty"`      // For cycling
		// Mobility has no additional fields
		// Sandbag has no additional fields yet
	} `yaml:"sessions"`
}

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

			// Start a transaction
			tx, err := db.Begin()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			// Insert new plan into database
			result, err := tx.Exec(`
				INSERT INTO training_plans (name, workout_type_id, created_at)
				VALUES (?, ?, ?)
			`, name, workoutTypeID, time.Now())

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Get the ID of the newly inserted plan
			planID, err := result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Process YAML sessions if provided
			yamlData := r.FormValue("yaml_sessions")
			if yamlData != "" {
				var sessions SessionsYAML
				if err := yaml.Unmarshal([]byte(yamlData), &sessions); err != nil {
					http.Error(w, "Invalid YAML format: "+err.Error(), http.StatusBadRequest)
					return
				}

				// Insert all sessions
				for _, s := range sessions.Sessions {
					// Start with base session insertion
					result, err := tx.Exec(`
						INSERT INTO training_sessions (plan_id, session_order, description, date)
						VALUES (?, ?, ?, ?)
					`, planID, s.Order, s.Description, s.Date)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					sessionID, err := result.LastInsertId()
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					// Handle type-specific fields based on workout type
					switch workoutTypeID {
					case "1": // cycling
						if s.HFMax != "" {
							_, err = tx.Exec(`
								INSERT INTO cycling_sessions (session_id, hfmax)
								VALUES (?, ?)
							`, sessionID, s.HFMax)
							if err != nil {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
						}
					case "2": // mobility
						_, err = tx.Exec(`
							INSERT INTO mobility_sessions (session_id)
							VALUES (?)
						`, sessionID)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
					case "3": // sandbag
						_, err = tx.Exec(`
							INSERT INTO sandbag_sessions (session_id)
							VALUES (?)
						`, sessionID)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
					}
				}
			}

			// Commit the transaction
			if err := tx.Commit(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect to plan view
			http.Redirect(w, r, fmt.Sprintf("/plans/%d", planID), http.StatusSeeOther)
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

		// Get training sessions with type-specific details
		rows, err := db.Query(`
			SELECT 
				ts.id, 
				ts.session_order, 
				ts.description, 
				ts.date,
				CASE 
					WHEN cs.session_id IS NOT NULL THEN cs.hfmax
					ELSE NULL
				END as hfmax
			FROM training_sessions ts
			LEFT JOIN cycling_sessions cs ON ts.id = cs.session_id
			LEFT JOIN mobility_sessions ms ON ts.id = ms.session_id
			LEFT JOIN sandbag_sessions ss ON ts.id = ss.session_id
			WHERE ts.plan_id = ? 
			ORDER BY ts.session_order`, plan.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sessions []struct {
			models.TrainingSession
			HFMax string `json:"hfmax,omitempty"`
		}

		for rows.Next() {
			var session struct {
				models.TrainingSession
				HFMax string `json:"hfmax,omitempty"`
			}
			if err := rows.Scan(
				&session.ID,
				&session.SessionOrder,
				&session.Description,
				&session.Date,
				&session.HFMax,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sessions = append(sessions, session)
		}

		data := struct {
			Plan            models.TrainingPlan
			WorkoutTypeName string
			Sessions        []struct {
				models.TrainingSession
				HFMax string `json:"hfmax,omitempty"`
			}
			WorkoutTypeID   int64
		}{
			Plan:            plan,
			WorkoutTypeName: workoutTypeName,
			Sessions:        sessions,
			WorkoutTypeID:   plan.WorkoutTypeID,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
