# SwitchBot API Client Example: Custom JSON Handler

This example demonstrates how to use the SwitchBot Go client library with a custom JSON handler. It specifically shows how to replace the default `encoding/json` with `json-iterator/go`.

## Purpose

To illustrate the flexibility of the client library in allowing users to plug in their preferred JSON processing library, potentially for performance reasons.

## Prerequisites

1.  **Go:** Ensure Go is installed (version 1.18 or later recommended).
2.  **SwitchBot Client Package:** The `switchbot` package (located in the parent directory, e.g., `your_module_path/switchbot`) must be available in your Go workspace. Replace `your_module_path/switchbot` in the `import` statement with your actual module path if necessary.
3.  **API Credentials:** You need your SwitchBot API Token and Secret Key. Set them as environment variables:
    ```bash
    export SWITCHBOT_TOKEN="YOUR_TOKEN_HERE"
    export SWITCHBOT_SECRET="YOUR_SECRET_KEY_HERE"
    ```
    Replace `YOUR_TOKEN_HERE` and `YOUR_SECRET_KEY_HERE` with your actual credentials.
4.  **json-iterator:** This example requires the `json-iterator/go` library. Install it using:
    ```bash
    go get github.com/json-iterator/go
    ```

## How to Run

1.  **Set Environment Variables:** Make sure `SWITCHBOT_TOKEN` and `SWITCHBOT_SECRET` are set.
2.  **Install Dependency:** Run `go get github.com/json-iterator/go`.
3.  **Navigate to Directory:** Open your terminal and change the directory to `examples/`.
4.  **Run the Code:** Execute the following command:
    ```bash
    go run json.go
    ```

## Code Overview

-   Imports the `switchbot` client package and `json-iterator/go`.
-   Defines a `JsonIterHandler` struct that implements the `switchbot.JSONHandler` interface using `jsoniter.Marshal` and `jsoniter.Unmarshal`.
-   Retrieves API credentials from environment variables.
-   Initializes the client using `switchbot.NewClient()` and passes the custom handler via the `switchbot.WithJSONHandler()` option:
    ```go
    customJsonHandler := &JsonIterHandler{}
    client, err := switchbot.NewClient(token, secret,
        switchbot.WithJSONHandler(customJsonHandler),
    )
    ```
-   Calls `client.GetDevices()` to fetch the device list. The client internally uses the provided `JsonIterHandler` for request/response processing.
-   Prints the retrieved device list (or an error message) to the console.

## Expected Output

The program will print logs indicating the start and end of the process, and that it's using the `JsonIterHandler`. If successful, it will output:
-   A message indicating successful retrieval.
-   A pretty-printed JSON representation of the device list.
-   The total count of physical and infrared devices found.

The final output should be functionally identical to the "Normal Usage" example, but the internal JSON processing uses `json-iterator`. Error messages will be logged if issues occur.