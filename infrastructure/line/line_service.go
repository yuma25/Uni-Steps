package line

import (
	"context"
	"fmt"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/yuma25/Uni-Steps/domain"
)

// LineService は domain.NotificationService インターフェースを実装し，
// LINE Messaging API を使用して通知を送信する構造体である．
type LineService struct {
	client *messaging_api.MessagingApiAPI // LINE API と通信するためのクライアントである．
}

// コンパイル時にインターフェースの実装をチェックする．
var _ domain.NotificationService = (*LineService)(nil)

// NewLineService は LineService の新しいインスタンスを生成する．
func NewLineService(channelToken string) (*LineService, error) {
	client, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		return nil, fmt.Errorf("LINE クライアントの作成に失敗した： %w", err)
	}

	return &LineService{
		client: client,
	}, nil
}

// SendGroupMessage は指定された LINE グループ ID に対してテキストメッセージを送信する．
func (s *LineService) SendGroupMessage(ctx context.Context, targetID string, message string) error {
	// 送信するテキストメッセージのオブジェクトを作成する．
	textMessage := messaging_api.TextMessage{
		Text: message,
	}

	// PushMessage リクエストを作成する．
	// PushMessage は任意のタイミングで指定した宛先（targetID）へメッセージを送る API である．
	pushRequest := &messaging_api.PushMessageRequest{
		To:       targetID,
		Messages: []messaging_api.MessageInterface{textMessage},
	}

	// API を呼び出して送信を実行する．
	// 第二引数の xLineRetryKey はリトライ時の重複実行を防ぐためのキーである（今回は使用しないため空文字）．
	_, err := s.client.PushMessage(pushRequest, "")
	if err != nil {
		return fmt.Errorf("LINE への Push Message 送信に失敗した： %w", err)
	}

	return nil
}

// SendDirectMessage は Web Push など別機能で実装するため，ここではエラーを返すかログを出力する．
func (s *LineService) SendDirectMessage(ctx context.Context, userID string, message string) error {
	// 今回の設計では，個人宛の直接通知は Web Push を使用する想定であるため，
	// LINE サービスでは未実装（Not Implemented）とする．
	return fmt.Errorf("LineService では SendDirectMessage はサポートされていない")
}
