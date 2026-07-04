package main

import (
	"fmt"
	"log"

	"github.com/SherClockHolmes/webpush-go"
)

// このスクリプトは Web Push に必要な VAPID キーペア（公開鍵・秘密鍵）を一度だけ生成するためのものである．
func main() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("VAPID キーの生成に失敗した： %v", err)
	}

	fmt.Println("=== VAPID キーペアの生成に成功した ===")
	fmt.Println("以下の値を .env ファイルに追記すること：")
	fmt.Println()
	fmt.Printf("VAPID_PUBLIC_KEY=\"%s\"\n", publicKey)
	fmt.Printf("VAPID_PRIVATE_KEY=\"%s\"\n", privateKey)
	fmt.Println("======================================")
}
