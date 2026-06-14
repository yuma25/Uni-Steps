# Uni-Steps インフラ層 リファレンスマニュアル

本ドキュメントは，`infrastructure/` 配下で定義されている各パッケージ・ファイルごとの構造体とメソッドの仕様を網羅したものである．

---

## 1. infrastructure/ai (AI サービス実装)

### 📦 `type GeminiService struct`
Google Gemini API を使用して，ドメイン層の `AIService` インターフェースを実装する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `client` | `*genai.Client` | Google 公式の Gemini API クライアント． |
| `model` | `*genai.GenerativeModel` | 生成に使用する特定の AI モデル（Flash等）． |

*   ⚙️ **関数: `NewGeminiService(client, modelName) *GeminiService`**
    *   **概要**: 認証済みクライアントとモデル名（例: "gemini-2.0-flash"）を指定してインスタンスを生成する．
    *   **引数**: `client` (`*genai.Client`), `modelName` (`string`)．
    *   **戻り値**: 生成された `*GeminiService`．
*   🔧 **メソッド: `GenerateRemindMessage(ctx, task, style) (string, error)`**
    *   **概要**: 性格設定に応じたプロンプトを組み立て，Gemini API を呼び出してリマインド文を生成する．
    *   **引数**: `ctx`, `task` (`*domain.Task`), `style` (`string`)．
    *   **戻り値**: 生成されたテキスト，およびエラー．
*   🔧 **メソッド: `GenerateGroupSummaryMessage(ctx, workloadSummary, style) (string, error)`**
    *   **概要**: グループ全体の進捗状況（集計済みテキスト）を元に，Gemini API を呼び出して要約文を生成する．
    *   **引数**: `ctx`, `workloadSummary` (`string`), `style` (`string`)．
    *   **戻り値**: 生成されたテキスト，およびエラー．

---

## 2. infrastructure/db (リポジトリ実装)

GORM を用いて，ドメイン層で定義された各リポジトリインターフェースを PostgreSQL に対して具現化する．

### 📁 infrastructure/db/task_repository.go

### 📦 `type taskRepository struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `db` | `*gorm.DB` | GORM データベース接続インスタンス． |

*   ⚙️ **関数: `NewTaskRepository(db) domain.TaskRepository`**
    *   **引数**: `db` (`*gorm.DB`)．
    *   **戻り値**: 実装された `domain.TaskRepository` インターフェース．
*   🔧 **メソッド: `Save(ctx, task) error`**
    *   **概要**: 課題の基本情報とメンバー進捗（`UserProgress`）を保存する．
    *   **こだわり**: トランザクションを使用し，`UserProgress` は一度全削除してから再登録することで，メンバーからの離脱も安全に同期する．
    *   **引数**: `ctx`, `task` (`*domain.Task`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `FindByID(ctx, id) (*domain.Task, error)`**
    *   **概要**: 指定 ID の課題を 1 件，進捗データも含めて取得する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 課題データ，およびエラー．
*   🔧 **メソッド: `FindByExternalID(ctx, externalID) (*domain.Task, error)`**
    *   **概要**: Classroom 側の課題 ID をキーに検索する．
    *   **引数**: `ctx`, `externalID` (`string`)．
    *   **戻り値**: 課題データ，およびエラー．
*   🔧 **メソッド: `FindByGroupID(ctx, groupID) ([]*domain.Task, error)`**
    *   **概要**: 部屋内の全課題を期限昇順（期限なしは末尾）で取得する．
    *   **引数**: `ctx`, `groupID` (`string`)．
    *   **戻り値**: 課題リスト，およびエラー．
*   🔧 **メソッド: `FindApproachingDeadlines(ctx, until) ([]*domain.Task, error)`**
    *   **概要**: 期限が迫っている未完了課題を抽出する（AI リマインド用）．
    *   **引数**: `ctx`, `until` (`time.Time`)．
    *   **戻り値**: 課題リスト，およびエラー．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: トランザクションを使用して，課題本体と紐づく進捗データを完全に抹消する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．

### 📁 infrastructure/db/group_repository.go

### 📦 `type groupRepository struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `db` | `*gorm.DB` | GORM データベース接続インスタンス． |

*   ⚙️ **関数: `NewGroupRepository(db) domain.GroupRepository`**
    *   **引数**: `db` (`*gorm.DB`)．
    *   **戻り値**: 実装された `domain.GroupRepository` インターフェース．
*   🔧 **メソッド: `Save(ctx, group) error`**
    *   **概要**: 部屋の設定情報を永続化する．
    *   **こだわり**: `Omit("Users")` により，設定変更時にメンバー構成が意図せず書き換わるのを防ぐ．新規作成時は自動的にレコードを作成する．
    *   **引数**: `ctx`, `group` (`*domain.Group`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `FindAllGroups(ctx) ([]*domain.Group, error)`**
    *   **概要**: システム上の全ての部屋を取得する．
    *   **引数**: `ctx`．
    *   **戻り値**: 部屋リスト，およびエラー．
*   🔧 **メソッド: `FindByID(ctx, id) (*domain.Group, error)`**
    *   **概要**: 部屋の情報を取得し，所属メンバー（`Users`）も Preload して展開する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 取得した部屋データ，およびエラー．
*   🔧 **メソッド: `FindByInviteCode(ctx, code) (*domain.Group, error)`**
    *   **概要**: 招待コードで部屋を特定する．メンバー情報も Preload する．
    *   **引数**: `ctx`, `code` (`string`)．
    *   **戻り値**: 取得した部屋データ，およびエラー．
*   🔧 **メソッド: `FindByUserID(ctx, userID) ([]*domain.Group, error)`**
    *   **概要**: `user_groups` 中間テーブルを JOIN し，その人が入っている全グループを返す．メンバー情報も Preload する．
    *   **引数**: `ctx`, `userID` (`string`)．
    *   **戻り値**: 部屋リスト，およびエラー．
*   🔧 **メソッド: `RemoveUser(ctx, groupID, userID) error`**
    *   **概要**: GORM の Association Delete を使い，中間テーブルからユーザーとグループの紐付けを解除する．
    *   **引数**: `ctx`, `groupID` (`string`), `userID` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: トランザクションを使用して，部屋の全メンバーとの紐付けを解除してから部屋本体を削除する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．

### 📁 infrastructure/db/user_repository.go

### 📦 `type userRepository struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `db` | `*gorm.DB` | GORM データベース接続インスタンス． |

*   ⚙️ **関数: `NewUserRepository(db) domain.UserRepository`**
    *   **引数**: `db` (`*gorm.DB`)．
    *   **戻り値**: 実装された `domain.UserRepository` インターフェース．
*   🔧 **メソッド: `Save(ctx, user) error`**
    *   **概要**: ユーザー情報（トークン含む）を保存・更新する．
    *   **引数**: `ctx`, `user` (`*domain.User`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `UpdateWebPushToken(ctx, userID, token) error`**
    *   **概要**: パフォーマンス向上のため，`web_push_token` カラムのみを直接 Update する．
    *   **引数**: `ctx`, `userID` (`string`), `token` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `FindByID(ctx, id) (*domain.User, error)`**
    *   **概要**: ID を指定してユーザーを取得する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 取得したユーザーデータ，およびエラー．
*   🔧 **メソッド: `FindByEmail(ctx, email) (*domain.User, error)`**
    *   **概要**: ログイン時の存在チェックとしてメールアドレス検索を提供する．
    *   **引数**: `ctx`, `email` (`string`)．
    *   **戻り値**: 取得したユーザーデータ，およびエラー．

### 📁 infrastructure/db/wakeup_repository.go

### 📦 `type wakeupRepository struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `db` | `*gorm.DB` | GORM データベース接続インスタンス． |

*   ⚙️ **関数: `NewWakeupRepository(db) domain.WakeupRepository`**
    *   **引数**: `db` (`*gorm.DB`)．
    *   **戻り値**: 実装された `domain.WakeupRepository` インターフェース．
*   🔧 **メソッド: `Save(ctx, check) error`**
    *   **概要**: 起床確認レコードを保存または更新する．
    *   **引数**: `ctx`, `check` (`*domain.WakeupCheck`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: 指定 ID の起床確認レコードを削除する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `FindByID(ctx, id) (*domain.WakeupCheck, error)`**
    *   **概要**: ID を指定して特定の起床確認レコードを取得する．
    *   **引数**: `ctx`, `id` (`string`)．
    *   **戻り値**: 取得したデータ，およびエラー．
*   🔧 **メソッド: `FindPendingByTime(ctx, now) ([]*domain.WakeupCheck, error)`**
    *   **概要**: 判定時刻を過ぎても status が `pending` の起床予約を探し出す．
    *   **引数**: `ctx`, `now` (`time.Time`)．
    *   **戻り値**: 起床予約リスト，およびエラー．
*   🔧 **メソッド: `FindActiveByUser(ctx, userID) ([]*domain.WakeupCheck, error)`**
    *   **概要**: 特定ユーザーの未完了な（`pending` 状態の）起床予約を取得する．
    *   **引数**: `ctx`, `userID` (`string`)．
    *   **戻り値**: 起床予約リスト，およびエラー．
*   🔧 **メソッド: `FindActiveByGroup(ctx, groupID) ([]*domain.WakeupCheck, error)`**
    *   **概要**: 本日の日付分（target_time が今日以降）であれば，完了済みも含めてすべて取得し，ダッシュボードに表示させる．
    *   **引数**: `ctx`, `groupID` (`string`)．
    *   **戻り値**: 起床状況リスト，およびエラー．

### 📁 infrastructure/db/notification_log_repository.go

### 📦 `type notificationLogRepository struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `db` | `*gorm.DB` | GORM データベース接続インスタンス． |

*   ⚙️ **関数: `NewNotificationLogRepository(db) domain.NotificationLogRepository`**
    *   **引数**: `db` (`*gorm.DB`)．
    *   **戻り値**: 実装された `domain.NotificationLogRepository` インターフェース．
*   🔧 **メソッド: `Save(ctx, log) error`**
    *   **概要**: 新しい通知ログを永続化する．
    *   **引数**: `ctx`, `log` (`*domain.NotificationLog`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `FindByGroupID(ctx, groupID, limit) ([]*domain.NotificationLog, error)`**
    *   **概要**: 最新の通知履歴を日付の降順で取得する．
    *   **引数**: `ctx`, `groupID` (`string`), `limit` (`int`)．
    *   **戻り値**: ログリスト，およびエラー．

---

## 3. infrastructure/line (LINE 通知実装)

### 📦 `type LineService struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `groupRepo` | `domain.GroupRepository` | グループごとのトークン取得に使用． |

*   ⚙️ **関数: `NewLineService(gr) *LineService`**
    *   **概要**: `LineService` の新しいインスタンスを生成する．
    *   **引数**: `gr` (`domain.GroupRepository`)．送信時にグループごとの個別トークンを取得するために必要．
    *   **戻り値**: 生成された `*LineService`．
*   🔧 **メソッド: `SendGroupMessage(ctx, targetID, message) error`**
    *   **概要**: 指定されたグループ ID に紐づく LINE Bot 設定を取得し，その Bot としてグループへプッシュ通知を送信する（BYOT）．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `targetID`: `string` 型．送信対象となるグループの内部 ID．
        *   `message`: `string` 型．送信するメッセージ本文．
    *   **戻り値**: 成功時は nil，設定未完了時はスキップして nil，通信失敗時はエラーを返す．
*   🔧 **メソッド: `SendDirectMessage(ctx, userID, message, targetURL) error`**
    *   **概要**: 個人宛の LINE 通知．現在の仕様では未サポート．
    *   **注意**: 呼び出すと常にエラーを返す．
    *   **引数**: `ctx`, `userID`, `message`, `targetURL`．
    *   **戻り値**: 常にエラーを返す．

---

## 4. infrastructure/lms (外部システム連携)

### 📦 `type GoogleClassroomService struct`
Google Classroom API を使用して，ドメイン層の `LMSService` インターフェースを実装する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | リフレッシュされたトークンの保存に使用． |
| `oauthCfg` | `*oauth2.Config` | Google OAuth 2.0 の設定情報． |

*   ⚙️ **関数: `NewGoogleClassroomService(ur, cfg) *GoogleClassroomService`**
    *   **概要**: 必要なリポジトリと設定を注入し，インスタンスを生成する．
    *   **引数**: `ur` (`domain.UserRepository`), `cfg` (`*oauth2.Config`)．
    *   **戻り値**: 生成された `*GoogleClassroomService`．
*   🔧 **メソッド: `FetchTasks(ctx, userID) ([]*Task, error)`**
    *   **概要**: ユーザーのアクセストークンを用いて Google Classroom API を叩き，全コースの課題（CourseWork）と提出状況（StudentSubmissions）を取得・変換する．
    *   **パフォーマンス最適化**: Goroutine と WaitGroup を用いて，各コースのデータ取得を並列化している．これによりコース数が多い場合でも同期時間が劇的に短縮される．
    *   **自動リフレッシュ**: トークンが期限切れの場合，内部の `persistentTokenSource` が自動的にリフレッシュを行い，新しいトークンをデータベースへ再保存する．
    *   **引数**: `ctx`, `userID` (`string`)．
    *   **戻り値**: 取得した `domain.Task` リスト（進捗状況含む），およびエラー．
*   🔧 **メソッド: `GetProviderName() string`**
    *   **概要**: このサービスが提供する課題の出典名（プロバイダ名）を返す．
    *   **役割**: ここで返される値（`google_classroom`）は，`Task` エンティティの `Source` フィールドに記録される．これにより，フロントエンドでの「Classroom アイコン」の表示や，「外部連携課題なのでタイトル編集不可」といった権限制御の判定基準となる．
    *   **戻り値**: `domain.SourceGoogleClassroom` ("google_classroom")．

### 📦 `type persistentTokenSource struct` (内部ヘルパー)
`oauth2.TokenSource` インターフェースをラップし，トークンリフレッシュ時の永続化を自動化する内部構造体である．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ctx` | `context.Context` | 保存処理等に使用するコンテキスト． |
| `userID` | `string` | トークンを更新する対象ユーザーの ID． |
| `userRepo` | `domain.UserRepository` | 新しいトークンを DB へ保存するために使用． |
| `oauthCfg` | `*oauth2.Config` | リフレッシュ処理用の OAuth2 設定． |
| `source` | `oauth2.TokenSource` | 元となる Google 認証の TokenSource． |

*   🔧 **メソッド: `Token() (*oauth2.Token, error)`**
    *   **概要**: 有効なトークンを取得する．内部で `source.Token()` を呼び出し，もしトークンが新しく発行（リフレッシュ）されていた場合は，その場で `userRepo.Save` を実行して DB を最新状態に保つ．
    *   **戻り値**: 有効なアクセストークン，およびエラー．

---

## 5. infrastructure/notification (通知統合)

### 📦 `type CompositeNotificationService struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `lineService` | `*line.LineService` | LINE 通知担当． |
| `webPushService` | `*webpush.WebPushService` | Web Push 通知担当． |

*   ⚙️ **関数: `NewCompositeNotificationService(ls, wps) *CompositeNotificationService`**
    *   **引数**: `ls` (`*line.LineService`), `wps` (`*webpush.WebPushService`)．
    *   **戻り値**: 生成された `*CompositeNotificationService`．
*   🔧 **メソッド: `SendGroupMessage(ctx, targetID, message) error`**
    *   **概要**: LINE サービスに処理を委譲する．
    *   **引数**: `ctx`, `targetID` (`string`), `message` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `SendDirectMessage(ctx, userID, message, targetURL) error`**
    *   **概要**: Web Push サービスに処理を委譲する．
    *   **引数**: `ctx`, `userID` (`string`), `message` (`string`), `targetURL` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．

---

## 6. infrastructure/scheduler (タイマー管理)

### 📦 `type InMemScheduler struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | 利用者データの取得に使用． |
| `groupRepo` | `domain.GroupRepository` | 部屋データの取得に使用． |
| `aiService` | `domain.AIService` | メッセージ生成に使用． |
| `notifService` | `domain.NotificationService` | 送信に使用． |
| `logRepo` | `domain.NotificationLogRepository` | 送信履歴の保存に使用． |
| `timers` | `map[string]*time.Timer` | メモリ上のアクティブなタイマー管理マップ． |

*   ⚙️ **関数: `NewInMemScheduler(ur, gr, ai, ns, lr) *InMemScheduler`**
    *   **引数**: `ur`, `gr`, `ai`, `ns`, `lr`．
    *   **戻り値**: 生成された `*InMemScheduler`．
*   🔧 **メソッド: `ScheduleTaskRemind(ctx, task, userID, interval, style, runAt) error`**
    *   **概要**: 未来の時刻にリマインドを実行するタイマーをセットする．実行時に AI で文面を生成し送信する．
    *   **引数**: `ctx`, `task` (`*domain.Task`), `userID` (`string`), `interval` (`int`), `style` (`string`), `runAt` (`time.Time`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `CancelTaskReminds(ctx, taskID, userID) error`**
    *   **概要**: 指定課題・ユーザーの予約をすべて破棄する．
    *   **引数**: `ctx`, `taskID` (`string`), `userID` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `ScheduleWakeupSOS(ctx, wakeupID, userID, groupID, runAt) error`**
    *   **概要**: 起床失敗時の SOS 送信タイマーをセットする．
    *   **引数**: `ctx`, `wakeupID` (`string`), `userID` (`string`), `groupID` (`string`), `runAt` (`time.Time`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．
*   🔧 **メソッド: `CancelWakeupSOS(ctx, wakeupID) error`**
    *   **概要**: SOS 予約を取り消す．
    *   **引数**: `ctx`, `wakeupID` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．

---

## 7. infrastructure/webpush (ブラウザ通知実装)

### 📦 `type WebPushService struct`
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | ユーザーの通知トークン取得に使用． |
| `publicKey` | `string` | VAPID 公開鍵． |
| `privateKey` | `string` | VAPID 秘密鍵． |
| `contact` | `string` | 連絡先情報． |

*   ⚙️ **関数: `NewWebPushService(ur, pubKey, privKey, contact) *WebPushService`**
    *   **引数**: `ur`, `pubKey`, `privKey`, `contact`．
    *   **戻り値**: 生成された `*WebPushService`．
*   🔧 **メソッド: `SendDirectMessage(ctx, userID, message, targetURL) error`**
    *   **概要**: ユーザーのトークンを JSON デコードして送信を実行する．
    *   **自己修復（Auto-Healing）**: 
        *   **サーバー側**: 410 (Gone) や 404 を受信した場合，トークンが無効であると判断し，即座にデータベースからそのトークンを削除（空文字に更新）する．
        *   **フロントエンド連携**: フロントエンド側はトークン消滅を検知すると，ブラウザの権限がある場合に限り「サイレント再登録（handleSilentResubscribe）」を試みるため，通知機能は自動的に復旧する．
    *   **引数**: `ctx`, `userID` (`string`), `message` (`string`), `targetURL` (`string`)．
    *   **戻り値**: 成功時は nil，失敗時はエラー．

---
*最終更新日: 2026年6月14日*
