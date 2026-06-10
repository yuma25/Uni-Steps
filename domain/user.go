package domain

import (
	"time"
)

type User struct {
	ID                 string    `json:"id" gorm:"primaryKey"`                 // ユーザーの一意識別子（UUID 等）である．
	Name               string    `json:"name"`                                 // ユーザーの表示名である．
	Email              string    `json:"email" gorm:"unique"`                  // ユーザーのメールアドレスである（Google ログインのキー）．
	WebPushToken       string    `json:"web_push_token"`                       // ブラウザ通知用のトークンである．
	GoogleAccessToken  string    `json:"google_access_token"`                  // Google Classroom 連携用の OAuth アクセストークンである．
	GoogleRefreshToken string    `json:"google_refresh_token"`                 // Google Classroom 連携用の OAuth リフレッシュトークンである．
	LastCheckInAt      time.Time `json:"last_check_in_at"`                     // 最終起床確認時刻である．
	Groups             []*Group  `json:"groups" gorm:"many2many:user_groups;"` // ユーザーが所属しているグループのリストである．
}
