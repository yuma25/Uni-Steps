package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
)

// UserHandler はユーザー情報に関する HTTP リクエストを受け付ける窓口である．
type UserHandler struct {
	userRepo domain.UserRepository
}

// NewUserHandler はハンドラーを初期化し，ルーティングを登録する．
func NewUserHandler(e *echo.Echo, ur domain.UserRepository) {
	h := &UserHandler{
		userRepo: ur,
	}
	e.GET("/api/users/:id", h.GetUser)
}

// GetUser は指定された ID のユーザー情報を取得する．
func (h *UserHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	user, err := h.userRepo.FindByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の取得に失敗した"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "ユーザーが見つからない"})
	}

	// パスワードやトークン自体は見せず，存在有無だけを返す安全な構造体
	response := map[string]interface{}{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"has_push_token": user.WebPushToken != "",
	}

	return c.JSON(http.StatusOK, response)
}
