package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/usecase"
)

// GroupHandler はグループ管理に関する HTTP リクエストを受け付ける窓口である．
type GroupHandler struct {
	groupUsecase *usecase.GroupUsecase // グループ管理のビジネスロジックを保持する．
	lmsService   domain.LMSService     // 外部 LMS 連携用サービスである．
}

// NewGroupHandler はハンドラーを初期化し，ルーティングを登録する．
func NewGroupHandler(e *echo.Echo, gu *usecase.GroupUsecase, ls domain.LMSService) {
	h := &GroupHandler{
		groupUsecase: gu,
		lmsService:   ls,
	}
	e.POST("/api/groups", h.CreateGroup)
	e.GET("/api/users/:userId/groups", h.ListUserGroups)
	e.POST("/api/users/:userId/groups/sync", h.SyncLMSGroups)
}

// CreateGroup は新しいグループの作成を受け付ける．
func (h *GroupHandler) CreateGroup(c echo.Context) error {
	var req struct {
		Name    string `json:"name"`
		OwnerID string `json:"owner_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	group, err := h.groupUsecase.CreateGroup(c.Request().Context(), req.Name, req.OwnerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, group)
}

// ListUserGroups はユーザーが所属するグループ一覧を返す．
func (h *GroupHandler) ListUserGroups(c echo.Context) error {
	userId := c.Param("userId")
	groups, err := h.groupUsecase.ListUserGroups(c.Request().Context(), userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}

// SyncLMSGroups は外部 LMS からコース情報を同期するリクエストを受け付ける．
func (h *GroupHandler) SyncLMSGroups(c echo.Context) error {
	userId := c.Param("userId")
	groups, err := h.groupUsecase.SyncLMSGroups(c.Request().Context(), userId, h.lmsService)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}
