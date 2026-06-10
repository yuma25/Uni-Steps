package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

// GoogleCallback は Google からの認可コードを受け取り，ユーザー情報の取得・保存・ログインを行う．
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

	// 2．Google の UserInfo API からユーザーのプロフィールを取得する．
	client := h.oauthCfg.Client(c.Request().Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の取得に失敗した"})
	}
	defer resp.Body.Close()

	var googleUser struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の解析に失敗した"})
	}

	// 3．メールアドレスを元に，既存ユーザーか新規ユーザーかを判定する．
	user, err := h.userRepo.FindByEmail(c.Request().Context(), googleUser.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "データベース検索に失敗した"})
	}

	if user == nil {
		// 新規ユーザー登録（サインアップ）
		user = &domain.User{
			ID:    uuid.New().String(),
			Name:  googleUser.Name,
			Email: googleUser.Email,
		}
	}

	// 4．トークン情報を更新して保存する．
	user.GoogleAccessToken = token.AccessToken
	if token.RefreshToken != "" {
		user.GoogleRefreshToken = token.RefreshToken
	}
	if err := h.userRepo.Save(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の保存に失敗した"})
	}

	// 5．ダッシュボードへリダイレクトする．
	// フロントエンドの URL へリダイレクト（暫定的に localhost:5173 を使用）
	frontendUrl := "http://localhost:5173/dashboard?user_id=" + user.ID
	return c.Redirect(http.StatusTemporaryRedirect, frontendUrl)
}
