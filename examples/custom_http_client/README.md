# SwitchBot API Client Example: Custom HTTP Client

This example demonstrates how to configure the SwitchBot Go client library to use a custom `net/http.Client` instance. This allows for fine-grained control over HTTP behavior, such as setting custom timeouts, transport settings, or proxy configurations.

## Purpose

To show how to provide a pre-configured `http.Client` to the SwitchBot client, enabling customization of the underlying HTTP transport layer.

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
    go run http_client.go
    ```

## Code Overview

-   Imports the `switchbot` client package and the standard `net/http` package.
-   Retrieves API credentials from environment variables.
-   Creates a custom `http.Client` instance (in this example, with a 15-second timeout).
    ```go
    customHttpClient := &http.Client{
        Timeout: 15 * time.Second,
    }
    ```
-   Initializes the SwitchBot client using `switchbot.NewClient()` and passes the custom HTTP client via the `switchbot.WithHTTPClient()` option:
    ```go
    client, err := switchbot.NewClient(token, secret,
        switchbot.WithHTTPClient(customHttpClient),
    )
    ```
-   Calls `client.GetDevices()` to fetch the list of devices. All API requests made by the client will now use the provided `customHttpClient`.
-   Prints the retrieved device list (or an error message) to the console.

## Expected Output

The program will print logs indicating the start and end of the process, and that it's using a custom HTTP client. If successful, it will output:
-   A message indicating successful retrieval.
-   A pretty-printed JSON representation of the device list.
-   The total count of physical and infrared devices found.

If a request takes longer than the custom timeout (15 seconds in this example), a timeout error will be logged. Other potential errors (credentials, network) will also be logged.