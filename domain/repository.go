package domain

import (
	"context"
	"time"
)

// TaskRepository は課題データの永続化に関する約束事（インターフェース）である．
// DDDでは，具体的なDBの実装（SQLなど）はここに書かず，ビジネスロジックが必要な「機能」だけを定義する．
type TaskRepository interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByExternalID(ctx context.Context, externalID string) (*Task, error) // 外部 LMS の ID で検索する．
	FindByGroupID(ctx context.Context, groupID string) ([]*Task, error)
	// FindApproachingDeadlines は指定された日時までに期限を迎える，未完了のタスクを取得する．
	FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*Task, error)
}

// UserRepository はユーザーデータの永続化に関する約束事である．
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	UpdateWebPushToken(ctx context.Context, userID string, token string) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error) // メールアドレスでユーザーを検索する．
}

// GroupRepository はグループデータの永続化に関する約束事である．
type GroupRepository interface {
	Save(ctx context.Context, group *Group) error
	FindByID(ctx context.Context, id string) (*Group, error)
	FindByInviteCode(ctx context.Context, code string) (*Group, error) // 招待コードでグループを検索する．
	FindByUserID(ctx context.Context, userID string) ([]*Group, error) // ユーザー ID に紐づくグループ一覧を取得する．
}

// WakeupRepository は起床確認データの永続化に関する約束事である．
type WakeupRepository interface {
	Save(ctx context.Context, check *WakeupCheck) error
	Delete(ctx context.Context, id string) error
	FindPendingByTime(ctx context.Context, now time.Time) ([]*WakeupCheck, error)
	FindActiveByUser(ctx context.Context, userID string) ([]*WakeupCheck, error)
	FindActiveByGroup(ctx context.Context, groupID string) ([]*WakeupCheck, error) // グループ内の現在進行中の起床確認を取得する．
}
