package usecase

import (
	"context"
	"fmt"

	"github.com/yuma25/Uni-Steps/domain"
)

// TaskUsecase は課題管理に関するビジネスロジックを担当する構造体である．
type TaskUsecase struct {
	taskRepo  domain.TaskRepository // 課題データの永続化を担うリポジトリである．
	aiService domain.AIService      // AI による解析や生成を担うサービスである．
}

// NewTaskUsecase は TaskUsecase の新しいインスタンスを生成する．
func NewTaskUsecase(tr domain.TaskRepository, ai domain.AIService) *TaskUsecase {
	return &TaskUsecase{
		taskRepo:  tr,
		aiService: ai,
	}
}

// RegisterTask はユーザーの生テキストを受け取り，AI 解析を経て課題を登録するユースケースである．
func (uc *TaskUsecase) RegisterTask(ctx context.Context, userID string, groupID string, rawText string) (*domain.Task, error) {
	// 1. AI を呼び出してテキストを解析する．
	// ここで「何をするか」というロジックに集中し，「どう解析するか（Geminiの呼び出し方等）」は AI Service に任せる．
	task, err := uc.aiService.AnalyzeTask(ctx, rawText)
	if err != nil {
		return nil, fmt.Errorf("AI 解析に失敗した： %w", err)
	}

	// 2. 解析結果にユーザーIDとグループIDを紐付ける．
	task.UserID = userID
	task.GroupID = groupID
	task.RawText = rawText

	// 3. データベースに保存する．
	// 「どう保存するか（SQL等）」は Repository に任せる．
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("タスクの保存に失敗した： %w", err)
	}

	return task, nil
}

// ListGroupTasks は特定のグループの課題一覧を取得するユースケースである．
func (uc *TaskUsecase) ListGroupTasks(ctx context.Context, groupID string) ([]*domain.Task, error) {
	return uc.taskRepo.FindByGroupID(ctx, groupID)
}
