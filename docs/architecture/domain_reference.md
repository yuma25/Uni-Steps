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
| `Source` | `string` | 入力元 (`manual`, `google_classroom`, `web_class`) |
| `ExternalID` | `string` | 外部 LMS 側での一意なID (UK/重複防止用) |
| `RawText` | `string` | ユーザーが入力した生の文章 (AI 解析時のみ) |
| `CreatorID` | `string` | 課題の作成者 ID（手動課題のみ保持） |
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
| `OwnerID` | `string` | 現在のオーナー（管理権限保持者）のユーザーID (FK) |
| `LineChannelToken` | `string` | BYOT 用の LINE Bot アクセストークン |
| `LineGroupID` | `string` | 通知先の LINE グループ ID |
| `LastSyncedAt` | `time.Time` | 同期処理を最後に実行した時刻 |
| `LMSLastUpdatedAt` | `time.Time` | 外部 LMS 側で最後に更新が検知された時刻 |
| `InviteCode` | `string` | 8 桁の参加用招待コード (UK) |
| `RemindIntervals` | `[]int` | リマインド通知を飛ばすタイミング（分前）のリストである．最大 3 つまで保持する． |
| `AICharacter` | `string` | AI の性格設定 (`default`, `strict`, `kind`, `cool`) である． |
| `SummaryMorningTime` | `string` | 朝のサマリー送信時刻（HH:mm形式）である． |
| `SummaryEveningTime` | `string` | 夜のサマリー送信時刻（HH:mm形式）である． |
| `Users` | `[]*User` | 所属メンバーのリスト (M:M) |

### 📦 `type NotificationLog struct`
送信された通知の履歴を記録するエンティティである．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | ログの一意識別子 (PK) |
| `GroupID` | `string` | 発生元のグループ ID |
| `UserID` | `string` | 対象ユーザー（または原因となったユーザー）の ID |
| `Type` | `string` | 通知の種別 (`remind`, `sos`, `summary`) |
| `Message` | `string` | 送信されたメッセージの本文 |
| `CreatedAt` | `time.Time` | 送信日時 |

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
*   `Save(ctx, task) error`: 保存・更新．`UserProgress` を含めた同期を保証する．
*   `FindByID(ctx, id) (*Task, error)`: 内部 ID で検索．
*   `FindByExternalID(ctx, extID) (*Task, error)`: Classroom 側の ID で検索．
*   `FindByGroupID(ctx, groupID) ([]*Task, error)`: 部屋の課題を期限順（ASC）で全件取得．
*   `FindApproachingDeadlines(ctx, until) ([]*Task, error)`: 期限間近の課題を抽出．
*   `Delete(ctx, id) error`: 課題とそれに関連する進捗データを完全に削除する．

### 🔹 `type UserRepository interface`
*   `Save(ctx, user) error`: 保存．
*   `UpdateWebPushToken(ctx, userID, token) error`: 通知トークンの更新．
*   `FindByID(ctx, id) (*User, error)`: ID で検索．
*   `FindByEmail(ctx, email) (*User, error)`: メールアドレスで検索．

### 🔹 `type GroupRepository interface`
*   `Save(ctx, group) error`: 保存（名称，性格，オーナー等）．
*   `FindByID(ctx, id) (*Group, error)`: ID で検索（メンバー情報も Preload）．
*   `FindAllGroups(ctx) ([]*Group, error)`: 全部屋の取得．
*   `FindByInviteCode(ctx, code) (*Group, error)`: 招待コードで特定．
*   `FindByUserID(ctx, userID) ([]*Group, error)`: 所属部屋を一覧．
*   `RemoveUser(ctx, groupID, userID) error`: 部屋から特定のメンバーを脱退させる．
*   `Delete(ctx, id) error`: 部屋を完全に削除し，紐付けをクリアする．

---

## 4. 外部サービス (domain/ai.go, domain/lms.go, domain/notification.go)

### 🔹 `type AIService interface`
*   `GenerateRemindMessage(ctx, task, style) (string, error)`: リマインド用文章生成．
*   `GenerateWakeupSOSMessage(ctx, user, group, style) (string, error)`: 寝坊 SOS 用文章生成．
*   `GenerateGroupSummaryMessage(ctx, group, mode, style) (string, error)`: サマリー用文章生成．

### 🔹 `type LMSService interface`
*   `FetchTasks(ctx, userID) ([]*Task, error)`: 課題と提出状況の一括取得．
*   `GetProviderName() string`: プロバイダ名（例: "google_classroom"）の取得．

### 🔹 `type NotificationService interface`
*   `SendGroupMessage(ctx, targetID, msg) error`: LINE 等のグループ通知．
*   `SendDirectMessage(ctx, userID, msg) error`: Web Push 等の個人通知．

### 🔹 `type SchedulerService interface`
未来の時刻に通知を予約する管理窓口である．
*   `ScheduleTaskRemind(ctx, task, userID, interval, style, runAt)`: 課題通知の予約．
*   `CancelTaskReminds(ctx, taskID, userID)`: 指定課題・ユーザーの全予約を取消．
*   `ScheduleWakeupSOS(ctx, checkID, userID, groupID, runAt)`: 起床 SOS の予約．
*   `CancelWakeupSOS(ctx, checkID)`: SOS 予約の取消．

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

---
*最終更新日: 2026年6月14日*
