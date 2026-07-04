package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
)

// NotificationHandler は通知設定に関する HTTP リクエストを受け付ける窓口である．
type NotificationHandler struct {
	userRepo     domain.UserRepository
	aiService    domain.AIService
	notifService domain.NotificationService
}

// NewNotificationHandler はハンドラーを初期化し，ルーティングを登録する．
func NewNotificationHandler(e *echo.Echo, ur domain.UserRepository, ai domain.AIService, ns domain.NotificationService) {
	h := &NotificationHandler{
		userRepo:     ur,
		aiService:    ai,
		notifService: ns,
	}
	e.POST("/api/notifications/subscribe", h.SubscribeWebPush)
	e.POST("/api/notifications/test", h.SendTestNotification)
}

// SendTestNotification は動作確認用のテスト通知を即座に送信する．
func (h *NotificationHandler) SendTestNotification(c echo.Context) error {
	var req struct {
		UserID      string `json:"user_id"`
		GroupID     string `json:"group_id"`
		AICharacter string `json:"ai_character"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	// ダミーの課題情報を作成して AI にメッセージを作らせる．
	dummyTask := &domain.Task{
		Title: "テスト通知の確認",
	}

	msg, err := h.aiService.GenerateRemindMessage(c.Request().Context(), dummyTask, req.AICharacter)
	if err != nil {
		msg = "【Uni-Steps】これはテスト通知です．正常に受信できています！"
	}

	// 遷移先 URL の決定（部屋 ID があればダッシュボードへ，なければ選択画面へ）
	targetURL := "/select-group?user_id=" + req.UserID
	if req.GroupID != "" {
		targetURL = "/dashboard?user_id=" + req.UserID + "&group_id=" + req.GroupID
	}

	err = h.notifService.SendDirectMessage(c.Request().Context(), req.UserID, msg, targetURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "通知の送信に失敗した：" + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "テスト通知を送信した"})
}

// SubscribeWebPush はクライアントから Web Push の購読情報を受け取り，ユーザー情報に保存する．
func (h *NotificationHandler) SubscribeWebPush(c echo.Context) error {
	// リクエストボディの構造体定義
	var req struct {
		UserID       string `json:"user_id"`
		Subscription string `json:"subscription"` // JSON 文字列化されたトークン
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	if req.UserID == "" || req.Subscription == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id と subscription は必須である"})
	}

	// ユーザーを取得し，トークンを更新して保存する．
	user, err := h.userRepo.FindByID(c.Request().Context(), req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の取得に失敗した"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "ユーザーが見つからない"})
	}

	user.WebPushToken = req.Subscription
	if err := h.userRepo.Save(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "トークンの保存に失敗した"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "購読情報の保存に成功した"})
}
