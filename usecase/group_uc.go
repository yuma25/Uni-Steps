package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// GroupUsecase はグループ（部屋）の作成や参加に関するビジネスロジックを担当する構造体である．
type GroupUsecase struct {
	groupRepo domain.GroupRepository // グループデータの永続化を担うリポジトリである．
	userRepo  domain.UserRepository  // ユーザー情報を確認するためのリポジトリである．
}

// NewGroupUsecase は GroupUsecase の新しいインスタンスを生成する．
func NewGroupUsecase(gr domain.GroupRepository, ur domain.UserRepository) *GroupUsecase {
	return &GroupUsecase{
		groupRepo: gr,
		userRepo:  ur,
	}
}

// CreateGroup は新しいグループを作成し，作成者をオーナーとして登録する．
func (uc *GroupUsecase) CreateGroup(ctx context.Context, name string, ownerID string) (*domain.Group, error) {
	// 1．オーナーとなるユーザーが存在するか確認する．
	user, err := uc.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("オーナー情報の取得に失敗した： %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("オーナーとなるユーザーが見つからない")
	}

	// 2．新しいグループ構造体を作成する．
	group := &domain.Group{
		ID:      uuid.New().String(),
		Name:    name,
		OwnerID: ownerID,
		Users:   []*domain.User{user}, // 作成者を最初のメンバーとして追加する．
	}

	// 3．データベースに保存する．
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループの保存に失敗した： %w", err)
	}

	return group, nil
}

// ListUserGroups は指定されたユーザーが所属しているグループの一覧を取得する．
func (uc *GroupUsecase) ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	// 本来は中間テーブル user_groups をクエリして，そのユーザーが所属する全グループを返す．
	// 現時点ではプロトタイプ用として，将来的にリポジトリに FindByUserID を追加して実装する．
	// 暫定的に，全てのグループを返す（要修正）．
	// TODO: repository.go に FindByUserID(userID string) ([]*Group, error) を追加すること．
	return nil, fmt.Errorf("ListUserGroups は未実装である")
}
