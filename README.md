# SwitchBot Go Client

[![GoDoc](https://godoc.org/github.com/mktbsh/switchbot-go?status.svg)](https://godoc.org/github.com/mktbsh/switchbot-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/mktbsh/switchbot-go)](https://goreportcard.com/report/github.com/mktbsh/switchbot-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
<!-- Add Build Status badge if CI is set up -->

Go client library for the [SwitchBot API v1.1](https://github.com/OpenWonderLabs/SwitchBotAPI).

## Overview

This library provides a convenient way to interact with your SwitchBot devices (Bot, Curtain, Meter, Plug, etc.), scenes, and webhooks programmatically using Go. It handles the v1.1 authentication (token, secret, signing with timestamp and nonce) automatically.

The library is structured into several files for better organization:
- `client.go`: Core client structure, request handling, and options.
- `auth.go`: Authentication header generation logic.
- `devices.go`: Device-related API methods and types.
- `scenes.go`: Scene-related API methods and types.
- `webhook.go`: Webhook-related API methods and types.
- `errors.go`: Custom API error type definition.
- `utils.go`: Utility functions (e.g., nonce generation).

## Features

-   Supports SwitchBot API **v1.1**.
-   Automatic request signing using your Token and Secret Key (`auth.go`).
-   UUIDv7 based nonce generation for improved uniqueness (`utils.go`).
-   **Devices API:** (`devices.go`)
    -   Get device list (physical & virtual infrared).
    -   Get device status.
    -   Send device commands.
-   **Scenes API:** (`scenes.go`)
    -   Get manual scene list.
    -   Execute manual scenes.
-   **Webhook API:** (`webhook.go`)
    -   Setup, query, update, and delete webhook configurations.
-   **Customizable:** (`client.go`)
    -   Provide your own `http.Client` (e.g., for custom timeouts, transport) using `WithHTTPClient`.
    -   Provide your own JSON marshaling (`JSONMarshal`) and unmarshaling (`JSONUnmarshal`) functions using `WithJSONEncoder` and `WithJSONDecoder`.
-   Basic API error handling (`errors.go`, `APIError` type).

## Installation

```bash
go get github.com/mktbsh/switchbot-go
```

## Getting Started

### 1. Obtain API Credentials

You need your **Token** and **Secret Key** from the SwitchBot mobile app:
1.  Open the SwitchBot app.
2.  Go to `Profile` > `Preferences`.
3.  Tap `App Version` 10 times to enable Developer Options.
4.  Tap `Developer Options`.
5.  Tap `Get Token` to see your Token and Secret Key.

### 2. Basic Usage

Store your credentials securely, for example, using environment variables.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mktbsh/switchbot-go" // Use 'import "github.com/mktbsh/switchbot-go"'
)

func main() {
	// 1. Load credentials
	token := os.Getenv("SWITCHBOT_TOKEN")
	secret := os.Getenv("SWITCHBOT_SECRET")

	if token == "" || secret == "" {
		log.Fatal("Error: SWITCHBOT_TOKEN and SWITCHBOT_SECRET must be set.")
	}

	// 2. Create a new client (using default settings)
	client, err := switchbot.NewClient(token, secret)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	// 3. Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 4. Call an API method (e.g., Get Devices)
	fmt.Println("Fetching device list...")
	devicesResp, err := client.GetDevices(ctx)
	if err != nil {
		// Handle potential API errors
		if apiErr, ok := err.(*switchbot.APIError); ok {
			log.Printf("API Error: StatusCode=%d, Message='%s'", apiErr.StatusCode, apiErr.Message)
		} else {
			log.Printf("Generic Error: %v", err)
		}
		return
	}

	// 5. Process the response
	fmt.Printf("Found %d physical devices and %d infrared devices.\n",
		len(devicesResp.DeviceList), len(devicesResp.InfraredRemoteList))

	if len(devicesResp.DeviceList) > 0 {
		if name, ok := devicesResp.DeviceList[0]["deviceName"].(string); ok {
			fmt.Printf("First physical device name: %s\n", name)
		}
	}
}
```
*(See [examples/normal/normal.go](./examples/normal/normal.go) for a runnable version)*

## Authentication

Authentication for API v1.1 is handled automatically within the `auth.go` file. It generates the `Authorization`, `sign`, `t` (timestamp), and `nonce` (UUIDv7) headers for each request.

## API Coverage

The library aims to cover the main functionalities of the SwitchBot API v1.1:

-   **Devices API (`devices.go`):** List, Status, Commands.
-   **Scenes API (`scenes.go`):** List, Execute.
-   **Webhook API (`webhook.go`):** Setup, Query, Update, Delete. (*Receiving webhook events requires a separate server implementation.*)

Response bodies for device lists/statuses (`map[string]interface{}`) and command responses (`map[string]interface{}`) use flexible types due to API variations. Type assertions are typically needed to access specific fields.

## Configuration Options

Customize the client using options passed to `switchbot.NewClient()`.

### Custom `http.Client`

Use `switchbot.WithHTTPClient()` to provide your own `http.Client` instance.

```go
import "net/http"
import "time"
// ...

customHttpClient := &http.Client{
    Timeout: 15 * time.Second,
}

client, err := switchbot.NewClient(token, secret,
    switchbot.WithHTTPClient(customHttpClient),
)
// ...
```
See [examples/custom_http_client/custom_http_client.go](./examples/custom_http_client/custom_http_client.go).

### Custom JSON Encoder/Decoder

Use `switchbot.WithJSONEncoder()` and `switchbot.WithJSONDecoder()` to provide custom marshaling (`switchbot.JSONMarshal`) and unmarshaling (`switchbot.JSONUnmarshal`) functions, allowing integration with libraries like `json-iterator/go`.

```go
import jsoniter "github.com/json-iterator/go"
// ...

// Define functions matching the required types
customEncoder := func(v any) ([]byte, error) {
    return jsoniter.Marshal(v)
}
customDecoder := func(data []byte, v any) error {
    return jsoniter.Unmarshal(data, v)
}

// Provide them during client creation
client, err := switchbot.NewClient(token, secret,
    switchbot.WithJSONEncoder(customEncoder),
    switchbot.WithJSONDecoder(customDecoder),
)
// ...
```
See [examples/json/json.go](./examples/json/json.go).

## Error Handling

API methods return an `error`. Errors from the SwitchBot API or HTTP errors are typically of type `*switchbot.APIError` (defined in `errors.go`). Use a type assertion to access detailed error information.

```go
resp, err := client.GetDevices(ctx)
if err != nil {
    if apiErr, ok := err.(*switchbot.APIError); ok {
        fmt.Printf("API Error Detected!\n")
        fmt.Printf("  Status Code: %d\n", apiErr.StatusCode)
        fmt.Printf("  Message:     %s\n", apiErr.Message)
        if len(apiErr.Body) > 0 {
             fmt.Printf("  Raw Body:    %s\n", string(apiErr.Body))
        }
        // ... handle apiErr.Err if needed ...
    } else {
        fmt.Printf("Non-API Error: %v\n", err)
    }
    // ...
}
```
Refer to `errors.go` and the official SwitchBot API documentation for status code meanings.

## Examples

Runnable examples are located in the `examples/` directory:

-   **[examples/normal/](./examples/normal/)**: Basic usage with defaults. ([README](./examples/normal/README.md))
-   **[examples/json/](./examples/json/)**: Using custom JSON functions. ([README](./examples/json/README.md))
-   **[examples/custom_http_client/](./examples/custom_http_client/)**: Using a custom `http.Client`. ([README](./examples/custom_http_client/README.md))

Remember to set the `SWITCHBOT_TOKEN` and `SWITCHBOT_SECRET` environment variables before running examples.

## Contributing

Contributions are welcome! Please submit issues or pull requests for bug fixes, improvements, or new features following the standard GitHub workflow (fork, branch, commit, push, PR). Ensure code is formatted with `go fmt`.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.