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
	aiService    domain.AIService           // リマインド時の「煽り」や「励まし」のメッセージを生成するための AI サービスである．
	notifService domain.NotificationService // 実際にメッセージを送信するための通知サービスである．
}

// NewMonitorUsecase は MonitorUsecase の新しいインスタンスを生成する．
func NewMonitorUsecase(tr domain.TaskRepository, ai domain.AIService, ns domain.NotificationService) *MonitorUsecase {
	return &MonitorUsecase{
		taskRepo:     tr,
		aiService:    ai,
		notifService: ns,
	}
}

// StartMonitoring は無限ループでデータベースを監視する処理を開始する．
// 実際の運用では，main 関数から Goroutine（go uc.StartMonitoring()）として呼び出される．
func (uc *MonitorUsecase) StartMonitoring(ctx context.Context) {
	log.Println("監視プロセス（Goroutine）が起動した．")

	// 5分おきに実行するタイマー（Ticker）を設定する．
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// コンテキストがキャンセルされた（アプリ終了時など）場合はループを抜ける．
			log.Println("監視プロセスを終了する．")
			return
		case t := <-ticker.C:
			// タイマーが発火するたびに実行される処理．
			log.Printf("[%s] 定期監視を実行中...\n", t.Format(time.RFC3339))

			// 1. 今から 24 時間以内に期限を迎える，未完了のタスクを検索する．
			targetTime := time.Now().Add(24 * time.Hour)
			tasks, err := uc.taskRepo.FindApproachingDeadlines(ctx, targetTime)
			if err != nil {
				log.Printf("エラー: 監視中のタスク取得に失敗した: %v\n", err)
				continue // エラーが起きても監視は続ける
			}

			// 2. 見つかったタスクに対して通知処理を行う．
			for _, task := range tasks {
				// AI に「煽り」メッセージを生成させる（今回は少し厳しめのスタイルを指定）．
				msg, err := uc.aiService.GenerateRemindMessage(ctx, task, "厳しい")
				if err != nil {
					log.Printf("エラー: AI メッセージ生成に失敗した: %v\n", err)
					msg = fmt.Sprintf("【リマインド】課題「%s」の期限が %s に迫っているぞ！", task.Title, task.Deadline.Format("1月2日 15時04分"))
				}

				// 通知を送信する（現在はモック的にログ出力のみ，実際の targetID 取得ロジックは別途必要）
				// 本来は task.GroupID に紐づく LINE グループ ID を Group リポジトリから取得して送信する．
				log.Printf(">>> 送信予定メッセージ: %s\n", msg)

				// LINE への実際の送信処理（※ group_id が実際の LINE の ID ではない場合があるため，要変換）
				// err = uc.notifService.SendGroupMessage(ctx, task.GroupID, msg)
			}
		}
	}
}
