package domain

import "context"

// AIService は自然言語解析を行う AI エンジンに関するインターフェースである．
// クリーンアーキテクチャでは，具体的な AI（Gemini等）の実装は infrastructure レイヤーで行い，
// ドメインレイヤーではその「能力」だけを定義する．
type AIService interface {
	// AnalyzeTask はユーザーの入力テキストを解析し，タスク構造体に変換する．
	AnalyzeTask(ctx context.Context, text string) (*Task, error)
	// GenerateRemindMessage はタスクの内容に基づいてリマインド文を生成する．
	GenerateRemindMessage(ctx context.Context, task *Task, style string) (string, error)
}
