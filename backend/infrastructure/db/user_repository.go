package db

import (
	"context"
	"errors"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// userRepository は domain.UserRepository インターフェースを実装する構造体である．
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository は userRepository の新しいインスタンスを生成する．
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Save はユーザーをデータベースに保存する．
func (r *userRepository) Save(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return err
	}
	return nil
}

// UpdateWebPushToken はユーザーの Web Push トークンを明示的に更新する（空文字含む）．
func (r *userRepository) UpdateWebPushToken(ctx context.Context, userID string, token string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("web_push_token", token).Error
}

// FindByID は指定された ID のユーザーをデータベースから取得する．
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail は指定されたメールアドレスのユーザーをデータベースから取得する．
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
