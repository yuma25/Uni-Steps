package domain

import (
	"context"
	"time"
)

// ReminderJob は予約されたリマインド通知を表すデータ構造である．
type ReminderJob struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	TaskID     string    `json:"task_id"`
	UserID     string    `json:"user_id"`
	TargetTime time.Time `json:"target_time"`
	Message    string    `json:"message"`
	Status     string    `json:"status"` // pending, sent, cancelled
}

// NotificationService はユーザーやグループへの通知を担うインターフェースである．
type NotificationService interface {
	SendGroupMessage(ctx context.Context, targetID string, message string) error
	SendDirectMessage(ctx context.Context, userID string, message string) error
}

// SchedulerService は未来の時刻に特定の処理を予約するためのインターフェースである．
type SchedulerService interface {
	// ScheduleTaskRemind は特定の課題に対して，指定されたタイミング（何分前か）でリマインドを予約する．
	ScheduleTaskRemind(ctx context.Context, task *Task, userID string, intervalMinutes int, runAt time.Time) error

	// CancelTaskReminds は予約済みのすべてのリマインドを取り消す．
	CancelTaskReminds(ctx context.Context, taskID string, userID string) error
}
