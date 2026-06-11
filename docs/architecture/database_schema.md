# Uni-Steps データベース設計書 (ER図)

本ドキュメントは，Uni-Steps システムにおけるデータベースの構造，テーブル間の関係性，および各カラムの詳細を定義したものである．

## 1. データベース概要
- **エンジン**: PostgreSQL (Supabase)
- **ORM**: GORM (Go Object Relational Mapping)
- **タイムゾーン**: 全てのタイムスタンプは日本時間 (Asia/Tokyo) を基準とする．

## 2. ER 図 (Entity Relationship Diagram)

```mermaid
erDiagram
    USERS ||--o{ USER_GROUPS : "所属する"
    GROUPS ||--o{ USER_GROUPS : "含まれる"
    GROUPS ||--o{ TASKS : "所有する"
    TASKS ||--o{ TASK_USER_PROGRESS : "進捗を管理"
    USERS ||--o{ TASK_USER_PROGRESS : "実行する"
    USERS ||--o{ WAKEUP_CHECKS : "起床確認を行う"

    USERS {
        string id PK "UUID"
        string name "ユーザー名"
        string email "メールアドレス"
        string google_access_token "Google 認証トークン"
        string google_refresh_token "Google リフレッシュトークン"
        string web_push_token "Web Push 購読情報 (JSON)"
    }

    GROUPS {
        string id PK "UUID"
        string name "部屋名"
        string owner_id FK "作成者のユーザー ID"
        string invite_code UK "8 桁の招待コード"
        string line_channel_token "LINE Messaging API トークン"
        string line_group_id "通知先 LINE グループ ID"
        timestamp last_synced_at "最終同期時刻"
        timestamp lms_last_updated_at "LMS 側の最終更新時刻"
    }

    USER_GROUPS {
        string user_id PK, FK "ユーザー ID"
        string group_id PK, FK "グループ ID"
    }

    TASKS {
        string id PK "UUID"
        string group_id FK "所属グループ ID"
        string source "入力元 (manual / google_classroom)"
        string external_id UK "外部 LMS 側の課題 ID"
        string title "課題のタイトル"
        timestamp deadline "提出期限 (1/1/1 は未定扱い)"
        json recurrence "繰り返し設定 (type, custom_dates 含む JSON)"
    }

    TASK_USER_PROGRESS {
        string task_id PK, FK "課題 ID"
        string user_id PK, FK "ユーザー ID"
        string user_name "表示用ユーザー名"
        bool is_completed "完了フラグ"
        timestamp updated_at "ステータス更新日時"
    }

    WAKEUP_CHECKS {
        string id PK "UUID"
        string user_id FK "ユーザー ID"
        timestamp scheduled_time "確認予定時刻"
        timestamp confirmed_at "実際の確認時刻"
        string status "状態 (pending / success / failed)"
    }
```

## 3. テーブル定義詳細

### 3.1 USERS (ユーザー)
システムを利用する個人を管理する．
- `web_push_token`: ブラウザから取得した通知用エンドポイント情報を JSON 文字列として保持する．

### 3.2 GROUPS (部屋)
複数のユーザーが課題を共有する単位である．
- `invite_code`: ユニーク制約を持ち，他のユーザーが部屋に参加するためのキーとなる．
- `last_synced_at`: Google Classroom との自動同期におけるクールダウン制御やログとして使用する．

### 3.3 TASKS (課題)
LMS から同期された課題，または手動で作成された課題の基本情報を保持する．
- `external_id`: Google Classroom 側での ID．これを用いて同期時の重複登録を防止する．
- `deadline`: 未設定の場合は Go のゼロ値 (`0001-01-01`) が入り，UI 上は「未定」と表示される．
- `recurrence`: 繰り返しのタイプ（none/weekly 等）と，カスタム日付のリストを包含する JSON オブジェクトである．
「1つの課題に対して複数人の進捗」を管理するための中間テーブル的役割を持つ．
- `task_id` と `user_id` の複合主キーにより，同一ユーザーが同一課題に複数の進捗を持つことを防ぐ．
- `is_completed`: このフラグが `true` になった際，`updated_at` に完了時刻が記録される．

### 3.5 USER_GROUPS (ユーザー・グループ紐付け)
どのユーザーがどの部屋に所属しているかを管理する多対多の交差テーブルである．

---
*最終更新日: 2026年6月11日*
