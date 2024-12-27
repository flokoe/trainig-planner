package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type CalendarDay struct {
	Date     time.Time
	Sessions []SessionWithPlan
}

type SessionWithPlan struct {
	ID          int64
	PlanID      int64
	PlanName    string
	Description string
	Date        time.Time
}

type CalendarData struct {
	Days        []CalendarDay
	CurrentWeek time.Time
	WeekOffset  int
	WeekNumber  int
	Year        int
}

func handleCalendar(db *sql.DB) http.HandlerFunc {
	// Register template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
	}
	
	tmpl := template.Must(template.New("calendar.html").Funcs(funcMap).ParseFiles("internal/templates/calendar.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get week offset from query parameter
		weekOffset := 0
		if offsetStr := r.URL.Query().Get("weekOffset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err == nil {
				weekOffset = offset
			}
		}

		// Get Monday of the requested week
		now := time.Now()
		monday := now.AddDate(0, 0, -int(now.Weekday())+1)
		monday = monday.AddDate(0, 0, weekOffset*7)
		
		// Create slice for 7 days
		days := make([]CalendarDay, 7)
		
		// For each day of the week
		for i := 0; i < 7; i++ {
			currentDate := monday.AddDate(0, 0, i)
			
			// Get sessions with plan names for this day
			rows, err := db.Query(`
				SELECT ts.id, ts.plan_id, p.name, ts.description, ts.date 
				FROM training_sessions ts 
				JOIN training_plans p ON ts.plan_id = p.id
				WHERE DATE(ts.date) = DATE(?)
				ORDER BY ts.date
			`, currentDate.Format("2006-01-02"))
			
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var sessions []SessionWithPlan
			for rows.Next() {
				var session SessionWithPlan
				err := rows.Scan(&session.ID, &session.PlanID, &session.PlanName, &session.Description, &session.Date)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				sessions = append(sessions, session)
			}

			days[i] = CalendarDay{
				Date:     currentDate,
				Sessions: sessions,
			}
		}

		year, week := monday.ISOWeek()
		data := CalendarData{
			Days:        days,
			CurrentWeek: monday,
			WeekOffset:  weekOffset,
			WeekNumber:  week,
			Year:        year,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
