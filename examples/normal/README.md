# SwitchBot API Client Example: Normal Usage

This example demonstrates the basic usage of the SwitchBot Go client library with default settings. It uses the standard Go `net/http` client and `encoding/json` for JSON handling.

## Purpose

To show how to initialize the client with minimal configuration and fetch the device list from the SwitchBot API.

## Prerequisites

1.  **Go:** Ensure Go is installed (version 1.18 or later recommended).
2.  **SwitchBot Client Package:** The `switchbot` package (located in the parent directory, e.g., `your_module_path/switchbot`) must be available in your Go workspace. Replace `your_module_path/switchbot` in the `import` statement with your actual module path if necessary.
3.  **API Credentials:** You need your SwitchBot API Token and Secret Key. Set them as environment variables:
    ```bash
    export SWITCHBOT_TOKEN="YOUR_TOKEN_HERE"
    export SWITCHBOT_SECRET="YOUR_SECRET_KEY_HERE"
    ```
    Replace `YOUR_TOKEN_HERE` and `YOUR_SECRET_KEY_HERE` with your actual credentials.

## How to Run

1.  **Set Environment Variables:** Make sure `SWITCHBOT_TOKEN` and `SWITCHBOT_SECRET` are set as shown above.
2.  **Navigate to Directory:** Open your terminal and change the directory to `examples/`.
3.  **Run the Code:** Execute the following command:
    ```bash
    go run normally.go
    ```

## Code Overview

-   Imports the `switchbot` client package.
-   Retrieves API credentials from environment variables.
-   Initializes the client using `switchbot.NewClient(token, secret)`. This uses default options for the HTTP client and JSON handler.
-   Calls `client.GetDevices()` to fetch the list of devices associated with the account.
-   Prints the retrieved device list (or an error message) to the console.

## Expected Output

The program will print logs indicating the start and end of the process. If successful, it will output:
-   A message indicating successful retrieval.
-   A pretty-printed JSON representation of the device list (physical and infrared devices).
-   The total count of physical and infrared devices found.

If an error occurs (e.g., invalid credentials, network issues), an error message detailing the problem will be logged.