package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// InMemScheduler はメモリ上でタイマーを管理する SchedulerService の実装である．
type InMemScheduler struct {
	taskRepo     domain.TaskRepository
	userRepo     domain.UserRepository
	groupRepo    domain.GroupRepository
	aiService    domain.AIService
	notifService domain.NotificationService
	logRepo      domain.NotificationLogRepository
	timers       map[string]*time.Timer // key: "task:taskID:interval" or "wakeup:wakeupID"
	mu           sync.Mutex
}

// NewInMemScheduler は InMemScheduler の新しいインスタンスを生成する．
func NewInMemScheduler(tr domain.TaskRepository, ur domain.UserRepository, gr domain.GroupRepository, ai domain.AIService, ns domain.NotificationService, lr domain.NotificationLogRepository) *InMemScheduler {
	return &InMemScheduler{
		taskRepo:     tr,
		userRepo:     ur,
		groupRepo:    gr,
		aiService:    ai,
		notifService: ns,
		logRepo:      lr,
		timers:       make(map[string]*time.Timer),
	}
}

// ScheduleTaskRemind は指定された時刻にリマインドを実行するタイマーをセットする．
func (s *InMemScheduler) ScheduleTaskRemind(ctx context.Context, task *domain.Task, intervalMinutes int, style string, runAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("task:%s:%d", task.ID, intervalMinutes)

	if oldTimer, ok := s.timers[key]; ok {
		oldTimer.Stop()
	}

	delay := time.Until(runAt)
	if delay <= 0 {
		log.Printf("[Scheduler] 期限が過去のため予約をスキップ: %s (at %v)\n", task.Title, runAt)
		return nil
	}

	log.Printf("[Scheduler] リマインド予約完了: %s (in %v, interval %d)\n", task.Title, delay, intervalMinutes)

	timer := time.AfterFunc(delay, func() {
		runCtx := context.Background()

		// 最新のタスク情報を取得
		currentTask, err := s.taskRepo.FindByID(runCtx, task.ID)
		if err != nil || currentTask == nil {
			log.Printf("[Scheduler] タスクが存在しないか取得に失敗したためリマインドをスキップ: ID=%s\n", task.ID)
			s.mu.Lock()
			delete(s.timers, key)
			s.mu.Unlock()
			return
		}

		// 未完了のユーザーを抽出
		var pendingUsers []string
		for _, up := range currentTask.UserProgress {
			if !up.IsCompleted {
				pendingUsers = append(pendingUsers, up.UserID)
			}
		}

		// 未完了ユーザーがいない場合は何もしない
		if len(pendingUsers) == 0 {
			log.Printf("[Scheduler] すべてのユーザーが完了しているためリマインドをスキップ: %s\n", currentTask.Title)
			s.mu.Lock()
			delete(s.timers, key)
			s.mu.Unlock()
			return
		}

		// AIメッセージを1回だけ生成
		msg, err := s.aiService.GenerateRemindMessage(runCtx, currentTask, style)
		if err != nil {
			msg = fmt.Sprintf("【リマインド】課題「%s」の期限まであと %d 分です！", currentTask.Title, intervalMinutes)
		}

		// 未完了の全ユーザーへ同じメッセージを送信
		for _, uID := range pendingUsers {
			targetURL := fmt.Sprintf("/dashboard?user_id=%s&group_id=%s", uID, currentTask.GroupID)
			_ = s.notifService.SendDirectMessage(runCtx, uID, msg, targetURL)

			// 履歴を保存する
			_ = s.logRepo.Save(runCtx, &domain.NotificationLog{
				ID:        uuid.New().String(),
				GroupID:   currentTask.GroupID,
				UserID:    uID,
				Type:      domain.NotificationTypeRemind,
				Message:   msg,
				CreatedAt: time.Now(),
			})
		}

		log.Printf("[Scheduler] リマインド送信完了: %s (送信先 %d 人)\n", currentTask.Title, len(pendingUsers))

		s.mu.Lock()
		delete(s.timers, key)
		s.mu.Unlock()
	})

	s.timers[key] = timer
	return nil
}

// CancelTaskReminds は特定の課題に紐づくすべてのタイマーを停止する．
func (s *InMemScheduler) CancelTaskReminds(ctx context.Context, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := fmt.Sprintf("task:%s:", taskID)
	count := 0
	for key, timer := range s.timers {
		if strings.HasPrefix(key, prefix) {
			timer.Stop()
			delete(s.timers, key)
			count++
		}
	}

	if count > 0 {
		log.Printf("[Scheduler] リマインドキャンセル完了: TaskID=%s (計 %d 件)\n", taskID, count)
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

		// 履歴を保存する
		_ = s.logRepo.Save(runCtx, &domain.NotificationLog{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			UserID:    userID, // 対象となった（寝坊した）ユーザーのID
			Type:      domain.NotificationTypeSOS,
			Message:   alertMsg,
			CreatedAt: time.Now(),
		})

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
