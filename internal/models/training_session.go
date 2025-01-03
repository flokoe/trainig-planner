package models

import "time"

type TrainingSession struct {
	ID           int64     `json:"id"`
	PlanID       int64     `json:"plan_id"`
	SessionOrder *int      `json:"session_order"`
	Description  string    `json:"description"`
	Date         time.Time `json:"date"`
}

type CyclingSession struct {
	SessionID int64 `json:"session_id"`
	HFMax     string `json:"hfmax"`
}

type MobilitySession struct {
	SessionID int64 `json:"session_id"`
}

type SandbagSession struct {
	SessionID int64 `json:"session_id"`
}

type CoreSession struct {
	SessionID int64 `json:"session_id"`
}
