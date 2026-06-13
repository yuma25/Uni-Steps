package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
)

// SummaryUsecase はグループ全体の状況を要約し，通知するロジックを担当する．
type SummaryUsecase struct {
	groupRepo    domain.GroupRepository
	taskRepo     domain.TaskRepository
	aiService    domain.AIService
	notifService domain.NotificationService
	logRepo      domain.NotificationLogRepository
}

// NewSummaryUsecase は SummaryUsecase の新しいインスタンスを生成する．
func NewSummaryUsecase(gr domain.GroupRepository, tr domain.TaskRepository, ai domain.AIService, ns domain.NotificationService, lr domain.NotificationLogRepository) *SummaryUsecase {
	return &SummaryUsecase{
		groupRepo:    gr,
		taskRepo:     tr,
		aiService:    ai,
		notifService: ns,
		logRepo:      lr,
	}
}

// SendAllSummaries は全グループをチェックし，現在時刻が設定時刻と一致する場合に送信を実行する．
// 外部のループ（1分おき）から呼ばれることを想定している．
func (uc *SummaryUsecase) SendAllSummaries(ctx context.Context, now time.Time) error {
	groups, err := uc.groupRepo.FindAllGroups(ctx)
	if err != nil {
		return err
	}

	currentTime := now.Format("15:04") // HH:mm

	for _, group := range groups {
		if group.SummaryMorningTime == currentTime {
			_ = uc.ProcessSingleGroupSummary(ctx, group.ID, domain.SummaryTypeMorning)
		}
		if group.SummaryEveningTime == currentTime {
			_ = uc.ProcessSingleGroupSummary(ctx, group.ID, domain.SummaryTypeEvening)
		}
	}

	return nil
}

// ProcessSingleGroupSummary は特定のグループに対してサマリーを生成・送信する．
func (uc *SummaryUsecase) ProcessSingleGroupSummary(ctx context.Context, groupID string, summaryType string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil || group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// 1．グループに紐づく全ての課題を取得
	tasks, err := uc.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return err
	}

	// 2．「対象」の課題をメンバーごとに集計
	now := time.Now()
	var targetEnd time.Time
	var typeLabel string

	if summaryType == domain.SummaryTypeMorning {
		// 朝：今日が期限のものを集計
		targetEnd = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
		typeLabel = "今日が期限の課題"
	} else {
		// 夜：明日が期限のものを集計
		tomorrow := now.Add(24 * time.Hour)
		targetEnd = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 0, time.Local)
		typeLabel = "明日までの課題"
	}

	summaryItems := []string{}
	for _, user := range group.Users {
		userTaskCount := 0
		for _, task := range tasks {
			for _, up := range task.UserProgress {
				if up.UserID == user.ID && !up.IsCompleted {
					if !task.Deadline.IsZero() && task.Deadline.Before(targetEnd) {
						userTaskCount++
					}
				}
			}
		}
		if userTaskCount > 0 {
			summaryItems = append(summaryItems, fmt.Sprintf("- %s: 残り %d 件", user.Name, userTaskCount))
		}
	}

	if len(summaryItems) == 0 {
		log.Printf("[Summary] グループ %s (%s) は対象課題がないためスキップ\n", group.Name, summaryType)
		return nil
	}

	// 3．AI にサマリー文を作らせる
	statusText := strings.Join(summaryItems, "\n")
	content := fmt.Sprintf("%s の状況:\n%s", typeLabel, statusText)

	msg, err := uc.aiService.GenerateGroupSummaryMessage(ctx, content, group.AICharacter)
	if err != nil {
		log.Printf("[Summary] AI サマリー生成失敗: %v\n", err)
		msg = fmt.Sprintf("【%sサマリー】\n%s\n頑張りましょう！", summaryType, statusText)
	}

	// 4．LINE グループへ送信（設定されている場合）
	if group.LineChannelToken != "" && group.LineGroupID != "" {
		_ = uc.notifService.SendGroupMessage(ctx, group.ID, msg)
	}

	// 5．Web Push でも全員に飛ばす（スマホ通知用）
	for _, member := range group.Users {
		targetURL := fmt.Sprintf("/dashboard?user_id=%s&group_id=%s", member.ID, group.ID)
		_ = uc.notifService.SendDirectMessage(ctx, member.ID, msg, targetURL)
	}

	// 6．履歴にも残す
	_ = uc.logRepo.Save(ctx, &domain.NotificationLog{
		ID:        uuid.New().String(),
		GroupID:   group.ID,
		UserID:    "system",
		Type:      domain.NotificationTypeSummary,
		Message:   msg,
		CreatedAt: time.Now(),
	})

	log.Printf("[Summary] グループ %s への %s サマリー送信が完了した．\n", group.Name, summaryType)
	return nil
}
