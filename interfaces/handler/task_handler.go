package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/usecase"
)

// TaskHandler は Echo を使った HTTP リクエストの窓口である．
type TaskHandler struct {
	taskUsecase *usecase.TaskUsecase // 課題管理のビジネスロジックを保持する．
	syncUsecase *usecase.SyncUsecase // 外部 LMS との同期ロジックを保持する．
}

// NewTaskHandler はハンドラーを初期化し，ルーティングを登録する．
func NewTaskHandler(e *echo.Echo, tu *usecase.TaskUsecase, su *usecase.SyncUsecase) {
	h := &TaskHandler{
		taskUsecase: tu,
		syncUsecase: su,
	}
	e.GET("/api/groups/:id/tasks", h.ListTasks)
	e.POST("/api/tasks/manual", h.CreateManualTask)
	e.POST("/api/tasks/sync", h.SyncTasks)
}

// ListTasks はグループ内の課題一覧を返す．
func (h *TaskHandler) ListTasks(c echo.Context) error {
	groupID := c.Param("id")
	tasks, err := h.taskUsecase.ListGroupTasks(c.Request().Context(), groupID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, tasks)
}

// CreateManualTask は UI からの直接入力を受け付ける．
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

// SyncTasks は外部 LMS（Google Classroom等）から課題を同期するリクエストを受け付ける．
func (h *TaskHandler) SyncTasks(c echo.Context) error {
	// リクエストボディの構造体定義
	var req struct {
		GroupID string `json:"group_id"` // 同期対象のグループ（コース）IDである．
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	if req.GroupID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "group_id は必須である"})
	}

	syncedTasks, err := h.syncUsecase.SyncTasks(c.Request().Context(), req.GroupID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 同期されたタスクのリストを返す．
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "同期が完了した",
		"tasks":   syncedTasks,
	})
}
