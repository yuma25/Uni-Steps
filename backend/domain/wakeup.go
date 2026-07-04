package domain

import (
	"time"
)

const (
	WakeupStatusPending   = "pending"   // 起床確認待ちである．
	WakeupStatusConfirmed = "confirmed" // 起床が確認された（成功）．
	WakeupStatusAlerted   = "alerted"   // 時間を過ぎても確認できず，アラートを送信済みである（失敗）．
)

// WakeupCheck は特定の時間の起床を監視するためのエンティティである．
type WakeupCheck struct {
	ID           string    `json:"id" gorm:"primaryKey"` // 一意識別子である．
	UserID       string    `json:"user_id"`              // 対象となるユーザーの ID である．
	GroupID      string    `json:"group_id"`             // 起きなかった場合に通知を飛ばすグループの ID である．
	TargetTime   time.Time `json:"target_time"`          // 起床予定時刻である．
	GraceMinutes int       `json:"grace_minutes"`        // 予定時刻から何分まで待つかの猶予期間である．
	Status       string    `json:"status"`               // 現在の状態（pending, confirmed, alerted）である．
	CreatedAt    time.Time `json:"created_at"`           // 作成日時である．
}
