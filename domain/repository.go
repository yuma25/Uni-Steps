package domain

import "context"

// TaskRepository は課題データの永続化に関する約束事（インターフェース）である．
// DDDでは，具体的なDBの実装（SQLなど）はここに書かず，ビジネスロジックが必要な「機能」だけを定義する．
type TaskRepository interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByGroupID(ctx context.Context, groupID string) ([]*Task, error)
}

// UserRepository はユーザーデータの永続化に関する約束事である．
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
}
