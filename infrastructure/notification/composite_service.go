package notification

import (
	"context"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/infrastructure/line"
	"github.com/yuma25/Uni-Steps/infrastructure/webpush"
)

// CompositeNotificationService は LINE と Web Push の両方を束ねて
// 一つの NotificationService として振る舞う構造体である．
type CompositeNotificationService struct {
	lineService    *line.LineService       // グループ宛の LINE 通知を担当する．
	webPushService *webpush.WebPushService // 個人宛の Web Push 通知を担当する．
}

var _ domain.NotificationService = (*CompositeNotificationService)(nil)

// NewCompositeNotificationService は新しい CompositeNotificationService を生成する．
func NewCompositeNotificationService(ls *line.LineService, wps *webpush.WebPushService) *CompositeNotificationService {
	return &CompositeNotificationService{
		lineService:    ls,
		webPushService: wps,
	}
}

// SendGroupMessage は LINE サービスに処理を委譲する．
func (s *CompositeNotificationService) SendGroupMessage(ctx context.Context, targetID string, message string) error {
	return s.lineService.SendGroupMessage(ctx, targetID, message)
}

// SendDirectMessage は Web Push サービスに処理を委譲する．
func (s *CompositeNotificationService) SendDirectMessage(ctx context.Context, userID string, message string) error {
	return s.webPushService.SendDirectMessage(ctx, userID, message)
}
