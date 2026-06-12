package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// InMemScheduler はメモリ上でタイマーを管理する SchedulerService の実装である．
type InMemScheduler struct {
	userRepo     domain.UserRepository
	groupRepo    domain.GroupRepository
	aiService    domain.AIService
	notifService domain.NotificationService
	timers       map[string]*time.Timer // key: "task:taskID:userID:interval" or "wakeup:wakeupID"
	mu           sync.Mutex
}

// NewInMemScheduler は InMemScheduler の新しいインスタンスを生成する．
func NewInMemScheduler(ur domain.UserRepository, gr domain.GroupRepository, ai domain.AIService, ns domain.NotificationService) *InMemScheduler {
	return &InMemScheduler{
		userRepo:     ur,
		groupRepo:    gr,
		aiService:    ai,
		notifService: ns,
		timers:       make(map[string]*time.Timer),
	}
}

// ScheduleTaskRemind は指定された時刻にリマインドを実行するタイマーをセットする．
func (s *InMemScheduler) ScheduleTaskRemind(ctx context.Context, task *domain.Task, userID string, intervalMinutes int, style string, runAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("task:%s:%s:%d", task.ID, userID, intervalMinutes)

	if oldTimer, ok := s.timers[key]; ok {
		oldTimer.Stop()
	}

	delay := time.Until(runAt)
	if delay <= 0 {
		log.Printf("[Scheduler] 期限が過去のため予約をスキップ: %s (at %v)\n", task.Title, runAt)
		return nil
	}

	log.Printf("[Scheduler] リマインド予約完了: %s (in %v)\n", task.Title, delay)

	timer := time.AfterFunc(delay, func() {
		runCtx := context.Background()
		msg, err := s.aiService.GenerateRemindMessage(runCtx, task, style)
		if err != nil {
			msg = fmt.Sprintf("【リマインド】課題「%s」の期限まであと %d 分です！", task.Title, intervalMinutes)
		}

		targetURL := fmt.Sprintf("/dashboard?user_id=%s&group_id=%s", userID, task.GroupID)
		_ = s.notifService.SendDirectMessage(runCtx, userID, msg, targetURL)

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

	prefix := fmt.Sprintf("task:%s:%s:", taskID, userID)
	count := 0
	for key, timer := range s.timers {
		if strings.HasPrefix(key, prefix) {
			timer.Stop()
			delete(s.timers, key)
			count++
		}
	}

	if count > 0 {
		log.Printf("[Scheduler] リマインドキャンセル完了: %s (計 %d 件)\n", taskID, count)
	}
	return nil
}

// ScheduleWakeupSOS は起床確認失敗時の SOS 通知を予約する．
func (s *InMemScheduler) ScheduleWakeupSOS(ctx context.Context, wakeupID string, userID string, groupID string, runAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := "wakeup:" + wakeupID

	if oldTimer, ok := s.timers[key]; ok {
		oldTimer.Stop()
	}

	delay := time.Until(runAt)
	if delay <= 0 {
		log.Printf("[Scheduler] 起床時刻が過去のため SOS 予約をスキップ (at %v)\n", runAt)
		return nil
	}

	log.Printf("[Scheduler] SOS 通知を予約: ID=%s (in %v)\n", wakeupID, delay)

	timer := time.AfterFunc(delay, func() {
		runCtx := context.Background()

		// 1．対象ユーザーとグループの情報を取得する．
		user, _ := s.userRepo.FindByID(runCtx, userID)
		group, _ := s.groupRepo.FindByID(runCtx, groupID)
		if user == nil || group == nil {
			log.Printf("[Scheduler] SOS 実行失敗: ユーザーまたはグループが見つからない\n")
			return
		}

		// 2．緊急メッセージを作成する．
		alertMsg := fmt.Sprintf("【緊急】%s さんが起床予定時刻を過ぎてもチェックインしていません！誰か連絡を取ってください！", user.Name)

		// 3．Web Push で個人宛に通知を飛ばす（本人を含む）．
		sentCount := 0
		for _, member := range group.Users {
			targetURL := fmt.Sprintf("/dashboard?user_id=%s&group_id=%s", member.ID, group.ID)
			_ = s.notifService.SendDirectMessage(runCtx, member.ID, alertMsg, targetURL)
			sentCount++
		}

		// 4．設定されていれば，LINE グループにも一斉送信する．
		_ = s.notifService.SendGroupMessage(runCtx, group.ID, alertMsg)

		log.Printf("[Scheduler] SOS 発信完了: 対象=%s, 送信先=%d 人（Web Push）\n", user.Name, sentCount)

		s.mu.Lock()
		delete(s.timers, key)
		s.mu.Unlock()
	})

	s.timers[key] = timer
	return nil
}

// CancelWakeupSOS は予約済みの SOS 通知を取り消す．
func (s *InMemScheduler) CancelWakeupSOS(ctx context.Context, wakeupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := "wakeup:" + wakeupID
	if timer, ok := s.timers[key]; ok {
		timer.Stop()
		delete(s.timers, key)
		log.Printf("[Scheduler] SOS 通知予約をキャンセル: ID=%s\n", wakeupID)
	}

	return nil
}
