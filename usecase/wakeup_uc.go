package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// WakeupUsecase は起床確認の登録や判定に関するビジネスロジックを担当する．
type WakeupUsecase struct {
	wakeupRepo domain.WakeupRepository
	scheduler  domain.SchedulerService
}

// NewWakeupUsecase は WakeupUsecase の新しいインスタンスを生成する．
func NewWakeupUsecase(wr domain.WakeupRepository, sch domain.SchedulerService) *WakeupUsecase {
	return &WakeupUsecase{
		wakeupRepo: wr,
		scheduler:  sch,
	}
}

// RequestWakeup は新しい起床見守りを予約し，同時にスケジューラーに SOS 発信を登録する．
func (uc *WakeupUsecase) RequestWakeup(ctx context.Context, userID, groupID string, targetTime time.Time, graceMinutes int) (*domain.WakeupCheck, error) {
	// 既存のアクティブな見守りがあればキャンセルして上書きする（変更機能）
	_ = uc.CancelWakeup(ctx, userID)

	check := &domain.WakeupCheck{
		ID:           uuid.New().String(),
		UserID:       userID,
		GroupID:      groupID,
		TargetTime:   targetTime,
		GraceMinutes: graceMinutes,
		Status:       domain.WakeupStatusPending,
		CreatedAt:    time.Now(),
	}

	// 1．データベースに記録を保存する．
	if err := uc.wakeupRepo.Save(ctx, check); err != nil {
		return nil, fmt.Errorf("起床確認の保存に失敗した： %w", err)
	}

	// 2．スケジューラーに SOS 送信を予約する（起床予定時刻 ＋ 猶予期間）．
	runAt := targetTime.Add(time.Duration(graceMinutes) * time.Minute)
	if err := uc.scheduler.ScheduleWakeupSOS(ctx, check.ID, userID, groupID, runAt); err != nil {
		return nil, fmt.Errorf("SOS 通知の予約に失敗した： %w", err)
	}

	return check, nil
}

// CancelWakeup は進行中の見守りをキャンセルし，データベースから削除する．
func (uc *WakeupUsecase) CancelWakeup(ctx context.Context, userID string) error {
	checks, err := uc.wakeupRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, check := range checks {
		// スケジューラーの予約をキャンセル
		_ = uc.scheduler.CancelWakeupSOS(ctx, check.ID)
		// データベースから削除
		_ = uc.wakeupRepo.Delete(ctx, check.ID)
	}

	return nil
}

// ConfirmWakeup は本人の起床を確認し，予約されていた SOS 通知をキャンセルする．
func (uc *WakeupUsecase) ConfirmWakeup(ctx context.Context, userID string) error {
	// 現在進行中の（未確認の）起床確認を取得する．
	checks, err := uc.wakeupRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, check := range checks {
		// 1．ステータスを更新する．
		check.Status = domain.WakeupStatusConfirmed
		if err := uc.wakeupRepo.Save(ctx, check); err != nil {
			continue
		}

		// 2．スケジューラーの予約（SOS 発信）をキャンセルする．
		_ = uc.scheduler.CancelWakeupSOS(ctx, check.ID)
	}

	return nil
}

// GetActiveChecks はユーザーの進行中の起床確認を取得する．
func (uc *WakeupUsecase) GetActiveChecks(ctx context.Context, userID string) ([]*domain.WakeupCheck, error) {
	return uc.wakeupRepo.FindActiveByUser(ctx, userID)
}

// GetActiveGroupChecks は指定されたグループの進行中の起床確認一覧を取得する．
func (uc *WakeupUsecase) GetActiveGroupChecks(ctx context.Context, groupID string) ([]*domain.WakeupCheck, error) {
	return uc.wakeupRepo.FindActiveByGroup(ctx, groupID)
}
