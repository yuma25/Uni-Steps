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

// GetProviderName はこのサービスが Google Classroom であることを返す．
func (s *GoogleClassroomService) GetProviderName() string {
	return domain.SourceGoogleClassroom
}
