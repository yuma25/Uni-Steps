package db

import (
	"context"
	"errors"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

type wakeupRepository struct {
	db *gorm.DB
}

// NewWakeupRepository は WakeupRepository の新しいインスタンスを生成する．
func NewWakeupRepository(db *gorm.DB) domain.WakeupRepository {
	return &wakeupRepository{
		db: db,
	}
}

func (r *wakeupRepository) Save(ctx context.Context, check *domain.WakeupCheck) error {
	return r.db.WithContext(ctx).Save(check).Error
}

func (r *wakeupRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.WakeupCheck{}).Error
}

func (r *wakeupRepository) FindByID(ctx context.Context, id string) (*domain.WakeupCheck, error) {
	var check domain.WakeupCheck
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&check).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &check, nil
}

func (r *wakeupRepository) FindPendingByTime(ctx context.Context, now time.Time) ([]*domain.WakeupCheck, error) {
	checks := []*domain.WakeupCheck{}
	// ターゲット時刻 + 猶予期間 を過ぎても pending のものを取得する．
	err := r.db.WithContext(ctx).
		Where("status = ? AND target_time <= ?", domain.WakeupStatusPending, now).
		Find(&checks).Error
	return checks, err
}

func (r *wakeupRepository) FindActiveByUser(ctx context.Context, userID string) ([]*domain.WakeupCheck, error) {
	checks := []*domain.WakeupCheck{}
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.WakeupStatusPending).
		Find(&checks).Error
	return checks, err
}

func (r *wakeupRepository) FindActiveByGroup(ctx context.Context, groupID string) ([]*domain.WakeupCheck, error) {
	checks := []*domain.WakeupCheck{}
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND status = ?", groupID, domain.WakeupStatusPending).
		Find(&checks).Error
	return checks, err
}
