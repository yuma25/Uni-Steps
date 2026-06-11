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

// GoogleClassroomService は domain.LMSService インターフェースを実装する構造体である．
type GoogleClassroomService struct {
	userRepo domain.UserRepository
	oauthCfg *oauth2.Config
}

// NewGoogleClassroomService は GoogleClassroomService の新しいインスタンスを生成する．
func NewGoogleClassroomService(ur domain.UserRepository, cfg *oauth2.Config) *GoogleClassroomService {
	return &GoogleClassroomService{
		userRepo: ur,
		oauthCfg: cfg,
	}
}

// FetchTasks は Google Classroom からユーザーの全アクティブコースの課題一覧を取得する．
func (s *GoogleClassroomService) FetchTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil || user.GoogleAccessToken == "" {
		return nil, fmt.Errorf("Google 連携が行われていないか，トークンが存在しない")
	}

	initialToken := &oauth2.Token{
		AccessToken:  user.GoogleAccessToken,
		RefreshToken: user.GoogleRefreshToken,
		Expiry:       user.GoogleTokenExpiry,
		TokenType:    "Bearer",
	}

	// 自動リフレッシュ機能付きの TokenSource を作成する．
	// トークンが更新された際にデータベースへ保存するラップ処理を行う．
	ts := &persistentTokenSource{
		ctx:      ctx,
		userID:   userID,
		userRepo: s.userRepo,
		oauthCfg: s.oauthCfg,
		source:   s.oauthCfg.TokenSource(ctx, initialToken),
	}

	client := oauth2.NewClient(ctx, ts)
	srv, err := classroom.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("Classroom サービスの作成に失敗した： %w", err)
	}

	coursesResp, err := srv.Courses.List().CourseStates("ACTIVE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("コース一覧の取得に失敗した： %w", err)
	}

	tasks := []*domain.Task{}

	for _, course := range coursesResp.Courses {
		resp, err := srv.Courses.CourseWork.List(course.Id).Context(ctx).Do()
		if err != nil {
			continue
		}

		submissionsResp, err := srv.Courses.CourseWork.StudentSubmissions.List(course.Id, "-").Context(ctx).Do()
		submissionStatus := make(map[string]bool)
		if err == nil {
			for _, sub := range submissionsResp.StudentSubmissions {
				if sub.State == "TURNED_IN" || sub.State == "RETURNED" {
					submissionStatus[sub.CourseWorkId] = true
				}
			}
		}

		for _, cw := range resp.CourseWork {
			var deadline time.Time
			if cw.DueDate != nil {
				hour, min := 23, 59
				if cw.DueTime != nil {
					hour, min = int(cw.DueTime.Hours), int(cw.DueTime.Minutes)
				}
				deadlineUTC := time.Date(int(cw.DueDate.Year), time.Month(cw.DueDate.Month), int(cw.DueDate.Day),
					hour, min, 0, 0, time.UTC)
				deadline = deadlineUTC.Local()
			}

			lmsUpdateTime, _ := time.Parse(time.RFC3339, cw.UpdateTime)
			lmsUpdateTime = lmsUpdateTime.Local()

			task := &domain.Task{
				Title:            cw.Title,
				ExternalID:       cw.Id,
				Deadline:         deadline,
				IsLMSDeadlineSet: cw.DueDate != nil,
				LMSUpdateTime:    lmsUpdateTime,
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

// persistentTokenSource は oauth2.TokenSource をラップし，
// 新しいトークンが発行された際に自動的にデータベースを更新する構造体である．
type persistentTokenSource struct {
	ctx      context.Context
	userID   string
	userRepo domain.UserRepository
	oauthCfg *oauth2.Config
	source   oauth2.TokenSource
}

func (ts *persistentTokenSource) Token() (*oauth2.Token, error) {
	token, err := ts.source.Token()
	if err != nil {
		return nil, err
	}

	// 取得したトークンをデータベースの最新状態と比較（簡易的に毎回保存する設計とする）．
	// 本来は AccessToken の変更を検知して保存するのが効率的だが，
	// ここでは安全のためリフレッシュの可能性がある場合は常に保存を試みる．
	user, err := ts.userRepo.FindByID(ts.ctx, ts.userID)
	if err == nil && user != nil {
		// トークンが更新されているか，有効期限が DB より先であれば保存する．
		if user.GoogleAccessToken != token.AccessToken {
			user.GoogleAccessToken = token.AccessToken
			if token.RefreshToken != "" {
				user.GoogleRefreshToken = token.RefreshToken
			}
			user.GoogleTokenExpiry = token.Expiry
			_ = ts.userRepo.Save(ts.ctx, user)
		}
	}

	return token, nil
}

// GetProviderName はこのサービスが Google Classroom であることを返す．
func (s *GoogleClassroomService) GetProviderName() string {
	return domain.SourceGoogleClassroom
}
