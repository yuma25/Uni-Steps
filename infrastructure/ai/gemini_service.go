package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/yuma25/Uni-Steps/domain"
)

// GeminiService は Google Gemini API を使用して AI サービスを提供する構造体である．
type GeminiService struct {
	client *genai.Client          // Gemini API と通信するためのクライアントである．
	model  *genai.GenerativeModel // 使用する生成モデル（gemini-2.0-flash 等）である．
}

// NewGeminiService は GeminiService の新しいインスタンスを生成する．
func NewGeminiService(client *genai.Client, modelName string) *GeminiService {
	model := client.GenerativeModel(modelName)
	return &GeminiService{
		client: client,
		model:  model,
	}
}

// GenerateRemindMessage はタスクの内容に基づいて AI にリマインド文を作らせる．
func (s *GeminiService) GenerateRemindMessage(ctx context.Context, task *domain.Task, style string) (string, error) {
	var characterPrompt string
	switch style {
	case domain.AICharacterStrict:
		characterPrompt = "あなたは非常に厳しい軍隊の指導官です．遅延は一切許されません．"
	case domain.AICharacterKind:
		characterPrompt = "あなたはユーザーをいつも心配している，優しくてお節介な幼馴染です．"
	case domain.AICharacterCool:
		characterPrompt = "あなたは冷徹で仕事が完璧な執事です．感情を表に出さず，論理的に追い詰めてください．"
	default:
		characterPrompt = "あなたは親切なアシスタントです．"
	}

	prompt := fmt.Sprintf(`%s
以下の課題について，ユーザーのやる気を引き出すようなリマインドメッセージを作成せよ．
課題名: %s
期限: %s
条件:
- 100文字以内で，短く心に刺さる言葉にすること．
- 相手の心拍数が上がるような，切迫感のある表現を含めること．
- キャラクターの設定を徹底すること．`, characterPrompt, task.Title, task.Deadline.Format("1月2日 15時04分"))

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Gemini へのリクエストに失敗した： %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("AI からの応答が空である")
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil
}

// GenerateGroupSummaryMessage はグループ全体の状況を要約したメッセージを生成する．
func (s *GeminiService) GenerateGroupSummaryMessage(ctx context.Context, workloadSummary string, style string) (string, error) {
	var characterPrompt string
	switch style {
	case domain.AICharacterStrict:
		characterPrompt = "あなたは非常に厳しい軍隊の指導官です．部下たちの不甲斐なさを叱責しつつ，奮起させてください．"
	case domain.AICharacterKind:
		characterPrompt = "あなたは仲間思いでお節介な幼馴染です．みんなの体調を気遣いつつ，優しく背中を押してください．"
	case domain.AICharacterCool:
		characterPrompt = "あなたは冷徹で論理的な執事です．淡々と状況を報告し，最善の行動を促してください．"
	default:
		characterPrompt = "あなたは親切なチームアシスタントです．"
	}

	prompt := fmt.Sprintf(`%s
以下のグループ全体の今日の課題状況を見て，朝の挨拶とまとめのメッセージを作成してください．
状況:
%s
条件:
- 150文字以内で，簡潔かつ印象的にまとめること．
- メンバー全員の名前を呼びかける必要はなく，チーム全体へのメッセージにすること．
- キャラクター設定を徹底すること．`, characterPrompt, workloadSummary)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Gemini へのサマリー生成依頼に失敗した： %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", fmt.Errorf("AI からの応答が空である")
	}

	return fmt.Sprint(resp.Candidates[0].Content.Parts[0]), nil
}
