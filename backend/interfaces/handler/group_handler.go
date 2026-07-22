package handler

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/usecase"
)

// GroupHandler はグループ管理に関する HTTP リクエストを受け付ける窓口である．
type GroupHandler struct {
	groupUsecase *usecase.GroupUsecase // グループ管理のビジネスロジックを保持する．
}

// NewGroupHandler はハンドラーを初期化し，ルーティングを登録する．
func NewGroupHandler(e *echo.Echo, gu *usecase.GroupUsecase) {
	h := &GroupHandler{
		groupUsecase: gu,
	}
	e.POST("/api/groups", h.CreateGroup)
	e.POST("/api/groups/join", h.JoinGroupByInviteCode)
	e.GET("/api/users/:userId/groups", h.ListUserGroups)
	e.PATCH("/api/groups/:groupId/settings", h.UpdateGroupSettings)
	e.PUT("/api/groups/:groupId/owner", h.TransferOwnership)
	e.DELETE("/api/groups/:groupId", h.DeleteGroup)
	e.DELETE("/api/groups/:groupId/users/:userId", h.LeaveGroup)
	e.GET("/api/groups/:groupId/notifications", h.GetNotificationLogs)
}

// TransferOwnership はオーナー権限の譲渡リクエストを処理する．
func (h *GroupHandler) TransferOwnership(c echo.Context) error {
	groupID := c.Param("groupId")
	var req struct {
		CurrentOwnerID string `json:"current_owner_id"`
		NewOwnerID     string `json:"new_owner_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	err := h.groupUsecase.TransferOwnership(c.Request().Context(), groupID, req.CurrentOwnerID, req.NewOwnerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "オーナー権限を譲渡した"})
}

// DeleteGroup は部屋の削除リクエストを処理する．
func (h *GroupHandler) DeleteGroup(c echo.Context) error {
	groupID := c.Param("groupId")
	userID := c.QueryParam("user_id") // オーナー確認用

	err := h.groupUsecase.DeleteGroup(c.Request().Context(), groupID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "部屋を完全に削除した"})
}

// LeaveGroup はユーザーの部屋退出リクエストを処理する．
func (h *GroupHandler) LeaveGroup(c echo.Context) error {
	groupID := c.Param("groupId")
	userID := c.Param("userId")

	err := h.groupUsecase.LeaveGroup(c.Request().Context(), groupID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "部屋を退出した"})
}

// GetNotificationLogs はグループの通知履歴を取得する．
func (h *GroupHandler) GetNotificationLogs(c echo.Context) error {
	groupID := c.Param("groupId")
	limit := 50 // デフォルトで直近50件を取得

	logs, err := h.groupUsecase.GetNotificationLogs(c.Request().Context(), groupID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, logs)
}

// UpdateGroupSettings は部屋の設定（リマインド間隔など）を更新する．
func (h *GroupHandler) UpdateGroupSettings(c echo.Context) error {
	groupID := c.Param("groupId")
	var req struct {
		Name               string `json:"name"`
		RemindIntervals    []int  `json:"remind_intervals"`
		AICharacter        string `json:"ai_character"`
		UserID             string `json:"user_id"`
		LineChannelToken   string `json:"line_channel_token"`
		LineGroupID        string `json:"line_group_id"`
		SummaryMorningTime string `json:"summary_morning_time"`
		SummaryEveningTime string `json:"summary_evening_time"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	log.Printf("[DEBUG] UpdateGroupSettings: GroupID=%s, Name=%s, Intervals=%v\n", groupID, req.Name, req.RemindIntervals)

	err := h.groupUsecase.UpdateSettings(c.Request().Context(), groupID, req.UserID, req.Name, req.RemindIntervals, req.AICharacter, req.LineChannelToken, req.LineGroupID, req.SummaryMorningTime, req.SummaryEveningTime)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "設定を更新した"})
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

// ListUserGroups はユーザーが所属するグループ一覧を返す．
func (h *GroupHandler) ListUserGroups(c echo.Context) error {
	userId := c.Param("userId")
	groups, err := h.groupUsecase.ListUserGroups(c.Request().Context(), userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, groups)
}
