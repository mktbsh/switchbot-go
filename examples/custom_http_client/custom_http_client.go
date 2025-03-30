package main

import (
	"context"
	"encoding/json" // For pretty printing
	"fmt"
	"log"
	"net/http" // 標準のhttpパッケージ
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

// APIエラーハンドリング関数 (normally.go と同じ)
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
	log.Println("--- [Custom HTTP Client Example] Start ---")

	// 1. 環境変数から認証情報を取得
	token := os.Getenv("SWITCHBOT_TOKEN")
	secret := os.Getenv("SWITCHBOT_SECRET")
	if token == "" || secret == "" {
		log.Fatal("Error: SWITCHBOT_TOKEN and SWITCHBOT_SECRET environment variables must be set.")
	}
	log.Println("Credentials loaded.")

	// 2. カスタムHTTPクライアントの作成 (例: タイムアウト15秒)
	customHttpClient := &http.Client{
		Timeout: 15 * time.Second,
		// 他のTransport設定なども可能
		// Transport: &http.Transport{...},
	}
	log.Println("Custom HTTP client created with 15s timeout.")

	// 3. SwitchBotクライアントの作成 (カスタムHTTPクライアントを指定)
	client, err := switchbot.NewClient(token, secret,
		switchbot.WithHTTPClient(customHttpClient), // カスタムクライアントを注入
	)
	if err != nil {
		log.Fatalf("Error creating SwitchBot client: %v", err)
	}
	log.Println("SwitchBot client created with custom HTTP client.")

	// 4. Contextの作成 (API呼び出し個別ではなく、ここでの操作全体のタイムアウト)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 5. デバイス一覧の取得
	fmt.Println("\n===== Getting Devices =====")
	devicesResp, err := client.GetDevices(ctx)
	if err != nil {
		handleAPIError("getting devices", err)
	} else {
		fmt.Printf("Successfully retrieved device list:\n%s\n", prettyPrint(devicesResp))
		fmt.Printf("Total physical devices: %d\n", len(devicesResp.DeviceList))
		fmt.Printf("Total infrared devices: %d\n", len(devicesResp.InfraredRemoteList))
	}

	log.Println("--- [Custom HTTP Client Example] End ---")
}
