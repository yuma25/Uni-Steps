# Uni-Steps バックエンドエントリーポイント (backend/main.go) の詳細解説

本書は、Uni-Stepsプロジェクトのバックエンドの起動エントリーポイントである backend/main.go の全容を、コードの先頭から末尾まで詳細に解説したドキュメントです。

---

## 1. 概要

backend/main.go は、Go言語で実装されたAPIサーバーの主要な起動処理を担っています。本システムはクリーンアーキテクチャの設計思想に基づいて構築されており、以下のプロセスを順次実行してWebアプリケーションを稼働させます。

1. システム環境（タイムゾーン・環境変数）のセットアップ
2. データベース接続の確立とデータベースマイグレーション
3. インフラ構造体（リポジトリ、外部APIサービス）の初期化
4. ユースケース（ビジネスロジック）の初期化
5. バックグラウンドタスク（定期サマリ通知・自動同期）の常駐起動
6. HTTPサーバー（Echo）の設定・ミドルウェア適用・ハンドラー登録と起動

---

## 2. コード各セクションの逐次解説

以下、ファイルの行順に沿って具体的なコードの役割を解説します。

### 2.1 パッケージ定義とインポート (L1 - L30)

```go
package main

import (
	"context"
	"log"
	...
)
```

* **役割**: このファイルがプログラム全体の起動口（エントリーポイント）であることを示す `package main` が定義され、標準ライブラリおよび外部ライブラリをインポートしています。
* **主要な外部パッケージ**:
  * `github.com/google/generative-ai-go/genai`: Gemini APIとの通信を担う公式SDK。
  * `github.com/labstack/echo/v4`: 軽量で高速なWebフレームワーク「Echo」。
  * `gorm.io/gorm`: PostgreSQLなどのリレーショナルデータベース操作を容易にするORM。
* **プロジェクト内部モジュール**:
  * `domain` (ドメインモデル定義), `infrastructure` (DBやLINE/WebPushなどの外部インターフェース実装), `usecase` (業務ロジックのフロー管理), `interfaces/handler` (HTTPリクエスト処理)。

### 2.2 タイムゾーンの設定 (L32 - L39)

```go
func main() {
	// 0. タイムゾーンを日本時間 (JST) に設定する．
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("タイムゾーンの読み込みに失敗した（デフォルトを使用する）： %v\n", err)
	} else {
		time.Local = loc
	}
```

* **役割**: アプリケーション内部のシステム時刻(`time.Local`)を `Asia/Tokyo`（日本時間: JST）に固定します。これにより、起床予定時刻や課題の期限の判定処理を日本標準時基準で正確に行うことができます。

### 2.3 環境変数のロード (L41 - L49)

```go
	// 1. 環境変数の読み込み（環境に応じてファイルを切り替える）
	env := os.Getenv("GO_ENV")
	if env == "development" {
		_ = godotenv.Load(".env.development")
	}
	_ = godotenv.Load() // デフォルトの .env 読み込む
```

* **役割**: 実行環境を示す環境変数 `GO_ENV` の値に基づき `.env.development` ファイルをロードして環境変数を設定します。最後に共通の `.env` もロードします。

### 2.4 データベース接続とマイグレーション (L50 - L97)

```go
	// 2. データベース接続の確立
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL が設定されていない．.env ファイルを確認すること．")
	}

	if !strings.Contains(dbURL, "TimeZone=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&TimeZone=Asia/Tokyo"
		} else {
			dbURL += "?TimeZone=Asia/Tokyo"
		}
	}

	log.Println("データベースへの接続を開始中（最大5秒待機）...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, // プリペアドステートメントを完全に無効化
	}), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("データベース接続に失敗した: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("データベースへの Ping に失敗した: %v", err)
	}

	// 古い単一ユニークインデックスを削除し，新しい複合インデックスが適用されるようにする．
	_ = gormDB.Exec("DROP INDEX IF EXISTS idx_tasks_external_id").Error

	err = gormDB.AutoMigrate(
		&domain.User{},
		&domain.Group{},
		&domain.Task{},
		&domain.TaskUserProgress{},
		&domain.WakeupCheck{},
		&domain.NotificationLog{},
	)
	if err != nil {
		log.Fatalf("マイグレーションに失敗した: %v", err)
	}
```

* **DSN設定とタイムゾーン**:
  環境変数 `DATABASE_URL` から接続情報を取得し、明示的に `TimeZone=Asia/Tokyo` パラメータを末尾に連結させて、データベースレベルでの時刻処理が日本標準時（JST）で行われるよう保証します。
* **接続制限時間（タイムアウト）の設定**:
  `context.WithTimeout` を使用して、データベースの接続試行に「最大5秒」の時間制限を設定しています。ネットワーク切断などで無限に接続待ち（フリーズ）状態になるのを防ぎ、タイムアウトしたらエラーとしてプロセスを安全に停止させます。
* **接続オプション（Supabase等のサーバーレス対策）**:
  GORM経由で接続を開く際、`PreferSimpleProtocol: true` および `PrepareStmt: false` を設定します。これにより「プリペアドステートメント（SQLの事前解析）」機能を無効化します。Supabaseや外部のコネクションプーラー（PgBouncer等）を経由する環境では、プリペアドステートメントが共有接続内で競合してエラーを引き起こしやすいため、これを無効化して安全性の高い「単純プロトコル」での通信を強制しています。
* **疎通確認 (Ping)**:
  `gormDB.DB()` から生のSQLコネクションを取得し、`PingContext` を用いて、実際にデータベースサーバーとのパケット通信が確立し応答があるかをテストします。応答がなければそこでプログラムを強制終了させます。
* **スキーマ変更（古いインデックスのクリーンアップ）**:
  `DROP INDEX IF EXISTS idx_tasks_external_id` を直接実行します。これは旧設計での「外部タスクIDが全体で一意」という制約を外し、新設計の「グループごとでタスクIDを一意にする（複合制約）」ための調整用です。古いインデックスが残っているとスキーマ衝突エラーとなるため事前に破棄します。
* **自動マイグレーション (AutoMigrate)**:
  `User`, `Group`, `Task` などのGo言語の構造体設計図（`domain`パッケージ内）を順にスキャンし、テーブルが未作成であれば作成し、不足しているカラム（列）があれば自動でテーブルスキーマに追記・同期します。もしマイグレーションに失敗した場合は、起動時の不整合を防ぐためその場で起動処理を落とします。

### 2.5 インフラ層（リポジトリとサービス）の初期化 (L98 - L139)

```go
	// --- インフラ層（道具）の初期化 ---
	taskRepo := db.NewTaskRepository(gormDB)
	userRepo := db.NewUserRepository(gormDB)
	...
```

* **DBリポジトリ**: データベース操作を行うリポジトリクラスを初期化します。
* **Gemini AI サービス**:
  環境変数 `GEMINI_API_KEY` および `GEMINI_MODEL`（指定なしの場合はデフォルト `gemini-2.0-flash`）を使用してクライアントを立ち上げます。
* **通知機能（LINE / WebPush）**:
  ブラウザプッシュ用のVAPIDキー情報を元に `WebPushService` を生成し、グループ情報を元に `LineService` を生成。最終的にこれらを統合した `CompositeNotificationService` を構築し、両チャネルへ透過的に通知できる構成を作ります。
* **Google Classroom API 設定**:
  課題の取得に必要なOAuth2設定（クライアントID、シークレット、リダイレクトURI、権限スコープ）を定義し、Classroom API連携を行うサービスを初期化します。
* **スケジューラー**:
  アプリ内部で管理される通知タイマーやタスク実行のためのメモリ内スケジューラー（`InMemScheduler`）を準備します。

### 2.6 ユースケース層（ビジネスロジック）の初期化 (L140 - L146)

```go
	// --- ユースケース（現場監督）の初期化 ---
	taskUsecase := usecase.NewTaskUsecase(taskRepo, groupRepo, aiService, schService)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, groupRepo, lmsService, schService)
	...
```

* **役割**: リポジトリやサービスなどの「インフラ部品」を組み合わせ、実際の業務手続き（課題登録、同期、グループ操作、起床チェック、サマリ配信など）を実行するユースケースインスタンスを作成します。コントローラー（ハンドラー）とインフラ層の橋渡し役です。

### 2.7 バックグラウンド定期処理（ゴルーチン） (L147 - L180)

```go
	// 毎分，全グループのサマリー送信設定を確認するバックグラウンド処理
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		...
		if err := summaryUsecase.SendAllSummaries(ctx, now); err != nil { ... }
	}()
```

* **サマリ配信 (毎分監視)**: `time.NewTicker(1 * time.Minute)` を使用し、バックグラウンドの別スレッド（ゴルーチン）で毎分、各グループのサマリ通知（朝刊・夕刊）の配信設定時刻に達したかを確認・送信します。
* **自動同期 (6時間ごと)**:
  ```go
  go func() {
      ticker := time.NewTicker(6 * time.Hour)
      ...
      if err := syncUsecase.SyncAllGroups(ctx); err != nil { ... }
  }()
  ```

  6時間ごとにGoogle Classroom等から最新の課題情報をバックグラウンドで自動取得・同期します。

### 2.8 Echo Webサーバーの設定・ミドルウェア適用 (L182 - L199)

```go
	// Echo サーバーの初期化
	e := echo.New()

	// ログ出力を人間が読みやすい形式にカスタマイズ
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "{\n  \"time\": \"${time_rfc3339}\",\n  ... \n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(...)
```

* **CORS設定**: フロントエンド（ローカル開発環境の `http://localhost:5173` や 本番環境の `https://uni-steps.vercel.app`）からのクロスオリジンリクエストを許可します。
* **ヘルスチェックエンドポイント**: `/health` へのリクエストに対して正常稼働メッセージを返却する基本ルートを定義します。

### 2.9 ハンドラーのルーティング登録と起動 (L200 - L213)

```go
	// ハンドラー（窓口）の初期化とルーティング登録
	handler.NewTaskHandler(e, taskUsecase, syncUsecase)
	handler.NewNotificationHandler(e, userRepo, aiService, compositeNotifService)
	...

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	e.Logger.Fatal(e.Start(":" + port))
}
```

* **ハンドラーの登録**:
  Echoインスタンス `e` と初期化したユースケース層を、各種ハンドラー（Task, Notification, Auth, Group, Wakeup, User）に渡し、APIの各URLエンドポイントと処理を結びつけます。
* **サーバー起動**:
  環境変数 `PORT` が指定されていればその値を、なければデフォルトの `8080`番ポートを使用し、HTTPサーバーを立ち上げます。起動に致命的なエラーが発生した場合はログに記録した上でプロセスを強制終了させます。

---

## 3. クリーンアーキテクチャとしての役割整理

この backend/main.go の記述により、本プロジェクトの構成要素は以下のようなレイヤー依存に沿って綺麗に結合されます。

* **main.go の責任範囲**:
  各レイヤーを構成する具象クラス（`db.NewTaskRepository` や `usecase.NewTaskUsecase`）のインスタンスを生成し、これらを実行可能な状態に組み立てる（依存性の注入: DI）コンポーザの役割を果たしています。
