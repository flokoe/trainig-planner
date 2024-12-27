package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Plans handlers
	mux.HandleFunc("/", listPlansHandler(db))
	mux.HandleFunc("/plans/new", newPlanHandler(db))
	mux.HandleFunc("/plans/create", createPlanHandler(db))
	mux.HandleFunc("/plans/view/", viewPlanHandler(db))
	
	// Sessions handlers
	mux.HandleFunc("/sessions/new/", newSessionHandler(db))
	mux.HandleFunc("/sessions/create", createSessionHandler(db))
	mux.HandleFunc("/sessions/edit/", editSessionHandler(db))
}
