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
	// これにより time.Now() などがデフォルトで日本時間を返すようになる．
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

	// 接続文字列にタイムゾーン設定を追加する（PostgreSQL 用）．
	// すでにクエリパラメータがある場合を考慮しつつ追加する．
	if !strings.Contains(dbURL, "TimeZone=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&TimeZone=Asia/Tokyo"
		} else {
			dbURL += "?TimeZone=Asia/Tokyo"
		}
	}

	log.Println("データベースへの接続を開始中（最大5秒待機）...")

	// 接続タイムアウトを設定したコンテキストを作成する．
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Supabase のコネクションプーラー（6543ポート）を使用する場合，
	// プリペアドステートメントの衝突を防ぐために PreferSimpleProtocol を true に設定する必要がある．
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, // シンプルプロトコルを強制する．
	}), &gorm.Config{
		ConnPool: nil,
	})

	// 接続確認（Ping）を行い，実際に通信できるかテストする．
	if err == nil {
		sqlDB, _ := gormDB.DB()
		err = sqlDB.PingContext(ctx)
	}

	if err != nil {
		log.Fatalf("データベースの接続に失敗した．設定またはネットワークを確認すること: %v", err)
	}
	log.Println("データベースの接続に成功した．")

	// 2.5 データベースの自動マイグレーション
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

	// 2.6 デモ用データの投入 (Seeding)
	// フロントエンドの初期開発で使用するデフォルトグループが存在しない場合は作成する．
	var groupCount int64
	gormDB.Model(&domain.Group{}).Where("id = ?", "default-group-id").Count(&groupCount)
	if groupCount == 0 {
		gormDB.Create(&domain.Group{
			ID:      "default-group-id",
			Name:    "デフォルトグループ",
			OwnerID: "system-admin", // 初期データ用の管理者 ID である．
		})
		log.Println("デモ用のデフォルトグループを作成した．")
	}

	// 3. 依存性の注入（DI: Dependency Injection）
	log.Println("各サービスの初期化と依存性の注入を開始中...")

	// --- インフラストラクチャ（道具）の初期化 ---
	taskRepo := db.NewTaskRepository(gormDB)
	userRepo := db.NewUserRepository(gormDB)
	groupRepo := db.NewGroupRepository(gormDB)
	wakeupRepo := db.NewWakeupRepository(gormDB)

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

	// Web Push サービスの初期化
	vapidPub := os.Getenv("VAPID_PUBLIC_KEY")
	vapidPriv := os.Getenv("VAPID_PRIVATE_KEY")
	vapidContact := os.Getenv("VAPID_CONTACT") // 例: "mailto:admin@example.com"
	webPushService := webpush.NewWebPushService(userRepo, vapidPub, vapidPriv, vapidContact)

	// 通知サービス（LINE + Web Push の複合）の初期化
	compositeNotifService := notification.NewCompositeNotificationService(lineService, webPushService)

	// LMS サービス（Google Classroom）の初期化
	// 環境変数から OAuth 2.0 クライアント ID とシークレットを取得して設定する．
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	// Classroom API の読み取り権限およびユーザー識別用スコープを要求する．
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

	// --- ユースケース（現場監督）の初期化 ---
	taskUsecase := usecase.NewTaskUsecase(taskRepo, aiService)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, groupRepo, lmsService)
	monitorUsecase := usecase.NewMonitorUsecase(taskRepo, userRepo, groupRepo, wakeupRepo, aiService, compositeNotifService)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo)

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
	handler.NewNotificationHandler(e, userRepo)
	handler.NewAuthHandler(e, userRepo, oauthCfg)
	handler.NewGroupHandler(e, groupUsecase, lmsService)
	handler.NewWakeupHandler(e, wakeupRepo) // lmsService を追加

	log.Println("全てのコンポーネントの初期化が完了した．サーバーを起動する．")

	// 4. サーバーの起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
