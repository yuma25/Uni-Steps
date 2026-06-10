package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/domain"
)

// WakeupHandler は起床確認の予約やチェックインを受け付ける窓口である．
type WakeupHandler struct {
	wakeupRepo domain.WakeupRepository
}

// NewWakeupHandler はハンドラーを初期化し，ルーティングを登録する．
func NewWakeupHandler(e *echo.Echo, wr domain.WakeupRepository) {
	h := &WakeupHandler{
		wakeupRepo: wr,
	}
	e.POST("/api/wakeup/request", h.RequestWakeupCheck)
	e.POST("/api/wakeup/checkin", h.CheckIn)
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

	grace := req.GraceMinutes
	if grace == 0 {
		grace = 5 // デフォルト5分
	}

	check := &domain.WakeupCheck{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		GroupID:      req.GroupID,
		TargetTime:   targetTime,
		GraceMinutes: grace,
		Status:       domain.WakeupStatusPending,
		CreatedAt:    time.Now(),
	}

	if err := h.wakeupRepo.Save(c.Request().Context(), check); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "予約に失敗した"})
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

	// ユーザーの進行中の起床確認を取得する．
	checks, err := h.wakeupRepo.FindActiveByUser(c.Request().Context(), req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "データの取得に失敗した"})
	}

	for _, check := range checks {
		check.Status = domain.WakeupStatusConfirmed
		_ = h.wakeupRepo.Save(c.Request().Context(), check)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "起床を確認した．おはよう！"})
}
