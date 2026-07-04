package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

/**
 * 実際に Gemini API へ接続し，利用可能なモデルの一覧を取得，
 * および設定されたモデルでの疎通テストを行う診断プログラムである．
 */
func main() {
	// .env ファイルを読み込む
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("\n[!] エラー: GEMINI_API_KEY が .env に設定されていません．")
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("[!] クライアントの初期化に失敗しました: %v", err)
	}
	defer client.Close()

	// --- 1. モデル一覧の表示 ---
	fmt.Println("\n===========================================")
	fmt.Println("   1. Gemini API 利用可能モデル一覧   ")
	fmt.Println("===========================================")

	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("[!] モデル一覧の取得に失敗しました: %v", err)
		}
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				fmt.Printf(" ・ %-25s (%s)\n", m.Name, m.DisplayName)
				break
			}
		}
	}

	// --- 2. 疎通テスト ---
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "models/gemini-1.5-flash" // デフォルト
	}
	// 前後の空白を削除
	modelName = strings.TrimSpace(modelName)

	fmt.Println("\n===========================================")
	fmt.Printf("   2. 疎通テスト開始 (モデル: %s)\n", modelName)
	fmt.Println("===========================================")

	model := client.GenerativeModel(modelName)
	resp, err := model.GenerateContent(ctx, genai.Text("こんにちは！Uni-Steps のテストです．短い挨拶を返してください．"))

	if err != nil {
		fmt.Printf(" [✘] テスト失敗: %v\n", err)
		if strings.Contains(err.Error(), "404") {
			fmt.Println("     -> モデル名が間違っている可能性があります．")
		} else if strings.Contains(err.Error(), "429") {
			fmt.Println("     -> 利用制限（クォータ）に達しています．少し待ってください．")
		}
	} else if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		fmt.Println(" [✓] テスト成功！AI からの応答:")
		fmt.Printf("     「%v」\n", resp.Candidates[0].Content.Parts[0])
	} else {
		fmt.Println(" [?] テスト結果: 応答が空でした．")
	}

	fmt.Println("===========================================")
}
