package lms

import (
	"context"
	"fmt"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

// GoogleClassroomService は Google Classroom API と連携する LMSService の実装である．
type GoogleClassroomService struct {
	service *classroom.Service // Classroom API と通信するためのサービスである．
}

// NewGoogleClassroomService は GoogleClassroomService の新しいインスタンスを生成する．
func NewGoogleClassroomService(ctx context.Context, apiKey string) (*GoogleClassroomService, error) {
	// API キーを使用して Classroom サービスを初期化する．
	// ※実際の運用では OAuth 2.0 のトークン（ユーザーごとの認可）が必要になる場合が多いが，
	// ここでは設計として API キーによる初期化の形をとる．
	srv, err := classroom.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Classroom サービスの作成に失敗した： %w", err)
	}

	return &GoogleClassroomService{
		service: srv,
	}, nil
}

// FetchTasks は Google Classroom から指定されたコース（グループ）の課題一覧を取得する．
func (s *GoogleClassroomService) FetchTasks(ctx context.Context, courseID string) ([]*domain.Task, error) {
	// 1．Classroom API を呼び出して，指定されたコースの課題 (CourseWork) を取得する．
	resp, err := s.service.Courses.CourseWork.List(courseID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("コースワークの取得に失敗した： %w", err)
	}

	var tasks []*domain.Task

	// 2．取得した CourseWork をドメインモデルの Task に変換する．
	for _, cw := range resp.CourseWork {
		// Classroom API の DueDate (年，月，日) と DueTime (時，分，秒) を結合して time.Time に変換する．
		var deadline time.Time
		if cw.DueDate != nil && cw.DueTime != nil {
			// タイムゾーンは UTC として解釈（要調整）
			deadline = time.Date(int(cw.DueDate.Year), time.Month(cw.DueDate.Month), int(cw.DueDate.Day),
				int(cw.DueTime.Hours), int(cw.DueTime.Minutes), 0, 0, time.UTC)
		}

		task := &domain.Task{
			Title:      cw.Title,
			ExternalID: cw.Id, // Classroom 側の課題 ID を重複防止のために保持する．
			Deadline:   deadline,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetProviderName はこのサービスが Google Classroom であることを返す．
func (s *GoogleClassroomService) GetProviderName() string {
	return domain.SourceGoogleClassroom
}
