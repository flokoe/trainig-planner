package models

import (
	"time"
)

// TrainingPlan represents a complete training program
type TrainingPlan struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

// TrainingSession represents a planned workout
type TrainingSession struct {
	ID             int64     `db:"id"`
	PlanID         int64     `db:"plan_id"`
	ScheduledDate  time.Time `db:"scheduled_date"`
	Type        string    `db:"type"` // e.g., endurance, intervals, recovery
	Description string    `db:"description"`
	Intensity   int       `db:"intensity"` // percentage of HFmax
}

// CompletedSession represents an actual completed workout
type CompletedSession struct {
	ID               int64     `db:"id"`
	TrainingSessionID int64     `db:"training_session_id"`
	ActualDate       time.Time `db:"actual_date"`
	ActualDuration   int      `db:"actual_duration"` // in minutes
	ActualDistance   float64  `db:"actual_distance"` // in kilometers
	Notes            string    `db:"notes"`
}
