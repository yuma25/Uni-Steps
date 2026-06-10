package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// SyncUsecase は外部の学習管理システム（LMS）との同期ロジックを担当する構造体である．
type SyncUsecase struct {
	taskRepo   domain.TaskRepository  // 課題データの永続化を担うリポジトリである．
	groupRepo  domain.GroupRepository // グループ情報の取得・更新を担うリポジトリである．
	lmsService domain.LMSService      // 外部 LMS と通信するためのサービスである．
}

// NewSyncUsecase は SyncUsecase の新しいインスタンスを生成する．
func NewSyncUsecase(tr domain.TaskRepository, gr domain.GroupRepository, lms domain.LMSService) *SyncUsecase {
	return &SyncUsecase{
		taskRepo:   tr,
		groupRepo:  gr,
		lmsService: lms,
	}
}

// SyncTasks は指定されたグループに対して，外部 LMS から最新の課題を取得し保存する．
// 取得はリクエストを行ったユーザー（userID）の権限（OAuth トークン等）を用いて実行される．
func (uc *SyncUsecase) SyncTasks(ctx context.Context, userID string, groupID string) ([]*domain.Task, error) {
	// 1．グループ情報を取得し，クールダウン期間（5分）をチェックする．
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループ情報の取得に失敗した： %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("指定されたグループが見つからない")
	}

	// 前回の同期実行から 5 分経過しているか確認する．
	if time.Since(group.LastSyncedAt) < 5*time.Minute {
		return nil, fmt.Errorf("前回の同期から時間が経過していない（5分間は再試行不可）")
	}

	// 2．外部 LMS から課題の一覧を取得する．
	tasks, err := uc.lmsService.FetchTasks(ctx, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("LMS からの課題取得に失敗した： %w", err)
	}

	// 3．差分検知：取得した課題の中で最新の更新時刻を確認する．
	var maxLMSUpdate time.Time
	for _, t := range tasks {
		if t.LMSUpdateTime.After(maxLMSUpdate) {
			maxLMSUpdate = t.LMSUpdateTime
		}
	}

	// Google 側で情報更新がない場合は，DB への書き込みを行わず終了する．
	if !maxLMSUpdate.After(group.LMSLastUpdatedAt) {
		// 同期処理自体は「成功」として，更新がなかった旨を記録して返す．
		group.LastSyncedAt = time.Now()
		_ = uc.groupRepo.Save(ctx, group)
		return nil, nil // 空のリストを返すことで「更新なし」を表現する
	}

	var savedTasks []*domain.Task

	// 4．更新があった場合のみ，データベースへ保存し，同期状態を更新する．
	for _, task := range tasks {
		task.GroupID = groupID
		task.Source = uc.lmsService.GetProviderName()

		if err := uc.taskRepo.Save(ctx, task); err != nil {
			return nil, fmt.Errorf("タスク（外部ID: %s）の保存に失敗した： %w", task.ExternalID, err)
		}
		savedTasks = append(savedTasks, task)
	}

	// 最終同期時刻と最終更新検知時刻を更新する．
	group.LastSyncedAt = time.Now()
	group.LMSLastUpdatedAt = maxLMSUpdate
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループの同期状態の更新に失敗した： %w", err)
	}

	return savedTasks, nil
}
