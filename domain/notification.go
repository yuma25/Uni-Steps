package domain

import "context"

// NotificationService はユーザーやグループへの通知を担うインターフェースである．
// クリーンアーキテクチャに基づき，LINE や Web Push といった具体的な送信手段は隠蔽する．
type NotificationService interface {
	// SendGroupMessage は指定されたグループ（LINE 等）に対してメッセージを送信する．
	SendGroupMessage(ctx context.Context, targetID string, message string) error

	// SendDirectMessage は特定の個人のデバイス（Web Push 等）に対して直接メッセージを送信する．
	SendDirectMessage(ctx context.Context, userID string, message string) error
}
