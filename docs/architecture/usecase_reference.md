# Uni-Steps ユースケース層 リファレンスマニュアル

本ドキュメントは，`usecase/` 配下で定義されている各ファイルごとの構造体（頭脳）とメソッド（手順）の仕様を網羅したものである．

---

## 1. usecase/group_uc.go (部屋管理のロジック)

### 📦 `type GroupUsecase struct`
グループ（部屋）のライフサイクルおよび設定管理を担当する構造体である．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `groupRepo` | `domain.GroupRepository` | 部屋データの保存・取得を担当． |
| `userRepo` | `domain.UserRepository` | 利用者データの確認を担当． |
| `logRepo` | `domain.NotificationLogRepository` | 通知履歴の取得を担当． |

*   ⚙️ **関数: `NewGroupUsecase(gr, ur, lr) *GroupUsecase`**
    *   **概要**: 必要なリポジトリを注入し，`GroupUsecase` のインスタンスを生成する．
    *   **引数**: `gr` (GroupRepo), `ur` (UserRepo), `lr` (LogRepo)．
    *   **戻り値**: 生成された `*GroupUsecase`．
*   🔧 **メソッド: `CreateGroup(ctx, name, ownerID) (*domain.Group, error)`**
    *   **概要**: 新しい部屋を作成し，作成したユーザーをオーナー兼最初のメンバーとして登録する．
    *   **引数**:
        *   `name`: `string` 型．作成する部屋の名前．
        *   `ownerID`: `string` 型．作成者（オーナー）のユーザー ID．
    *   **戻り値**: 作成された `Group` オブジェクト，およびエラー．
*   🔧 **メソッド: `JoinGroupByInviteCode(ctx, code, userID) (*domain.Group, error)`**
    *   **概要**: 8 桁の招待コードを使用して，既存の部屋に新しいメンバーを参加させる．
    *   **引数**:
        *   `code`: `string` 型．招待コード．
        *   `userID`: `string` 型．参加しようとしているユーザーの ID．
    *   **戻り値**: 参加した `Group` オブジェクト，およびエラー．
*   🔧 **メソッド: `ListUserGroups(ctx, userID) ([]*domain.Group, error)`**
    *   **概要**: 特定のユーザーが現在所属している全ての部屋を一覧取得する．
    *   **引数**: `userID` (`string`)．
    *   **戻り値**: 所属グループのリスト，およびエラー．
*   🔧 **メソッド: `UpdateSettings(ctx, groupID, userID, name, intervals, aiCharacter, ...) error`**
    *   **概要**: 部屋の名前や AI の性格，LINE 連携情報，通知タイミングなどの設定を一括更新する．
    *   **権限**: `userID` がその部屋のオーナーである場合のみ実行可能．
    *   **引数**: `groupID`, `userID`（実行者）, `name`, `intervals`, `aiCharacter`, `lineToken`, `lineGroupID`, `morningTime`, `eveningTime`．
    *   **戻り値**: 成功時は nil，権限不足や失敗時は error．
*   🔧 **メソッド: `LeaveGroup(ctx, groupID, userID) error`**
    *   **概要**: ユーザーを部屋から脱退させる．
    *   **制限**: オーナーが脱退する場合，他にメンバーがいるならば事前に `TransferOwnership` を行う必要がある．
    *   **引数**: `groupID`, `userID`（脱退するユーザー）．
    *   **戻り値**: 成功時は nil，条件未達時は error．
*   🔧 **メソッド: `TransferOwnership(ctx, groupID, currentOwnerID, newOwnerID) error`**
    *   **概要**: 部屋の管理権限（オーナー）を別の既存メンバーへ譲渡する．
    *   **引数**: `groupID`, `currentOwnerID`（現在のオーナー）, `newOwnerID`（新しいオーナー）．
    *   **戻り値**: 成功時は nil，失敗時は error．
*   🔧 **メソッド: `DeleteGroup(ctx, groupID, userID) error`**
    *   **概要**: 部屋を完全に削除し，関連する課題やデータとの紐付けをすべて消去する．
    *   **権限**: オーナーのみ実行可能．
    *   **引数**: `groupID`, `userID`（実行者）．
*   🔧 **メソッド: `GetNotificationLogs(ctx, groupID, limit) ([]*domain.NotificationLog, error)`**
    *   **概要**: ダッシュボードの活動履歴（タイムライン）を表示するために，最新の通知ログを取得する．
    *   **引数**: `groupID`, `limit`（取得件数）．
    *   **戻り値**: ログのリスト，およびエラー．

---

## 2. usecase/task_uc.go (課題管理のロジック)

### 📦 `type TaskUsecase struct`
手動課題の登録，編集，削除，および進捗（完了状態）の更新を統合管理する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `taskRepo` | `domain.TaskRepository` | 課題データの永続化を担当． |
| `groupRepo` | `domain.GroupRepository` | 部屋の設定（リマインド間隔等）の参照を担当． |
| `aiService` | `domain.AIService` | リマインド文等の AI 生成を担当． |
| `scheduler` | `domain.SchedulerService` | 未来の通知予約の管理を担当． |

*   ⚙️ **関数: `NewTaskUsecase(tr, gr, ai, sch) *TaskUsecase`**
    *   **概要**: `TaskUsecase` の新しいインスタンスを生成する．
    *   **引数**: `tr` (TaskRepo), `gr` (GroupRepo), `ai` (AIService), `sch` (SchedulerService)．
    *   **戻り値**: 生成された `*TaskUsecase`．
*   🔧 **メソッド: `RegisterManualTask(ctx, task) (*domain.Task, error)`**
    *   **概要**: ユーザーが入力した情報を元に新しい課題を登録し，同時にグループ設定に基づいた未来のリマインド予約を自動生成する．
    *   **引数**: `task` (`*domain.Task`)．タイトル，期限，担当者情報が含まれる．
    *   **戻り値**: 登録後の課題オブジェクト，およびエラー．
*   🔧 **メソッド: `ListGroupTasks(ctx, groupID) ([]*domain.Task, error)`**
    *   **概要**: 部屋 ID をキーに，所属する全課題を取得する．
    *   **戻り値**: 課題リスト，およびエラー．
*   🔧 **メソッド: `UpdateTask(ctx, taskID, input, operatorID) (*domain.Task, error)`**
    *   **概要**: 既存課題を更新する．権限チェックを厳格に行い，必要に応じて通知の再予約やキャンセルを行う．
    *   **引数**: `taskID`, `input`（新しい情報）, `operatorID`（実行者）．
    *   **戻り値**: 更新後の課題オブジェクト（削除時は nil），およびエラー．
*   🔧 **メソッド: `DeleteTask(ctx, taskID, operatorID) error`**
    *   **概要**: 課題を物理削除する．削除前に予約されていた全てのリマインド通知をキャンセルする．
    *   **引数**: `taskID`, `operatorID`．
*   🔧 **メソッド: `ToggleUserCompletion(ctx, taskID, userID) error`**
    *   **概要**: 特定のユーザーにおける「完了 / 未完了」を切り替える．
    *   **引数**: `taskID`, `userID`．

---

## 3. usecase/sync_uc.go (LMS同期のロジック)

### 📦 `type SyncUsecase struct`
外部学習管理システム（LMS）から情報を取得し，内部データベースと同期する責務を担う．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `taskRepo` | `domain.TaskRepository` | 同期したデータの保存を担当． |
| `groupRepo` | `domain.GroupRepository` | 同期時刻の更新などを担当． |
| `lmsService` | `domain.LMSService` | 外部 API との通信を担当． |
| `scheduler` | `domain.SchedulerService` | 同期後のリマインド自動予約を担当． |

*   ⚙️ **関数: `NewSyncUsecase(tr, gr, lms, sch) *SyncUsecase`**
    *   **概要**: `SyncUsecase` の新しいインスタンスを生成する．
    *   **引数**: `tr` (TaskRepo), `gr` (GroupRepo), `lms` (LMSService), `sch` (SchedulerService)．
    *   **戻り値**: 生成された `*SyncUsecase`．
*   🔧 **メソッド: `SyncTasks(ctx, userID, groupID) ([]*domain.Task, error)`**
    *   **概要**: 指定されたユーザーの権限を用いて Google Classroom 等から課題を取得し，現在の部屋のデータとマージする．
    *   **パフォーマンス最適化**:
        *   **N+1 問題の解消**: 同期ループ内で個別にデータベース検索を行わず，事前にグループ内の全課題を取得してメモリ上のマップで照合することで，クエリ回数を劇的に削減している．
        *   **同期内重複排除**: LMS から取得したデータ自体に重複 ID が含まれている場合に備え，セッション内での `ExternalID` チェックを行い，一意制約違反（`idx_group_external`）を未然に防止している．
    *   **引数**: `userID`（取得者）, `groupID`（保存先の部屋）．
    *   **戻り値**: 同期・保存された課題のリスト，およびエラー．

---

## 4. usecase/wakeup_uc.go (起床確認のロジック)

### 📦 `type WakeupUsecase struct`
起床予定の登録，判定，および SOS 発信の連動制御を担当する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `wakeupRepo` | `domain.WakeupRepository` | 起床確認データの保存・取得を担当． |
| `scheduler` | `domain.SchedulerService` | SOS 通知の予約・取消を担当． |

*   ⚙️ **関数: `NewWakeupUsecase(wr, sch) *WakeupUsecase`**
    *   **概要**: `WakeupUsecase` の新しいインスタンスを生成する．
    *   **引数**: `wr` (WakeupRepo), `sch` (SchedulerService)．
    *   **戻り値**: 生成された `*WakeupUsecase`．
*   🔧 **メソッド: `RequestWakeup(ctx, userID, groupID, targetTime, graceMinutes) (*domain.WakeupCheck, error)`**
    *   **概要**: 新しい起床予定を登録し，同時に指定された猶予時間後の SOS 通知をスケジューラーに予約する．
    *   **引数**: `userID`, `groupID`, `targetTime`, `graceMinutes`．
    *   **戻り値**: 登録された起床確認オブジェクト，およびエラー．
*   🔧 **メソッド: `ConfirmWakeup(ctx, userID) error`**
    *   **概要**: ユーザーの起床を確認し，状態を `confirmed` に更新する．連動して，予約されていた SOS 通知をキャンセルする．
    *   **引数**: `userID`．
*   🔧 **メソッド: `CancelWakeup(ctx, userID) error`**
    *   **概要**: 起床予定を完全に取り消し，予約されていた SOS 通知も破棄する．
    *   **引数**: `userID`．
*   🔧 **メソッド: `GetActiveGroupChecks(ctx, groupID) ([]*domain.WakeupCheck, error)`**
    *   **概要**: 指定されたグループの本日分のメンバー全員の起床状況を取得する．
    *   **戻り値**: 起床状況のリスト，およびエラー．

---

## 5. usecase/summary_uc.go (状況要約のロジック)

### 📦 `type SummaryUsecase struct`
設定された時刻にグループ全体の進捗状況を集計し，AI による要約メッセージを配信する機能を担当する．
| 内部フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `groupRepo` | `domain.GroupRepository` | 配信対象グループの検索を担当． |
| `taskRepo` | `domain.TaskRepository` | 集計対象課題の取得を担当． |
| `aiService` | `domain.AIService` | 要約メッセージの生成を担当． |
| `notifService` | `domain.NotificationService` | 実際のメッセージ配信を担当． |
| `logRepo` | `domain.NotificationLogRepository` | 送信したサマリーの記録を担当． |

*   ⚙️ **関数: `NewSummaryUsecase(gr, tr, ai, ns, lr) *SummaryUsecase`**
    *   **概要**: `SummaryUsecase` の新しいインスタンスを生成する．
    *   **引数**: `gr`, `tr`, `ai`, `ns`, `lr`（各種依存サービス）．
    *   **戻り値**: 生成された `*SummaryUsecase`．
*   🔧 **メソッド: `SendAllSummaries(ctx, now) error`**
    *   **概要**: システム全体の全グループを巡回し，現在時刻が各グループの配信設定と一致する場合に配信をキックする．
    *   **引数**: `now`（現在時刻）．
*   🔧 **メソッド: `ProcessSingleGroupSummary(ctx, groupID, summaryType) error`**
    *   **概要**: 特定のグループに対してサマリーを生成・送信する実処理．
    *   **引数**: `groupID`, `summaryType`．

---
*最終更新日: 2026年6月15日*
