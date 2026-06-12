package domain

import (
	"context"
	"time"
)

const (
	NotificationTypeRemind = "remind" // 通常の課題リマインド
	NotificationTypeSOS    = "sos"    // 起床見守り失敗時の SOS
)

// NotificationLog は送信された通知の履歴を保持するエンティティである．
type NotificationLog struct {
	ID        string    `json:"id" gorm:"primaryKey"`  // ログの一意識別子である．
	GroupID   string    `json:"group_id" gorm:"index"` // どの部屋で発生した通知か．
	UserID    string    `json:"user_id" gorm:"index"`  // 誰宛ての（または誰が原因の）通知か．
	Type      string    `json:"type"`                  // 通知の種類（remind, sos）である．
	Message   string    `json:"message"`               // AI が生成したメッセージの内容である．
	CreatedAt time.Time `json:"created_at"`            // 送信日時である．
}

// NotificationLogRepository は通知履歴の保存・取得に関する約束事である．
type NotificationLogRepository interface {
	Save(ctx context.Context, log *NotificationLog) error
	FindByGroupID(ctx context.Context, groupID string, limit int) ([]*NotificationLog, error)
}
