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

	var maxLMSUpdate time.Time
	for _, t := range tasks {
		if t.LMSUpdateTime.After(maxLMSUpdate) {
			maxLMSUpdate = t.LMSUpdateTime
		}
	}

	savedTasks := []*domain.Task{}
	for _, task := range tasks {
		task.GroupID = groupID
		task.Source = uc.lmsService.GetProviderName()

		existing, _ := uc.taskRepo.FindByExternalID(ctx, task.ExternalID)
		if existing != nil {
			task.ID = existing.ID
			existing.IsLMSDeadlineSet = task.IsLMSDeadlineSet
			if task.Deadline.IsZero() && !existing.Deadline.IsZero() {
				task.Deadline = existing.Deadline
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
