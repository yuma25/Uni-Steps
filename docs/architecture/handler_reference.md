# Uni-Steps インターフェース（ハンドラー）層 リファレンスマニュアル

本ドキュメントは，`interfaces/handler/` 配下で定義されている構造体（ハンドラー）および，フロントエンドからの HTTP リクエストを処理する全ての API エンドポイントの仕様を網羅したものである．

---

## 1. 認証管理 (AuthHandler)
Google OAuth 2.0 を用いたログインおよびユーザー登録を管理する．

### 🔧 `GET /api/auth/google/login` (`GoogleLogin`)
Google の認可画面へユーザーをリダイレクトさせる．

### 🔧 `GET /api/auth/google/callback` (`GoogleCallback`)
Google での承認後に呼び出されるコールバック窓口である．

---

## 2. 部屋管理 (GroupHandler)

### 🔧 `POST /api/groups` (`CreateGroup`)
新しい部屋を手動で作成する．
- **ボディ**: `{"name": "部屋名", "owner_id": "ユーザーID"}`

### 🔧 `POST /api/groups/join` (`JoinGroupByInviteCode`)
招待コードを用いて部屋に参加する．

### 🔧 `GET /api/users/:userId/groups` (`ListUserGroups`)
ユーザーが所属している部屋の一覧を取得する．

### 🔧 `PATCH /api/groups/:groupId/settings` (`UpdateGroupSettings`)
部屋の設定（名称，リマインド間隔，AI 性格，LINE 連携，サマリー時刻等）を更新する．
- **権限**: 部屋のオーナーのみ許可される．
- **ボディ**: `{"name": "新部屋名", "remind_intervals": [...], "ai_character": "...", ...}`

### 🔧 `PUT /api/groups/:groupId/owner` (`TransferOwnership`)
部屋のオーナー権限を他のメンバーに譲渡する．
- **ボディ**: `{"current_owner_id": "旧ID", "new_owner_id": "新ID"}`

### 🔧 `DELETE /api/groups/:groupId` (`DeleteGroup`)
部屋を完全に削除する．
- **権限**: オーナーのみ可能．
- **クエリ**: `?user_id=オーナーID` (確認用)

### 🔧 `DELETE /api/groups/:groupId/users/:userId` (`LeaveGroup`)
部屋から退出する．

### 🔧 `GET /api/groups/:groupId/notifications` (`GetNotificationLogs`)
通知履歴（AI ログ等）を取得する．

---

## 3. 課題管理 (TaskHandler)

### 🔧 `GET /api/groups/:id/tasks` (`ListTasks`)
部屋の全課題を取得する．

### 🔧 `POST /api/tasks/manual` (`CreateManualTask`)
手動課題を登録する．

### 🔧 `PUT /api/tasks/:id` (`UpdateTask`)
課題を更新する．
- **権限**: タイトル・期限の変更は「作成者」または「部屋のオーナー」のみ可能．該当者の変更は誰でも可能．
- **クエリ**: `?user_id=実行者ID` (権限チェック用)

### 🔧 `DELETE /api/tasks/:id` (`DeleteTask`)
課題を削除する．
- **権限**: 「作成者」または「部屋のオーナー」のみ可能．
- **クエリ**: `?user_id=実行者ID`

### 🔧 `PATCH /api/tasks/:id/toggle-completion` (`ToggleTaskCompletion`)
完了状態を切り替える（手動課題のみ有効）．

### 🔧 `POST /api/tasks/sync` (`SyncTasks`)
Google Classroom 等との同期を実行する．

---

## 4. 起床管理 (WakeupHandler)

### 🔧 `POST /api/wakeup/request` (`RequestWakeupCheck`)
起床見守りを予約する．

### 🔧 `POST /api/wakeup/checkin` (`CheckIn`)
起床報告を行う．

### 🔧 `DELETE /api/wakeup/cancel` (`CancelWakeup`)
予約をキャンセルする．

### 🔧 `GET /api/groups/:groupId/wakeups/active` (`GetActiveGroupChecks`)
メンバー全員の起床状況を取得する．
- **詳細**: 「本日」のデータであれば，`confirmed`（起きた）状態のものも含まれる．

---

## 5. その他 (NotificationHandler, UserHandler)

### 🔧 `POST /api/notifications/subscribe` (`SubscribeWebPush`)
### 🔧 `POST /api/notifications/test` (`SendTestNotification`)
### 🔧 `GET /api/users/:id` (`GetUser`)

---
*最終更新日: 2026年6月14日*
