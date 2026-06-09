package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
)

// NotificationHandler は通知設定に関する HTTP リクエストを受け付ける窓口である．
type NotificationHandler struct {
	userRepo domain.UserRepository
}

// NewNotificationHandler はハンドラーを初期化し，ルーティングを登録する．
func NewNotificationHandler(e *echo.Echo, ur domain.UserRepository) {
	h := &NotificationHandler{
		userRepo: ur,
	}
	e.POST("/api/notifications/subscribe", h.SubscribeWebPush)
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
