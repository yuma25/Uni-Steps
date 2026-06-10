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

// FetchTasks は Google Classroom からユーザーの全アクティブコースの課題一覧を取得する．
// userID に紐づく OAuth トークンを用いて認証を行う．
func (s *GoogleClassroomService) FetchTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	// 1．データベースからユーザー情報を取得する．
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil || user.GoogleAccessToken == "" {
		return nil, fmt.Errorf("Google 連携が行われていないか，トークンが存在しない")
	}

	// 2．OAuth2 トークンを準備する．
	token := &oauth2.Token{
		AccessToken:  user.GoogleAccessToken,
		RefreshToken: user.GoogleRefreshToken,
		TokenType:    "Bearer",
	}

	// 3．Classroom サービスを生成する．
	client := s.oauthCfg.Client(ctx, token)
	srv, err := classroom.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Classroom サービスの作成に失敗した： %w", err)
	}

	// 4．まずアクティブなコース一覧を取得する．
	coursesResp, err := srv.Courses.List().CourseStates("ACTIVE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("コース一覧の取得に失敗した： %w", err)
	}

	tasks := []*domain.Task{}

	// 5．各コースの課題を取得する．
	for _, course := range coursesResp.Courses {
		resp, err := srv.Courses.CourseWork.List(course.Id).Context(ctx).Do()
		if err != nil {
			continue
		}

		// ユーザーの提出状況を取得して完了フラグを判定する．
		submissionsResp, err := srv.Courses.CourseWork.StudentSubmissions.List(course.Id, "-").Context(ctx).Do()
		submissionStatus := make(map[string]bool)
		if err == nil {
			for _, sub := range submissionsResp.StudentSubmissions {
				// TURNED_IN（提出済み）または RETURNED（返却済み）なら完了とみなす．
				if sub.State == "TURNED_IN" || sub.State == "RETURNED" {
					submissionStatus[sub.CourseWorkId] = true
				}
			}
		}

		for _, cw := range resp.CourseWork {
			var deadline time.Time
			if cw.DueDate != nil {
				// Google Classroom API の DueDate/DueTime は常に UTC として解釈されるべきである．
				hour, min := 23, 59
				if cw.DueTime != nil {
					hour, min = int(cw.DueTime.Hours), int(cw.DueTime.Minutes)
				}

				// UTC で time オブジェクトを作成する．
				deadlineUTC := time.Date(int(cw.DueDate.Year), time.Month(cw.DueDate.Month), int(cw.DueDate.Day),
					hour, min, 0, 0, time.UTC)

				// アプリケーション全体の Local (JST) に変換する．
				deadline = deadlineUTC.Local()
			}

			// LMS 側の更新時刻（RFC3339）を解析し，Local に変換する．
			lmsUpdateTime, _ := time.Parse(time.RFC3339, cw.UpdateTime)
			lmsUpdateTime = lmsUpdateTime.Local()

			task := &domain.Task{
				Title:         cw.Title,
				ExternalID:    cw.Id,
				Deadline:      deadline,
				LMSUpdateTime: lmsUpdateTime,
				UserProgress: []*domain.TaskUserProgress{
					{
						UserID:      userID,
						UserName:    user.Name,
						IsCompleted: submissionStatus[cw.Id],
						UpdatedAt:   time.Now(),
					},
				},
			}
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// FetchCourses は Google Classroom からユーザーが所属しているコース一覧を取得する．
func (s *GoogleClassroomService) FetchCourses(ctx context.Context, userID string) ([]*domain.Group, error) {
	// 1．データベースからユーザー情報を取得する（トークン取得のため）．
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil || user.GoogleAccessToken == "" {
		return nil, fmt.Errorf("Google 連携が行われていない")
	}

	// 2．トークンを用いて Classroom サービスを生成する．
	token := &oauth2.Token{
		AccessToken:  user.GoogleAccessToken,
		RefreshToken: user.GoogleRefreshToken,
		TokenType:    "Bearer",
	}
	client := s.oauthCfg.Client(ctx, token)
	srv, err := classroom.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Classroom サービスの作成に失敗した： %w", err)
	}

	// 3．アクティブなコース一覧を取得する（アーカイブ済みを除外）．
	resp, err := srv.Courses.List().CourseStates("ACTIVE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Google Classroom API からのデータ取得に失敗した（権限や API 設定を確認してほしい）： %w", err)
	}

	groups := []*domain.Group{}
	for _, c := range resp.Courses {
		groups = append(groups, &domain.Group{
			ID:      c.Id,
			Name:    c.Name,
			OwnerID: userID, // 現在のユーザーがインポートした形とする
		})
	}

	return groups, nil
}

// GetProviderName はこのサービスが Google Classroom であることを返す．
func (s *GoogleClassroomService) GetProviderName() string {
	return domain.SourceGoogleClassroom
}
