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
	taskRepo  domain.TaskRepository // 課題データを検索・更新するためのリポジトリである．
	aiService domain.AIService      // リマインド時の「煽り」や「励まし」のメッセージを生成するための AI サービスである．
}

// NewMonitorUsecase は MonitorUsecase の新しいインスタンスを生成する．
func NewMonitorUsecase(tr domain.TaskRepository, ai domain.AIService) *MonitorUsecase {
	return &MonitorUsecase{
		taskRepo:  tr,
		aiService: ai,
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

			// 本来はここで「期限が近いタスク」や「未起床のユーザー」をデータベースから検索する．
			// TODO: taskRepo に期限を指定して検索するメソッドを追加し，通知処理（LINE/Web Push）を実装する．

			// (デモ用ダミー処理)
			fmt.Println(" -> データベースのチェックが完了した．異常なし．")
		}
	}
}
