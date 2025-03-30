package main

import (
	"context"
	"encoding/json" // For pretty printing only
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mktbsh/switchbot-go"
)

func customJsonMarshal(v interface{}) ([]byte, error) {
	// As a custom JSON Marshal
	return json.Marshal(v)
}
func customJsonUnmarshal(data []byte, v interface{}) error {
	// As a custom JSON Un Marshal
	return json.Unmarshal(data, v)
}

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

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
	log.Println("--- [Custom JSON Handler Example] Start ---")

	token := os.Getenv("SWITCHBOT_TOKEN")
	secret := os.Getenv("SWITCHBOT_SECRET")
	if token == "" || secret == "" {
		log.Fatal("Error: SWITCHBOT_TOKEN and SWITCHBOT_SECRET environment variables must be set.")
	}
	log.Println("Credentials loaded.")

	client, err := switchbot.NewClient(token, secret,
		switchbot.WithJSONEncoder(customJsonMarshal),
		switchbot.WithJSONDecoder(customJsonUnmarshal),
	)
	if err != nil {
		log.Fatalf("Error creating SwitchBot client: %v", err)
	}
	log.Println("SwitchBot client created with JsonIterHandler.")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n===== Getting Devices =====")
	devicesResp, err := client.GetDevices(ctx)
	if err != nil {
		handleAPIError("getting devices", err)
	} else {
		fmt.Printf("Successfully retrieved device list:\n%s\n", prettyPrint(devicesResp))
		fmt.Printf("Total physical devices: %d\n", len(devicesResp.DeviceList))
		fmt.Printf("Total infrared devices: %d\n", len(devicesResp.InfraredRemoteList))
	}

	log.Println("--- [Custom JSON Handler Example] End ---")
}
