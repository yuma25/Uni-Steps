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
	// GORM の Save メソッドは，ID が存在すれば UPDATE，存在しなければ INSERT を行う．
	// WithContext でコンテキストを渡し，タイムアウト等に対応する．
	if err := r.db.WithContext(ctx).Save(task).Error; err != nil {
		return err
	}
	return nil
}

// FindByID は指定された ID のタスクをデータベースから取得する．
func (r *taskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	// First メソッドで 1 件だけ取得する．
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		// 見つからなかった場合は GORM 固有のエラーを返す．
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 見つからない場合はエラーではなく nil を返す設計とする場合もある（要検討）
		}
		return nil, err
	}
	return &task, nil
}

// FindByGroupID は指定されたグループ ID に紐づくタスク一覧を取得する．
func (r *taskRepository) FindByGroupID(ctx context.Context, groupID string) ([]*domain.Task, error) {
	var tasks []*domain.Task
	// Find メソッドで複数件を取得する．
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindApproachingDeadlines は指定された日時までに期限を迎える，未完了のタスクを取得する．
func (r *taskRepository) FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*domain.Task, error) {
	var tasks []*domain.Task
	// "deadline <= ?" で期限が until より前，かつ "is_completed = false" のものを検索する．
	err := r.db.WithContext(ctx).
		Where("deadline <= ? AND is_completed = ?", until, false).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
