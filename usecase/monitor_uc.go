package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// MonitorUsecase は定期的にデータベースを監視し，リマインドや生存確認を行うロジックである．
type MonitorUsecase struct {
	taskRepo     domain.TaskRepository      // 課題データを検索・更新するためのリポジトリである．
	userRepo     domain.UserRepository      // ユーザー情報を検索・更新するためのリポジトリである．
	groupRepo    domain.GroupRepository     // グループの所属メンバーを確認するためのリポジトリである．
	wakeupRepo   domain.WakeupRepository    // 起床確認状態を確認するためのリポジトリである．
	aiService    domain.AIService           // リマインド時のメッセージを生成するための AI サービスである．
	notifService domain.NotificationService // メッセージを送信するための通知サービスである．
}

// NewMonitorUsecase は MonitorUsecase の新しいインスタンスを生成する．
func NewMonitorUsecase(tr domain.TaskRepository, ur domain.UserRepository, gr domain.GroupRepository, wr domain.WakeupRepository, ai domain.AIService, ns domain.NotificationService) *MonitorUsecase {
	return &MonitorUsecase{
		taskRepo:     tr,
		userRepo:     ur,
		groupRepo:    gr,
		wakeupRepo:   wr,
		aiService:    ai,
		notifService: ns,
	}
}

func (uc *MonitorUsecase) checkApproachingTasks(ctx context.Context) {
	targetTime := time.Now().Add(24 * time.Hour)
	tasks, err := uc.taskRepo.FindApproachingDeadlines(ctx, targetTime)
	if err != nil {
		log.Printf("エラー: タスク取得に失敗: %v\n", err)
		return
	}

	for _, task := range tasks {
		msg, err := uc.aiService.GenerateRemindMessage(ctx, task, "厳しい")
		if err != nil {
			msg = fmt.Sprintf("【リマインド】課題「%s」の期限が %s に迫っているぞ！", task.Title, task.Deadline.Format("1月2日 15時04分"))
		}
		log.Printf(">>> 送信予定リマインド: %s\n", msg)
		// LINE グループへの送信等は将来的に実装する．
	}
}

func (uc *MonitorUsecase) checkWakeupStatuses(ctx context.Context) {
	now := time.Now()
	// 猶予期間を過ぎても pending の起床確認を検索する．
	pendingChecks, err := uc.wakeupRepo.FindPendingByTime(ctx, now)
	if err != nil {
		log.Printf("エラー: 起床確認データの取得に失敗: %v\n", err)
		return
	}

	for _, check := range pendingChecks {
		// 1．対象ユーザーの名前を取得する．
		user, _ := uc.userRepo.FindByID(ctx, check.UserID)
		userName := "あるメンバー"
		if user != nil {
			userName = user.Name
		}

		// 2．対象グループの全メンバーを取得する．
		group, _ := uc.groupRepo.FindByID(ctx, check.GroupID)
		if group == nil {
			continue
		}

		// 3．AI に「〇〇さんがまだ起きていない」という緊急メッセージを生成させる．
		alertMsg := fmt.Sprintf("【緊急】%s さんが起床予定時刻を過ぎてもチェックインしていません！誰か連絡を取ってください！", userName)
		// 本来は AI に生成させるが，ここでは固定メッセージとする．

		// 4．グループメンバー全員に Web Push で通知を飛ばす．
		for _, member := range group.Users {
			if member.ID == check.UserID {
				continue // 本人には送らない
			}
			_ = uc.notifService.SendDirectMessage(ctx, member.ID, alertMsg)
		}

		// 5．状態を alerted に更新する．
		check.Status = domain.WakeupStatusAlerted
		_ = uc.wakeupRepo.Save(ctx, check)
		log.Printf(">>> 生存確認アラートをグループ %s に送信した（対象: %s）\n", group.Name, userName)
	}
}

// StartMonitoring は無限ループでデータベースを監視する処理を開始する．
func (uc *MonitorUsecase) StartMonitoring(ctx context.Context) {
	log.Println("監視プロセス（Goroutine）が起動した．")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("監視プロセスを終了する．")
			return
		case t := <-ticker.C:
			log.Printf("[%s] 定期監視を実行中...\n", t.Format(time.RFC3339))

			// A. 期限間近の課題リマインド（既存ロジック）
			uc.checkApproachingTasks(ctx)

			// B. 起床確認のチェック（新規ロジック）
			uc.checkWakeupStatuses(ctx)
		}
	}
}
