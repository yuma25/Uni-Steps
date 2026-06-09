package domain

import (
	"time"
)

type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	WebPushToken  string    `json:"web_push_token"`
	LastCheckInAt time.Time `json:"last_check_in_at"`
}
