package db

import (
	"context"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// notificationLogRepository は domain.NotificationLogRepository を実装する構造体である．
type notificationLogRepository struct {
	db *gorm.DB
}

// NewNotificationLogRepository は新しいインスタンスを生成する．
func NewNotificationLogRepository(db *gorm.DB) domain.NotificationLogRepository {
	return &notificationLogRepository{
		db: db,
	}
}

// Save は新しい通知ログを保存する．
func (r *notificationLogRepository) Save(ctx context.Context, log *domain.NotificationLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// FindByGroupID は指定された部屋の過去の通知ログを新しい順に取得する．
func (r *notificationLogRepository) FindByGroupID(ctx context.Context, groupID string, limit int) ([]*domain.NotificationLog, error) {
	var logs []*domain.NotificationLog
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
