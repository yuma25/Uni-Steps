package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/infrastructure/ai"
	"github.com/yuma25/Uni-Steps/infrastructure/db"
	"github.com/yuma25/Uni-Steps/infrastructure/line"
	"github.com/yuma25/Uni-Steps/interfaces/handler"
	"github.com/yuma25/Uni-Steps/usecase"
	"google.golang.org/api/option"
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

	// --- インフラストラクチャ（道具）の初期化 ---
	taskRepo := db.NewTaskRepository(gormDB)

	// AI サービスの初期化
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	genaiClient, err := genai.NewClient(context.Background(), option.WithAPIKey(geminiAPIKey))
	if err != nil {
		log.Fatalf("Gemini クライアントの作成に失敗した: %v", err)
	}
	defer genaiClient.Close()
	aiService := ai.NewGeminiService(genaiClient, "gemini-2.0-flash")

	// LINE サービスの初期化
	lineToken := os.Getenv("LINE_CHANNEL_TOKEN")
	lineService, err := line.NewLineService(lineToken)
	if err != nil {
		log.Fatalf("LINE サービスの作成に失敗した: %v", err)
	}
	// TODO: LMS サービス（Google Classroom等）の初期化もここで行う．

	// --- ユースケース（現場監督）の初期化 ---
	taskUsecase := usecase.NewTaskUsecase(taskRepo, aiService)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, nil) // LMS 実装待ちのため nil
	monitorUsecase := usecase.NewMonitorUsecase(taskRepo, aiService, lineService)

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
