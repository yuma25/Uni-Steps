# 【第3回】Uni-Stepsで学ぶクリーンアーキテクチャ：Interfaces と Infrastructure

Uni-Stepsのコードをベースにしたクリーンアーキテクチャ解説連載の第3回である．

前回は，データベースやフレームワークに依存しない，純粋なビジネスロジックである **Domainレイヤー** と **UseCaseレイヤー** について解説した．
今回は，それらのビジネスロジックを外界と接続するための **Interfaces（インターフェース）レイヤー** と **Infrastructure（インフラストラクチャ）レイヤー**，およびこれらすべてを結びつける **依存関係の注入（DI：Dependency Injection）** について解説する．

---

## 1. Interfaces（Handler）レイヤー：外部からのリクエストの受付窓口

同心円の外側に向かって進むと，まず現れるのが **Interfacesレイヤー** である．
本アプリでは `interfaces/handler/` 配下に配置されており，Webフレームワーク（Echo）を使用してHTTPリクエストを受け取り，UseCaseへ橋渡しをする役割を持つ．

### 役割：プロトコルの変換器
*   **入力の変換**: HTTPリクエストのボディ（JSON）をパースし，Goの構造体（ドメインモデル）に変換する．
*   **ユースケースの呼び出し**: 変換したデータをUseCaseに渡し，処理を実行させる．
*   **出力の変換**: UseCaseから返ってきたデータやエラーを，HTTPステータスコードやJSONレスポンスに変換してクライアント（フロントエンド）へ返却する．

### 実例：TaskHandler (interfaces/handler/task_handler.go)
実際のコード（一部抜粋）を示す．

```go
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/usecase"
)

// TaskHandler は Echo を使った HTTP リクエストの窓口である．
type TaskHandler struct {
	taskUsecase *usecase.TaskUsecase // UseCase層に依存
	syncUsecase *usecase.SyncUsecase
}

// NewTaskHandler はハンドラーを初期化し，ルーティングを登録する．
func NewTaskHandler(e *echo.Echo, tu *usecase.TaskUsecase, su *usecase.SyncUsecase) {
	h := &TaskHandler{
		taskUsecase: tu,
		syncUsecase: su,
	}
	// ルーティングの設定
	e.POST("/api/tasks/manual", h.CreateManualTask)
	e.PUT("/api/tasks/:id", h.UpdateTask)
}

// CreateManualTask は UI からの手動タスク登録リクエストを受け付ける．
func (h *TaskHandler) CreateManualTask(c echo.Context) error {
	task := new(domain.Task)
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}
	createdTask, err := h.taskUsecase.RegisterManualTask(c.Request().Context(), task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, createdTask)
}

// UpdateTask は課題の情報を更新する．
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	// 1. 各種パラメータ（引数）をリクエストから抽出する
	taskID := c.Param("id")           // URLパスの ":id" から抽出
	userID := c.QueryParam("user_id") // クエリパラメータ "?user_id=..." から抽出
	task := new(domain.Task)          // リクエストボディのJSONをバインドする構造体
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	// 2. 抽出した引数を渡して，UseCase（現場監督）を呼び出す
	updatedTask, err := h.taskUsecase.UpdateTask(c.Request().Context(), taskID, task, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if updatedTask == nil {
		return c.JSON(http.StatusOK, map[string]string{"message": "担当者がいなくなったため，課題を自動削除した"})
	}

	// 3. 成功レスポンス（HTTP 200 OK）と更新されたオブジェクトを返す
	return c.JSON(http.StatusOK, updatedTask)
}
```

---

### NewTaskHandlerの解説（初期化とルーティング）

`NewTaskHandler` は，この `TaskHandler` （受付窓口）を起動可能な状態にし，どのURLでどの処理を実行するかを定義する関数である．

1.  **依存関係の注入（DI）**：
    引数として，Webフレームワークの本体である `*echo.Echo` と，ビジネスロジックである UseCase レイヤーのインスタンス（`tu`, `su`）を受け取る．これを `TaskHandler` 構造体内部にセットすることで，ハンドラーがいつでもUseCase（現場監督）を呼び出せる状態にする．これが「依存関係の注入」である．
2.  **ルーティング（交通整理）の登録**：
    `e.POST("/api/tasks/manual", h.CreateManualTask)` は，「クライアントから `/api/tasks/manual` というURLパスに対して `POST`（新規作成）メソッドでアクセスが来たら，このハンドラーの `CreateManualTask` という関数を実行してね」という指示を Echo に登録している．
    `e.PUT("/api/tasks/:id", h.UpdateTask)` も同様に，タスクの更新要求（`PUT`）が来た際，URL内の課題ID（`:id`）を自動で読み取って `UpdateTask` 関数へ処理を流すように登録している．

---

### CreateManualTaskの解説（手動課題の登録）

登録用のURLにアクセスが来た際，`CreateManualTask` 関数の中では以下の3つのステップが実行される．

1.  **リクエストのバインド (`c.Bind(task)`)**：
    フロントエンドから送られてきた HTTP リクエストのボディ（JSONデータ）を，Goの構造体 `domain.Task` に自動的にマッピング（デシリアライズ）する処理である．もし，JSONのデータ型が不整合であったり，壊れたJSONが送信されたりした場合はエラーとなり，即座に `HTTP 400 Bad Request`（リクエスト形式が不正である）をクライアントに返却して処理を中断する．
2.  **UseCase의 実行 (`h.taskUsecase.RegisterManualTask(...)`)**：
    第2回で作成した UseCase レイヤー（現場監督）を呼び出す処理である．第1引数として `c.Request().Context()`（リクエストのコンテキスト）を渡すことで，処理の途中でクライアントがブラウザを閉じたりタイムアウトが発生したりした際に，DBの処理や通知の予約を安全にキャンセル（中断）できるように制御している．
3.  **レスポンスの生成 (`c.JSON(...)`)**：
    *   **エラー発生時**：`HTTP 500 Internal Server Error` とともに，エラーメッセージを格納した JSON （`{"error": "..."}`）を返却する．
    *   **成功時**：`HTTP 201 Created`（作成成功）とともに，保存完了した課題オブジェクト（DBで生成されたIDや更新日時が含まれる）を JSON に変換して返却する．フロントエンドは，このレスポンスデータを受け取ることで画面を即座に最新状態に更新できる．

---

### UpdateTaskの解説（課題の更新とContextの抽出）

課題の更新要求（`PUT`）が届いた際，クライアントから送信される各種パラメータ（引数）はリクエストの別々の場所（URLパス，URLの末尾，HTTPボディ）に散らばっている．Handlerはこれらを個別に抽出して整理し，UseCaseに引き渡す．

1.  **URLパラメータの抽出 (`c.Param("id")`)**：
    `e.PUT("/api/tasks/:id", ...)` の `:id` 部分に入っている実際の課題ID（例: `tsk-go`）をURLパスから文字列として抽出する．
2.  **クエリパラメータの抽出 (`c.QueryParam("user_id")`)**：
    URLの末尾に付与されている `?user_id=usr-alice` といったクエリパラメータから，操作を実行したユーザーのIDを抽出する．
3.  **リクエストボディのバインド (`c.Bind(task)`)**：
    JSON形式で送られてきた「更新後の課題のタイトルや提出期限」のデータを，`domain.Task` 構造体に変換する．
4.  **UseCaseへの受け渡し**：
    抽出した `taskID`（文字列），`task`（構造体ポインタ），および `userID`（文字列）を引数として，UseCase層の `h.taskUsecase.UpdateTask(...)` を実行する．

ここで重要なのは，**「Webフレームワーク（Echo）に依存したオブジェクト（`echo.Context`）を，内側のUseCase層に直接引き渡さない」**ということである．

これは，リクエストの生存期間やキャンセルを制御する `context.Context` （コンテキスト）の扱いについても同様である．
Echoフレームワークを使用している場合，Handler関数には `c echo.Context` が渡されるが，これをそのまま内側のUseCaseに渡すことはしない．代わりに `c.Request().Context()` を使って，**Go標準パッケージの `context.Context`** を抽出し，それをUseCaseへ引き渡す．

#### もしフレームワークを Gin に変更する場合どうなるか？
もし，Webフレームワークを Echo から **Gin** に移行することになったとする．その場合，Handler関数の引数は `c *gin.Context` に変わる．
しかし，Handlerから内側に引き渡す際は，Gin独自のコンテキストから `c.Request.Context()` として**Go標準の `context.Context`** を抽出し，UseCaseに渡す．

この「フレームワーク特有のコンテキスト」から「Go標準のコンテキスト」への変換（抽出）を最外周（Handler）で行うことで，内側のUseCaseやDomain，Infrastructure（データベース）のレイヤーは，フレームワークがEchoからGinに変わったことを一切知る必要がない．それらは標準の `context.Context` のみを使ってタイムアウトやキャンセル処理を記述できる．

Handlerがリクエストから必要なパラメータ（ただの文字列，構造体ポインタ，およびGo標準의 Context）だけを「抽出し，翻訳」してUseCaseへ渡すことで，内側のUseCaseレイヤーが特定のWebフレームワーク（Echo）に汚染されるのを防ぎ，システムの疎結合（クリーンさ）を保っている．

このハンドラー自体には，「タイトルが空ならエラーにする」といったビジネスルールは記述されていない．それはUseCaseの役割だからである．ハンドラーはあくまで，**「HTTPという通信規約を，Goの関数呼び出しに翻訳する」**という役割（プロトコル変換）に特化している．

---

## 2. Infrastructure（インフラストラクチャ）レイヤー：技術的詳細の具象化

同心円の最も外側に位置するのが **Infrastructureレイヤー** である．
ここには，データベース操作（GORM），AIの呼び出し（Gemini SDK），外部サービス連携（LINE通知，Google Classroom同期）といった，具体的なテクノロジーに関するコードが記述される．

### 役割：インターフェースの具象化（DIPの受け皿）
第2回でDomainレイヤーに定義した `domain.TaskRepository` インターフェースを，このレイヤーで実際に実装（実装コードを記述）する．

### 実例：GORMを用いたリポジトリ実装 (infrastructure/db/task_repository.go)

```go
package db

import (
	"context"
	"errors"

	"github.com/yuma25/Uni-Steps/domain"
	"gorm.io/gorm"
)

// taskRepository は domain.TaskRepository インターフェースを実装する構造体である．
type taskRepository struct {
	db *gorm.DB // GORM の接続情報を保持
}

// NewTaskRepository はリポジトリの具象クラスを生成し，インターフェース型として返す．
func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &taskRepository{db: db}
}

// Save はタスクをDBに保存または更新する（具体的な技術実装）
func (r *taskRepository) Save(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Task 本体を保存（関連テーブルを除外してシンプルに保存）
		if err := tx.Omit("UserProgress").Save(task).Error; err != nil {
			return err
		}

		// 進捗状況（UserProgress）を同期
		if err := tx.Where("task_id = ?", task.ID).Delete(&domain.TaskUserProgress{}).Error; err != nil {
			return err
		}

		if len(task.UserProgress) > 0 {
			for i := range task.UserProgress {
				task.UserProgress[i].TaskID = task.ID
			}
			if err := tx.Create(task.UserProgress).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID はIDでタスクを検索する
func (r *taskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	err := r.db.WithContext(ctx).Preload("UserProgress").Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // レコードが見つからなければ nil を返す（Domainの契約に従う）
		}
		return nil, err
	}
	return &task, nil
}
```

### プログラムの具体的な挙動解説

このリポジトリの実装内では，GORMが提供する以下の強力なデータベース操作機能が使用されている．

1.  **コンテキストの伝播 (`WithContext(ctx)`)**：
    すべてのDB操作の前に `.WithContext(ctx)` を挟むことで，Handler層から引き渡されてきたHTTPリクエストのコンテキストをデータベース処理に伝えている．これにより，もしクライアントが途中でブラウザを閉じたりしてリクエストがキャンセルされた場合，実行中のSQLクエリも即座にデータベース側で安全に中断される．
2.  **トランザクション制御 (`Transaction(func(tx *gorm.DB) error { ... })`)**：
    `Save` メソッド内で行われている「親タスクの保存」「古い進捗データの削除」「新しい進捗データの挿入」という一連の処理を，1つの**トランザクション**として実行している．万が一，進捗データの挿入（STEP 3）でエラーが発生した場合，すでに実行された親タスクの保存（STEP 1）も自動的に巻き戻され（ロールバック），データベース内のデータの不整合（親タスクだけが残って進捗データがない状態）が発生するのを防いでいる．
3.  **関連データの保存除外 (`Omit("UserProgress")`)**：
    親タスクを保存する際，GORMの自動アソシエーション保存機能を使わず，あえて `.Omit("UserProgress")` を指定して親テーブル（`tasks`）だけを単体で保存している．これは，進捗データ（子）の同期（一度全削除してから新規追加する処理）を，明示的にコントロールするためである．
4.  **関連データの事前読み込み (`Preload("UserProgress")`)**：
    `FindByID` でタスクを取得する際，`.Preload("UserProgress")` を呼び出している．これは**イーガーローディング（Eager Loading）**と呼ばれる機能であり，親タスクの取得と同時に，外部キー（`TaskID`）で繋がっている子レコード（`task_user_progresses`）をデータベースから自動的に一括取得し，`Task.UserProgress` フィールドに詰め込んで返却する．これにより，N+1問題（レコード数だけ無駄なクエリが走る問題）を回避している．

データベースへの接続やSQLの発行（GORMの操作）はすべてこのレイヤーに隠蔽されている．DomainやUseCaseは，このリポジトリクラスの内部がGORMで書かれていることすら知らない．もし将来，GORMから別のORM（SQLBoilerなど）に変更したり，データベースをMySQLやMongoDBに変えたりしたとしても，ビジネスロジックには影響を及ぼさず，変更箇所はこのレイヤーだけに留めることができる．

---

## 3. 全てを結びつける：main.go における依存関係の注入（DI）

各レイヤーが独立して作られているため，アプリを起動する際にはそれらを正しく組み立てて繋ぎ合わせる必要がある．この「組み立て」を担当するのが，プログラムのエントリーポイントである `main.go` である．

---

### 依存関係の注入（DI）とは何か？

「依存関係の注入（Dependency Injection：DI）」という言葉は難しく聞こえるが，例えるなら**「ゲーム機本体に，別売りのコントローラーやソフトを外からカチッと差し込むこと」**である．

もし，ゲーム機の中にコントローラーが「直接ハンダ付けされて固定（密結合）」されていたら，コントローラーが壊れたり，別のコントローラーに変えたくなったりした時に，ゲーム機ごと買い換える必要が出てしまう．
外から差し込める（DIする）ようになっていれば，コントローラーが壊れても，そこだけを差し替えることができる．

クリーンアーキテクチャでも同様に，各クラスは自力で「データベース接続」などの具体的な道具を作らない．**「必要な道具は，すべて外から引数として差し込んでもらう」**という受け身の姿勢を貫く．その差し込み作業（DI）を一手に引き受けるのが，アプリの組み立て工場である `main.go` である．

---

### main.go での組み立て手順（ステップ・バイ・ステップ）

`main.go` では，以下の順序で下位のパーツ（道具）から順に組み立てを行い，外側のインスタンスを内側のインスタンスに差し込んで（注入して）いく．

```
【 組み立てのフロー図 】

 1. 道具の実体（Infrastructure）を作る
    gormDB ➔ taskRepo
       │
       ▼
 2. 現場監督（UseCase）を作り，道具のインターフェース（差し込み口）に実体を差し込む (DIPによる逆転)
    NewTaskUsecase(taskRepo)
       │
       ▼
 3. 窓口（Handler）の初期化時に，現場監督を差し込む
    NewTaskHandler(taskUsecase)
```

#### コラム：このフロー図の「差し込みの向き」のねじれに気づきましたか？

このフロー図をじっくり見ると，奇妙な「ねじれ（矛盾）」が起きていることに気づく．

*   **STEP 2**: UseCase（内周）の中に，道具である `taskRepo`（外周）を差し込んでいる（**内に外を差し込む**）．
*   **STEP 3**: Handler（外周）の中に，現場監督である `taskUsecase`（内周）を差し込んでいる（**外に内を差し込む**）．

なぜ，ステップ2とステップ3で「差し込み（注入）の向き」が真逆になっているのだろうか？
これこそが，まさに第2回で解説した **「依存関係逆転の原則（DIP）」が物理的に作動している証拠** である．

本来，プログラムを呼び出す側が，呼び出される側に依存するため，DIPを使わない通常の設計であれば，ステップ2も「外（UseCase）に内（DB）を差し込む」という向きになるはずである．
しかし，DIPによって依存方向を「外から内へ」ひっくり返した結果，**「内側（UseCase）が要求するインターフェース（規格）に対して，外側の実体（インフラ）を上からカチッと差し込む」**という逆転現象が起きている．

このねじれこそが，クリーンアーキテクチャがインフラ（技術）の変更からビジネスロジックを守るために仕組んだ「依存の逆転」の正体なのである．

---

#### STEP 1：本物の道具（Infrastructure）の準備
まず，データベースの実態（PostgreSQL）に接続し，`gorm.DB` インスタンスを作成する．そして，それを利用する具体的なDB操作用の道具（`taskRepo`）を作る．
```go
// データベースに接続
gormDB, _ := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

// 本物の道具を作成
taskRepo := db.NewTaskRepository(gormDB)
```

#### STEP 2：現場監督（UseCase）の作成と、道具の差し込み（DI）
ビジネスロジックを実行する現場監督（`taskUsecase`）を作成する．この時，STEP 1で作った `taskRepo` をコンストラクタの引数に渡すことで，**「外から差し込む」**．
```go
// taskRepo という道具を、引数（外）から差し込んで組み立てる
taskUsecase := usecase.NewTaskUsecase(taskRepo, groupRepo, aiService, schService)
```
*(※ UseCaseは `domain.TaskRepository` というインターフェースを要求しており，`taskRepo` はその約束を満たしているため，問題なく差し込める)*

#### コラム：なぜ内側のUseCaseに，外側のインフラ（道具）を差し込めるのか？

クリーンアーキテクチャのレイヤー図では，**UseCaseは内側**にあり，**Infrastructure（データベースなど）は外側**にある．
「なぜ内側であるUseCaseの中に，外側の道具（`taskRepo`）を差し込むことができるのか？ それは依存のルール（外から内へ）に違反していないか？」と疑問に思うかもしれない．

ここが，クリーンアーキテクチャの最も巧妙なポイントである．
実際は，**UseCaseは外側の道具を直接受け取っていない．**

`NewTaskUsecase` が引数として要求しているのは，外側の `taskRepo` ではなく，最も内側の **Domainレイヤーで定義されたインターフェース（`domain.TaskRepository`）** である．
つまり，UseCase（内）は，Domain（中心）のインターフェースという「差し込み口（規格）」だけを用意して待っている．

`main.go` は起動時に，その規格に適合した外側の本物の道具（`taskRepo`）をカチッと差し込む．
これにより，**ソースコード上は「内側（Domain）にのみ依存している」状態を維持したまま，実行時には「外側の本物の道具」と繋がる**という構造（依存性の逆転）が成立している．

#### コラム：組み立て（初期化）の順序が「内 ➔ 外」になる理由

もう一つの疑問として，「組み立てる順番が，なぜ UseCase ➔ Handler の順なのか？ Handlerが外側なのだから，Handlerを先に作るべきではないか？」と思うかもしれない．

プログラムのインスタンスを生成する際は，**「差し込まれるパーツ（引数）」を先に作っておかなければ，呼び出す側の引数に渡せない**という物理的なルールがある．

*   `TaskHandler`（窓口/外）を作るためには，引数として `TaskUsecase`（現場監督/内）が必要である．
*   そのため，先に `TaskUsecase` を作成して実体化させておき，それを `NewTaskHandler` に手渡す必要がある．

結果として，`main.go` における起動時の組み立てフローは，**「依存の先端（道具） ➔ ユースケース（内） ➔ ハンドラー（外）」** という順序になる．

#### STEP 3：窓口（Handler）の作成と、現場監督の差し込み（DI）
最後に，HTTPリクエストを受け取る窓口（`TaskHandler`）を作成する．この時，STEP 2で作った `taskUsecase` を引数に渡すことで，**「外から差し込む」**．
```go
// taskUsecase という現場監督を、引数（外）から差し込んで組み立てる
handler.NewTaskHandler(echoInstance, taskUsecase, syncUsecase)
```

---

### なぜ main.go だけがこの組み立てを行うのか？

もし，`usecase` や `handler` のコードの中で，自力で `gorm.Open` などを書いてデータベース接続を作ってしまった場合，そのレイヤーは「PostgreSQLやGORM」という具体的な技術に直接縛られてしまう．

データベースなどの「本物の技術」を生成・管理する責任は，アプリケーションの最も外側の境界にある `main.go` だけに持たせる．
他のレイヤーは，ただ**「外から渡されたインターフェース（道具）をそのまま使う」**という形にすることで，それぞれのレイヤーが技術に汚染されず，テスト可能でクリーンな状態を維持できるのである．

---

### main.go のコードイメージ

```go
func main() {
    // 1. データベース接続
    gormDB, _ := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

    // 2. インフラ層（道具）の初期化
    taskRepo  := db.NewTaskRepository(gormDB)
    userRepo  := db.NewUserRepository(gormDB)
    groupRepo := db.NewGroupRepository(gormDB)
    aiService := ai.NewGeminiService(genaiClient, "gemini-2.0-flash")
    schService := scheduler.NewInMemScheduler(...)

    // 3. ユースケース層（現場監督）の初期化（依存関係の注入）
    taskUsecase := usecase.NewTaskUsecase(taskRepo, groupRepo, aiService, schService)

    // 4. Echo サーバーの起動
    e := echo.New()

    // 5. ハンドラー（窓口）の初期化（ユースケースの注入）
    handler.NewTaskHandler(e, taskUsecase, syncUsecase)

    // サーバー起動
    e.Start(":8080")
}
```
