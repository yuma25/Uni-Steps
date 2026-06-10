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
		return nil, fmt.Errorf("オーナーとなるユーザーが見つからない（ログイン状態を確認してほしい）")
	}

	// 2．新しいグループ構造体を作成する．
	group := &domain.Group{
		ID:      uuid.New().String(),
		Name:    name,
		OwnerID: ownerID,
		// 最初は空のリストにして，部屋の保存を優先する．
		Users: []*domain.User{},
	}

	// 3．まずデータベースに「部屋」だけを保存する．
	log.Println("DEBUG: データベースへのグループ保存（1段階目）を実行する．")
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		log.Printf("ERROR: グループ自体の保存に失敗した: %v\n", err)
		return nil, fmt.Errorf("グループの保存に失敗した： %w", err)
	}

	// 4．保存された部屋に対して，オーナー（ユーザー）を所属させる．
	log.Println("DEBUG: グループへのメンバー紐付け（2段階目）を実行する．")
	group.Users = append(group.Users, user)
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		log.Printf("ERROR: メンバーの紐付けに失敗した: %v\n", err)
		return nil, fmt.Errorf("メンバーの紐付けに失敗した： %w", err)
	}

	log.Printf("DEBUG: グループ作成とメンバー登録に成功した (ID: %s)\n", group.ID)
	return group, nil
}

// FetchAvailableLMSCourses は外部 LMS からユーザーが利用可能なコース一覧を取得する．
func (uc *GroupUsecase) FetchAvailableLMSCourses(ctx context.Context, userID string, lmsService domain.LMSService) ([]*domain.Group, error) {
	log.Printf("DEBUG: ユーザー %s の利用可能コース取得を開始する．\n", userID)
	groups, err := lmsService.FetchCourses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("LMS からのコース取得に失敗した： %w", err)
	}
	return groups, nil
}

// SyncLMSGroups は外部 LMS（Google Classroom 等）からコース一覧を取得し，
// 必要に応じてデータベースに保存するユースケースである．
func (uc *GroupUsecase) SyncLMSGroups(ctx context.Context, userID string, lmsService domain.LMSService) ([]*domain.Group, error) {
	log.Printf("DEBUG: ユーザー %s の LMS コース同期を開始する．\n", userID)

	// 1．外部 LMS からコース一覧を取得する．
	groups, err := lmsService.FetchCourses(ctx, userID)
	if err != nil {
		log.Printf("ERROR: LMS からのコース取得に失敗した: %v\n", err)
		return nil, fmt.Errorf("LMS からのコース取得に失敗した： %w", err)
	}
	log.Printf("DEBUG: LMS から %d 件のコースを取得した．\n", len(groups))

	// 2．取得したコースをデータベースに保存する．
	for _, g := range groups {
		if err := uc.groupRepo.Save(ctx, g); err != nil {
			log.Printf("WARNING: グループ %s の保存に失敗した: %v\n", g.Name, err)
		}
	}

	return groups, nil
}

// LinkLMSCourse は特定の部屋に対して外部 LMS のコース ID を紐付ける．
func (uc *GroupUsecase) LinkLMSCourse(ctx context.Context, groupId string, lmsCourseId string) error {
	// 1．対象のグループを取得する．
	group, err := uc.groupRepo.FindByID(ctx, groupId)
	if err != nil {
		return fmt.Errorf("グループ情報の取得に失敗した： %w", err)
	}
	if group == nil {
		return fmt.Errorf("指定されたグループが見つからない")
	}

	// 2．コース ID を設定して保存する．
	group.LMSCourseID = lmsCourseId
	if err := uc.groupRepo.Save(ctx, group); err != nil {
		return fmt.Errorf("LMS コースの紐付けに失敗した： %w", err)
	}

	return nil
}

// ListUserGroups は指定されたユーザーが所属しているグループの一覧を取得する．
func (uc *GroupUsecase) ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	groups, err := uc.groupRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("所属グループ一覧の取得に失敗した： %w", err)
	}
	return groups, nil
}
