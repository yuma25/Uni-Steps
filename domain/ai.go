package domain

import "context"

// AIService は AI エンジンによる文章生成等に関するインターフェースである．
// 本システムでは，ユーザーへの効果的なリマインド文の作成等に AI を活用する．
type AIService interface {
	// GenerateRemindMessage はタスクの内容に基づいてリマインド文を生成する．
	// task: リマインド対象の課題情報
	// style: 生成する文章のトーン（例：「熱血」「厳しい」「優しく」等）
	GenerateRemindMessage(ctx context.Context, task *Task, style string) (string, error)
}

// 「AIエンジンに対して, どんな命令を出してどんな結果を得るか」というインターフェース
//  具体的な Geminiへの接続方法はここには書かず, あくまで「こういう機能が欲しい」という設計図
