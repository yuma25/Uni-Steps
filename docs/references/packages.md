# Uni-Steps 技術リファレンス集

開発で使用している主要な Go パッケージの公式リファレンス（pkg.go.dev）の内容を詳細にまとめている．

---

## 1. Web フレームワーク (Echo v4)
**Package**: `github.com/labstack/echo/v4`  
**Reference**: [pkg.go.dev/.../echo/v4](https://pkg.go.dev/github.com/labstack/echo/v4)

### `func New() *Echo`
新しい Echo インスタンスを生成し，ポインタを返す．これがサーバーの基盤となる．

### `func (e *Echo) Start(address string) error`
指定されたアドレス（例: `:8080`）で HTTP サーバーを起動する．

### `func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route`
### `func (e *Echo) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route`
指定されたパスにハンドラー関数を登録する．Echo のルーティングは非常に高速である．

### `type Context interface`
HTTP リクエストの全ての情報を保持するインターフェースである．
*   `func (c Context) Param(name string) string`: URL 内のパラメータ（`:id` 等）を取得する．
*   `func (c Context) Bind(i interface{}) error`: リクエストボディ（JSON等）を構造体に自動で流し込む．
*   `func (c Context) JSON(code int, i interface{}) error`: 構造体を JSON に変換してレスポンスとして送信する．

---

## 2. 生成 AI (Gemini)
**Package**: `github.com/google/generative-ai-go/genai`  
**Reference**: [pkg.go.dev/.../genai](https://pkg.go.dev/github.com/google/generative-ai-go/genai)

### `func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error)`
Gemini API と通信するためのクライアントを生成する．

### `func (c *Client) GenerativeModel(name string) *GenerativeModel`
使用するモデル（例: `gemini-2.0-flash`）を指定して，モデルインスタンスを取得する．

### `func (m *GenerativeModel) CountTokens(ctx context.Context, parts ...Part) (*CountTokensResponse, error)`
指定したテキストのトークン数を計算する．課金は発生しない．

### `func (m *GenerativeModel) GenerateContent(ctx context.Context, parts ...Part) (*GenerateContentResponse, error)`
AI にテキストを生成させる（チャットや解析に使用する）．

---

## 3. 環境変数管理 (godotenv)
**Package**: `github.com/joho/godotenv`  
**Reference**: [pkg.go.dev/github.com/joho/godotenv](https://pkg.go.dev/github.com/joho/godotenv)

### `func Load(filenames ...string) (err error)`
`.env` ファイルを読み込み，現在のプロセスの環境変数に展開する．
*   引数なしの場合，実行パスの `.env` を読み込む．
*   **注意**: 既に設定されている環境変数は上書きしない（開発用のデフォルト値を設定するのに適している）．
*   プログラムの開始直後（`main` 関数内）で呼び出すことが推奨される．

---

## 4. Google API 通信オプション
**Package**: `google.golang.org/api/option`  
**Reference**: [pkg.go.dev/google.golang.org/api/option](https://pkg.go.dev/google.golang.org/api/option)

### `func WithAPIKey(apiKey string) ClientOption`
API キーを使用して認証するためのオプションを返す．`genai.NewClient` の引数に渡す．

---

## 5. 標準ライブラリ
### `package context`
API リクエストの期限（Timeout）やキャンセルを制御するために全ての API 呼び出しで使用する．
*   `func Background() Context`: 空のコンテキストを生成する．
*   `func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)`: 指定時間で打ち切るコンテキストを生成する．
