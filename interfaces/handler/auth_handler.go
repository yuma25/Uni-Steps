package handler

import (
	"encoding/json"
	"net/http"
	"time"

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
	// CSRF 対策としてランダムな文字列（state）を生成する．
	state := uuid.New().String()

	// state を Cookie に保存する（ブラウザ側で保持させ，Callback 時に検証する）．
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(15 * time.Minute), // 15 分間有効
		HttpOnly: true,                             // JavaScript からアクセス不可
		Secure:   false,                            // 本番環境（HTTPS）では true にすべきである
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback は Google からの認可コードを受け取り，ユーザー情報の取得・保存・ログインを行う．
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	// state の検証を行う（CSRF 対策）．
	state := c.QueryParam("state")
	cookie, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != cookie.Value {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "不正なリクエスト（state 不一致）が検出された"})
	}

	// 使用済みの Cookie を削除する．
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	})

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
	user.GoogleTokenExpiry = token.Expiry

	if err := h.userRepo.Save(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の保存に失敗した"})
	}

	// 5．ダッシュボードへリダイレクトする．
	// フロントエンドの URL へリダイレクト（暫定的に localhost:5173 を使用）
	frontendUrl := "http://localhost:5173/select-group?user_id=" + user.ID
	return c.Redirect(http.StatusTemporaryRedirect, frontendUrl)
}
