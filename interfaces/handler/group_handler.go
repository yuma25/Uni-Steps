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
	e.POST("/api/groups/join", h.JoinGroupByInviteCode)
	e.GET("/api/users/:userId/groups", h.ListUserGroups)
	e.POST("/api/users/:userId/groups/sync", h.SyncLMSGroups)
	e.GET("/api/users/:userId/lms/courses", h.FetchAvailableLMSCourses)
	e.PATCH("/api/groups/:groupId/link-lms", h.LinkLMSCourse)
}

// JoinGroupByInviteCode は招待コードを用いて部屋に参加する．
func (h *GroupHandler) JoinGroupByInviteCode(c echo.Context) error {
	var req struct {
		InviteCode string `json:"invite_code"`
		UserID     string `json:"user_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	group, err := h.groupUsecase.JoinGroupByInviteCode(c.Request().Context(), req.InviteCode, req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, group)
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

// FetchAvailableLMSCourses は外部 LMS から利用可能なコース一覧を取得する．
func (h *GroupHandler) FetchAvailableLMSCourses(c echo.Context) error {
	userId := c.Param("userId")
	groups, err := h.groupUsecase.FetchAvailableLMSCourses(c.Request().Context(), userId, h.lmsService)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}

// LinkLMSCourse は特定の部屋に対して外部 LMS のコースを紐付ける．
func (h *GroupHandler) LinkLMSCourse(c echo.Context) error {
	groupId := c.Param("groupId")
	var req struct {
		LMSCourseID string `json:"lms_course_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	err := h.groupUsecase.LinkLMSCourse(c.Request().Context(), groupId, req.LMSCourseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "LMS コースの紐付けに成功した"})
}
