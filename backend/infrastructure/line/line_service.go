package line

import (
	"context"
	"fmt"
	"log"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/yuma25/Uni-Steps/domain"
)

// LineService は domain.NotificationService インターフェースを実装し，
// LINE Messaging API を使用してグループ通知を送信する構造体である．
type LineService struct {
	groupRepo domain.GroupRepository // BYOT 方式のため，グループごとのトークンを取得する．
}

// コンパイル時にインターフェースの実装をチェックする．
var _ domain.NotificationService = (*LineService)(nil)

// NewLineService は LineService の新しいインスタンスを生成する．
func NewLineService(gr domain.GroupRepository) *LineService {
	return &LineService{
		groupRepo: gr,
	}
}

// SendGroupMessage は指定されたグループ（targetID = グループID）の LINE 設定に基づいてメッセージを送信する．
func (s *LineService) SendGroupMessage(ctx context.Context, targetID string, message string) error {
	group, err := s.groupRepo.FindByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("グループの取得に失敗した： %w", err)
	}
	if group == nil {
		return fmt.Errorf("グループが存在しない")
	}

	// LINE 連携が設定されていない場合はスキップ
	if group.LineChannelToken == "" || group.LineGroupID == "" {
		log.Printf("[LineService] グループ %s に LINE 連携が設定されていないためスキップした\n", targetID)
		return nil
	}

	client, err := messaging_api.NewMessagingApiAPI(group.LineChannelToken)
	if err != nil {
		return fmt.Errorf("LINE クライアントの作成に失敗した： %w", err)
	}

	textMessage := messaging_api.TextMessage{
		Text: message,
	}

	pushRequest := &messaging_api.PushMessageRequest{
		To:       group.LineGroupID,
		Messages: []messaging_api.MessageInterface{textMessage},
	}

	_, err = client.PushMessage(pushRequest, "")
	if err != nil {
		return fmt.Errorf("LINE への Push Message 送信に失敗した： %w", err)
	}

	log.Printf("[LineService] グループ %s の LINE にメッセージを送信した\n", targetID)
	return nil
}

// SendDirectMessage は Web Push など別機能で実装するため，ここではサポートしない．
func (s *LineService) SendDirectMessage(ctx context.Context, userID string, message string, targetURL string) error {
	return fmt.Errorf("LineService では SendDirectMessage はサポートされていない")
}
