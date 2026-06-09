# Uni-Steps 技術リファレンス集

開発で使用している主要な Go パッケージの公式リファレンス（pkg.go.dev）の内容を詳細にまとめている．

## 📜 記述ルール
1.  **詳細な仕様**: 使用しているインターフェース，構造体，およびそのメソッドまで詳細に記述する．
2.  **種別の明記**: 以下のアイコンとタグを用いて，要素の種類が一目で判別できるように表記する．
    *   🔹 **[Interface]**: インターフェース
    *   📦 **[Struct]**: 構造体
    *   ⚙️ **[Function]**: 関数（構造体に紐づかないもの）
    *   🔧 **[Method]**: メソッド（構造体やインターフェースに紐づくもの）
    *   🏷️ **[Constant / Field]**: 定数や構造体のフィールド変数
3.  **既存内容の維持**: 既存の記述を削除せず，常に追記・更新を行う．
4.  **記述スタイル**: 「である」調を使用し，句読点は「．」「，」に統一する．
5.  **リンクの明示**: パッケージ名および公式リファレンスへの URL を必ず記載する．

---

## 1. Web フレームワーク (Echo v4)
**Package**: `github.com/labstack/echo/v4`  
**Reference**: [pkg.go.dev/github.com/labstack/echo/v4](https://pkg.go.dev/github.com/labstack/echo/v4)

### 📦 `type Echo struct`
Echo フレームワークのメインとなる構造体である．
*   ⚙️ `func New() *Echo`: 新しい Echo インスタンスを生成し，ポインタを返す．
*   🔧 `func (e *Echo) Start(address string) error`: 指定されたアドレスで HTTP サーバーを起動する．
*   🔧 `func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route`: GET リクエストのパスを登録する．
*   🔧 `func (e *Echo) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route`: POST リクエストのパスを登録する．
*   🔧 `func (e *Echo) Use(middleware ...MiddlewareFunc)`: ミドルウェアを登録する．

### 🔹 `type Context interface`
各リクエストの情報を保持するインターフェースである．
*   🔧 `func (c Context) Param(name string) string`: URL 内のパラメータ（`:id` 等）を取得する．
*   🔧 `func (c Context) Bind(i interface{}) error`: リクエストボディ（JSON 等）を構造体にバインドする．
*   🔧 `func (c Context) JSON(code int, i interface{}) error`: 構造体を JSON 変換してレスポンスとして送信する．
*   🔧 `func (c Context) Request() *http.Request`: 現在の HTTP リクエストオブジェクトを返す．

---

## 2. 生成 AI (Gemini)
**Package**: `github.com/google/generative-ai-go/genai`  
**Reference**: [pkg.go.dev/github.com/google/generative-ai-go/genai](https://pkg.go.dev/github.com/google/generative-ai-go/genai)

### 📦 `type Client struct`
Gemini API クライアントである．
*   ⚙️ `func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error)`: クライアントを生成する．
*   🔧 `func (c *Client) GenerativeModel(name string) *GenerativeModel`: モデルを指定して取得する．
*   🔧 `func (c *Client) Close() error`: クライアントとの接続を閉じる．

### 📦 `type GenerativeModel struct`
AI によるコンテンツ生成を行うモデルである．
*   🔧 `func (m *GenerativeModel) GenerateContent(ctx context.Context, parts ...Part) (*GenerateContentResponse, error)`: AI にコンテンツ生成を依頼する．
*   🔧 `func (m *GenerativeModel) CountTokens(ctx context.Context, parts ...Part) (*CountTokensResponse, error)`: 指定したテキストのトークン数を計算する．

### 📦 `type GenerateContentResponse struct`
生成結果を保持する構造体である．
*   🏷️ `Candidates []*Candidate`: 生成された候補のリストである．

### 📦 `type Candidate struct`
生成結果の候補である．
*   🏷️ `Content *Content`: 生成された内容である．

---

## 3. 環境変数管理 (godotenv)
**Package**: `github.com/joho/godotenv`  
**Reference**: [pkg.go.dev/github.com/joho/godotenv](https://pkg.go.dev/github.com/joho/godotenv)

### ⚙️ `func Load(filenames ...string) (err error)`
`.env` ファイルを読み込み，環境変数としてロードする．
*   **注意**: 既存の環境変数は上書きしない．

---

## 4. Google API 通信オプション
**Package**: `google.golang.org/api/option`  
**Reference**: [pkg.go.dev/google.golang.org/api/option](https://pkg.go.dev/google.golang.org/api/option)

### ⚙️ `func WithAPIKey(apiKey string) ClientOption`
API キーによる認証オプションを生成する．

---

## 5. 標準ライブラリ (context)
**Package**: `context`  
**Reference**: [pkg.go.dev/context](https://pkg.go.dev/context)

### 🔹 `type Context interface`
デッドラインやキャンセル信号を運ぶインターフェースである．
*   ⚙️ `func Background() Context`: 空の Context を返す．
*   ⚙️ `func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)`: タイムアウト付き Context を生成する．

---

## 6. 文字列操作・フォーマット (fmt)
**Package**: `fmt`  
**Reference**: [pkg.go.dev/fmt](https://pkg.go.dev/fmt)

### ⚙️ `func Errorf(format string, a ...any) error`
エラーを生成する．`%w` でエラーのラップが可能である．

### ⚙️ `func Printf(format string, a ...any) (n int, err error)`
標準出力にフォーマット済み文字列を表示する．

---

## 7. JSON のエンコード・デコード (encoding/json)
**Package**: `encoding/json`  
**Reference**: [pkg.go.dev/encoding/json](https://pkg.go.dev/encoding/json)

### ⚙️ `func Unmarshal(data []byte, v any) error`
JSON データを構造体に変換する．

### ⚙️ `func Marshal(v any) ([]byte, error)`
構造体を JSON データに変換する．

---

## 8. Google Classroom API
**Package**: `google.golang.org/api/classroom/v1`  
**Reference**: [pkg.go.dev/google.golang.org/api/classroom/v1](https://pkg.go.dev/google.golang.org/api/classroom/v1)

### 📦 `type Service struct`
Google Classroom と通信するためのメインサービスである．
*   ⚙️ `func NewService(ctx context.Context, opts ...option.ClientOption) (*Service, error)`: サービスを生成する．

### 📦 `type CoursesService struct`
コース関連の操作を提供する構造体である．
*   🏷️ `CourseWork *CoursesCourseWorkService`: コース内の課題（コースワーク）を操作するサービスへアクセスする．

### 📦 `type CoursesCourseWorkService struct`
課題（コースワーク）に対する具体的な操作を提供する構造体である．
*   🔧 `func (r *CoursesCourseWorkService) List(courseId string) *CoursesCourseWorkListCall`: 指定したコースの課題一覧を取得するための呼び出しを作成する．
*   🔧 `func (c *CoursesCourseWorkListCall) Context(ctx context.Context) *CoursesCourseWorkListCall`: コンテキストを設定する．
*   🔧 `func (c *CoursesCourseWorkListCall) Do(opts ...googleapi.CallOption) (*ListCourseWorkResponse, error)`: 実際の API リクエストを実行し，課題一覧（`CourseWork`）を取得する．

---

## 9. HTTP クライアント・サーバー (net/http)
**Package**: `net/http`  
**Reference**: [pkg.go.dev/net/http](https://pkg.go.dev/net/http)

### 🏷️ 定数 (Status Codes)
*   `StatusOK (200)`
*   `StatusCreated (201)`
*   `StatusBadRequest (400)`
*   `StatusInternalServerError (500)`

### 📦 `type Request struct`
HTTP リクエストを表す構造体である．
*   🔧 `func (r *Request) Context() context.Context`: リクエストのコンテキストを返す．

---

## 10. 時間操作 (time)
**Package**: `time`  
**Reference**: [pkg.go.dev/time](https://pkg.go.dev/time)

### 📦 `type Time struct`
日時を表現する構造体である．
*   ⚙️ `func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time`: 指定した日時要素から Time 構造体を生成する．
*   🔧 `func (t Time) Format(layout string) string`: 日時を指定したレイアウト文字列（例：`time.RFC3339`）でフォーマットする．
*   ⚙️ `func Now() Time`: 現在のローカル時間を返す．

## 11. データベース ORM (gorm)
**Package**: `gorm.io/gorm`  
**Reference**: [pkg.go.dev/gorm.io/gorm](https://pkg.go.dev/gorm.io/gorm)

### 📦 `type DB struct`
データベースとの接続や操作（クエリの構築，実行）を管理するメインの構造体である．
*   🔧 `func (db *DB) AutoMigrate(dst ...interface{}) error`: 指定された構造体（モデル）に基づいてデータベースのテーブルを自動的に作成・更新する．
*   🔧 `func (db *DB) WithContext(ctx context.Context) *DB`: リクエストのキャンセルやタイムアウトを制御するためのコンテキストを設定した新しい DB インスタンスを返す．
*   🔧 `func (db *DB) Save(value interface{}) *DB`: オブジェクトを保存する．主キー（ID）が存在する場合は UPDATE，存在しない場合は INSERT を実行する．
*   🔧 `func (db *DB) Where(query interface{}, args ...interface{}) *DB`: SQL の WHERE 句に相当する条件を指定する．
*   🔧 `func (db *DB) First(dest interface{}, conds ...interface{}) *DB`: 条件に一致する最初の1件を取得し，`dest` に格納する．見つからない場合は `gorm.ErrRecordNotFound` を返す．
*   🔧 `func (db *DB) Find(dest interface{}, conds ...interface{}) *DB`: 条件に一致する複数件を取得し，スライス `dest` に格納する．

### 🏷️ エラー定数
*   `ErrRecordNotFound`: データベース検索時（`First` 等）にレコードが見つからなかった場合に返される標準エラーである．

---

## 12. LINE Messaging API
**Package**: `github.com/line/line-bot-sdk-go/v8/linebot/messaging_api`  
**Reference**: [pkg.go.dev/github.com/line/line-bot-sdk-go/v8/linebot/messaging_api](https://pkg.go.dev/github.com/line/line-bot-sdk-go/v8/linebot/messaging_api)

### 📦 `type MessagingApiAPI struct`
LINE Messaging API と通信するためのメインクライアントである．
*   ⚙️ `func NewMessagingApiAPI(channelToken string) (*MessagingApiAPI, error)`: チャンネルアクセストークンを用いてクライアントを生成する．
*   🔧 `func (client *MessagingApiAPI) PushMessage(pushMessageRequest *PushMessageRequest, xLineRetryKey string) (*PushMessageResponse, error)`: 指定した宛先に対してメッセージを送信する．第二引数は重複実行防止用のキー（任意）である．

### 📦 `type PushMessageRequest struct`
Push メッセージ送信時のリクエストボディを表す構造体である．
*   🏷️ `To string`: 送信先の ID（ユーザーID，グループIDなど）である．
*   🏷️ `Messages []MessageInterface`: 送信するメッセージオブジェクトの配列である（最大 5 件まで）．

### 📦 `type TextMessage struct`
テキストメッセージを表す構造体である．（`MessageInterface` を満たす）
*   🏷️ `Text string`: 送信するテキスト内容である．