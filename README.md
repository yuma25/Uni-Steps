# Uni-Steps

AIと仲間が支える課題管理プラットフォームである．

## 📁 プロジェクト構造とファイル内容

各ディレクトリの役割と，現在作成済みの主要ファイルの内容である．

### 1. `domain/` (ドメインレイヤー)
ビジネスロジックの核となるデータ構造とインターフェースである．
*   `ai.go`: AI による文章生成インターフェース定義である．
*   `group.go`: グループ情報および同期状態の構造体定義である．
*   `lms.go`: 外部 LMS 連携インターフェース定義である．
*   `notification.go`: 通知サービス（LINE/Web Push）のインターフェース定義である．
*   `repository.go`: データベース操作（Task, User, Group）のインターフェース定義である．
*   `task.go`: 課題・予定情報の詳細な構造体定義である．
*   `user.go`: ユーザー情報および OAuth トークンの構造体定義である．

### 2. `usecase/` (ユースケースレイヤー)
アプリの具体的なビジネスシナリオを組み立てる場所である．
*   `monitor_uc.go`: 定期的なデータベース監視とリマインド通知のロジックである．
*   `sync_uc.go`: 外部 LMS からの課題同期ロジックである（クールダウン・差分検知付き）．
*   `task_uc.go`: 手動による課題登録・一覧取得のロジックである．

### 3. `interfaces/` (インターフェースレイヤー)
外部（HTTPリクエスト等）との窓口である．
*   `handler/auth_handler.go`: Google OAuth 2.0 認証（ログイン・コールバック）のハンドラーである．
*   `handler/notification_handler.go`: Web Push 購読情報の登録ハンドラーである．
*   `handler/task_handler.go`: 課題の登録（手動）・同期・取得のハンドラーである．

### 4. `infrastructure/` (インフラレイヤー)
データベースや外部 API との具体的な通信処理である．
*   `ai/gemini_service.go`: Google Gemini API を用いたリマインド文生成の実装である．
*   `db/`: GORM を用いた PostgreSQL (Supabase) への各リポジトリ実装である．
*   `line/line_service.go`: LINE Messaging API を用いたグループ通知の実装である．
*   `lms/google_classroom.go`: Google Classroom API を用いた課題取得の実装である．
*   `notification/composite_service.go`: LINE と Web Push を統合した通知サービスの実装である．
*   `webpush/webpush_service.go`: Web Push プロトコルによる個人宛通知の実装である．

### 5. `cmd/` (コマンド)
*   `vapid/main.go`: Web Push に必要な VAPID キーペアを生成するユーティリティである．

### 6. `docs/` (ドキュメント)
*   `overview.md`: アプリのコンセプトと機能概要である．
*   `architecture/design.md`: クリーンアーキテクチャと処理フローの図解である．
*   `infrastructure/`: 各種外部サービス（LINE, Google Classroom）の設定ガイドである．
*   `references/packages.md`: 使用ライブラリの公式リファレンス集である．
*   `references/commands.md`: 開発・実行用コマンドのまとめである．

---

## 🛠 エンジニアの開発工程

実際のプロの現場で行われる開発のステップである．詳細は `docs/engineering_workflow.md` を参照すること．
