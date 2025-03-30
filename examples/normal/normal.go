// examples/normally.go
package main

import (
	"context"
	"encoding/json" // For pretty printing
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mktbsh/switchbot-go"
)

// ヘルパー関数: 整形表示
func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

// APIエラーハンドリング関数
func handleAPIError(action string, err error) {
	if apiErr, ok := err.(*switchbot.APIError); ok {
		log.Printf("API Error %s: StatusCode=%d, Message='%s', Body=%s",
			action, apiErr.StatusCode, apiErr.Message, string(apiErr.Body))
		if apiErr.Err != nil {
			log.Printf("  Underlying cause: %v", apiErr.Err)
		}
	} else {
		log.Printf("Error %s: %v", action, err)
	}
}

func main() {
	log.Println("--- [Normally Example] Start ---")

	// 1. 環境変数から認証情報を取得
	token := os.Getenv("SWITCHBOT_TOKEN")
	secret := os.Getenv("SWITCHBOT_SECRET")
	if token == "" || secret == "" {
		log.Fatal("Error: SWITCHBOT_TOKEN and SWITCHBOT_SECRET environment variables must be set.")
	}
	log.Println("Credentials loaded.")

	// 2. SwitchBotクライアントの作成 (デフォルト設定)
	client, err := switchbot.NewClient(token, secret)
	if err != nil {
		log.Fatalf("Error creating SwitchBot client: %v", err)
	}
	log.Println("SwitchBot client created with default settings.")

	// 3. Contextの作成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 4. デバイス一覧の取得
	fmt.Println("\n===== Getting Devices =====")
	devicesResp, err := client.GetDevices(ctx)
	if err != nil {
		handleAPIError("getting devices", err)
	} else {
		fmt.Printf("Successfully retrieved device list:\n%s\n", prettyPrint(devicesResp))
		fmt.Printf("Total physical devices: %d\n", len(devicesResp.DeviceList))
		fmt.Printf("Total infrared devices: %d\n", len(devicesResp.InfraredRemoteList))
	}

	log.Println("--- [Normally Example] End ---")
}
