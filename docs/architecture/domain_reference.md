# Uni-Steps ドメイン層 リファレンスマニュアル

本ドキュメントは，`domain/` 配下で定義されている各ファイルごとの構造体（エンティティ）とインターフェース（契約）の仕様を網羅したものである．

---

## 1. domain/ai.go (知能の定義)

### 🔹 `type AIService interface`
AI による文章生成の窓口である．具体的なモデル（Gemini 等）への接続はインフラ層で行う．

*   🔧 **メソッド: `GenerateRemindMessage(ctx, task, style) (string, error)`**
    *   **概要**: 課題の内容と期限に基づき，ユーザーのやる気を引き出す（あるいは焦らせる）個別の励ましメッセージを作成する．
    *   **引数**:
        *   `ctx`: コンテキスト（タイムアウトやキャンセル管理）．
        *   `task`: `*domain.Task` 型．通知対象の課題データ（タイトルや期限）．
        *   `style`: `string` 型．AI の性格設定（`strict`, `kind` 等）．
    *   **戻り値**: AI が生成したリマインド文，およびエラー．
*   🔧 **メソッド: `GenerateGroupSummaryMessage(ctx, workloadSummary, style) (string, error)`**
    *   **概要**: チーム全体の未完了課題の溜まり具合を鳥瞰し，朝や夜に配信するための「現状報告と応援」の要約メッセージを作成する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `workloadSummary`: `string` 型．グループの課題状況をテキスト化したもの．
        *   `style`: `string` 型．AI の性格設定．
    *   **戻り値**: チームの現状を把握しやる気を引き出すサマリー文，およびエラー．

---

## 2. domain/group.go (部屋の定義)

### 📦 `type Group struct`
共同作業の単位である「部屋」の情報．グループ内での通知ルールや AI の性格を決定する中心的なエンティティである．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 部屋のユニークID (PK/UUID) |
| `Name` | `string` | 部屋の名称（例：「卒論用」「月曜2限」） |
| `OwnerID` | `string` | 現在のオーナー（管理権限保持者）のユーザーID．オーナーのみが設定変更や部屋の削除を行える． |
| `LineChannelToken` | `string` | **BYOT (Bring Your Own Token) 方式**: ユーザーが自分で作成した LINE Bot のトークンを設定する． |
| `LineGroupID` | `string` | 通知の送信先となる LINE グループの内部 ID． |
| `LastSyncedAt` | `time.Time` | 最後に LMS との同期処理が行われた時刻．連続実行を防ぐための判定にも使われる． |
| `LMSLastUpdatedAt` | `time.Time` | 外部 LMS 側で最後に課題の更新が検知された時刻． |
| `InviteCode` | `string` | 8 桁の参加用招待コード．これが分かれば誰でもグループに参加できる． |
| `RemindIntervals` | `[]int` | リマインド通知を飛ばすタイミング（分前）の配列．例: `[1440, 60]` なら「1日前」と「1時間前」に通知する． |
| `AICharacter` | `string` | AI がメッセージを生成する際の性格設定（`default`, `strict`, `kind`, `cool`）． |
| `SummaryMorningTime` | `string` | 朝のサマリー（今日が期限の課題）を送信する時刻（"HH:mm"）． |
| `SummaryEveningTime` | `string` | 夜のサマリー（明日以降の課題展望）を送信する時刻（"HH:mm"）． |
| `Users` | `[]*User` | 所属している全メンバーのリスト（多対多の関係）． |

#### 🏷️ 定数定義 (domain/group.go)
AI の振る舞いやサマリーのモードを切り替えるために使用される．

*   **AI 性格設定 (`AICharacter`)**:
    *   `AICharacterDefault` ("default"): 標準的な親切なアシスタント．
    *   `AICharacterStrict` ("strict"): 厳しい軍隊の教官．遅延を一切許さず，ユーザーを厳しく奮起させるスタイル．
    *   `AICharacterKind` ("kind"): 心配性な幼馴染．ユーザーの体調やメンタルを気遣いつつ，優しくサポートするスタイル．
    *   `AICharacterCool` ("cool"): 冷徹で仕事が完璧な執事．感情を表に出さず，論理的かつ淡々と状況を報告するスタイル．
*   **サマリー種別 (`SummaryType`)**:
    *   `SummaryTypeMorning` ("morning"): 朝に配信される「今日一日の締切」に特化した要約メッセージ．
    *   `SummaryTypeEvening` ("evening"): 夜に配信される「明日以降の課題展望」を見据えたまとめメッセージ．

---

## 3. domain/lms.go (外部連携の定義)

### 🔹 `type LMSService interface`
外部学習管理システム（LMS）との連携を抽象化する．具体的な API 通信はインフラ層で実装される．

*   🔧 **メソッド: `FetchTasks(ctx, userID) ([]*Task, error)`**
    *   **概要**: ユーザーに紐づくすべての有効なコースを巡回し，未完了課題の情報と提出ステータスを最新の状態で引き抜く．
    *   **引数**: `userID` (`string`)．課題を取得したいユーザーの ID．
    *   **戻り値**: 取得した課題データのスライス（ドメインモデル変換済み），およびエラー．
*   🔧 **メソッド: `GetProviderName() string`**
    *   **概要**: 連携先の識別名を取得する．
    *   **戻り値**: プロバイダの識別名（例: "google_classroom"）．課題の出典（Source フィールド）として記録される．

---

## 4. domain/notification.go (通知・予約의定義)

### 🔹 `type NotificationService interface`
外部への通知送信を抽象化する窓口である．

*   🔧 **メソッド: `SendGroupMessage(ctx, targetID, message) error`**
    *   **概要**: LINE グループなどの「共有された空間」に対してメッセージを投稿する．
    *   **引数**: `targetID` (送信先ID)，`message` (本文)．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `SendDirectMessage(ctx, userID, message, targetURL) error`**
    *   **概要**: 特定の個人に対し，Web Push などの非公開な手段で通知を届ける．
    *   **引数**: `userID` (宛先ID)，`message` (本文)，`targetURL` (通知クリック時の遷移先相対パス)．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．

### 📦 `type ReminderJob struct`
予約されたリマインド通知の状態を保持するための内部モデルである．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | ジョブの一意識別子 (PK/UUID)． |
| `TaskID` | `string` | 紐づく課題の ID． |
| `UserID` | `string` | 宛先となるユーザーの ID． |
| `TargetTime` | `time.Time` | メッセージを送信する予定時刻． |
| `Message` | `string` | 送信する予定のメッセージ本文． |
| `Status` | `string` | **状態**: `pending`（待機中）, `sent`（送信済み）, `cancelled`（取消済み）． |

### 🔹 `type SchedulerService interface`
未来の特定の時刻に処理を予約する機能を抽象化する．Uni-Steps の「見守り」機能の核心を担う．

*   🔧 **メソッド: `ScheduleTaskRemind(ctx, task, userID, interval, style, runAt) error`**
    *   **概要**: 課題の締切前（例えば1時間前）に自動的にリマインドが飛ぶよう，未来のタイマーを仕掛ける．
    *   **引数**: `task` (対象課題)，`userID` (宛先)，`interval` (何分前か)，`style` (AI性格)，`runAt` (実行時刻)．
*   🔧 **メソッド: `CancelTaskReminds(ctx, taskID, userID) error`**
    *   **概要**: 課題を終わらせた場合などに，不要になった未来のリマインド予約をすべて解除する．
    *   **引数**: `taskID` (課題ID)，`userID` (ユーザーID)．
*   🔧 **メソッド: `ScheduleWakeupSOS(ctx, checkID, userID, groupID, runAt) error`**
    *   **概要**: 起床予定時刻を過ぎてもチェックインがない場合に，グループへ SOS を発信する予約を仕掛ける．
    *   **引数**: `checkID` (起床確認ID)，`userID` (本人)，`groupID` (通知先)，`runAt` (SOS発信時刻)．
*   🔧 **メソッド: `CancelWakeupSOS(ctx, checkID) error`**
    *   **概要**: 無事に起きた場合などに，予約されていた SOS 通知の送信を取り消す．
    *   **引数**: `checkID` (起床確認ID)．

---

## 5. domain/notification_log.go (履歴の定義)

### 📦 `type NotificationLog struct`
システムから送信された通知の履歴を記録するエンティティである．ダッシュボードの「活動履歴」のデータソースとなる．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | ログの一意識別子 (PK/UUID)． |
| `GroupID` | `string` | 通知が発生したグループの ID．どの部屋のタイムラインに表示するかを決定する． |
| `UserID` | `string` | 通知の対象となったユーザー，または SOS 発信の原因となったユーザーの ID． |
| `Type` | `string` | **通知種別**: メッセージの性質（`remind`, `sos`, `summary`）を表し，UI 上のアイコン出し分けに使われる． |
| `Message` | `string` | AI が生成した，あるいはシステムが自動作成したメッセージの本文． |
| `CreatedAt` | `time.Time` | メッセージが送信され，記録された日時． |

### 🔹 `type NotificationLogRepository interface`
通知履歴の保存と取得に関する契約を定義する．

*   🔧 **メソッド: `Save(ctx, log) error`**
    *   **概要**: 送信されたメッセージの内容とタイミングをデータベースに永続化する．
    *   **引数**: `log` (`*domain.NotificationLog`)．保存したいログオブジェクト．
*   🔧 **メソッド: `FindByGroupID(ctx, groupID, limit) ([]*NotificationLog, error)`**
    *   **概要**: ダッシュボードのタイムラインに表示するために，特定の部屋で発生した最新の通知履歴を時系列で取得する．
    *   **引数**: `groupID` (対象部屋)，`limit` (最大取得件数)．
    *   **戻り値**: ログリスト（日付降順），およびエラー．

#### 🏷️ 定数定義 (domain/notification_log.go)
*   `NotificationTypeRemind` ("remind"): 通常の課題リマインド通知．
*   `NotificationTypeSOS` ("sos"): 起床見守り失敗時の緊急 SOS 通知．
*   `NotificationTypeSummary` ("summary"): 朝刊・夕刊などの状況要約通知．

---

## 6. domain/repository.go (永続化の契約)
データベースに対する操作の「約束事」を定義する．具体的な実装はインフラ層（GORM 等）で行われる．

### 🔹 `type TaskRepository interface`
*   🔧 **メソッド: `Save(ctx, task) error`**
    *   **概要**: 課題の基本情報と，紐づく全メンバーの個別の進捗状況（`UserProgress`）をセットで保存または更新する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `task`: `*domain.Task` 型．保存・更新対象の課題データ．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `FindByID(ctx, id) (*Task, error)`**
    *   **概要**: 指定された内部 ID を持つ課題を 1 件取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．取得したい課題の UUID．
    *   **戻り値**: 取得した課題データ，およびエラー．
*   🔧 **メソッド: `FindByExternalID(ctx, extID) (*Task, error)`**
    *   **概要**: Google Classroom などの外部 ID をキーに検索し，課題が既に登録済みかチェックする．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `extID`: `string` 型．外部 LMS 側の一意な課題 ID．
    *   **戻り値**: 取得した課題データ，およびエラー．
*   🔧 **メソッド: `FindByGroupID(ctx, groupID) ([]*Task, error)`**
    *   **概要**: 特定の部屋に属するすべての課題を，期限の近い順に並べて取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `groupID`: `string` 型．対象となるグループの ID．
    *   **戻り値**: 課題データのリスト，およびエラー．
*   🔧 **メソッド: `FindApproachingDeadlines(ctx, until) ([]*Task, error)`**
    *   **概要**: 現時刻から指定された時刻までの間に期限を迎える，未完了の課題をバッチ処理用に抽出する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `until`: `time.Time` 型．期限の検索終了時刻．
    *   **戻り値**: 条件に合致する課題リスト，およびエラー．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: 課題本体と，それに関連付けられたメンバーの進捗データを物理的に削除する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．削除したい課題の ID．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．

### 🔹 `type UserRepository interface`
*   🔧 **メソッド: `Save(ctx, user) error`**
    *   **概要**: ユーザー情報やアクセストークンの情報を新規作成または更新する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `user`: `*domain.User` 型．保存・更新対象のユーザーデータ．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `UpdateWebPushToken(ctx, userID, token) error`**
    *   **概要**: 頻繁に更新される可能性があるブラウザの通知用トークンのみをピンポイントで更新する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `userID`: `string` 型．対象ユーザーの ID．
        *   `token`: `string` 型．Web Push 用の JSON トークン文字列．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `FindByID(ctx, id) (*User, error)`**
    *   **概要**: ユーザー ID をキーに 1 名の情報を取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．取得したいユーザーの ID．
    *   **戻り値**: 取得したユーザーデータ，およびエラー．
*   🔧 **メソッド: `FindByEmail(ctx, email) (*User, error)`**
    *   **概要**: ログイン時のメールアドレスをキーにユーザーを特定する．登録済みかどうかの判定に使われる．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `email`: `string` 型．検索キーとなるメールアドレス．
    *   **戻り値**: 取得したユーザーデータ，およびエラー．

### 🔹 `type GroupRepository interface`
*   🔧 **メソッド: `Save(ctx, group) error`**
    *   **概要**: 部屋の名前，AI性格設定，LINE連携情報などの設定を保存または更新する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `group`: `*domain.Group` 型．保存・更新対象の部屋データ．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `FindByID(ctx, id) (*Group, error)`**
    *   **概要**: 部屋 ID をキーに 1 件の部屋情報を取得する（所属メンバーも含む）．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．取得したい部屋の ID．
    *   **戻り値**: 取得した部屋データ，およびエラー．
*   🔧 **メソッド: `FindAllGroups(ctx) ([]*Group, error)`**
    *   **概要**: システム内のすべての有効な部屋をスキャンする．定期サマリーの配信対象を探すために使われる．
    *   **引数**:
        *   `ctx`: コンテキスト．
    *   **戻り値**: 全ての部屋データのリスト，およびエラー．
*   🔧 **メソッド: `FindByInviteCode(ctx, code) (*Group, error)`**
    *   **概要**: 8 桁の招待コードに基づき，特定の部屋を特定する．参加機能の入り口となる．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `code`: `string` 型．検索したい招待コード．
    *   **戻り値**: 一致する部屋データ，およびエラー．
*   🔧 **メソッド: `FindByUserID(ctx, userID) ([]*Group, error)`**
    *   **概要**: 指定したユーザーが所属しているすべての部屋を一覧で取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `userID`: `string` 型．検索対象となるユーザーの ID．
    *   **戻り値**: 所属している部屋リスト，およびエラー．
*   🔧 **メソッド: `RemoveUser(ctx, groupID, userID) error`**
    *   **概要**: ユーザーと部屋の紐付けを解除する（退出処理）．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `groupID`: `string` 型．対象の部屋 ID．
        *   `userID`: `string` 型．脱退させるユーザーの ID．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: 部屋を完全に削除し，紐づく全メンバーとの関係性（中間テーブル）も解消する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．削除したい部屋の ID．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．

### 🔹 `type WakeupRepository interface`
*   🔧 **メソッド: `Save(ctx, check) error`**
    *   **概要**: 起床見守りスケジュールを新規作成または状態更新する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `check`: `*domain.WakeupCheck` 型．保存・更新対象の起床確認データ．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `Delete(ctx, id) error`**
    *   **概要**: 起床見守り予約を物理的に削除する（キャンセルの際など）．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `id`: `string` 型．削除したい起床確認データの ID．
    *   **戻り値**: 成功時は nil，失敗時は error を返す．
*   🔧 **メソッド: `FindPendingByTime(ctx, now) ([]*WakeupCheck, error)`**
    *   **概要**: 指定時刻を過ぎても status が `pending` のままの（起きた報告がない）予約を抽出し，SOS 発信のトリガーとする．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `now`: `time.Time` 型．判定の基準となる現在時刻．
    *   **戻り値**: 条件に合致する起床確認リスト，およびエラー．
*   🔧 **メソッド: `FindActiveByUser(ctx, userID) ([]*WakeupCheck, error)`**
    *   **概要**: 指定ユーザーの，現在進行中（未完了）の見守り予約を 1 件取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `userID`: `string` 型．対象ユーザーの ID．
    *   **戻り値**: 起床確認データのリスト（通常は1件），およびエラー．
*   🔧 **メソッド: `FindActiveByGroup(ctx, groupID) ([]*WakeupCheck, error)`**
    *   **概要**: 指定グループの本日分の起床状況を全メンバー分スキャンし，ダッシュボード表示用に取得する．
    *   **引数**:
        *   `ctx`: コンテキスト．
        *   `groupID`: `string` 型．対象グループの ID．
    *   **戻り値**: 起床状況のリスト，およびエラー．

---

## 7. domain/task.go (課題の定義)

### 📦 `type Task struct`
課題の基本情報を保持するエンティティ．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 課題のユニークID (PK/UUID) |
| `GroupID` | `string` | 所属する部屋の ID． |
| `Source` | `string` | **出典**: `manual` (手動), `google_classroom` 等．出典により編集権限が制御される． |
| `ExternalID` | `string` | 外部 LMS 側の ID．重複登録を防止するためのユニークキーとなる． |
| `RawText` | `string` | ユーザーが入力した生のテキスト（AI 解析用）． |
| `CreatorID` | `string` | 手動課題の作成者のユーザー ID． |
| `Title` | `string` | 課題のタイトル． |
| `Deadline` | `time.Time` | 提出期限．西暦 1 年（Zero値）の場合は「期限未定」として扱われる． |
| `IsLMSDeadlineSet` | `bool` | 元々外部 LMS 側で期限が設定されていたか． |
| `LMSUpdateTime` | `time.Time` | 外部 LMS 側での最終更新日時． |
| `Recurrence` | `RecurrenceSettings` | 繰り返し設定（JSON型）． |
| `UserProgress` | `[]*TaskUserProgress` | **進捗リスト**: 課題にアサインされた各メンバーの完了状態を保持する（Has-Many関係）． |

### 📦 `type TaskUserProgress struct`
特定の課題に対する，個別のユーザーの完了状態を管理する．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `TaskID` | `string` | 対象課題の ID (PK/FK)． |
| `UserID` | `string` | 対象ユーザーの ID (PK/FK)． |
| `UserName` | `string` | 表示用のユーザー名（進捗確認時に使用）． |
| `IsCompleted` | `bool` | 完了フラグ．その人が課題を終えたかを表す． |
| `UpdatedAt` | `time.Time` | 進捗が最後に更新された日時． |

#### 📦 `type RecurrenceSettings struct`
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `Type` | `string` | `none`, `weekly`, `biweekly`, `custom` |
| `CustomDates` | `[]time.Time` | `custom` 時のみ使用する特定日付のリスト |

#### 🏷️ 定数定義 (domain/task.go)
課題の発生元を識別するために使用される．

*   **出典 (`Source`)**:
    *   `SourceManual` ("manual"): ユーザーがアプリの UI から直接タイトルや期限を入力して作成した課題．
    *   `SourceGoogleClassroom` ("google_classroom"): Google Classroom API を通じて自動的に同期された課題．
    *   `SourceWebClass` ("web_class"): Web Class 等の外部学習システムから同期された課題（将来拡張用）．

---

## 8. domain/user.go (ユーザーの定義)

### 📦 `type User struct`
ユーザーの身分情報および，外部サービスとの連携状態を保持するエンティティ．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | ユーザーの一意識別子 (PK/UUID)． |
| `Name` | `string` | 表示名（Google アカウントの氏名など）． |
| `Email` | `string` | メールアドレス (UK)．ログイン時のキーおよび一意性の保証に使われる． |
| `WebPushToken` | `string` | ブラウザの Web Push 購読情報（JSON形式）．個人宛通知の送信先となる． |
| `GoogleAccessToken` | `string` | Classroom API にアクセスするための認可トークン． |
| `GoogleRefreshToken` | `string` | アクセストークンを更新するためのリフレッシュトークン． |
| `GoogleTokenExpiry` | `time.Time` | アクセストークンの有効期限．期限が近い場合はリフレッシュが行われる． |
| `LastCheckInAt` | `time.Time` | 最後に「起床確認」または「ログイン」が確認された時刻．生活リズムの生存確認基準． |
| `Groups` | `[]*Group` | ユーザーが所属している全ての部屋のリスト（多対多の関係）． |

---

## 9. domain/wakeup.go (起床確認の定義)

### 📦 `type WakeupCheck struct`
起床予定と現在の状態を保持するエンティティ．
| フィールド | 型 | 説明 |
| :--- | :--- | :--- |
| `ID` | `string` | 一意識別子 (PK/UUID)． |
| `UserID` | `string` | 対象ユーザーの ID． |
| `GroupID` | `string` | 起きなかった場合に通知を飛ばす先の部屋 ID． |
| `TargetTime` | `time.Time` | 起床を約束した予定時刻． |
| `GraceMinutes` | `int` | 予定時刻から SOS を出すまでの猶予期間（分）． |
| `Status` | `string` | **現在の状態**: `pending`, `confirmed`, `alerted` のいずれか． |
| `CreatedAt` | `time.Time` | 予約が作成された日時． |

#### 🏷️ 定数定義 (domain/wakeup.go)
*   `WakeupStatusPending` ("pending"): 起床確認の待機中．
*   `WakeupStatusConfirmed` ("confirmed"): ユーザーが無事に起床報告を完了した状態（成功）．
*   `WakeupStatusAlerted` ("alerted"): 猶予期間を過ぎても報告がなく，SOS 通知が送信された状態（失敗）．

---
*最終更新日: 2026年6月14日*
