package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/usecase"
)

// WakeupHandler は起床確認の予約やチェックインを受け付ける窓口である．
type WakeupHandler struct {
	wakeupUsecase *usecase.WakeupUsecase
}

// NewWakeupHandler はハンドラーを初期化し，ルーティングを登録する．
func NewWakeupHandler(e *echo.Echo, uc *usecase.WakeupUsecase) {
	h := &WakeupHandler{
		wakeupUsecase: uc,
	}
	e.POST("/api/wakeup/request", h.RequestWakeupCheck)
	e.POST("/api/wakeup/checkin", h.CheckIn)
	e.GET("/api/wakeup/active", h.GetActiveChecks)
}

// ... existing RequestWakeupCheck and CheckIn ...

// GetActiveChecks は進行中の起床確認を取得する．
func (h *WakeupHandler) GetActiveChecks(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id が必要である"})
	}

	checks, err := h.wakeupUsecase.GetActiveChecks(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, checks)
}

// RequestWakeupCheck は新しい起床確認を予約する．
func (h *WakeupHandler) RequestWakeupCheck(c echo.Context) error {
	var req struct {
		UserID       string `json:"user_id"`
		GroupID      string `json:"group_id"`
		TargetTime   string `json:"target_time"` // ISO8601 形式
		GraceMinutes int    `json:"grace_minutes"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	targetTime, err := time.Parse(time.RFC3339, req.TargetTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "時刻の形式が不正である"})
	}

	check, err := h.wakeupUsecase.RequestWakeup(c.Request().Context(), req.UserID, req.GroupID, targetTime, req.GraceMinutes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, check)
}

// CheckIn は起床を報告し，進行中の監視を完了させる．
func (h *WakeupHandler) CheckIn(c echo.Context) error {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "リクエスト形式が不正である"})
	}

	err := h.wakeupUsecase.ConfirmWakeup(c.Request().Context(), req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "起床を確認した．おはよう！"})
}
