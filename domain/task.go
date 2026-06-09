package domain

import (
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	UserID      string    `json:"user_id"`
	RawText     string    `json:"raw_text"`
	Title       string    `json:"title"`
	Deadline    time.Time `json:"deadline"`
	IsCompleted bool      `json:"is_completed"`
	IsCritical  bool      `json:"is_critical"`
}
