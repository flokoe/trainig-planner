package models

import "time"

type TrainingPlan struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	WorkoutTypeID int64     `json:"workout_type_id"`
	CreatedAt     time.Time `json:"created_at"`
}
