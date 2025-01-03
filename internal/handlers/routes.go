package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Session completion handler
	mux.HandleFunc("/complete-session/", handleCompleteSession(db))
	
	// Plans handlers
	mux.HandleFunc("/plans", handleListPlans(db))
	mux.HandleFunc("/plans/create", handleCreatePlan(db))
	mux.HandleFunc("/plans/", handleViewPlan(db))
	
	// Sessions handlers
	mux.HandleFunc("/sessions/create/", handleCreateSession(db))
	
	// Calendar handler
	mux.HandleFunc("/", handleCalendar(db))
}
