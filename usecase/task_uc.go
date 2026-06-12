package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// TaskUsecase は課題管理に関するビジネスロジックを担当する構造体である．
type TaskUsecase struct {
	taskRepo  domain.TaskRepository   // 課題データの永続化を担うリポジトリである．
	groupRepo domain.GroupRepository  // グループ設定を取得するためのリポジトリである．
	aiService domain.AIService        // AI による文章生成（リマインド等）を担うサービスである．
	scheduler domain.SchedulerService // リマインド予約を管理するサービスである．
}

// NewTaskUsecase は TaskUsecase の新しいインスタンスを生成する．
func NewTaskUsecase(tr domain.TaskRepository, gr domain.GroupRepository, ai domain.AIService, sch domain.SchedulerService) *TaskUsecase {
	return &TaskUsecase{
		taskRepo:  tr,
		groupRepo: gr,
		aiService: ai,
		scheduler: sch,
	}
}

// RegisterManualTask は UI から直接入力された情報に基づいて課題を登録するユースケースである．
func (uc *TaskUsecase) RegisterManualTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	task.Source = domain.SourceManual
	if task.Title == "" {
		return nil, fmt.Errorf("タイトルは必須である")
	}
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// 新規作成時，各ユーザーの進捗データにも親の課題 ID を確実にセットする．
	for _, up := range task.UserProgress {
		up.TaskID = task.ID
		if up.UpdatedAt.IsZero() {
			up.UpdatedAt = time.Now()
		}
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("手動タスクの保存に失敗した： %w", err)
	}

	// 部屋の設定を取得してリマインドを予約する．
	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)
	if group != nil && !task.Deadline.IsZero() {
		for _, up := range task.UserProgress {
			if !up.IsCompleted {
				for _, interval := range group.RemindIntervals {
					_ = uc.scheduler.ScheduleTaskRemind(ctx, task, up.UserID, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
				}
			}
		}
	}
	return task, nil
}

// ListGroupTasks は特定のグループの課題一覧を取得するユースケースである．
func (uc *TaskUsecase) ListGroupTasks(ctx context.Context, groupID string) ([]*domain.Task, error) {
	return uc.taskRepo.FindByGroupID(ctx, groupID)
}

// UpdateTask は課題の情報を更新するユースケースである．
func (uc *TaskUsecase) UpdateTask(ctx context.Context, taskID string, input *domain.Task) (*domain.Task, error) {
	existing, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("更新対象の課題が見つからない")
	}

	existing.Title = input.Title
	existing.Deadline = input.Deadline

	// 部屋の設定を取得
	group, _ := uc.groupRepo.FindByID(ctx, existing.GroupID)

	// 進捗状況（該当者）の更新
	if input.UserProgress != nil {
		inputUserIDs := make(map[string]*domain.TaskUserProgress)
		for _, up := range input.UserProgress {
			inputUserIDs[up.UserID] = up
		}
		newProgress := []*domain.TaskUserProgress{}
		for _, ep := range existing.UserProgress {
			if _, ok := inputUserIDs[ep.UserID]; ok {
				newProgress = append(newProgress, ep)
				delete(inputUserIDs, ep.UserID)
			} else {
				// 該当者から外れた場合は全ての予約をキャンセル
				_ = uc.scheduler.CancelTaskReminds(ctx, taskID, ep.UserID)
			}
		}
		for _, up := range inputUserIDs {
			up.TaskID = taskID
			if up.UpdatedAt.IsZero() {
				up.UpdatedAt = time.Now()
			}
			newProgress = append(newProgress, up)
		}
		existing.UserProgress = newProgress
	}

	if err := uc.taskRepo.Save(ctx, existing); err != nil {
		return nil, err
	}

	// リマインドの再予約
	if group != nil && !existing.Deadline.IsZero() {
		for _, up := range existing.UserProgress {
			if !up.IsCompleted {
				for _, interval := range group.RemindIntervals {
					_ = uc.scheduler.ScheduleTaskRemind(ctx, existing, up.UserID, interval, group.AICharacter, existing.Deadline.Add(-time.Duration(interval)*time.Minute))
				}
			} else {
				_ = uc.scheduler.CancelTaskReminds(ctx, taskID, up.UserID)
			}
		}
	} else {
		for _, up := range existing.UserProgress {
			_ = uc.scheduler.CancelTaskReminds(ctx, taskID, up.UserID)
		}
	}

	return existing, nil
}

// ToggleUserCompletion は特定のユーザーの課題完了状態を反転させる．
func (uc *TaskUsecase) ToggleUserCompletion(ctx context.Context, taskID, userID string) error {
	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("課題が見つからない")
	}

	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)

	found := false
	for i, up := range task.UserProgress {
		if up.UserID == userID {
			task.UserProgress[i].IsCompleted = !task.UserProgress[i].IsCompleted
			task.UserProgress[i].UpdatedAt = time.Now()

			if task.UserProgress[i].IsCompleted {
				_ = uc.scheduler.CancelTaskReminds(ctx, taskID, userID)
			} else if group != nil && !task.Deadline.IsZero() {
				for _, interval := range group.RemindIntervals {
					_ = uc.scheduler.ScheduleTaskRemind(ctx, task, userID, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
				}
			}
			found = true
			break
		}
	}

	if !found {
		task.UserProgress = append(task.UserProgress, &domain.TaskUserProgress{
			TaskID:      taskID,
			UserID:      userID,
			IsCompleted: true,
			UpdatedAt:   time.Now(),
		})
		_ = uc.scheduler.CancelTaskReminds(ctx, taskID, userID)
	}

	return uc.taskRepo.Save(ctx, task)
}
