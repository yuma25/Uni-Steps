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
    GROUPS ||--o{ NOTIFICATION_LOGS : "履歴を残す"

    USERS {
        string id PK "UUID"
        string name "ユーザー名"
        string email "メールアドレス"
        string google_access_token "Google 認証トークン"
        string google_refresh_token "Google リフレッシュトークン"
        timestamp google_token_expiry "Google トークン有効期限"
        string web_push_token "Web Push 購読情報 (JSON)"
        timestamp last_check_in_at "最終チェックイン時刻"
        json groups "所属グループ一覧 (M:M)"
    }

    GROUPS {
        string id PK "UUID"
        string name "部屋名"
        string owner_id FK "作成者のユーザー ID"
        string invite_code UK "8 桁の招待コード"
        string ai_character "AI の性格設定"
        string line_channel_token "LINE Messaging API トークン"
        string line_group_id "通知先 LINE グループ ID"
        string summary_morning_time "朝刊の送信時刻"
        string summary_evening_time "夕刊の送信時刻"
        json remind_intervals "通知タイミングの配列 (JSON)"
        timestamp last_synced_at "最終同期時刻"
        timestamp lms_last_updated_at "LMS 側の最終更新時刻"
        json users "所属メンバー一覧 (M:M)"
    }

    USER_GROUPS {
        string user_id PK, FK "ユーザー ID"
        string group_id PK, FK "グループ ID"
    }

    TASKS {
        string id PK "UUID"
        string group_id FK "所属グループ ID"
        string source "入力元 (manual / google_classroom)"
        string creator_id "手動課題の作成者 ID"
        string external_id UK "外部 LMS 側の課題 ID"
        string raw_text "AI 解析用原文"
        string title "課題のタイトル"
        timestamp deadline "提出期限 (1/1/1 は未定扱い)"
        bool is_lms_deadline_set "LMS 側に期限があったか"
        timestamp lms_update_time "LMS 側の更新日時"
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
        string group_id FK "グループ ID"
        timestamp target_time "起床予定時刻"
        int grace_minutes "猶予時間"
        string status "状態 (pending / confirmed / alerted)"
        timestamp created_at "作成日時"
    }

    NOTIFICATION_LOGS {
        string id PK "UUID"
        string group_id FK "グループ ID"
        string user_id FK "対象ユーザー ID"
        string type "種別 (remind / sos / summary)"
        string message "通知内容"
        timestamp created_at "送信日時"
    }
```

## 3. テーブル定義詳細

### 3.1 TASKS (課題)
- `external_id`: 外部 LMS における課題の一意識別子である．`group_id` との複合ユニークインデックス (`idx_group_external`) が設定されており，異なる部屋であれば同じ外部課題を重複して保持することが可能である．
- `creator_id`: 手動で登録された課題において，誰が作成したかを管理する．権限チェック（編集・削除）に使用される．

---
*最終更新日: 2026年6月15日*
