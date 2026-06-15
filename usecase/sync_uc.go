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
	taskRepo   domain.TaskRepository   // 課題データの永続化を担うリポジトリである．
	groupRepo  domain.GroupRepository  // グループ情報の取得・更新を担うリポジトリである．
	lmsService domain.LMSService       // 外部 LMS と通信するためのサービスである．
	scheduler  domain.SchedulerService // リマインド予約を管理するサービスである．
}

// NewSyncUsecase は SyncUsecase の新しいインスタンスを生成する．
func NewSyncUsecase(tr domain.TaskRepository, gr domain.GroupRepository, lms domain.LMSService, sch domain.SchedulerService) *SyncUsecase {
	return &SyncUsecase{
		taskRepo:   tr,
		groupRepo:  gr,
		lmsService: lms,
		scheduler:  sch,
	}
}

// SyncTasks は指定されたグループに対して，外部 LMS から最新の課題を取得し保存する．
func (uc *SyncUsecase) SyncTasks(ctx context.Context, userID string, groupID string) ([]*domain.Task, error) {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("グループ情報の取得に失敗した： %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("指定されたグループが見つからない")
	}

	tasks, err := uc.lmsService.FetchTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("LMS からの課題取得に失敗した： %w", err)
	}

	// N+1 問題の解消：グループ内の既存課題を一括取得してマップ化する．
	existingTasks, err := uc.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("既存課題の取得に失敗した： %w", err)
	}
	existingMap := make(map[string]*domain.Task)
	for _, t := range existingTasks {
		if t.ExternalID != "" {
			existingMap[t.ExternalID] = t
		}
	}

	var maxLMSUpdate time.Time
	for _, t := range tasks {
		if t.LMSUpdateTime.After(maxLMSUpdate) {
			maxLMSUpdate = t.LMSUpdateTime
		}
	}

	savedTasks := []*domain.Task{}
	processedExternalIDs := make(map[string]bool) // 同期セッション内での重複排除用

	for _, task := range tasks {
		if task.ExternalID != "" {
			if processedExternalIDs[task.ExternalID] {
				continue // 今回のループで既に処理済みの外部 ID はスキップする
			}
			processedExternalIDs[task.ExternalID] = true
		}

		task.GroupID = groupID
		task.Source = uc.lmsService.GetProviderName()

		// マップから既存の課題を検索（DB への問い合わせをループ内で行わない）．
		existing := existingMap[task.ExternalID]
		if existing != nil {
			task.ID = existing.ID

			// 既存の課題がローカルで（LMS以外で）設定され、かつゼロ値ではない期限を持つ場合、その期限を優先する。
			// これにより、ユーザーが手動で設定した期限がLMS同期によって上書きされるのを防ぐ。
			if !existing.IsLMSDeadlineSet && !existing.Deadline.IsZero() {
				task.Deadline = existing.Deadline // ローカルで設定された期限を保持
				task.IsLMSDeadlineSet = false     // これはローカルで管理されている期限であることを示す
			} else {
				// それ以外の場合は、LMSが提供する期限を使用する。
				// 'task'オブジェクトには既にLMSからの値が格納されている。
				// LMSが期限を提供している場合は 'IsLMSDeadlineSet' を true に設定し、そうでない場合は false に設定する。
				if !task.Deadline.IsZero() {
					task.IsLMSDeadlineSet = true // LMSが期限を提供した
				} else {
					task.IsLMSDeadlineSet = false // LMSは期限を提供しなかった
				}
			}
			if len(task.UserProgress) > 0 {
				newProgress := task.UserProgress[0]
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
			task.ID = uuid.New().String()
		}

		if err := uc.taskRepo.Save(ctx, task); err != nil {
			return nil, fmt.Errorf("タスク保存失敗： %w", err)
		}

		// 【予約方式】同期した課題のリマインドを予約（部屋の設定に基づく）
		// 期限が未来の場合のみ予約する
		if !task.Deadline.IsZero() && task.Deadline.After(time.Now()) {
			for _, up := range task.UserProgress {
				if !up.IsCompleted {
					for _, interval := range group.RemindIntervals {
						_ = uc.scheduler.ScheduleTaskRemind(ctx, task, up.UserID, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
					}
				} else {
					_ = uc.scheduler.CancelTaskReminds(ctx, task.ID, up.UserID)
				}
			}
		}

		savedTasks = append(savedTasks, task)
	}

	group.LastSyncedAt = time.Now()
	group.LMSLastUpdatedAt = maxLMSUpdate
	_ = uc.groupRepo.Save(ctx, group)

	return savedTasks, nil
}
