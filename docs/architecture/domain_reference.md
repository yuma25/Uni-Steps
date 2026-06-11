# Uni-Steps ドメイン層 リファレンスマニュアル

本ドキュメントは，`domain/` 配下で定義されている構造体（データモデル）とインターフェース（契約）の全ての仕様を網羅したものである．

---

## 1. 課題関連 (domain/task.go)

### 📦 `type Task struct`
システムで管理される「やるべきこと」の基本情報．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 課題のユニークID (PK/UUID) |
| `GroupID` | `string` | 所属する部屋のID (FK) |
| `Source` | `string` | 入力元 (`manual`, `ai`, `google_classroom`) |
| `ExternalID` | `string` | Google Classroom 側での一意なID (UK/重複防止用) |
| `RawText` | `string` | ユーザーが入力した生の文章 (AI 解析時のみ) |
| `Title` | `string` | 課題のタイトル |
| `Deadline` | `time.Time` | 提出期限 (西暦1年は「未定」扱い) |
| `IsLMSDeadlineSet` | `bool` | 外部 LMS 側で最初から期限があったかのフラグである． |
| `LMSUpdateTime` | `time.Time` | LMS 側での最終更新日時 |
| `Recurrence` | `RecurrenceSettings` | 繰り返しの設定 (JSON形式で一括管理) |
| `UserProgress` | `[]*TaskUserProgress` | メンバーごとの進捗状況リスト (1:N) |

### 📦 `type TaskUserProgress struct`
一つの課題に対し，各メンバーが「終わったかどうか」を個別に管理する．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `TaskID` | `string` | 対象課題のID (PK/FK) |
| `UserID` | `string` | 対象ユーザーのID (PK/FK) |
| `UserName` | `string` | 表示用のユーザー名 |
| `IsCompleted` | `bool` | 完了フラグ |
| `UpdatedAt` | `time.Time` | 最後にステータスを変更した日時 |

### 📦 `type RecurrenceSettings struct`
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `Type` | `string` | ルール種別 (`none`, `weekly`, `biweekly`, `custom`) |
| `CustomDates` | `[]time.Time` | `custom` 時のみ使用する特定日付のリスト |

---

## 2. 部屋・ユーザー関連 (domain/group.go, domain/user.go)

### 📦 `type Group struct`
共同作業の単位である「部屋」の情報．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 部屋のユニークID (PK/UUID) |
| `Name` | `string` | 部屋の名称 |
| `OwnerID` | `string` | 作成者のユーザーID (FK) |
| `LineChannelToken` | `string` | BYOT 用の LINE Bot アクセストークン |
| `LineGroupID` | `string` | 通知先の LINE グループ ID |
| `LastSyncedAt` | `time.Time` | 同期処理を最後に実行した時刻 |
| `LMSLastUpdatedAt` | `time.Time` | 外部 LMS 側で最後に更新が検知された時刻 |
| `InviteCode` | `string` | 8 桁の参加用招待コード (UK) |
| `Users` | `[]*User` | 所属メンバーのリスト (M:M) |

### 📦 `type User struct`
システム利用者（個人）の情報．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | ユーザーの一意識別子 (PK) |
| `Name` | `string` | 表示名 |
| `Email` | `string` | メールアドレス (UK/Googleログインキー) |
| `WebPushToken` | `string` | ブラウザ通知用トークン (JSON) |
| `GoogleAccessToken` | `string` | Classroom同期用アクセストークン |
| `GoogleRefreshToken` | `string` | トークン更新用リフレッシュトークン |
| `GoogleTokenExpiry` | `time.Time` | Google OAuth トークンの有効期限 |
| `LastCheckInAt` | `time.Time` | 最終起床確認（またはログイン）時刻 |
| `Groups` | `[]*Group` | 所属している部屋のリスト |

---

## 3. 保存・検索ルール (domain/repository.go)

### 🔹 `type TaskRepository interface`
*   `Save(ctx, task) error`: 保存・更新．`UserProgress` も含めて永続化する．
*   `FindByID(ctx, id) (*Task, error)`: 内部 ID で検索．
*   `FindByExternalID(ctx, extID) (*Task, error)`: Classroom 側の ID で検索．
*   `FindByGroupID(ctx, groupID) ([]*Task, error)`: 部屋の課題を期限順（ASC）で全件取得．
*   `FindApproachingDeadlines(ctx, until) ([]*Task, error)`: 期限間近の課題を抽出．

### 🔹 `type UserRepository interface`
*   `Save(ctx, user) error`: 保存．
*   `FindByID(ctx, id) (*User, error)`: ID で検索．
*   `FindByEmail(ctx, email) (*User, error)`: メールアドレスで検索．

### 🔹 `type GroupRepository interface`
*   `Save(ctx, group) error`: 保存．`Users` との紐付け（多対多）も管理．
*   `FindByID(ctx, id) (*Group, error)`: ID で検索（メンバー情報も Preload）．
*   `FindByInviteCode(ctx, code) (*Group, error)`: 招待コードで特定．
*   `FindByUserID(ctx, userID) ([]*Group, error)`: 所属部屋を一覧．

---

## 4. 外部サービス (domain/ai.go, domain/lms.go, domain/notification.go)

### 🔹 `type AIService interface`
*   `GenerateRemindMessage(ctx, task, style) (string, error)`: 文章生成．

### 🔹 `type LMSService interface`
*   `FetchTasks(ctx, userID) ([]*Task, error)`: 課題と提出状況の一括取得．
*   `GetProviderName() string`: プロバイダ名（例: "google_classroom"）の取得．

### 🔹 `type NotificationService interface`
*   `SendGroupMessage(ctx, targetID, msg) error`: グループ通知．
*   `SendDirectMessage(ctx, userID, msg) error`: 個人通知．

### 🔹 `type SchedulerService interface` (NEW: 予約方式)
未来の時刻に通知を予約する新方式の窓口である．
*   `ScheduleTaskRemind(ctx, task, userID, runAt)`: 指定時刻にリマインドを予約する．
*   `CancelTaskRemind(ctx, taskID, userID)`: 予約済みの通知を取り消す．

### 📦 `type ReminderJob struct` (NEW)
予約された通知の情報を保持するデータモデルである．
*   `TargetTime`: 通知を飛ばすべき正確な時刻．
*   `Status`: `pending`（待機中），`sent`（送信済み），`cancelled`（取消済み）．

---

## 5. 起床確認 (domain/wakeup.go)

### 📦 `type WakeupCheck struct`
起床見守りスケジュール．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 管理 ID |
| `UserID` | `string` | 対象ユーザー |
| `GroupID` | `string` | 通知先グループ |
| `TargetTime` | `time.Time` | 約束の起床時刻 |
| `GraceMinutes` | `int` | 猶予時間（分） |
| `Status` | `string` | 状態 (`pending`, `confirmed`, `alerted`) |

### 🔹 `type WakeupRepository interface`
*   `Save(ctx, check) error`: 保存．
*   `FindByID(ctx, id) (*WakeupCheck, error)`: 検索．
*   `FindPendingByTime(ctx, now) ([]*WakeupCheck, error)`: 寝坊者抽出．
*   `FindActiveByUser(ctx, userID) ([]*WakeupCheck, error)`: 進行中スケジュール取得．

---
*最終更新日: 2026年6月11日*
