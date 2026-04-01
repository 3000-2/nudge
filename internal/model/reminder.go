package model

import "time"

const (
	StatusActive    = "active"
	StatusCompleted = "completed"

	ScheduleTypeOnce    = "once"
	ScheduleTypeDaily   = "daily"
	ScheduleTypeWeekly  = "weekly"
	ScheduleTypeMonthly = "monthly"
)

type Reminder struct {
	ID        string     `json:"id"`
	Message   string     `json:"message"`
	Schedule  Schedule   `json:"schedule"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	FiredAt   *time.Time `json:"fired_at,omitempty"`
	PlistPath string     `json:"plist_path"`
}

type Schedule struct {
	Type       string  `json:"type"`
	Hour       int     `json:"hour"`
	Minute     int     `json:"minute"`
	Date       *string `json:"date,omitempty"`
	Weekday    *int    `json:"weekday,omitempty"`
	DayOfMonth *int    `json:"day_of_month,omitempty"`
}

type StorageFile struct {
	Version   int        `json:"version"`
	Reminders []Reminder `json:"reminders"`
}
