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
}

// NewTaskHandler はハンドラーを初期化し，ルーティングを登録する．
func NewTaskHandler(e *echo.Echo, tu *usecase.TaskUsecase) {
	h := &TaskHandler{
		taskUsecase: tu,
	}
	e.GET("/api/groups/:id/tasks", h.ListTasks)
	e.POST("/api/tasks/manual", h.CreateManualTask)
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
