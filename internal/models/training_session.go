package models

import "time"

type TrainingSession struct {
	ID           int64     `json:"id"`
	PlanID       int64     `json:"plan_id"`
	SessionOrder int       `json:"session_order"`
	Description  string    `json:"description"`
	Date         time.Time `json:"date"`
}
