# Uni-Steps ユースケース層 リファレンスマニュアル

本ドキュメントは，`usecase/` 配下で定義されている構造体，コンストラクタ（関数），およびメソッドの全ての仕様を網羅したものである．ビジネスロジックの実行手順を確認する際に参照すること．

---

## 1. 部屋管理 (GroupUsecase)

### 📦 `type GroupUsecase struct`
部屋の作成や参加に関するロジックを保持する構造体である．
- **フィールド**:
  - `groupRepo domain.GroupRepository`: 部屋データの保存・検索を担当する．
  - `userRepo domain.UserRepository`: オーナーや参加者の存在確認を担当する．

### ⚙️ `func NewGroupUsecase(gr, ur) *GroupUsecase`
`GroupUsecase` の新しいインスタンスを生成するコンストラクタである．

### 🔧 `func (uc *GroupUsecase) CreateGroup(ctx, name, ownerID) (*domain.Group, error)`
新しい部屋を作成し，作成者をオーナーとして登録する手順である．
- **引数**:
  - `ctx context.Context`: 中断制御用のコンテキスト．
  - `name string`: 作成する部屋の名称．
  - `ownerID string`: 作成者（オーナー）のユーザー ID．
- **戻り値**:
  - `*domain.Group`: 作成された部屋のオブジェクト．自動生成された 8 桁の `InviteCode` を含む．
  - `error`: 保存失敗やオーナー不在時のエラー．

### 🔧 `func (uc *GroupUsecase) JoinGroupByInviteCode(ctx, code, userID) (*domain.Group, error)`
8 桁の招待コードを用いて，既存の部屋に参加する手順である．
- **引数**:
  - `ctx context.Context`: コンテキスト．
  - `code string`: 参加対象の部屋に設定された招待コード．
  - `userID string`: 参加しようとしているユーザーの ID．
- **戻り値**:
  - `*domain.Group`: 参加に成功した部屋のオブジェクト．
  - `error`: コード無効，ユーザー不在，または保存失敗時のエラー．

### 🔧 `func (uc *GroupUsecase) ListUserGroups(ctx, userID) ([]*domain.Group, error)`
ユーザーが現在所属している部屋の一覧を取得する手順である．
- **引数**:
  - `userID string`: 取得対象のユーザー ID．
- **戻り値**:
  - `[]*domain.Group`: 所属している部屋のリスト．各部屋には `Users` リストが Preload（事前読み込み）されている．
  - `error`: 取得失敗時のエラー．

---

## 2. 課題管理 (TaskUsecase)

### 📦 `type TaskUsecase struct`
手動での課題操作や進捗更新を管理する構造体である．
- **フィールド**:
  - `taskRepo domain.TaskRepository`: 課題データの保存・検索を担当する．
  - `aiService domain.AIService`: メッセージ生成を担当する．

### ⚙️ `func NewTaskUsecase(tr, ai) *TaskUsecase`
`TaskUsecase` の新しいインスタンスを生成するコンストラクタである．

### 🔧 `func (uc *TaskUsecase) RegisterManualTask(ctx, task) (*domain.Task, error)`
ユーザーが UI から入力した情報に基づき，新しい課題を登録する手順である．
- **引数**:
  - `task *domain.Task`: 登録する課題のプロトタイプ．タイトルは必須．
- **戻り値**:
  - `*domain.Task`: 保存された課題オブジェクト．UUID が自動発行される．
  - `error`: バリデーション違反や保存失敗時のエラー．

### 🔧 `func (uc *TaskUsecase) ListGroupTasks(ctx, groupID) ([]*domain.Task, error)`
特定の部屋に属する課題一覧を取得する手順である．
- **引数**:
  - `groupID string`: 取得対象の部屋 ID．
- **戻り値**:
  - `[]*domain.Task`: 期限の近い順（昇順）に並んだ課題のリスト．

### 🔧 `func (uc *TaskUsecase) UpdateTask(ctx, taskID, input) (*domain.Task, error)`
既存の課題の内容や該当者リストを更新する手順である．
- **引数**:
  - `taskID string`: 更新対象の内部課題 ID．
  - `input *domain.Task`: 更新後の情報（タイトル，期限，該当者リスト）を持つオブジェクト．
- **戻り値**:
  - `*domain.Task`: 更新後の課題オブジェクト．既存メンバーの完了状態は維持される．
  - `error`: 対象不在や保存失敗時のエラー．

### 🔧 `func (uc *TaskUsecase) ToggleUserCompletion(ctx, taskID, userID) error`
特定のユーザーの課題完了状態を「完了 ↔ 未完了」で反転させる手順である．
- **引数**:
  - `taskID string`: 対象の課題 ID．
  - `userID string`: 操作を行うユーザーの ID．
- **戻り値**:
  - `error`: 保存失敗時のエラー．

---

## 3. LMS 同期 (SyncUsecase)

### 📦 `type SyncUsecase struct`
Google Classroom 等の外部システムから情報を統合するロジックを保持する．
- **フィールド**:
  - `taskRepo domain.TaskRepository`: 同期した課題の保存を担当する．
  - `groupRepo domain.GroupRepository`: 部屋の同期ステータス更新を担当する．
  - `lmsService domain.LMSService`: 外部 API との通信を担当する．

### ⚙️ `func NewSyncUsecase(tr, gr, lms) *SyncUsecase`
`SyncUsecase` のコンストラクタである．

### 🔧 `func (uc *SyncUsecase) SyncTasks(ctx, userID, groupID) ([]*domain.Task, error)`
外部 LMS から最新の課題と提出状況を取得し，特定の部屋へ保存・更新する手順である．
- **引数**:
  - `userID string`: 同期に使用する OAuth トークンを持つユーザーの ID．
  - `groupID string`: 課題を流し込む先の部屋 ID．
- **戻り値**:
  - `[]*domain.Task`: 今回の同期で新規作成または更新された課題のリスト．
  - `error`: 通信失敗やデータ不整合時のエラー．

---

## 4. 自動監視 (MonitorUsecase)

### 📦 `type MonitorUsecase struct`
バックエンドで常駐し，時間経過による自動処理を行う構造体である．
- **フィールド**:
  - リポジトリ群 (Task, User, Group, Wakeup) および AI, 通知サービスへの依存を持つ．

### ⚙️ `func NewMonitorUsecase(...) *MonitorUsecase`
`MonitorUsecase` のコンストラクタである．

### 🔧 `func (uc *MonitorUsecase) StartMonitoring(ctx)`
定期監視プロセス（Goroutine）を起動する手順である．
- **引数**:
  - `ctx context.Context`: このコンテキストが終了（Done）すると，監視も停止する．

### 🔧 `func (uc *MonitorUsecase) checkApproachingTasks(ctx)` (内部処理)
締め切り間近の課題に対し，AI が生成したリマインド文をログに出力する（将来的に送信）．

### 🔧 `func (uc *MonitorUsecase) checkWakeupStatuses(ctx)` (内部処理)
起床予定時刻を過ぎた「未確認」のユーザーに対し，グループメンバーへ SOS 通知を送信する．

---
*最終更新日: 2026年6月11日*
