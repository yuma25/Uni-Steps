package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
// ここでは，ユーザーに関連するすべての有効なコースから課題を取得するよう変更されている．
func (uc *SyncUsecase) SyncTasks(ctx context.Context, userID string, groupID string) ([]*domain.Task, error) {
	// 1．グループ情報を取得する．
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループ情報の取得に失敗した： %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("指定されたグループが見つからない")
	}

	// 2．外部 LMS からすべての課題一覧を取得する．
	tasks, err := uc.lmsService.FetchTasks(ctx, userID)
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

	// 4．データベースへ保存し，同期状態を更新する．
	savedTasks := []*domain.Task{}
	for _, task := range tasks {
		task.GroupID = groupID
		task.Source = uc.lmsService.GetProviderName()

		// すでに同じ外部 ID のタスクが存在するか確認する．
		existing, err := uc.taskRepo.FindByExternalID(ctx, task.ExternalID)
		if err != nil {
			return nil, fmt.Errorf("既存タスクの確認に失敗した： %w", err)
		}

		if existing != nil {
			// 既存の場合は，ID を引き継いで「更新」扱いにする．
			task.ID = existing.ID

			// LMS の元々の期限設定有無の状態を反映する
			existing.IsLMSDeadlineSet = task.IsLMSDeadlineSet

			// 期限の上書き防止ロジック：
			// Uni-Steps 側ですでに期限が設定されており，LMS 側が未定（Zero）の場合は，既存の期限を優先する．
			if task.Deadline.IsZero() && !existing.Deadline.IsZero() {
				task.Deadline = existing.Deadline
			}

			// 進捗状況の統合：今回のユーザーの情報を更新または追加し，他人の情報は維持する．
			if len(task.UserProgress) > 0 {
				newProgress := task.UserProgress[0] // FetchTasks は常に1人分（自分）を返す
				found := false
				for i, ep := range existing.UserProgress {
					if ep.UserID == newProgress.UserID {
						existing.UserProgress[i].IsCompleted = newProgress.IsCompleted
						existing.UserProgress[i].UpdatedAt = time.Now()
						found = true
						break
					}
				}
				if !found {
					existing.UserProgress = append(existing.UserProgress, newProgress)
				}
				task.UserProgress = existing.UserProgress
			}
		} else {
			// 新規の場合は，新しい UUID を発行する．
			task.ID = uuid.New().String()
		}

		// 保存（既存の場合は更新）
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
