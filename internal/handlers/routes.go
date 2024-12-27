package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Plans handlers
	mux.HandleFunc("/plans/create", handleCreatePlan(db))
	mux.HandleFunc("/plans/", handleViewPlan(db))
}
