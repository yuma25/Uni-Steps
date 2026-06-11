package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// InMemScheduler はメモリ上でタイマーを管理する SchedulerService の実装である．
type InMemScheduler struct {
	aiService    domain.AIService
	notifService domain.NotificationService
	timers       map[string]*time.Timer // key: "taskID:userID:interval"
	mu           sync.Mutex
}

// NewInMemScheduler は InMemScheduler の新しいインスタンスを生成する．
func NewInMemScheduler(ai domain.AIService, ns domain.NotificationService) *InMemScheduler {
	return &InMemScheduler{
		aiService:    ai,
		notifService: ns,
		timers:       make(map[string]*time.Timer),
	}
}

// ScheduleTaskRemind は指定された時刻にリマインドを実行するタイマーをセットする．
func (s *InMemScheduler) ScheduleTaskRemind(ctx context.Context, task *domain.Task, userID string, intervalMinutes int, runAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%d", task.ID, userID, intervalMinutes)

	// すでに同じ予約があれば上書き（再予約）する．
	if oldTimer, ok := s.timers[key]; ok {
		oldTimer.Stop()
	}

	delay := time.Until(runAt)
	if delay <= 0 {
		return nil // すでに過ぎている場合は何もしない
	}

	timer := time.AfterFunc(delay, func() {
		runCtx := context.Background()

		// AI メッセージの生成
		style := "厳しい"
		if intervalMinutes > 180 {
			style = "ふつう"
		}

		msg, err := s.aiService.GenerateRemindMessage(runCtx, task, style)
		if err != nil {
			msg = fmt.Sprintf("【リマインド】課題「%s」の期限まであと %d 分です！", task.Title, intervalMinutes)
		}

		_ = s.notifService.SendDirectMessage(runCtx, userID, msg)

		s.mu.Lock()
		delete(s.timers, key)
		s.mu.Unlock()
	})

	s.timers[key] = timer
	return nil
}

// CancelTaskReminds は特定の課題・ユーザーに紐づくすべてのタイマーを停止する．
func (s *InMemScheduler) CancelTaskReminds(ctx context.Context, taskID string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := taskID + ":" + userID + ":"
	for key, timer := range s.timers {
		if strings.HasPrefix(key, prefix) {
			timer.Stop()
			delete(s.timers, key)
		}
	}

	return nil
}
