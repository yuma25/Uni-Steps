# Uni-Steps インターフェース（ハンドラー）層 リファレンスマニュアル

本ドキュメントは，`interfaces/handler/` 配下で定義されている各ファイルごとの構造体（ハンドラー）と，外部からの HTTP リクエストを処理する API エンドポイントの仕様を網羅したものである．

---

## 1. interfaces/handler/auth_handler.go (認証管理)

### 📦 `type AuthHandler struct`
Google OAuth 2.0 を用いたログインおよびユーザー登録を管理するハンドラーである．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | ユーザー情報の検索・保存を担当． |
| `oauthCfg` | `*oauth2.Config` | Google OAuth の設定情報． |

*   ⚙️ **関数: `NewAuthHandler(e, ur, cfg)`**
    *   **概要**: ハンドラーを初期化し，認証に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `ur` (`domain.UserRepository`), `cfg` (`*oauth2.Config`)．
    *   **戻り値**: なし（Echo インスタンスへのルーティング登録を直接行う）．
*   🔧 **メソッド: `GoogleLogin(c) error`**
    *   **概要**: ユーザーを Google の認可画面へリダイレクトさせる．CSRF 対策として `state` を生成し Cookie に保存する．
    *   **エンドポイント**: `GET /api/auth/google/login`
    *   **引数**: `c` (`echo.Context`)．Echo のリクエストコンテキスト．
    *   **戻り値**: `error` 型．Google へのリダイレクト処理の結果（Echo 仕様）．
*   🔧 **メソッド: `GoogleCallback(c) error`**
    *   **概要**: Google からの認可コードを受け取り，アクセストークンの交換，プロフィール取得，ユーザー登録（またはログイン）を実行する．
    *   **エンドポイント**: `GET /api/auth/google/callback`
    *   **引数**: `c` (`echo.Context`)．
    *   **戻り値**: `error` 型．処理成功時はフロントエンドのポータル画面へリダイレクトし，失敗時はエラーを返す．

---

## 2. interfaces/handler/group_handler.go (部屋管理)

### 📦 `type GroupHandler struct`
部屋の作成，参加，設定変更などの HTTP リクエストを処理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `groupUsecase` | `*usecase.GroupUsecase` | 部屋管理のビジネスロジックを担当． |
| `lmsService` | `domain.LMSService` | 外部 LMS 連携（識別名取得等）を担当． |

*   ⚙️ **関数: `NewGroupHandler(e, gu, ls)`**
    *   **概要**: グループ管理に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `gu` (`*usecase.GroupUsecase`), `ls` (`domain.LMSService`)．
    *   **戻り値**: なし．
*   🔧 **メソッド: `CreateGroup(c) error`**
    *   **概要**: 新しい部屋を作成するリクエストを処理する．
    *   **エンドポイント**: `POST /api/groups`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `name` と `owner_id` を含む必要がある．
    *   **戻り値**: `error` 型．成功時は作成された `Group` オブジェクトを JSON で返却する．
*   🔧 **メソッド: `JoinGroupByInviteCode(c) error`**
    *   **概要**: 招待コードによる部屋への参加リクエストを処理する．
    *   **エンドポイント**: `POST /api/groups/join`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `invite_code` と `user_id` を含む必要がある．
    *   **戻り値**: `error` 型．成功時は参加後の `Group` オブジェクトを JSON で返却する．
*   🔧 **メソッド: `ListUserGroups(c) error`**
    *   **概要**: 指定されたユーザーが所属している部屋の一覧を返す．
    *   **エンドポイント**: `GET /api/users/:userId/groups`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:userId` を使用する．
    *   **戻り値**: `error` 型．成功時は所属グループのリストを返却する．
*   🔧 **メソッド: `UpdateGroupSettings(c) error`**
    *   **概要**: 部屋の名前や AI 設定などの一括更新を行う．オーナー権限が必要．
    *   **エンドポイント**: `PATCH /api/groups/:groupId/settings`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` と，更新内容を含むリクエストボディを使用する．
    *   **戻り値**: `error` 型．成功時は完了メッセージを返却する．
*   🔧 **メソッド: `TransferOwnership(c) error`**
    *   **概要**: オーナー権限を他のメンバーへ譲渡するリクエストを処理する．
    *   **エンドポイント**: `PUT /api/groups/:groupId/owner`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` と，現在のオーナー ID・新しいオーナー ID を含むボディを使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `DeleteGroup(c) error`**
    *   **概要**: 部屋を完全に削除するリクエストを処理する．
    *   **エンドポイント**: `DELETE /api/groups/:groupId`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` と，クエリパラメータ `user_id`（オーナー確認用）を使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `LeaveGroup(c) error`**
    *   **概要**: ユーザーの部屋退出リクエストを処理する．
    *   **エンドポイント**: `DELETE /api/groups/:groupId/users/:userId`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` と `:userId` を使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `GetNotificationLogs(c) error`**
    *   **概要**: 部屋のタイムライン表示用の通知履歴を取得する．
    *   **エンドポイント**: `GET /api/groups/:groupId/notifications`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` を使用する．
    *   **戻り値**: `error` 型．成功時は通知ログのスライスを返却する．

---

## 3. interfaces/handler/notification_handler.go (通知管理)

### 📦 `type NotificationHandler struct`
Web Push の購読やテスト通知のリクエストを処理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | ユーザーのトークン保存を担当． |
| `aiService` | `domain.AIService` | テスト通知用のメッセージ生成を担当． |
| `notifService` | `domain.NotificationService` | 実際の通知送信を担当． |

*   ⚙️ **関数: `NewNotificationHandler(e, ur, ai, ns)`**
    *   **概要**: 通知に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `ur` (`domain.UserRepository`), `ai` (`domain.AIService`), `ns` (`domain.NotificationService`)．
    *   **戻り値**: なし．
*   🔧 **メソッド: `SubscribeWebPush(c) error`**
    *   **概要**: クライアントから Web Push 購読情報を受け取り，ユーザーデータへ保存する．
    *   **エンドポイント**: `POST /api/notifications/subscribe`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `user_id` と `subscription`（トークン）を含む必要がある．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `SendTestNotification(c) error`**
    *   **概要**: 動作確認用のテスト通知を即座に送信するリクエストを処理する．
    *   **エンドポイント**: `POST /api/notifications/test`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `user_id`, `group_id`, `ai_character` を含む必要がある．
    *   **戻り値**: `error` 型．

---

## 4. interfaces/handler/task_handler.go (課題管理)

### 📦 `type TaskHandler struct`
課題の取得，登録，更新，削除，および同期リクエストを処理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `taskUsecase` | `*usecase.TaskUsecase` | 手動課題・進捗管理ロジックを担当． |
| `syncUsecase` | `*usecase.SyncUsecase` | 外部 LMS 同期ロジックを担当． |

*   ⚙️ **関数: `NewTaskHandler(e, tu, su)`**
    *   **概要**: 課題管理に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `tu` (`*usecase.TaskUsecase`), `su` (`*usecase.SyncUsecase`)．
    *   **戻り値**: なし．
*   🔧 **メソッド: `ListTasks(c) error`**
    *   **概要**: グループ内の全課題を取得するリクエストを処理する．
    *   **エンドポイント**: `GET /api/groups/:id/tasks`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:id`（グループ ID）を使用する．
    *   **戻り値**: `error` 型．成功時は課題リストを返却する．
*   🔧 **メソッド: `CreateManualTask(c) error`**
    *   **概要**: UI からの手動課題登録を受け付ける．
    *   **エンドポイント**: `POST /api/tasks/manual`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに課題データ（タイトル等）を含む必要がある．
    *   **戻り値**: `error` 型．成功時は作成された課題オブジェクトを返却する．
*   🔧 **メソッド: `UpdateTask(c) error`**
    *   **概要**: 課題のタイトル，期限，担当者などの更新リクエストを処理する．
    *   **エンドポイント**: `PUT /api/tasks/:id`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:id` と，クエリパラメータ `user_id`（権限チェック用），および更新データボディを使用する．
    *   **戻り値**: `error` 型．成功時は更新後の課題オブジェクト，または自動削除通知を返却する．
*   🔧 **メソッド: `DeleteTask(c) error`**
    *   **概要**: 課題の削除リクエストを処理する．
    *   **エンドポイント**: `DELETE /api/tasks/:id`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:id` と，クエリパラメータ `user_id`（権限チェック用）を使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `ToggleTaskCompletion(c) error`**
    *   **概要**: 特定ユーザーの完了状態切り替えリクエストを処理する．
    *   **エンドポイント**: `PATCH /api/tasks/:id/toggle-completion`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:id` と，リクエストボディに `user_id` を含む必要がある．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `SyncTasks(c) error`**
    *   **概要**: Google Classroom 等との同期実行リクエストを処理する．
    *   **エンドポイント**: `POST /api/tasks/sync`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `user_id` と `group_id` を含む必要がある．
    *   **戻り値**: `error` 型．成功時は同期完了メッセージとタスクリストを返却する．

---

## 5. interfaces/handler/user_handler.go (ユーザー管理)

### 📦 `type UserHandler struct`
ユーザー情報の取得リクエストを処理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `userRepo` | `domain.UserRepository` | ユーザー情報の取得を担当． |

*   ⚙️ **関数: `NewUserHandler(e, ur)`**
    *   **概要**: ユーザー管理に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `ur` (`domain.UserRepository`)．
    *   **戻り値**: なし．
*   🔧 **メソッド: `GetUser(c) error`**
    *   **概要**: 指定された ID のユーザープロフィール（安全な情報のみ）を取得する．
    *   **エンドポイント**: `GET /api/users/:id`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:id` を使用する．
    *   **戻り値**: `error` 型．成功時はユーザー ID，氏名，メールアドレス，プッシュ通知許可状態を JSON で返却する．

---

## 6. interfaces/handler/wakeup_handler.go (起床管理)

### 📦 `type WakeupHandler struct`
起床見守りの予約や報告（チェックイン）のリクエストを処理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `wakeupUsecase` | `*usecase.WakeupUsecase` | 起床管理のビジネスロジックを担当． |

*   ⚙️ **関数: `NewWakeupHandler(e, uc)`**
    *   **概要**: 起床管理に関連するエンドポイントを登録する．
    *   **引数**: `e` (`*echo.Echo`), `uc` (`*usecase.WakeupUsecase`)．
    *   **戻り値**: なし．
*   🔧 **メソッド: `RequestWakeupCheck(c) error`**
    *   **概要**: 新しい起床見守りの予約リクエストを処理する．
    *   **エンドポイント**: `POST /api/wakeup/request`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `user_id`, `group_id`, `target_time`, `grace_minutes` を含む必要がある．
    *   **戻り値**: `error` 型．成功時は作成された起床予約オブジェクトを返却する．
*   🔧 **メソッド: `CheckIn(c) error`**
    *   **概要**: 起床報告を受け付け，進行中の監視を完了させる．
    *   **エンドポイント**: `POST /api/wakeup/checkin`
    *   **引数**: `c` (`echo.Context`)．リクエストボディに `user_id` を含む必要がある．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `GetActiveChecks(c) error`**
    *   **概要**: ユーザー自身の進行中の起床予約を取得する．
    *   **エンドポイント**: `GET /api/wakeup/active`
    *   **引数**: `c` (`echo.Context`)．クエリパラメータ `user_id` を使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `CancelWakeup(c) error`**
    *   **概要**: 進行中の見守り予約をキャンセルするリクエストを処理する．
    *   **エンドポイント**: `DELETE /api/wakeup/cancel`
    *   **引数**: `c` (`echo.Context`)．クエリパラメータ `user_id` を使用する．
    *   **戻り値**: `error` 型．
*   🔧 **メソッド: `GetActiveGroupChecks(c) error`**
    *   **概要**: グループ全員の起床状況（本日分）を一括取得する．
    *   **エンドポイント**: `GET /api/groups/:groupId/wakeups/active`
    *   **引数**: `c` (`echo.Context`)．パスパラメータ `:groupId` を使用する．
    *   **戻り値**: `error` 型．成功時はグループ全員の状態リストを返却する．

---
*最終更新日: 2026年6月14日*
