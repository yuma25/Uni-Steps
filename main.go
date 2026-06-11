package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/infrastructure/ai"
	"github.com/yuma25/Uni-Steps/infrastructure/db"
	"github.com/yuma25/Uni-Steps/infrastructure/line"
	"github.com/yuma25/Uni-Steps/infrastructure/lms"
	"github.com/yuma25/Uni-Steps/infrastructure/notification"
	"github.com/yuma25/Uni-Steps/infrastructure/scheduler"
	"github.com/yuma25/Uni-Steps/infrastructure/webpush"
	"github.com/yuma25/Uni-Steps/interfaces/handler"
	"github.com/yuma25/Uni-Steps/usecase"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 0. タイムゾーンを日本時間 (JST) に設定する．
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("タイムゾーンの読み込みに失敗した（デフォルトを使用する）： %v\n", err)
	} else {
		time.Local = loc
	}

	// 1. 環境変数の読み込み
	_ = godotenv.Load()

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

	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		PrepareStmt: false, // プリペアドステートメントのキャッシュを無効化（Supabase などのプール環境でのエラー防止）
	})
	if err != nil {
		log.Fatalf("データベース接続に失敗した: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("データベースへの Ping に失敗した: %v", err)
	}
	log.Println("データベースの接続に成功した．")

	log.Println("データベースのマイグレーションを実行中...")
	err = gormDB.AutoMigrate(
		&domain.User{},
		&domain.Group{},
		&domain.Task{},
		&domain.TaskUserProgress{},
		&domain.WakeupCheck{},
	)
	if err != nil {
		log.Fatalf("マイグレーションに失敗した: %v", err)
	}
	log.Println("マイグレーションが完了した．")

	// デモ用の初期データ（必要に応じて）
	var count int64
	gormDB.Model(&domain.Group{}).Count(&count)
	if count == 0 {
		gormDB.Create(&domain.Group{ID: "default-group-id", Name: "デフォルトグループ", OwnerID: "system-admin"})
		log.Println("デモ用のデフォルトグループを作成した．")
	}

	log.Println("各サービスの初期化と依存性の注入を開始中...")

	// --- インフラ層（道具）の初期化 ---
	taskRepo := db.NewTaskRepository(gormDB)
	userRepo := db.NewUserRepository(gormDB)
	groupRepo := db.NewGroupRepository(gormDB)
	wakeupRepo := db.NewWakeupRepository(gormDB)

	// Gemini AI の初期化
	genaiClient, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalf("Gemini クライアントの初期化に失敗した: %v", err)
	}
	aiService := ai.NewGeminiService(genaiClient, "gemini-2.0-flash")

	// 通知サービスの初期化
	webPushService := webpush.NewWebPushService(userRepo, os.Getenv("VAPID_PUBLIC_KEY"), os.Getenv("VAPID_PRIVATE_KEY"), "mailto:admin@example.com")
	lineService, _ := line.NewLineService(os.Getenv("LINE_CHANNEL_TOKEN"))
	compositeNotifService := notification.NewCompositeNotificationService(lineService, webPushService)

	// Google Classroom API 設定
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	oauthCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/classroom.courses.readonly",
			"https://www.googleapis.com/auth/classroom.coursework.me.readonly",
		},
		Endpoint: google.Endpoint,
	}
	lmsService := lms.NewGoogleClassroomService(userRepo, oauthCfg)

	// スケジューラー（予約管理）の初期化
	schService := scheduler.NewInMemScheduler(aiService, compositeNotifService)

	// --- ユースケース（現場監督）の初期化 ---
	taskUsecase := usecase.NewTaskUsecase(taskRepo, groupRepo, aiService, schService)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, groupRepo, lmsService, schService)
	monitorUsecase := usecase.NewMonitorUsecase(taskRepo, userRepo, groupRepo, wakeupRepo, aiService, compositeNotifService)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo)

	// 3.5 監視プロセス（Goroutine）の起動
	go monitorUsecase.StartMonitoring(context.Background())

	// Echo サーバーの初期化
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Uni-Steps API is running")
	})

	// ハンドラー（窓口）の初期化とルーティング登録
	handler.NewTaskHandler(e, taskUsecase, syncUsecase)
	handler.NewNotificationHandler(e, userRepo)
	handler.NewAuthHandler(e, userRepo, oauthCfg)
	handler.NewGroupHandler(e, groupUsecase, lmsService)
	handler.NewWakeupHandler(e, wakeupRepo)

	log.Println("全てのコンポーネントの初期化が完了した．サーバーを起動する．")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
