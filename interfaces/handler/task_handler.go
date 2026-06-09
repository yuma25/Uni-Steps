package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
)

// TaskHandler は Echo を使った HTTP リクエストの窓口である．
type TaskHandler struct {
	// ここに Usecase（ビジネスロジック）を注入するが，
	// 今回は Echo の説明のために一旦構造体だけ定義する．
}

func NewTaskHandler(e *echo.Echo) {
	h := &TaskHandler{}
	// Echo のルーティング設定
	e.GET("/tasks/:id", h.GetTask)
	e.POST("/tasks", h.CreateTask)
}

// GetTask: 指定したIDの課題を取得する
// --- Echo と net/http の違い解説 ---
//  1. 引数: net/http は (w http.ResponseWriter, r *http.Request) の2つが必要であるが，
//     Echo は `echo.Context` 1つに集約されている．
//  2. パラメータ取得: net/http では URL パラメータの取得が面倒であるが，
//     Echo では `c.Param("id")` で一発で取得できる．
func (h *TaskHandler) GetTask(c echo.Context) error {
	id := c.Param("id") // URLの :id の部分を抽出

	// 本来はここで Usecase を呼び出すが，デモ用にダミーデータを返す．
	dummyTask := domain.Task{
		ID:    id,
		Title: "Echo の勉強をする",
	}

	// 3. レスポンス: net/http では JSON の変換（Marshal）と Header の設定を手動で行うが，
	//    Echo では `c.JSON` メソッドだけで「ステータスコード設定 + JSON変換 + 送信」を完遂できる．
	return c.JSON(http.StatusOK, dummyTask)
}

// CreateTask: 新しい課題を登録する
func (h *TaskHandler) CreateTask(c echo.Context) error {
	task := new(domain.Task)

	// 4. バインド: net/http ではリクエストボディの読み込みと JSON デコードを自前で書くが，
	//    Echo は `c.Bind(task)` を使うだけで，JSON データを構造体に自動で流し込んでくれる．
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// 5. エラーハンドリング: net/http は関数の戻り値がないが，
	//    Echo のハンドラーは `error` を返す．これにより Echo のミドルウェアで一括してエラー処理が可能である．
	return c.JSON(http.StatusCreated, task)
}
