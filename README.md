# Uni-Steps

AIと仲間が支える，次世代の課題管理・生活リズム支援プラットフォームである．

---

## 📂 詳細ディレクトリ構造とファイル解説

本プロジェクトはクリーンアーキテクチャを採用し，各レイヤーの責任を明確に分離している．

### 📁 Root
*   `main.go`: アプリケーションのエントリーポイントである．各部品の組み立て（依存性注入）とサーバーの起動を担う．
*   `.env.example`: プロジェクトに必要な環境変数のテンプレートである．
*   `go.mod` / `go.sum`: Go 言語の依存パッケージ管理ファイルである．

### 📁 domain/ (ドメイン層)
ビジネスロジックの核となるルールや「名詞」を定義する，最も内側のレイヤーである．
*   `user.go`: ユーザー情報の形状である．OAuth トークン等を含む．
*   `group.go`: 部屋（グループ）の形状と同期状態である．
*   `task.go`: 課題・予定の形状である．繰り返し設定や LMS 更新日時を持つ．
*   `wakeup.go`: 起床確認（WakeupCheck）の形状と状態である．
*   `repository.go`: データベース操作の抽象的なインターフェース（約束事）定義である．
*   `ai.go`: AI による文章生成機能のインターフェース定義である．
*   `lms.go`: 外部 LMS（Classroom 等）連携のインターフェース定義である．
*   `notification.go`: 通知送信機能のインターフェース定義である．

### 📁 usecase/ (ユースケース層)
ユーザーがやりたいこと（物語）の手順を組み立てるレイヤーである．
*   `task_uc.go`: 手動による課題の登録や一覧取得の手順である．
*   `sync_uc.go`: 外部 LMS から課題を同期するインテリジェントな手順である．
*   `group_uc.go`: 部屋の作成，参加，LMS 紐付けの手順である．
*   `monitor_uc.go`: 定期的な監視（課題リマインド，生存確認）の実行手順である．

### 📁 interfaces/ (インターフェース層)
外部（HTTP リクエスト等）との入出力を制御するレイヤーである．
*   `handler/`: Echo フレームワークを用いた HTTP ハンドラー群である．
    *   `auth_handler.go`: Google OAuth 2.0 認証（ログイン・コールバック）を管理する．
    *   `group_handler.go`: 部屋の作成，一覧取得，LMS 同期トリガーを管理する．
    *   `task_handler.go`: 課題の手動登録や取得を管理する．
    *   `wakeup_handler.go`: 起床予約とチェックイン（起床報告）を管理する．
    *   `notification_handler.go`: Web Push の購読情報の登録を管理する．

### 📁 infrastructure/ (インフラ層)
外部サービスやデータベースとの具体的な通信を担う最外周のレイヤーである．
*   `db/`: GORM を用いたデータベース（Supabase/PostgreSQL）へのアクセス実装である．
*   `ai/`: Google Gemini API を用いたリマインド文生成の実装である．
*   `line/`: LINE Messaging API を用いた通知送信の実装である．
*   `webpush/`: Web Push プロトコルによる個人デバイスへの通知実装である．
*   `notification/`: LINE と Web Push を統合した複合サービスの提供である．
*   `lms/`: Google Classroom API を用いた具体的な課題取得実装である．

### 📁 frontend/ (フロントエンド)
Vite + React + TypeScript を用いたユーザーインターフェースである．
*   `src/api/`: バックエンド API と通信するためのモジュール群である．
*   `src/pages/`: 各画面（ログイン，部屋選択，ダッシュボード）の実装である．
*   `src/types/`: TypeScript によるフロントエンド用の型定義である．

### 📁 cmd/ (ユーティリティコマンド)
*   `vapid/main.go`: Web Push に不可欠な VAPID キーペアを生成するための独立したツールである．

### 📁 docs/ (ドキュメント)
*   `overview.md`: プロジェクトの全体像とコア機能の解説である．
*   `architecture/`: 設計思想（クリーンアーキテクチャ）と処理フローの図解である．
*   `infrastructure/`: 各種サービス（Google, LINE）の設定手順書である．
*   `references/`: 技術仕様（packages.md）と開発コマンド（commands.md）の集約である．

---

## 🛠 エンジニアの開発工程
詳細は `docs/engineering_workflow.md` を参照すること．
1.  **Research**: 仕様と技術の調査である．
2.  **Strategy**: 全体像とアーキテクチャの設計である．
3.  **Execution**: 計画・実装・検証（CI/CD）の繰り返しである．
