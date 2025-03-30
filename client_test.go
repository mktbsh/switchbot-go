package switchbot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// --- Test NewClient ---

func TestNewClient(t *testing.T) {
	token := "valid-token"
	secret := "valid-secret"

	t.Run("ValidCredentials", func(t *testing.T) {
		client, err := NewClient(token, secret)
		if err != nil {
			t.Fatalf("NewClient() with valid credentials returned error: %v", err)
		}
		if client == nil {
			t.Fatal("NewClient() with valid credentials returned nil client")
		}
		if client.token != token {
			t.Errorf("client.token = %q; want %q", client.token, token)
		}
		if client.secret != secret {
			t.Errorf("client.secret = %q; want %q", client.secret, secret)
		}
		// Check defaults
		if client.httpClient != http.DefaultClient {
			t.Errorf("client.httpClient is not http.DefaultClient")
		}
		if client.baseURL.String() != DefaultBaseURL {
			t.Errorf("client.baseURL = %q; want %q", client.baseURL.String(), DefaultBaseURL)
		}
		// Check default JSON handlers (using reflection as they are functions)
		defaultEncoderPtr := reflect.ValueOf(json.Marshal).Pointer()
		defaultDecoderPtr := reflect.ValueOf(json.Unmarshal).Pointer()
		if reflect.ValueOf(client.jsonEncoder).Pointer() != defaultEncoderPtr {
			t.Errorf("client.jsonEncoder is not json.Marshal by default")
		}
		if reflect.ValueOf(client.jsonDecoder).Pointer() != defaultDecoderPtr {
			t.Errorf("client.jsonDecoder is not json.Unmarshal by default")
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		_, err := NewClient("", secret)
		if err == nil {
			t.Error("NewClient() with empty token did not return an error")
		} else if !strings.Contains(err.Error(), "token and secret must not be empty") {
			t.Errorf("NewClient() with empty token returned unexpected error: %v", err)
		}
	})

	t.Run("EmptySecret", func(t *testing.T) {
		_, err := NewClient(token, "")
		if err == nil {
			t.Error("NewClient() with empty secret did not return an error")
		} else if !strings.Contains(err.Error(), "token and secret must not be empty") {
			t.Errorf("NewClient() with empty secret returned unexpected error: %v", err)
		}
	})

	t.Run("WithOptions", func(t *testing.T) {
		customHTTPClient := &http.Client{Timeout: 5 * time.Second}
		customBaseURL := "http://localhost:8080"
		customEncoder := func(v any) ([]byte, error) { return []byte("encoded"), nil }
		customDecoder := func(data []byte, v any) error { return nil }

		client, err := NewClient(token, secret,
			WithHTTPClient(customHTTPClient),
			WithBaseURL(customBaseURL),
			WithJSONEncoder(customEncoder),
			WithJSONDecoder(customDecoder),
		)
		if err != nil {
			t.Fatalf("NewClient() with options returned error: %v", err)
		}
		if client.httpClient != customHTTPClient {
			t.Errorf("WithHTTPClient option was not applied")
		}
		if client.baseURL.String() != customBaseURL {
			t.Errorf("WithBaseURL option was not applied")
		}
		// Compare function pointers using reflection
		if reflect.ValueOf(client.jsonEncoder).Pointer() != reflect.ValueOf(customEncoder).Pointer() {
			t.Errorf("WithJSONEncoder option was not applied")
		}
		if reflect.ValueOf(client.jsonDecoder).Pointer() != reflect.ValueOf(customDecoder).Pointer() {
			t.Errorf("WithJSONDecoder option was not applied")
		}
	})

	t.Run("WithInvalidBaseURL", func(t *testing.T) {
		_, err := NewClient(token, secret, WithBaseURL("://invalid url"))
		if err == nil {
			t.Error("NewClient() with invalid base URL did not return an error")
		} else if !strings.Contains(err.Error(), "invalid base URL") {
			t.Errorf("NewClient() with invalid base URL returned unexpected error: %v", err)
		}
	})

	t.Run("WithNilHTTPClient", func(t *testing.T) {
		// Should default to http.DefaultClient
		client, err := NewClient(token, secret, WithHTTPClient(nil))
		if err != nil {
			t.Fatalf("NewClient() with nil HTTP client option returned error: %v", err)
		}
		if client.httpClient != http.DefaultClient {
			t.Errorf("WithHTTPClient(nil) did not default to http.DefaultClient")
		}
	})

	t.Run("WithNilEncoder", func(t *testing.T) {
		_, err := NewClient(token, secret, WithJSONEncoder(nil))
		if err == nil {
			t.Error("NewClient() with nil encoder did not return an error")
		} else if !strings.Contains(err.Error(), "JSONEncoder cannot be nil") {
			t.Errorf("NewClient() with nil encoder returned unexpected error: %v", err)
		}
	})

	t.Run("WithNilDecoder", func(t *testing.T) {
		_, err := NewClient(token, secret, WithJSONDecoder(nil))
		if err == nil {
			t.Error("NewClient() with nil decoder did not return an error")
		} else if !strings.Contains(err.Error(), "JSONDecoder cannot be nil") {
			t.Errorf("NewClient() with nil decoder returned unexpected error: %v", err)
		}
	})
}

// --- Test doRequest (via public methods like GetDevices) ---

// setupMockServer creates a httptest server and a client pointing to it.
// handlerFunc allows customizing the server's response for different tests.
func setupMockServer(t *testing.T, handlerFunc http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper() // Marks this as a test helper function

	server := httptest.NewServer(handlerFunc)
	// t.Cleanup ensures the server is closed even if the test fails
	t.Cleanup(server.Close)

	token := "mock-token"
	secret := "mock-secret"

	client, err := NewClient(token, secret, WithBaseURL(server.URL)) // Point client to mock server
	if err != nil {
		t.Fatalf("Failed to create client for mock server: %v", err)
	}

	return client, server
}

func TestDoRequest_Success(t *testing.T) {
	// Mocked response body content for GetDevices
	mockDevicesBody := GetDevicesResponse{
		DeviceList: []Device{
			{"deviceId": "D1", "deviceName": "Bot 1", "deviceType": "Bot"},
		},
		InfraredRemoteList: []InfraredRemoteDevice{
			{DeviceID: "IR1", DeviceName: "TV", RemoteType: "TV"},
		},
	}
	// Marshal the expected body structure for the mock server
	expectedBodyBytes, err := json.Marshal(mockDevicesBody)
	if err != nil {
		t.Fatalf("Failed to marshal mock devices body: %v", err)
	}
	// Construct the full API response string for the mock server
	mockResponse := fmt.Sprintf(`{"statusCode": 100, "message": "success", "body": %s}`, string(expectedBodyBytes))

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Optional: Verify method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1.1/devices") {
			t.Errorf("Expected path ending with /v1.1/devices, got %s", r.URL.Path)
		}

		// Optional: Verify headers were set
		if r.Header.Get("Authorization") == "" {
			t.Error("Authorization header missing in request")
		}
		// ... (other header checks)

		w.WriteHeader(http.StatusOK)
		// Write the full JSON response including statusCode, message, and body
		fmt.Fprintln(w, mockResponse)
	}

	client, _ := setupMockServer(t, handler)

	// Use GetDevices which calls doRequest and then unmarshals the body
	getDevicesResp, err := client.GetDevices(context.Background())
	if err != nil {
		t.Fatalf("GetDevices() returned error: %v", err)
	}

	if getDevicesResp == nil {
		t.Fatal("GetDevices() returned nil response")
	}

	// --- Compare the unmarshaled GetDevicesResponse structure ---
	// Use reflect.DeepEqual for comparing complex structs/slices/maps
	if !reflect.DeepEqual(getDevicesResp, &mockDevicesBody) {
		// Use pretty printing for better diff in error messages
		expectedJSON, _ := json.MarshalIndent(mockDevicesBody, "", "  ")
		actualJSON, _ := json.MarshalIndent(getDevicesResp, "", "  ")
		t.Errorf("GetDevices() response mismatch:\nGot:\n%s\n\nWant:\n%s", string(actualJSON), string(expectedJSON))
	}

	// Example of checking specific fields if needed
	if len(getDevicesResp.DeviceList) != 1 {
		t.Errorf("Expected 1 physical device, got %d", len(getDevicesResp.DeviceList))
	}
	if len(getDevicesResp.InfraredRemoteList) != 1 {
		t.Errorf("Expected 1 infrared device, got %d", len(getDevicesResp.InfraredRemoteList))
	}
	if name, _ := getDevicesResp.DeviceList[0]["deviceName"].(string); name != "Bot 1" {
		t.Errorf("First device name = %q; want %q", name, "Bot 1")
	}
}

func TestDoRequest_APIError(t *testing.T) {
	errorCode := 161 // Example: Device offline
	errorMessage := "device offline"
	mockResponse := fmt.Sprintf(`{"statusCode": %d, "message": "%s", "body": {}}`, errorCode, errorMessage)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // API error can happen even with HTTP 200 OK
		fmt.Fprintln(w, mockResponse)
	}

	client, _ := setupMockServer(t, handler)

	_, err := client.GetDevices(context.Background()) // Call any method
	if err == nil {
		t.Fatal("Expected an APIError, but got nil error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected error of type *APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != errorCode {
		t.Errorf("APIError StatusCode = %d; want %d", apiErr.StatusCode, errorCode)
	}
	if apiErr.Message != errorMessage {
		t.Errorf("APIError Message = %q; want %q", apiErr.Message, errorMessage)
	}
}

func TestDoRequest_HTTPError(t *testing.T) {
	httpStatusCode := http.StatusUnauthorized // 401
	errorMessage := "Unauthorized"            // Typical message for 401
	// SwitchBot might return a body even on HTTP error
	errorBody := `{"some":"details"}`
	mockResponse := fmt.Sprintf(`{"statusCode": 100, "message": "%s", "body": %s}`, errorMessage, errorBody) // API might try to report success internally

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(httpStatusCode)
		fmt.Fprintln(w, mockResponse)
	}

	client, _ := setupMockServer(t, handler)

	_, err := client.GetDevices(context.Background())
	if err == nil {
		t.Fatal("Expected an APIError for HTTP error, but got nil error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected error of type *APIError, got %T: %v", err, err)
	}

	// For HTTP errors, we prioritize the HTTP status code
	if apiErr.StatusCode != httpStatusCode {
		t.Errorf("APIError StatusCode = %d; want %d (HTTP Status)", apiErr.StatusCode, httpStatusCode)
	}
	// Message might come from the parsed body or be generated
	if apiErr.Message != errorMessage {
		// It might have generated a generic message if parsing failed or body was different
		t.Logf("APIError Message = %q; expected something like %q (might differ)", apiErr.Message, errorMessage)
	}
	// Check if the body was captured
	if string(apiErr.Body) != errorBody {
		t.Errorf("APIError Body = %s; want %s", string(apiErr.Body), errorBody)
	}
}

func TestDoRequest_NetworkError(t *testing.T) {
	// Create a server that immediately closes, simulating a network error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection immediately
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("Server does not support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("Hijack failed: %v", err)
		}
		conn.Close()
	}))
	server.Close() // Close the listener immediately after getting the URL

	token := "mock-token"
	secret := "mock-secret"
	client, err := NewClient(token, secret, WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetDevices(context.Background())
	if err == nil {
		t.Fatal("Expected a network error, but got nil error")
	}

	// Check if the error is a URL error (typical for connection issues)
	// Note: The exact error type might vary depending on OS and Go version
	if _, ok := err.(*url.Error); !ok {
		// Allow *APIError wrapping a network error too if doRequest wraps it
		if apiErr, okApi := err.(*APIError); !okApi || apiErr.Err == nil {
			t.Logf("Expected a network-related error (e.g., *url.Error or wrapped), got %T: %v", err, err)
		}
	}
	// Check if error message indicates connection refusal or similar
	errMsg := err.Error()
	if !strings.Contains(errMsg, "connection refused") && !strings.Contains(errMsg, "connection reset by peer") && !strings.Contains(errMsg, "EOF") {
		t.Logf("Error message %q doesn't clearly indicate a typical network error", errMsg)
	}

}

func TestDoRequest_CustomJSONHandler(t *testing.T) {
	encoderCalled := false
	decoderCalled := false

	customEncoder := func(v any) ([]byte, error) {
		encoderCalled = true
		return json.Marshal(v) // Use standard json for the test logic
	}
	customDecoder := func(data []byte, v any) error {
		decoderCalled = true
		return json.Unmarshal(data, v) // Use standard json for the test logic
	}

	// Mock server response
	expectedBody := `{"key":"value"}`
	mockResponse := fmt.Sprintf(`{"statusCode": 100, "message": "success", "body": %s}`, expectedBody)

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check if request body was encoded (for POST/PUT etc.)
		if r.Method == http.MethodPost {
			bodyBytes, _ := io.ReadAll(r.Body)
			// Here you could check if bodyBytes matches what customEncoder produced
			t.Logf("Received POST body: %s", string(bodyBytes))
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, mockResponse)
	}

	client, _ := setupMockServer(t, handler) // Gets a client pointed to the mock server

	// Apply custom handlers AFTER creating the client for this test
	// (Alternatively, create a new client with options)
	client.jsonEncoder = customEncoder
	client.jsonDecoder = customDecoder

	// Test with GET (only decoder should be called)
	_, err := client.GetDevices(context.Background()) // GetDevices uses GET
	if err != nil {
		t.Fatalf("GetDevices with custom handlers returned error: %v", err)
	}
	if !decoderCalled {
		t.Error("Custom JSON decoder was not called for GET request")
	}
	if encoderCalled {
		t.Error("Custom JSON encoder was unexpectedly called for GET request")
	}

	// Reset flags and test with POST
	encoderCalled = false
	decoderCalled = false
	requestBody := map[string]string{"command": "test"}
	// Use a method that uses POST, like SendDeviceCommand
	_, err = client.SendDeviceCommand(context.Background(), "dummyID", "test", requestBody, "command")
	if err != nil {
		t.Fatalf("SendDeviceCommand with custom handlers returned error: %v", err)
	}
	if !encoderCalled {
		t.Error("Custom JSON encoder was not called for POST request")
	}
	if !decoderCalled {
		t.Error("Custom JSON decoder was not called for POST request")
	}
}
