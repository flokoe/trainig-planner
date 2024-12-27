package models

type TrainingSession struct {
	ID           int64  `json:"id"`
	PlanID       int64  `json:"plan_id"`
	SessionOrder int    `json:"session_order"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Duration     int    `json:"duration"`
}
