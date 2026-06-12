package db

import (
	"context"
	"errors"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// taskRepository は domain.TaskRepository インターフェースを実装する構造体である．
// GORM を使用して PostgreSQL (Supabase) との通信を行う．
type taskRepository struct {
	db *gorm.DB // データベース接続を保持する GORM クライアントである．
}

// NewTaskRepository は taskRepository の新しいインスタンスを生成する．
// 引数として確立済みのデータベース接続（gorm.DB）を受け取る．
func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &taskRepository{
		db: db,
	}
}

// Save はタスクをデータベースに保存（新規作成または更新）する．
func (r *taskRepository) Save(ctx context.Context, task *domain.Task) error {
	// 1. まず Task 本体のみを保存・更新する（GORM の自動保存による外部キー制約エラーを防ぐため Omit する）
	if err := r.db.WithContext(ctx).Omit("UserProgress").Save(task).Error; err != nil {
		return err
	}

	// 2. UserProgress（該当者リスト）を完全に同期する．
	if err := r.db.WithContext(ctx).Model(task).Association("UserProgress").Replace(task.UserProgress); err != nil {
		return err
	}

	return nil
}

// FindByID は指定された ID のタスクをデータベースから取得する．
func (r *taskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	err := r.db.WithContext(ctx).Preload("UserProgress").Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// FindByExternalID は外部 LMS の ID をキーにしてタスクを取得する．
func (r *taskRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Task, error) {
	var task domain.Task
	err := r.db.WithContext(ctx).Preload("UserProgress").Where("external_id = ?", externalID).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// FindByGroupID は指定されたグループ ID に紐づくタスク一覧を取得する．
func (r *taskRepository) FindByGroupID(ctx context.Context, groupID string) ([]*domain.Task, error) {
	tasks := []*domain.Task{}
	// Preload で進捗状況も取得し，期限順に並べる．
	err := r.db.WithContext(ctx).
		Preload("UserProgress").
		Where("group_id = ?", groupID).
		Order("deadline ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindApproachingDeadlines は指定された日時までに期限を迎える，未完了のタスクを取得する．
func (r *taskRepository) FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*domain.Task, error) {
	tasks := []*domain.Task{}
	// 全員の完了状態ではなく，個別の通知ロジックが必要になるため，ここでは Preload しつつ取得する．
	err := r.db.WithContext(ctx).
		Preload("UserProgress").
		Where("deadline <= ?", until).
		Order("deadline ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
