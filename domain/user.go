package domain

import (
	"time"
)

type User struct {
	ID            string    `json:"id"`               // ユーザーの一意識別子である．
	Name          string    `json:"name"`             // ユーザーの表示名である．
	WebPushToken  string    `json:"web_push_token"`   // ブラウザ通知用のトークンである．
	LastCheckInAt time.Time `json:"last_check_in_at"` // 最終起床確認時刻である．
}
