package db

import (
	"context"
	"errors"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// groupRepository は domain.GroupRepository インターフェースを実装する構造体である．
type groupRepository struct {
	db *gorm.DB // データベース接続を保持する GORM クライアントである．
}

// NewGroupRepository は groupRepository の新しいインスタンスを生成する．
func NewGroupRepository(db *gorm.DB) domain.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

// Save はグループ情報を保存または更新する．
func (r *groupRepository) Save(ctx context.Context, group *domain.Group) error {
	if err := r.db.WithContext(ctx).Save(group).Error; err != nil {
		return err
	}
	return nil
}

// FindByID は指定された ID のグループ情報を取得する．
func (r *groupRepository) FindByID(ctx context.Context, id string) (*domain.Group, error) {
	var group domain.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}
