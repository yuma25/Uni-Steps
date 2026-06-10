# Uni-Steps 技術リファレンス集

開発で使用している主要なパッケージやライブラリの公式リファレンスを詳細にまとめている．

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
Echo フレームワークのメインとなる構造体である．
*   ⚙️ `func New() *Echo`: 新しい Echo インスタンスを生成する．
*   🔧 `func (e *Echo) Start(address string) error`: HTTP サーバーを起動する．
*   🔧 `func (e *Echo) GET/POST/PATCH/DELETE(...)`: 各 HTTP メソッドのルーティングを登録する．

### 🔹 `type Context interface`
各リクエストの情報を保持するインターフェースである．
*   🔧 `func (c Context) Param(name string) string`: URL パラメータを取得する．
*   🔧 `func (c Context) Bind(i interface{}) error`: リクエストボディを構造体に変換する．
*   🔧 `func (c Context) JSON(code int, i interface{}) error`: JSON レスポンスを送信する．
*   🔧 `func (c Context) Redirect(code int, url string) error`: 指定した URL へリダイレクトする．

---

## 2. 生成 AI (Gemini)
**Package**: `github.com/google/generative-ai-go/genai`  
**Reference**: [pkg.go.dev/github.com/google/generative-ai-go/genai](https://pkg.go.dev/github.com/google/generative-ai-go/genai)

### 📦 `type Client struct`
Gemini API クライアントである．
*   ⚙️ `func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error)`: クライアントを生成する．

### 📦 `type GenerativeModel struct`
*   🔧 `func (m *GenerativeModel) GenerateContent(ctx context.Context, parts ...Part) (*GenerateContentResponse, error)`: AI にコンテンツ生成を依頼する．

---

## 3. データベース ORM (gorm)
**Package**: `gorm.io/gorm`  
**Reference**: [pkg.go.dev/gorm.io/gorm](https://pkg.go.dev/gorm.io/gorm)

### 📦 `type DB struct`
*   🔧 `func (db *DB) AutoMigrate(dst ...interface{}) error`: テーブルを自動作成・更新する．
*   🔧 `func (db *DB) Save(value interface{}) *DB`: レコードを保存する．
*   🔧 `func (db *DB) Where(query interface{}, args ...interface{}) *DB`: 検索条件を指定する．

### 💡 GORM のシリアライザ (Serializers)
*   🏷️ `gorm:"serializer:json"`: Go のスライス等を JSON 文字列として DB に保存する機能である．

### 💡 トラブルシューティング：Supabase 接続
*   6543 ポート（プーラー）を使用する際は **`PreferSimpleProtocol: true`** を指定する必要がある．

---

## 4. LINE Messaging API
**Package**: `github.com/line/line-bot-sdk-go/v8/linebot/messaging_api`

### 📦 `type MessagingApiAPI struct`
*   🔧 `func (client *MessagingApiAPI) PushMessage(pushReq *PushMessageRequest, xRetryKey string)`: メッセージを送信する．

---

## 5. Web Push 通知 (webpush-go)
**Package**: `github.com/SherClockHolmes/webpush-go`

### ⚙️ `func SendNotification(message []byte, s *Subscription, options *Options)`
指定されたブラウザへ通知を送信する．

### 💡 配信の仕組み（グループ配信）
Web Push は本来個人宛だが，Uni-Steps では**ループ処理を用いてグループ全員のトークンへ個別に送信する**ことで，実質的な一斉発信を実現している．

---

## 6. 標準ライブラリ (Standard Libraries)

### ⚙️ `package context`
デッドラインやキャンセル信号を管理する．
*   `func WithTimeout(...)`: タイムアウト付き Context を生成する．

### ⚙️ `package fmt`
入出力や文字列フォーマットを行う．
*   `func Errorf(...)`: エラーをラップして生成する．

### ⚙️ `package io`
入出力の基本インターフェースを提供する．
*   🔹 `type Reader interface`: データの読み込みを抽象化する．

---

## 7. JSON (encoding/json)
*   📦 `type Decoder struct`: ストリーム（レスポンスボディ等）から JSON を効率的に読み取る．

---

## 8. UUID (github.com/google/uuid)
*   ⚙️ `func NewString() string`: ランダムな ID 文字列を生成する．

---

## 9. フロントエンド技術 (Frontend)

### 🛠️ Vite
ビルドおよび開発サーバーである．

### ⚛️ React
*   🎣 `useState`: コンポーネントの状態管理である．
*   🎣 `useEffect`: 副作用（API 通信等）の実行である．

### 📜 TypeScript
*   📜 `interface`: データ形状の定義である．
*   ⚙️ `import type`: 型のみを安全にインポートする構文である（画面真っ白エラーの防止に必須）．

### 📚 Axios
バックエンド API との通信用クライアントである．

### 📚 React Router
*   🎣 `useNavigate`: ページ遷移である．
*   🎣 `useSearchParams`: クエリパラメータの操作である．
