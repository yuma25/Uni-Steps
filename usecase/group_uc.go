package usecase

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("DEBUG: ユーザー %s によるグループ作成（名前: %s）を開始する．\n", ownerID, name)

	// 1．オーナーとなるユーザーが存在するか確認する．
	user, err := uc.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		log.Printf("ERROR: オーナー検索中にエラーが発生した: %v\n", err)
		return nil, fmt.Errorf("オーナー情報の取得に失敗した： %w", err)
	}
	if user == nil {
		log.Printf("ERROR: オーナーが見つからない (ID: %s)\n", ownerID)
		return nil, fmt.Errorf("オーナーとなるユーザーが見つからない")
	}

	// 2．新しいグループ構造体を作成する．
	// 招待コードは UUID の先頭 8 文字を簡易的に使用する．
	inviteCode := uuid.New().String()[:8]

	group := &domain.Group{
		ID:              uuid.New().String(),
		Name:            name,
		OwnerID:         ownerID,
		InviteCode:      inviteCode,
		RemindIntervals: []int{1440, 60}, // デフォルト設定：24時間前と1時間前
		AICharacter:     domain.AICharacterDefault,
		Users:           []*domain.User{},
	}

	// 3．まずデータベースに「部屋」だけを保存する．
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループの保存に失敗した： %w", err)
	}

	// 4．保存された部屋に対して，オーナー（ユーザー）を所属させる．
	group.Users = append(group.Users, user)
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("メンバーの紐付けに失敗した： %w", err)
	}

	return group, nil
}

// JoinGroupByInviteCode は招待コードを用いて既存のグループに参加する．
func (uc *GroupUsecase) JoinGroupByInviteCode(ctx context.Context, code string, userID string) (*domain.Group, error) {
	// 1．招待コードに該当するグループを探す．
	group, err := uc.groupRepo.FindByInviteCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("招待コードの検証に失敗した： %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("無効な招待コードである")
	}

	// 2．参加するユーザーが存在するか確認する．
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("ユーザーが見つからない")
	}

	// 3．既に参加していないか確認する．
	for _, u := range group.Users {
		if u.ID == userID {
			return group, nil // 既に参加済みの場合はそのまま成功とする
		}
	}

	// 4．ユーザーをグループに追加して保存する．
	group.Users = append(group.Users, user)
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループへの参加に失敗した： %w", err)
	}

	return group, nil
}

// ListUserGroups は指定されたユーザーが所属しているグループの一覧を取得する．
func (uc *GroupUsecase) ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	groups, err := uc.groupRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("所属グループ一覧の取得に失敗した： %w", err)
	}
	return groups, nil
}

// UpdateSettings は部屋の設定を更新する．オーナーのみ許可する．
func (uc *GroupUsecase) UpdateSettings(ctx context.Context, groupID string, userID string, intervals []int, aiCharacter string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	if group.OwnerID != userID {
		return fmt.Errorf("この部屋の設定を変更する権限がない（オーナーのみ可能）")
	}

	group.RemindIntervals = intervals
	group.AICharacter = aiCharacter
	return uc.groupRepo.Save(ctx, group)
}
