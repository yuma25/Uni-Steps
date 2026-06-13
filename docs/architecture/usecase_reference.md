# Uni-Steps ユースケース層 リファレンスマニュアル

本ドキュメントは，`usecase/` 配下で定義されている構造体，コンストラクタ（関数），およびメソッドの全ての仕様を網羅したものである．

---

## 1. 部屋管理 (GroupUsecase)

### 📦 `type GroupUsecase struct`
部屋の作成や参加に関するロジックを保持する構造体である．

### 🔧 `func (uc *GroupUsecase) UpdateSettings(ctx, groupID, userID, name, intervals, aiCharacter, ...) error`
部屋の設定（名称，リマインド間隔，AI 性格，LINE 連携，サマリー時刻）を更新する手順である．
- **権限**: 部屋のオーナーのみ許可される．

### 🔧 `func (uc *GroupUsecase) TransferOwnership(ctx, groupID, currentOwnerID, newOwnerID) error`
オーナー権限を別のメンバーに譲渡する．
- **仕様**: 譲渡先のユーザーがグループに所属している必要がある．

### 🔧 `func (uc *GroupUsecase) LeaveGroup(ctx, groupID, userID) error`
ユーザーを部屋から退出させる手順である．
- **制限**: オーナーが退出する場合は，事前に `TransferOwnership` で後継者を指名している必要がある（メンバーが他にいる場合）．メンバーが自分一人の場合は削除を推奨する．

### 🔧 `func (uc *GroupUsecase) DeleteGroup(ctx, groupID, userID) error`
部屋を完全に削除する．
- **権限**: オーナーのみ可能．

---

## 2. 課題管理 (TaskUsecase)

### 📦 `type TaskUsecase struct`
手動での課題操作や進捗更新を管理する構造体である．

### 🔧 `func (uc *TaskUsecase) RegisterManualTask(ctx, task) (*domain.Task, error)`
新しい課題を登録する．

### 🔧 `func (uc *TaskUsecase) UpdateTask(ctx, taskID, input, operatorID) (*domain.Task, error)`
既存の課題情報を更新する手順である．
- **権限管理**:
  - **タイトル・期限の変更**: 作成者（`CreatorID`）または「部屋のオーナー」のみ可能．
  - **該当者の変更**: 誰でも可能（自分自身や他人の追加・削除）．
- **自動クリーンアップ**: 更新の結果，担当者が 0 人になった場合，その課題は自動的にデータベースから削除される．

### 🔧 `func (uc *TaskUsecase) DeleteTask(ctx, taskID, operatorID) error`
課題を削除する．
- **権限**: 作成者または「部屋のオーナー」のみ可能．

### 🔧 `func (uc *TaskUsecase) ToggleUserCompletion(ctx, taskID, userID) error`
個人の完了状態を切り替える．
- **修正点**: データベースの Upsert ロジック（`infrastructure/db`）により，確実に状態が永続化されるようになった．

---

## 3. LMS 同期 & 起床確認 & 状況要約

### 🔧 `func (uc *SyncUsecase) SyncTasks(ctx, userID, groupID) ([]*domain.Task, error)`
外部 LMS（Google Classroom）から課題を同期する．

### 🔧 `func (uc *WakeupUsecase) GetActiveGroupChecks(ctx, groupID) ([]*domain.WakeupCheck, error)`
メンバー全員の起床状況を取得する．
- **表示仕様**: 「本日分」であれば，すでに起きた人（`confirmed`）の情報も返却され，画面に「起きた！」と表示し続けられる．

### 🔧 `func (uc *SummaryUsecase) SendAllSummaries(ctx, now) error`
設定時刻に達したグループへ朝夕のサマリーを送信する．

---
*最終更新日: 2026年6月14日*
