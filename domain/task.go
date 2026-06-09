package domain

import (
	"time"
)

type Task struct {
	ID          string    `json:"id"`           // 課題の一意識別子である．
	GroupID     string    `json:"group_id"`     // 所属するグループの ID である．
	UserID      string    `json:"user_id"`      // 担当するユーザーの ID である．
	RawText     string    `json:"raw_text"`     // ユーザーが入力した生の文章である．
	Title       string    `json:"title"`        // AI が解析した課題のタイトルである．
	Deadline    time.Time `json:"deadline"`     // 課題の期限である．
	IsCompleted bool      `json:"is_completed"` // 完了したかどうかのフラグである．
	IsCritical  bool      `json:"is_critical"`  // 起床確認が必要な重要課題かどうかのフラグである．
}
