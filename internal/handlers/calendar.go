package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type MonthDay struct {
    Date          time.Time
    IsCurrentMonth bool
    Sessions      []MonthSession
}

type MonthSession struct {
    PlanName    string
    WorkoutType string
    Date        time.Time
}

type MonthData struct {
    Days  []MonthDay
    Month time.Month
    Year  int
}

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
	WorkoutType string
	HFMax       sql.NullString  // For cycling
}

type CalendarData struct {
    Days        []CalendarDay
    CurrentWeek time.Time
    WeekOffset  int
    WeekNumber  int
    Year        int
    MonthData   MonthData
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
		"seq": func(start, end int) []int {
			seq := make([]int, end-start+1)
			for i := range seq {
				seq[i] = start + i
			}
			return seq
		},
		"multiply": func(a, b int) int {
			return a * b
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
				SELECT 
					ts.id, 
					ts.plan_id, 
					p.name, 
					ts.description, 
					ts.date,
					wt.name as workout_type,
					cs.hfmax
				FROM training_sessions ts 
				JOIN training_plans p ON ts.plan_id = p.id
				JOIN workout_types wt ON p.workout_type_id = wt.id
				LEFT JOIN cycling_sessions cs ON ts.id = cs.session_id
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
				err := rows.Scan(
					&session.ID, 
					&session.PlanID, 
					&session.PlanName, 
					&session.Description, 
					&session.Date,
					&session.WorkoutType,
					&session.HFMax,
				)
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

		// Get the first day of the current month
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		// Get the first day to display (might be from previous month)
		firstDisplayDay := firstOfMonth.AddDate(0, 0, -int(firstOfMonth.Weekday()))
		if firstOfMonth.Weekday() == 0 { // If month starts on Sunday
			firstDisplayDay = firstDisplayDay.AddDate(0, 0, -6)
		}

		// Create slice for up to 42 days (6 weeks)
		monthDays := make([]MonthDay, 42)

		// Get all sessions for the displayed date range
		monthSessions, err := db.Query(`
			SELECT p.name, wt.name, ts.date 
			FROM training_sessions ts 
			JOIN training_plans p ON ts.plan_id = p.id
			JOIN workout_types wt ON p.workout_type_id = wt.id
			WHERE ts.date >= ? AND ts.date <= ?
			ORDER BY ts.date
		`, firstDisplayDay.Format("2006-01-02"), firstDisplayDay.AddDate(0, 0, 41).Format("2006-01-02"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer monthSessions.Close()

		// Create a map to store sessions by date
		sessionsByDate := make(map[string][]MonthSession)
		for monthSessions.Next() {
			var session MonthSession
			var date time.Time
			err := monthSessions.Scan(&session.PlanName, &session.WorkoutType, &date)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dateKey := date.Format("2006-01-02")
			sessionsByDate[dateKey] = append(sessionsByDate[dateKey], session)
		}

		// Fill in the month days
		for i := 0; i < 42; i++ {
			currentDate := firstDisplayDay.AddDate(0, 0, i)
			dateKey := currentDate.Format("2006-01-02")
			
			monthDays[i] = MonthDay{
				Date:          currentDate,
				IsCurrentMonth: currentDate.Month() == now.Month(),
				Sessions:      sessionsByDate[dateKey],
			}
		}

		// Add month data to the calendar data
		data.MonthData = MonthData{
			Days:  monthDays,
			Month: now.Month(),
			Year:  now.Year(),
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
