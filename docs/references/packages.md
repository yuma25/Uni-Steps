# Uni-Steps 技術リファレンス集

本プロジェクトの開発で使用している全てのパッケージ，ライブラリ，およびそれらが提供するインターフェース，構造体，メソッドの詳細を網羅している．

## 📜 記述ルール
1.  **詳細な仕様**: 使用しているインターフェース，構造体，およびそのメソッドまで詳細に記述する．
2.  **種別の明記**: 以下のアイコンとタグを用いて，要素の種類が一目で判別できるように表記する．
    *   🔹 **[Interface]**: Go インターフェース
    *   📦 **[Struct]**: Go 構造体
    *   ⚙️ **[Function]**: 関数（独立したもの）
    *   🔧 **[Method]**: メソッド（構造体やクラスに紐づくもの）
    *   🏷️ **[Constant / Field]**: 定数や構造体のフィールド変数
    *   ⚛️ **[Component]**: React コンポーネント
    *   🎣 **[Hook]**: React フック
    *   📜 **[Type / Interface]**: TypeScript の型またはインターフェース
3.  **既存内容の維持**: 既存の記述を削除せず，常に追記・更新を行う．
4.  **記述スタイル**: 「である」調を使用し，句読点は「．」「，」に統一する．
5.  **リンクの明示**: パッケージ名および公式リファレンスへの URL を必ず記載する．

---

## 1. Web フレームワーク (Echo v4)
**Package**: `github.com/labstack/echo/v4`  
**Reference**: [pkg.go.dev/github.com/labstack/echo/v4](https://pkg.go.dev/github.com/labstack/echo/v4)

### 📦 `type Echo struct`
HTTP サーバーの本体を管理する構造体である．
*   ⚙️ `func New() *Echo`: インスタンスを生成する．
*   🔧 `func (e *Echo) Start(address string) error`: サーバーを起動する．
*   🔧 `func (e *Echo) GET/POST/PATCH/DELETE(path string, h HandlerFunc, ...)`: ルーティングを登録する．
*   🔧 `func (e *Echo) Use(middleware ...MiddlewareFunc)`: ミドルウェアを登録する．

### 🔹 `type Context interface`
各 HTTP リクエストのコンテキスト（状態）を保持するインターフェースである．
*   🔧 `func (c Context) Param(name string) string`: URL 内のパスパラメータ（`:id` 等）を取得する．
*   🔧 `func (c Context) QueryParam(name string) string`: クエリパラメータ（`?user_id=...` 等）を取得する．
*   🔧 `func (c Context) Bind(i interface{}) error`: リクエストボディ（JSON 等）を構造体に変換する．
*   🔧 `func (c Context) JSON(code int, i interface{}) error`: 構造体を JSON 変換して送信する．
*   🔧 `func (c Context) Request() *http.Request`: 現在の HTTP リクエストオブジェクトを返す．
*   🔧 `func (c Context) Redirect(code int, url string) error`: 指定した URL へリダイレクトする．

---

## 2. データベース ORM (GORM)
**Package**: `gorm.io/gorm`  
**Reference**: [pkg.go.dev/gorm.io/gorm](https://pkg.go.dev/gorm.io/gorm)
### 📦 `type DB struct`
データベース操作を管理するメインの構造体である．
*   ⚙️ `func Open(dialector Dialer, config *Config) (*DB, error)`: 接続を開く．
*   🔧 `func (db *DB) AutoMigrate(dst ...interface{}) error`: テーブルの自動作成・更新を行う．

### ⚙️ `package postgres`
PostgreSQL ドライバの設定を行う．
*   📦 `type Config struct`: 接続の詳細設定を保持する．
    *   🏷️ `PreferSimpleProtocol bool`: プリペアドステートメントを使用せず，シンプルなプロトコルを強制する設定である．Supabase 等の Transaction プーリング環境でのエラー防止に必須である．

---

## 3. 生成 AI (Google Generative AI)
*   🏷️ `gorm:"primaryKey"`: 主キーとして指定する．
*   🏷️ `gorm:"unique"`: ユニーク制約を付与する．
*   🏷️ `gorm:"many2many:table_name"`: 多対多の関係を中間テーブルとして定義する．
*   🏷️ `ErrRecordNotFound`: 検索結果が 0 件の際に返されるエラー定数である．

---

## 3. 生成 AI (Google Generative AI)
**Package**: `github.com/google/generative-ai-go/genai`  
**Reference**: [pkg.go.dev/github.com/google/generative-ai-go/genai](https://pkg.go.dev/github.com/google/generative-ai-go/genai)

### 📦 `type Client struct`
Gemini API との通信用クライアントである．
*   ⚙️ `func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error)`: クライアントを生成する．
*   🔧 `func (c *Client) GenerativeModel(name string) *GenerativeModel`: 特定のモデルを取得する．
*   🔧 `func (c *Client) Close() error`: 接続を終了する．

### 📦 `type GenerativeModel struct`
AI による文章生成を行う構造体である．
*   🔧 `func (m *GenerativeModel) GenerateContent(ctx context.Context, parts ...Part) (*GenerateContentResponse, error)`: テキストを生成する．
*   🏷️ `ResponseMIMEType string`: レスポンスの形式（`application/json` 等）を指定する．

---

## 4. LINE Messaging API
**Package**: `github.com/line/line-bot-sdk-go/v8/linebot/messaging_api`

### 📦 `type MessagingApiAPI struct`
*   ⚙️ `func NewMessagingApiAPI(channelToken string) (*MessagingApiAPI, error)`: クライアントを生成する．
*   🔧 `func (client *MessagingApiAPI) PushMessage(req *PushMessageRequest, xRetryKey string)`: 特定の ID（グループ等）にプッシュ通知を送る．

---

## 5. Web Push 通知 (webpush-go)
**Package**: `github.com/SherClockHolmes/webpush-go`

### ⚙️ `func GenerateVAPIDKeys() (privateKey, publicKey string, err error)`
キーペアを生成する．

### ⚙️ `func SendNotification(message []byte, s *Subscription, options *Options)`
指定されたブラウザ購読情報に対して通知を送信する．

### 📦 `type Subscription struct`
ブラウザから提供される通知宛先情報（Endpoint, Keys）を保持する．

---

## 6. OAuth 2.0 認証 (oauth2)
**Package**: `golang.org/x/oauth2`

### 📦 `type Config struct`
*   🔧 `func (c *Config) AuthCodeURL(state string, ...) string`: 認可 URL を生成する．
*   🔧 `func (c *Config) Exchange(ctx context.Context, code string, ...) (*Token, error)`: トークンを取得する．
*   🔧 `func (c *Config) Client(ctx context.Context, t *Token) *http.Client`: 認証済みクライアントを生成する．

---

## 7. Google Classroom API
**Package**: `google.golang.org/api/classroom/v1`

### 📦 `type Service struct`
*   ⚙️ `func NewService(ctx context.Context, opts ...option.ClientOption) (*Service, error)`: サービスを生成する．

### 📦 `type CoursesService struct`
*   🔧 `func (r *CoursesService) List() *CoursesListCall`: コース一覧取得を作成する．
*   🔧 `func (c *CoursesListCall) CourseStates(states ...string) *CoursesListCall`: 状態（`ACTIVE` 等）で絞り込む．

### 📦 `type CourseWorkService struct`
*   🔧 `func (r *CourseWorkService) List(courseId string) *CourseWorkListCall`: 指定したコース内の課題一覧を取得する．

### 📦 `type StudentSubmissionsService struct`
*   🔧 `func (r *StudentSubmissionsService) List(courseId string, courseWorkId string) *StudentSubmissionsListCall`: 指定した課題に対するユーザーの提出状況を取得する．
*   🔧 `func (c *StudentSubmissionsListCall) Context(ctx context.Context) *StudentSubmissionsListCall`: コンテキストを適用する．
*   🔧 `func (c *StudentSubmissionsListCall) Do(opts ...googleapi.CallOption) (*ListStudentSubmissionsResponse, error)`: API 呼び出しを実行する．

---

## 8. 標準ライブラリ (Standard Libraries)

### ⚙️ `package context`
*   ⚙️ `func Background() Context`: 起点のコンテキストを生成する．
*   ⚙️ `func WithTimeout(parent Context, timeout Duration) (Context, CancelFunc)`: 制限時間付きのコンテキストを生成する．
*   ⚙️ `func WithCancel(parent Context) (Context, CancelFunc)`: 手動キャンセル可能なコンテキストを生成する．

### ⚙️ `package time`
*   ⚙️ `func Now() Time`: 現在時刻を取得する．
*   ⚙️ `func LoadLocation(name string) (*Location, error)`: 指定した名前のタイムゾーン情報を取得する．
*   ⚙️ `func Parse(layout, value string) (Time, error)`: 文字列から時刻へ変換する．
*   🔧 `func (t Time) Format(layout string) string`: 時刻から文字列へ変換する．
*   🔧 `func (t Time) After(u Time) bool`: 時刻の前後を比較する．
*   🔧 `func (t Time) Local() Time`: ローカルタイム（`time.Local`）に変換する．
*   🔧 `func (t Time) IsZero() bool`: 時刻がゼロ値（`0001-01-01`）であるか判定する．
*   ⚙️ `func Since(t Time) Duration`: 経過時間を計測する．
*   ⚙️ `func NewTicker(d Duration) *Ticker`: 定期実行タイマーを生成する．
*   🏷️ `Local *Location`: Go プロセス全体のデフォルトタイムゾーンを保持する変数である．
*   🏷️ `RFC3339 string`: 標準的な日時フォーマットのレイアウト定数である．

### ⚙️ `package strings`
*   ⚙️ `func Contains(s, substr string) bool`: 文字列 `s` 内に `substr` が含まれているか判定する．

### ⚙️ `package fmt`
*   ⚙️ `func Errorf(format string, a ...any) error`: 元のエラーを `%w` でラップして新しいエラーを生成する．
*   ⚙️ `func Sprint(a ...any) string`: 引数を文字列に変換して連結する．

### ⚙️ `package log`
*   ⚙️ `func Println(v ...any)`: 標準出力にログを書き出す．
*   ⚙️ `func Fatalf(format string, v ...any)`: ログ出力後にプログラムを強制終了する．

### ⚙️ `package encoding/json`
*   ⚙️ `func Unmarshal(data []byte, v any) error`: JSON バイト列を構造体に変換する．
*   ⚙️ `func Marshal(v any) ([]byte, error)`: 構造体を JSON バイト列に変換する．
*   📦 `type Decoder struct`: ストリームから順次 JSON を読み取る．

---

## 9. ユーティリティ (uuid)
**Package**: `github.com/google/uuid`

### ⚙️ `func New().String() string`
ランダムな一意の ID 文字列を生成する．

---

## 10. ドメインモデル拡張 (Domain Specifics)

### 📦 `type Task struct`
課題の基本情報と進捗を管理する中心的な構造体である．
*   🏷️ `ExternalID string`: Google Classroom 側の ID（重複防止用）である．
*   🏷️ `Deadline time.Time`: 提出期限である．西暦 1 年（`IsZero`）は「未定」を意味する．
*   🏷️ `IsLMSDeadlineSet bool`: 外部 LMS 側で最初から期限があったかのフラグである．
*   🏷️ `Recurrence RecurrenceSettings`: 繰り返しのルールを保持する JSON オブジェクトである．
*   🏷️ `UserProgress []*TaskUserProgress`: 各メンバーの完了状態のリストである．

### 📦 `type TaskUserProgress struct`
特定の課題に対する個別のユーザーの完了状態を管理する構造体である．
*   🏷️ `TaskID string`: 紐づく課題の ID である．
*   🏷️ `UserID string`: 対象ユーザーの ID である．
*   🏷️ `UserName string`: 表示用のユーザー名である．
*   🏷️ `IsCompleted bool`: 完了フラグである．
*   🏷️ `UpdatedAt time.Time`: 完了ボタンが押された最新の時刻である．

### 📦 `type Group struct`
グループ（部屋）の情報を保持する構造体である．
*   🏷️ `InviteCode string`: 8 桁の参加用招待コードである．
*   🏷️ `RemindIntervals []int`: リマインド通知を飛ばすタイミング（分前）のリストである．
*   🏷️ `AICharacter string`: AI の性格設定である．
*   🏷️ `SummaryMorningTime string`: 朝のサマリー送信時刻（HH:mm）である．
*   🏷️ `SummaryEveningTime string`: 夜のサマリー送信時刻（HH:mm）である．
*   🏷️ `Users []*User`: 所属しているメンバーのリストである．

### 📦 `type NotificationLog struct`
送信された通知の履歴を保持する構造体である．
*   🏷️ `Type string`: 通知の種別（`remind`, `sos`, `summary`）である．
*   🏷️ `Message string`: 通知されたメッセージの本文である．

### 📦 `type User struct`
システム利用者（個人）の情報である．
*   🏷️ `GoogleTokenExpiry time.Time`: Google OAuth トークンの有効期限である．
*   🏷️ `Groups []*Group`: 所属している部屋のリストである．

### 📦 `type WakeupCheck struct`
起床確認のスケジュールと結果を保持する構造体である．
*   🏷️ `TargetTime time.Time`: 起床を約束した時刻である．
*   🏷️ `Status string`: `pending`（待ち），`confirmed`（成功），`alerted`（失敗）のいずれかである．

---

## 11. サービス・インターフェース

### 🔹 `type LMSService interface`
外部学習システム（現在は Google Classroom）との通信を抽象化する．
*   🔧 `func FetchTasks(ctx, userID) ([]*Task, error)`: ユーザーの全アクティブコースから課題を取得する．
*   🔧 `func GetProviderName() string`: "google_classroom" 等の識別名を返す．

### 🔹 `type AIService interface`
AI による文章生成の窓口である．
*   🔧 `func GenerateRemindMessage(ctx, task, style) (string, error)`: 課題内容に応じたリマインド文を生成する．
*   🔧 `func GenerateGroupSummaryMessage(ctx, workload, style) (string, error)`: グループ全体の課題状況を要約するメッセージを生成する．

### 🔹 `type NotificationService interface`
LINE や Web Push などの通知手段を抽象化する．
*   🔧 `func SendGroupMessage(ctx, targetID, message) error`: グループ全体へ通知する．
*   🔧 `func SendDirectMessage(ctx, userID, message, targetURL) error`: 個人へ直接通知し，クリック時の遷移先を指定する．

### 🔹 `type GroupRepository interface`
*   🔧 `func FindAllGroups(ctx context.Context) ([]*Group, error)`: 全てのグループを取得する（定期サマリー用）．
*   🔧 `func AddUserToGroup(ctx, groupID, userID) error`: 指定したユーザーをグループに関連付け，中間テーブルへ保存する．
*   🔧 `func FindByUserID(ctx, userID) ([]*Group, error)`: ユーザーが所属しているグループ一覧を取得する．

### 🔹 `type NotificationLogRepository interface`
*   🔧 `func Save(ctx context.Context, log *NotificationLog) error`: 通知ログを保存する．
*   🔧 `func FindByGroupID(ctx, groupID, limit) ([]*NotificationLog, error)`: 履歴を取得する．
