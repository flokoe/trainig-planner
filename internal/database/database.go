package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Schema contains the SQL statements to create the database tables
const Schema = `
CREATE TABLE IF NOT EXISTS training_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS training_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_id INTEGER NOT NULL,
    scheduled_date DATETIME NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    intensity INTEGER DEFAULT 0,
    FOREIGN KEY (plan_id) REFERENCES training_plans(id)
);

CREATE TABLE IF NOT EXISTS completed_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    training_session_id INTEGER NOT NULL,
    actual_date DATETIME NOT NULL,
    actual_duration INTEGER NOT NULL,
    actual_distance REAL,
    notes TEXT,
    FOREIGN KEY (training_session_id) REFERENCES training_sessions(id)
);`

// InitDB initializes the database connection and creates tables if they don't exist
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables
	_, err = db.Exec(Schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}
