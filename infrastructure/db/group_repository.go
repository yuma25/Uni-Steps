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
	// GORM の Save は全てのフィールドを更新するが，many-to-many などの扱いで意図しない挙動を防ぐため，
	// 明示的に Updates を使用した手法に切り替える（Users は除外）．
	err := r.db.WithContext(ctx).Model(group).Omit("Users").Updates(group).Error
	if err != nil {
		// レコードが存在しない場合は Create を試みる（新規作成時の互換性維持）．
		if errors.Is(err, gorm.ErrRecordNotFound) || group.ID == "" {
			return r.db.WithContext(ctx).Create(group).Error
		}
		return err
	}
	return nil
}

func (r *groupRepository) FindAllGroups(ctx context.Context) ([]*domain.Group, error) {
	var groups []*domain.Group
	err := r.db.WithContext(ctx).Preload("Users").Find(&groups).Error
	return groups, err
}

// FindByID は指定された ID のグループ情報を取得する．
func (r *groupRepository) FindByID(ctx context.Context, id string) (*domain.Group, error) {
	var group domain.Group
	err := r.db.WithContext(ctx).Preload("Users").Where("id = ?", id).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// FindByInviteCode は招待コードでグループを検索する．
func (r *groupRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Group, error) {
	var group domain.Group
	err := r.db.WithContext(ctx).Preload("Users").Where("invite_code = ?", code).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

// FindByUserID は指定されたユーザー ID が所属しているグループ一覧を取得する．
func (r *groupRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Group, error) {
	groups := []*domain.Group{}
	// 中間テーブル user_groups を結合して，特定のユーザー ID に紐づくグループを検索する．
	// Preload("Users") を追加してメンバー情報を取得する．
	err := r.db.WithContext(ctx).
		Preload("Users").
		Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// RemoveUser は部屋からユーザーの紐付けを削除する．
func (r *groupRepository) RemoveUser(ctx context.Context, groupID string, userID string) error {
	group := &domain.Group{ID: groupID}
	user := &domain.User{ID: userID}
	// 中間テーブル（many-to-many）から該当の組み合わせを削除する．
	return r.db.WithContext(ctx).Model(group).Association("Users").Delete(user)
}

// Delete はグループとその紐付けを完全に削除する．
func (r *groupRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1．中間テーブルの紐付けを削除する．
		if err := tx.Model(&domain.Group{ID: id}).Association("Users").Clear(); err != nil {
			return err
		}
		// 2．グループ本体を削除する．
		return tx.Where("id = ?", id).Delete(&domain.Group{}).Error
	})
}
