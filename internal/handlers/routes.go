package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Plans handlers
	mux.HandleFunc("/plans", handleListPlans(db))
	mux.HandleFunc("/plans/create", handleCreatePlan(db))
	mux.HandleFunc("/plans/", handleViewPlan(db))
	
	// Sessions handlers
	mux.HandleFunc("/sessions/new/", handleCreateSession(db))
}
