package domain

import "context"

// LMSService は外部の学習管理システム（LMS）との連携を担うインターフェースである．
// Google Classroom や Web Class などの差異をこのレイヤーで吸収する．
type LMSService interface {
	// FetchTasks は外部システムから最新の課題一覧を取得する．
	FetchTasks(ctx context.Context, userID string, groupID string) ([]*Task, error)
	// FetchCourses はユーザーが所属しているコース（授業）の一覧を取得する．
	FetchCourses(ctx context.Context, userID string) ([]*Group, error)
	// GetProviderName は連携先の名前（"google_classroom" 等）を返す．
	GetProviderName() string
}
