# Uni-Steps インターフェース（ハンドラー）層 リファレンスマニュアル

本ドキュメントは，`interfaces/handler/` 配下で定義されている構造体（ハンドラー）および，フロントエンドからの HTTP リクエストを処理する全ての API エンドポイントの仕様を網羅したものである．

---

## 1. 認証管理 (AuthHandler)
Google OAuth 2.0 を用いたログインおよびユーザー登録を管理する．

### 📦 `type AuthHandler struct`
- **依存**: `UserRepository`, `oauth2.Config`

### ⚙️ `func NewAuthHandler(e, ur, cfg)`
認証に関連する以下のルーティングを登録するコンストラクタである．

### 🔧 `GET /api/auth/google/login` (`GoogleLogin`)
Google の認可画面へユーザーをリダイレクトさせる．
- **処理**: CSRF 対策としてランダムな `state` を生成し，HttpOnly Cookie に保存した上で認可 URL を生成する．
- **戻り値**: `307 Temporary Redirect` (Google の認可ページへ)

### 🔧 `GET /api/auth/google/callback` (`GoogleCallback`)
Google での承認後に呼び出されるコールバック窓口である．
- **パラメータ**: `state` (Query), `code` (Query)
- **検証**: 送信された `state` が Cookie 内の期待値と一致するか確認する．不一致時は `403 Forbidden` を返す．
- **処理**: 認可コードをトークンに交換し，Google UserInfo API からプロフィールを取得．未登録なら `USERS` テーブルに新規作成する．
- **戻り値**: `307 Temporary Redirect` (フロントエンドのダッシュボードへ)

---

## 2. 部屋管理 (GroupHandler)
コミュニティ（部屋）の作成・参加・一覧取得を管理する．

### 📦 `type GroupHandler struct`
- **依存**: `GroupUsecase`, `LMSService`

### 🔧 `POST /api/groups` (`CreateGroup`)
新しい部屋を手動で作成する．
- **ボディ**: `{"name": "部屋名", "owner_id": "ユーザーID"}`
- **戻り値**: `201 Created` (作成された `Group` オブジェクト)

### 🔧 `POST /api/groups/join` (`JoinGroupByInviteCode`)
8 桁の招待コードを用いて部屋に参加する．
- **ボディ**: `{"invite_code": "xxxxxxxx", "user_id": "参加者ID"}`
- **戻り値**: `200 OK` (参加後の `Group` オブジェクト)

### 🔧 `GET /api/users/:userId/groups` (`ListUserGroups`)
指定したユーザーが所属している部屋の一覧を取得する．
- **パスパラメータ**: `userId`
- **戻り値**: `200 OK` (部屋リストの配列)

### 🔧 `PATCH /api/groups/:groupId/settings` (`UpdateGroupSettings`)
部屋の設定（リマインドタイミング等）を更新する．
- **権限**: オーナーのみ許可される．
- **パスパラメータ**: `groupId`
- **ボディ**: `{"remind_intervals": [60, 1440], "user_id": "本人ID"}`
- **戻り値**: `200 OK` (成功メッセージ)

---

## 3. 課題管理 (TaskHandler)
課題の操作全般（一覧，登録，編集，同期）を管理する．

### 📦 `type TaskHandler struct`
- **依存**: `TaskUsecase`, `SyncUsecase`

### 🔧 `GET /api/groups/:id/tasks` (`ListTasks`)
特定の部屋に紐づく全ての課題を期限順（昇順）で取得する．
- **パスパラメータ**: `id` (部屋 ID)
- **戻り値**: `200 OK` (課題リストの配列)

### 🔧 `POST /api/tasks/manual` (`CreateManualTask`)
手動で新しい課題を登録する．
- **ボディ**: `Task` 構造体の一部（タイトル，期限等）
- **戻り値**: `201 Created` (作成された `Task` オブジェクト)

### 🔧 `PUT /api/tasks/:id` (`UpdateTask`)
既存課題の情報（タイトル，期限，該当者）を更新する．
- **パスパラメータ**: `id` (課題 ID)
- **ボディ**: 更新後の課題データ
- **戻り値**: `200 OK` (更新後の `Task` オブジェクト)

### 🔧 `PATCH /api/tasks/:id/toggle-completion` (`ToggleTaskCompletion`)
ユーザー自身の完了状態を反転させる．
- **パスパラメータ**: `id` (課題 ID)
- **ボディ**: `{"user_id": "実行者ID"}`
- **戻り値**: `200 OK` (成功メッセージ)

### 🔧 `POST /api/tasks/sync` (`SyncTasks`)
Google Classroom から最新の課題を取得・統合する．
- **ボディ**: `{"user_id": "実行者ID", "group_id": "対象部屋ID"}`
- **戻り値**: `200 OK` (`tasks`: 更新された課題リスト)

---

## 4. 通知管理 (NotificationHandler)
Web Push 等の購読情報を管理する．

### 🔧 `POST /api/notifications/subscribe` (`SubscribeWebPush`)
ブラウザからのプッシュ通知許可（購読情報）を保存する．
- **ボディ**: `{"user_id": "ID", "subscription": "JSON文字列"}`
- **戻り値**: `200 OK`

---

## 5. 起床管理 (WakeupHandler)
起床確認のスケジュール登録とチェックインを管理する．

### 🔧 `POST /api/wakeup/request` (`RequestWakeupCheck`)
新しい起床確認を予約する．
- **ボディ**: `{"user_id": "ID", "group_id": "ID", "target_time": "時刻", "grace_minutes": 分}`
- **戻り値**: `201 Created`

### 🔧 `POST /api/wakeup/checkin` (`CheckIn`)
ユーザーが起床したことを報告する．
- **ボディ**: `{"user_id": "本人ID"}`
- **戻り値**: `200 OK` (進行中の全スケジュールを `confirmed` に更新)

---
*最終更新日: 2026年6月11日*
