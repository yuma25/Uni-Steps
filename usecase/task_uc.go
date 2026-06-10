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
	taskRepo  domain.TaskRepository // 課題データの永続化を担うリポジトリである．
	aiService domain.AIService      // AI による文章生成（リマインド等）を担うサービスである．
}

// NewTaskUsecase は TaskUsecase の新しいインスタンスを生成する．
func NewTaskUsecase(tr domain.TaskRepository, ai domain.AIService) *TaskUsecase {
	return &TaskUsecase{
		taskRepo:  tr,
		aiService: ai,
	}
}

// RegisterManualTask は UI から直接入力された情報に基づいて課題を登録するユースケースである．
func (uc *TaskUsecase) RegisterManualTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	// 1．入力元のソースを手動に設定する．
	task.Source = domain.SourceManual

	// 2．必要最低限のバリデーションを行う（タイトルが空でないか等）．
	if task.Title == "" {
		return nil, fmt.Errorf("タイトルは必須である")
	}

	// 3．新規 ID を発行する（ID が空の場合）．
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// 4．データベースに保存する．
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("手動タスクの保存に失敗した： %w", err)
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

	// Classroom 課題の場合は，タイトルや期限の編集を制限するなどの配慮が必要かもしれないが，
	// 今回は柔軟性を優先して更新を許可する．
	existing.Title = input.Title
	existing.Deadline = input.Deadline
	existing.IsCritical = input.IsCritical

	// 進捗状況（該当者）の更新
	if input.UserProgress != nil {
		// 入力にあるユーザー ID のリストをマップ化する．
		inputUserIDs := make(map[string]*domain.TaskUserProgress)
		for _, up := range input.UserProgress {
			inputUserIDs[up.UserID] = up
		}

		// 既存の進捗状況をフィルタリング・更新する．
		newProgress := []*domain.TaskUserProgress{}
		for _, ep := range existing.UserProgress {
			if _, ok := inputUserIDs[ep.UserID]; ok {
				// 継続して該当者の場合は残す（完了状態は維持するが，名前などは更新される可能性がある）
				newProgress = append(newProgress, ep)
				delete(inputUserIDs, ep.UserID)
			}
		}

		// 新しく追加された該当者を加える．
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

	found := false
	for i, up := range task.UserProgress {
		if up.UserID == userID {
			task.UserProgress[i].IsCompleted = !task.UserProgress[i].IsCompleted
			task.UserProgress[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		// 進捗レコードがない場合は作成する（手動追加された課題などの場合）．
		task.UserProgress = append(task.UserProgress, &domain.TaskUserProgress{
			TaskID:      taskID,
			UserID:      userID,
			IsCompleted: true,
			UpdatedAt:   time.Now(),
		})
	}

	return uc.taskRepo.Save(ctx, task)
}
