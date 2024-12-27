package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	tables := `
	CREATE TABLE IF NOT EXISTS workout_types (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS training_plans (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		workout_type_id INTEGER,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workout_type_id) REFERENCES workout_types(id)
	);

	CREATE TABLE IF NOT EXISTS training_sessions (
		id INTEGER PRIMARY KEY,
		plan_id INTEGER,
		session_order INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		duration INTEGER,
		FOREIGN KEY (plan_id) REFERENCES training_plans(id)
	);

	-- Insert default workout types if they don't exist
	INSERT OR IGNORE INTO workout_types (name) VALUES 
		('cycling'),
		('mobility'),
		('sandbag');
	`

	_, err := db.Exec(tables)
	return err
}
