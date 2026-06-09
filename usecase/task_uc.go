package usecase

import (
	"context"
	"fmt"

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

	// 3．データベースに保存する．
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("手動タスクの保存に失敗した： %w", err)
	}

	return task, nil
}

// ListGroupTasks は特定のグループの課題一覧を取得するユースケースである．
func (uc *TaskUsecase) ListGroupTasks(ctx context.Context, groupID string) ([]*domain.Task, error) {
	return uc.taskRepo.FindByGroupID(ctx, groupID)
}
