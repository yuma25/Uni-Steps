package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
	"golang.org/x/oauth2"
)

// AuthHandler は Google OAuth 2.0 による認証を管理するハンドラーである．
type AuthHandler struct {
	userRepo domain.UserRepository // ユーザー情報を保存するためのリポジトリである．
	oauthCfg *oauth2.Config        // Google OAuth 2.0 の設定情報である．
}

// NewAuthHandler は AuthHandler を初期化し，ルーティングを登録する．
func NewAuthHandler(e *echo.Echo, ur domain.UserRepository, cfg *oauth2.Config) {
	h := &AuthHandler{
		userRepo: ur,
		oauthCfg: cfg,
	}
	e.GET("/api/auth/google/login", h.GoogleLogin)
	e.GET("/api/auth/google/callback", h.GoogleCallback)
}

// GoogleLogin は Google の認可画面へリダイレクトする．
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	// 本来は CSRF 対策として state パラメータにランダムな文字列を設定すべきである．
	// 今回はプロトタイプのため固定文字列とする．
	url := h.oauthCfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback は Google からの認可コードを受け取り，トークンを取得・保存する．
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "認可コードが提供されていない"})
	}

	// 1．認可コードをアクセストークンと交換する．
	token, err := h.oauthCfg.Exchange(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "トークンの交換に失敗した"})
	}

	// 2．本来はここでユーザー ID を取得（ID トークンの解析等）する必要がある．
	// プロトタイプのため，固定のユーザー ID またはクエリ等から ID を受け取る想定とする．
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id が指定されていない（本来は認証済みのセッションから取得する）"})
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の取得に失敗した"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "ユーザーが見つからない"})
	}

	// 3．取得したトークンをユーザー情報に保存する．
	user.GoogleAccessToken = token.AccessToken
	user.GoogleRefreshToken = token.RefreshToken
	if err := h.userRepo.Save(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "トークンの保存に失敗した"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Google 連携に成功した"})
}
