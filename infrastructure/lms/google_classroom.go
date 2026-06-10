package lms

import (
	"context"
	"fmt"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"golang.org/x/oauth2"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

// GoogleClassroomService は Google Classroom API と連携する LMSService の実装である．
type GoogleClassroomService struct {
	userRepo domain.UserRepository // ユーザーの OAuth トークンを取得するためのリポジトリである．
	oauthCfg *oauth2.Config        // Google OAuth 2.0 の設定情報である．
}

// NewGoogleClassroomService は GoogleClassroomService の新しいインスタンスを生成する．
func NewGoogleClassroomService(ur domain.UserRepository, cfg *oauth2.Config) *GoogleClassroomService {
	return &GoogleClassroomService{
		userRepo: ur,
		oauthCfg: cfg,
	}
}

// FetchTasks は Google Classroom から指定されたコース（グループ）の課題一覧を取得する．
// userID に紐づく OAuth トークンを用いて認証を行う．
func (s *GoogleClassroomService) FetchTasks(ctx context.Context, userID string, courseID string) ([]*domain.Task, error) {
	// 1．データベースからユーザー情報を取得する．
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil || user.GoogleAccessToken == "" {
		return nil, fmt.Errorf("Google 連携が行われていないか，トークンが存在しない")
	}

	// 2．ユーザーのトークン情報から OAuth2 トークンを復元する．
	token := &oauth2.Token{
		AccessToken:  user.GoogleAccessToken,
		RefreshToken: user.GoogleRefreshToken,
		Expiry:       time.Now(), // 簡易的に常に期限切れとし，必要に応じてリフレッシュさせる設定（運用時に調整が必要）
		TokenType:    "Bearer",
	}

	// 3．トークンを用いて Classroom サービスを生成する．
	client := s.oauthCfg.Client(ctx, token)
	srv, err := classroom.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Classroom サービスの作成に失敗した： %w", err)
	}

	// 4．Classroom API を呼び出して，指定されたコースの課題 (CourseWork) を取得する．
	resp, err := srv.Courses.CourseWork.List(courseID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("コースワークの取得に失敗した： %w", err)
	}

	var tasks []*domain.Task

	// 5．取得した CourseWork をドメインモデル의 Task に変換する．
	for _, cw := range resp.CourseWork {
		var deadline time.Time
		if cw.DueDate != nil && cw.DueTime != nil {
			deadline = time.Date(int(cw.DueDate.Year), time.Month(cw.DueDate.Month), int(cw.DueDate.Day),
				int(cw.DueTime.Hours), int(cw.DueTime.Minutes), 0, 0, time.UTC)
		}

		// Classroom API の UpdateTime (RFC3339形式の文字列) を time.Time に変換する．
		lmsUpdateTime, _ := time.Parse(time.RFC3339, cw.UpdateTime)

		task := &domain.Task{
			Title:         cw.Title,
			ExternalID:    cw.Id, // Classroom 側の課題 ID
			Deadline:      deadline,
			LMSUpdateTime: lmsUpdateTime,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetProviderName はこのサービスが Google Classroom であることを返す．
func (s *GoogleClassroomService) GetProviderName() string {
	return domain.SourceGoogleClassroom
}
