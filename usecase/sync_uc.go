package usecase

import (
	"context"
	"fmt"

	"github.com/yuma25/Uni-Steps/domain"
)

// SyncUsecase は外部の学習管理システム（LMS）との同期ロジックを担当する構造体である．
type SyncUsecase struct {
	taskRepo   domain.TaskRepository // 課題データの永続化を担うリポジトリである．
	lmsService domain.LMSService     // 外部 LMS と通信するためのサービスである．
}

// NewSyncUsecase は SyncUsecase の新しいインスタンスを生成する．
func NewSyncUsecase(tr domain.TaskRepository, lms domain.LMSService) *SyncUsecase {
	return &SyncUsecase{
		taskRepo:   tr,
		lmsService: lms,
	}
}

// SyncTasks は指定されたグループに対して，外部 LMS から最新の課題を取得し保存する．
// 取得はリクエストを行ったユーザー（userID）の権限（OAuth トークン等）を用いて実行される．
func (uc *SyncUsecase) SyncTasks(ctx context.Context, userID string, groupID string) ([]*domain.Task, error) {
	// 1．外部 LMS から課題の一覧を取得する．
	tasks, err := uc.lmsService.FetchTasks(ctx, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("LMS からの課題取得に失敗した： %w", err)
	}

	var savedTasks []*domain.Task

	// 2．取得した課題を順番にデータベースへ保存する．
	for _, task := range tasks {
		// グループIDと入力ソースを設定する．
		task.GroupID = groupID
		task.Source = uc.lmsService.GetProviderName()

		// 重複チェックなどのロジックが本来は必要であるが，ここでは単純に保存する．
		if err := uc.taskRepo.Save(ctx, task); err != nil {
			// 一つの保存に失敗しても，他の保存は継続するようにエラーをログ出力する等も検討すべきである．
			// ここでは全体の失敗として扱う．
			return nil, fmt.Errorf("タスク（外部ID: %s）の保存に失敗した： %w", task.ExternalID, err)
		}
		savedTasks = append(savedTasks, task)
	}

	return savedTasks, nil
}
