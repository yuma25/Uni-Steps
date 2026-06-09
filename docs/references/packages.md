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

## 5. 標準ライブラリ (context)
**Package**: `context`  
**Reference**: [pkg.go.dev/context](https://pkg.go.dev/context)

### `type Context interface`
デッドライン，キャンセル信号，その他のリクエストスコープの値を API の境界を越えて運ぶためのインターフェースである．

### `func Background() Context`
空の Context を返す．通常，main 関数，初期化，テスト，およびトップレベルのリクエストの Context として使用される．

### `func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)`
親 Context のコピーを返すが，timeout が経過するとその Context はキャンセルされる．API リクエストのタイムアウト管理に必須である．

### `func WithCancel(parent Context) (ctx Context, cancel CancelFunc)`
親 Context のコピーを返し，cancel 関数が呼ばれると，返された Context の Done チャネルが閉じられる．

---

## 6. 文字列操作・フォーマット (fmt)
**Package**: `fmt`  
**Reference**: [pkg.go.dev/fmt](https://pkg.go.dev/fmt)

### `func Errorf(format string, a ...any) error`
フォーマットに従った文字列を持つエラーを生成する．`%w` を使用することで，元のエラーをラップ（包み込む）し，エラーの追跡（スタックトレースのようなもの）を可能にする．

### `func Printf(format string, a ...any) (n int, err error)`
フォーマットに従って標準出力に文字列を表示する．開発中のデバッグやログ出力に使用する．

---

## 7. Google Classroom API (Candidate)
**Package**: `google.golang.org/api/classroom/v1`  
**Reference**: [pkg.go.dev/google.golang.org/api/classroom/v1](https://pkg.go.dev/google.golang.org/api/classroom/v1)

### `func NewService(ctx context.Context, opts ...option.ClientOption) (*Service, error)`
Google Classroom と通信するためのサービスを生成する．

---

## 8. HTTP クライアント・サーバー (net/http)
**Package**: `net/http`  
**Reference**: [pkg.go.dev/net/http](https://pkg.go.dev/net/http)

### 定数 (Status Codes)
HTTP レスポンスの状態を表す定数である．
*   `StatusOK (200)`: リクエストが成功したことを示す．
*   `StatusCreated (201)`: リクエストが成功し，新しいリソース（タスク等）が作成されたことを示す．
*   `StatusBadRequest (400)`: クライアントのリクエストが不正（JSON 形式の誤り等）であることを示す．
*   `StatusInternalServerError (500)`: サーバー側で予期せぬエラーが発生したことを示す．

### `type Request struct`
サーバーが受信する，またはクライアントが送信する HTTP リクエストを表す．
*   `func (r *Request) Context() context.Context`: リクエストのコンテキストを返す．タイムアウトやキャンセルの制御に使用する．
