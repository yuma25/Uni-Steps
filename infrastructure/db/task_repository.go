package db

import (
	"context"
	"errors"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// taskRepository は domain.TaskRepository インターフェースを実装する構造体である．
type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &taskRepository{db: db}
}

// Save はタスクを保存または更新し，その進捗状況（UserProgress）を完全に同期する．
func (r *taskRepository) Save(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Task 本体を保存（UserProgress は除外）．
		if err := tx.Omit("UserProgress").Save(task).Error; err != nil {
			return err
		}

		// 2. UserProgress（該当者リスト）を同期する．
		if err := tx.Where("task_id = ?", task.ID).Delete(&domain.TaskUserProgress{}).Error; err != nil {
			return err
		}

		if len(task.UserProgress) > 0 {
			// 親 Task の ID を確実に各 UserProgress レコードにセットする．
			for i := range task.UserProgress {
				task.UserProgress[i].TaskID = task.ID
			}

			if err := tx.Create(task.UserProgress).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

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

func (r *taskRepository) FindByGroupID(ctx context.Context, groupID string) ([]*domain.Task, error) {
	var tasks []*domain.Task
	err := r.db.WithContext(ctx).Preload("UserProgress").Where("group_id = ?", groupID).Order("deadline ASC").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*domain.Task, error) {
	var tasks []*domain.Task
	err := r.db.WithContext(ctx).Preload("UserProgress").Where("deadline <= ?", until).Order("deadline ASC").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", id).Delete(&domain.TaskUserProgress{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&domain.Task{}).Error
	})
}
