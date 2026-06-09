package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/infrastructure/db"
	"github.com/yuma25/Uni-Steps/interfaces/handler"
	"github.com/yuma25/Uni-Steps/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. 環境変数の読み込み
	_ = godotenv.Load()

	// 2. データベース接続の確立
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL が設定されていない．.env ファイルを確認すること．")
	}

	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("データベースの接続に失敗した: %v", err)
	}

	// 2.5 データベースの自動マイグレーション
	// Go の構造体（Domain）を元に，必要なテーブルを自動で作成・更新する．
	log.Println("データベースのマイグレーションを実行中...")
	err = gormDB.AutoMigrate(
		&domain.User{},
		&domain.Group{},
		&domain.Task{},
	)
	if err != nil {
		log.Fatalf("マイグレーションに失敗した: %v", err)
	}
	log.Println("マイグレーションが完了した．")

	// 3. 依存性の注入（DI: Dependency Injection）
	// インフラ -> ユースケース -> ハンドラー の順に組み立てる．

	// インフラストラクチャ（道具）の初期化
	taskRepo := db.NewTaskRepository(gormDB)
	// TODO: LMS サービス（Google Classroom等）の初期化もここで行う．

	// ユースケース（現場監督）の初期化
	// ※現在 AIService と LMSService, NotificationService は nil を渡している（後ほど本実装と差し替える）
	taskUsecase := usecase.NewTaskUsecase(taskRepo, nil)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, nil)
	monitorUsecase := usecase.NewMonitorUsecase(taskRepo, nil, nil)

	// 3.5 監視プロセス（Goroutine）の起動
	// メインの HTTP サーバーの邪魔をしないように，`go` キーワードをつけて裏側（並行）で走らせる．
	go monitorUsecase.StartMonitoring(context.Background())

	// Echo サーバーの初期化
	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// ヘルスチェック
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Uni-Steps API is running")
	})

	// ハンドラー（窓口）の初期化とルーティング登録
	handler.NewTaskHandler(e, taskUsecase, syncUsecase)

	// 4. サーバーの起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
