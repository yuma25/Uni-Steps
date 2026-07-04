package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/yuma25/Uni-Steps/domain"
)

// WebPushService はブラウザに対してネイティブ通知を送るサービスである．
type WebPushService struct {
	userRepo   domain.UserRepository // ユーザーの通知トークンを取得するためのリポジトリである．
	publicKey  string                // VAPID 公開鍵である．
	privateKey string                // VAPID 秘密鍵である．
	contact    string                // 連絡先（mailto:メールアドレス等）である．
}

// NewWebPushService は WebPushService の新しいインスタンスを生成する．
func NewWebPushService(ur domain.UserRepository, pubKey, privKey, contact string) *WebPushService {
	return &WebPushService{
		userRepo:   ur,
		publicKey:  pubKey,
		privateKey: privKey,
		contact:    contact,
	}
}

// SendDirectMessage は指定されたユーザーのブラウザに対して Web Push 通知を送信する．
func (s *WebPushService) SendDirectMessage(ctx context.Context, userID string, message string, targetURL string) error {
	// 1．対象ユーザーの情報をデータベースから取得する．
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ユーザー情報の取得に失敗した： %w", err)
	}
	if user == nil || user.WebPushToken == "" {
		return fmt.Errorf("ユーザーの Web Push トークンが存在しない")
	}

	// 2．保存されている JSON 文字列のトークンを webpush.Subscription 構造体に変換する．
	var sub webpush.Subscription
	if err := json.Unmarshal([]byte(user.WebPushToken), &sub); err != nil {
		return fmt.Errorf("Web Push トークンの解析に失敗した： %w", err)
	}

	// 3．送信するメッセージを JSON 構造体にまとめる．
	payload := map[string]string{
		"title": "Uni-Steps",
		"body":  message,
		"url":   targetURL,
	}
	messageBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("通知データの JSON 変換に失敗した： %w", err)
	}

	// 4．Web Push 通知を送信する．
	resp, err := webpush.SendNotification(messageBytes, &sub, &webpush.Options{
		Subscriber:      s.contact,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             30, // 30秒間ブラウザに届かなければ破棄する（緊急通知のため短め）
	})
	if err != nil {
		return fmt.Errorf("Web Push の送信に失敗した： %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// 410 (Gone) や 404 (Not Found) の場合，トークンが無効なので削除する．
		if resp.StatusCode == 410 || resp.StatusCode == 404 {
			_ = s.userRepo.UpdateWebPushToken(ctx, userID, "")
			log.Printf("[WebPush] 無効なトークンを検知したため削除した．UserID: %s, Status: %d\n", userID, resp.StatusCode)
		}
		return fmt.Errorf("Web Push サーバーからエラーが返された： ステータスコード %d", resp.StatusCode)
	}

	log.Printf("[WebPush] 通知を正常に送信した．Status: %d, UserID: %s\n", resp.StatusCode, userID)
	return nil
}
