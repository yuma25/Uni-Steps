package domain

import "context"

// LMSService は外部の学習管理システム（LMS）との連携を担うインターフェースである．
// Google Classroom や Web Class などの差異をこのレイヤーで吸収する．
type LMSService interface {
	// FetchTasks は外部システムから最新の課題一覧を取得する．
	// userID を使ってデータベースから OAuth トークンを取得し，そのユーザーの権限で通信を行う．
	FetchTasks(ctx context.Context, userID string, groupID string) ([]*Task, error)
	// GetProviderName は連携先の名前（"google_classroom" 等）を返す．
	GetProviderName() string
}
