package domain

import "context"

// AIService は AI エンジンによる文章生成等に関するインターフェースである．
// 本システムでは，ユーザーへの効果的なリマインド文の作成等に AI を活用する．
type AIService interface {
	// GenerateRemindMessage はタスクの内容に基づいてリマインド文を生成する．
	GenerateRemindMessage(ctx context.Context, task *Task, style string) (string, error)

	// GenerateGroupSummaryMessage はグループ全体の状況を要約したメッセージを生成する．
	GenerateGroupSummaryMessage(ctx context.Context, workloadSummary string, style string) (string, error)
}
