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
	groupRepo domain.GroupRepository           // グループデータの永続化を担うリポジトリである．
	userRepo  domain.UserRepository            // ユーザー情報を確認するためのリポジトリである．
	logRepo   domain.NotificationLogRepository // 通知履歴を取得するためのリポジトリである．
}

// NewGroupUsecase は GroupUsecase の新しいインスタンスを生成する．
func NewGroupUsecase(gr domain.GroupRepository, ur domain.UserRepository, lr domain.NotificationLogRepository) *GroupUsecase {
	return &GroupUsecase{
		groupRepo: gr,
		userRepo:  ur,
		logRepo:   lr,
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
		ID:                 uuid.New().String(),
		Name:               name,
		OwnerID:            ownerID,
		InviteCode:         inviteCode,
		RemindIntervals:    []int{1440, 180, 60}, // デフォルト設定：24時間前と3時間前と1時間前
		AICharacter:        domain.AICharacterDefault,
		SummaryMorningTime: "06:00",
		SummaryEveningTime: "22:00",
		Users:              []*domain.User{},
	}

	// 3．まずデータベースに「部屋」だけを保存する．
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return nil, fmt.Errorf("グループの保存に失敗した： %w", err)
	}

	// 4．保存された部屋に対して，オーナー（ユーザー）を所属させる．
	if err := uc.groupRepo.AddUserToGroup(ctx, group.ID, user.ID); err != nil {
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
	if err := uc.groupRepo.AddUserToGroup(ctx, group.ID, user.ID); err != nil {
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
func (uc *GroupUsecase) UpdateSettings(ctx context.Context, groupID string, userID string, name string, intervals []int, aiCharacter string, lineToken string, lineGroupID string, morningTime string, eveningTime string) error {
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

	if name != "" {
		group.Name = name
	}
	group.RemindIntervals = intervals
	group.AICharacter = aiCharacter
	group.LineChannelToken = lineToken
	group.LineGroupID = lineGroupID
	group.SummaryMorningTime = morningTime
	group.SummaryEveningTime = eveningTime
	return uc.groupRepo.Save(ctx, group)
}

// LeaveGroup はユーザーを部屋から退出させる．
func (uc *GroupUsecase) LeaveGroup(ctx context.Context, groupID string, userID string) error {
	// 1．部屋が存在するか確認する．
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// オーナーが退出する場合は，事前に権譲譲が必要であることを伝える（Usecase レベルではエラーを返す）．
	if group.OwnerID == userID && len(group.Users) > 1 {
		return fmt.Errorf("オーナーは退出前に次のオーナーを指名する必要がある")
	}

	// 2．リポジトリ経由で紐付けを削除する．
	return uc.groupRepo.RemoveUser(ctx, groupID, userID)
}

// TransferOwnership は部屋のオーナー権限を別のユーザーに譲渡する．
func (uc *GroupUsecase) TransferOwnership(ctx context.Context, groupID string, currentOwnerID string, newOwnerID string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// 現在のオーナーであるか確認する．
	if group.OwnerID != currentOwnerID {
		return fmt.Errorf("オーナー権限の譲渡は現在のオーナーのみ可能である")
	}

	// 新しいオーナーがグループに所属しているか確認する．
	found := false
	for _, u := range group.Users {
		if u.ID == newOwnerID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("新しいオーナーはグループのメンバーである必要がある")
	}

	group.OwnerID = newOwnerID
	return uc.groupRepo.Save(ctx, group)
}

// DeleteGroup は部屋を完全に削除する．オーナーのみ許可する．
func (uc *GroupUsecase) DeleteGroup(ctx context.Context, groupID string, userID string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	if group.OwnerID != userID {
		return fmt.Errorf("この部屋を削除する権限がない（オーナーのみ可能）")
	}

	return uc.groupRepo.Delete(ctx, groupID)
}

// GetNotificationLogs は指定されたグループの通知履歴を取得する．
func (uc *GroupUsecase) GetNotificationLogs(ctx context.Context, groupID string, limit int) ([]*domain.NotificationLog, error) {
	return uc.logRepo.FindByGroupID(ctx, groupID, limit)
}
