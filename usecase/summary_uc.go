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

// SendDailyGroupSummary は全てのグループに対して，今日の課題状況を LINE へ通知する．
func (uc *SummaryUsecase) SendDailyGroupSummary(ctx context.Context) error {
	groups, err := uc.groupRepo.FindAllGroups(ctx)
	if err != nil {
		return fmt.Errorf("全グループの取得に失敗した： %w", err)
	}

	log.Printf("[Summary] %d 件のグループに対してサマリー処理を開始します...\n", len(groups))

	for _, group := range groups {
		if err := uc.ProcessSingleGroupSummary(ctx, group.ID); err != nil {
			log.Printf("[Summary] グループ %s (%s) のサマリー送信に失敗した: %v\n", group.Name, group.ID, err)
		}
	}

	return nil
}

// ProcessSingleGroupSummary は特定のグループに対してサマリーを生成・送信する．
func (uc *SummaryUsecase) ProcessSingleGroupSummary(ctx context.Context, groupID string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil || group == nil {
		return fmt.Errorf("グループが見つからない")
	}

	// LINE 連携がなければスキップ
	if group.LineChannelToken == "" || group.LineGroupID == "" {
		return nil
	}

	// 1．グループに紐づく全ての課題を取得
	tasks, err := uc.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return err
	}

	// 2．「今日が期限」または「未完了」の課題をメンバーごとに集計
	now := time.Now()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)

	summaryItems := []string{}
	for _, user := range group.Users {
		userTaskCount := 0
		for _, task := range tasks {
			for _, up := range task.UserProgress {
				if up.UserID == user.ID && !up.IsCompleted {
					// 今日までの課題かチェック
					if !task.Deadline.IsZero() && task.Deadline.Before(todayEnd) {
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
		return nil // 課題がなければ送らない
	}

	// 3．AI に状況を説明してサマリー文を作らせる
	statusText := strings.Join(summaryItems, "\n")
	msg, err := uc.aiService.GenerateGroupSummaryMessage(ctx, statusText, group.AICharacter)
	if err != nil {
		log.Printf("[Summary] AI サマリー生成失敗 (GroupID: %s): %v\n", groupID, err)
		msg = fmt.Sprintf("【朝の課題サマリー】\nおはようございます！今日の皆さんの状況です：\n%s\n今日も一日頑張りましょう！", statusText)
	}

	// 4．LINE グループへ送信
	_ = uc.notifService.SendGroupMessage(ctx, group.ID, msg)

	// 5．履歴にも残す
	_ = uc.logRepo.Save(ctx, &domain.NotificationLog{
		ID:        uuid.New().String(),
		GroupID:   group.ID,
		UserID:    "system",
		Type:      "summary",
		Message:   msg,
		CreatedAt: time.Now(),
	})

	log.Printf("[Summary] グループ %s への朝刊送信が完了した．\n", group.Name)
	return nil
}
