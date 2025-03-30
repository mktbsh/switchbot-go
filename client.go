package switchbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	DefaultBaseURL = "https://api.switch-bot.com"
	apiVersion     = "v1.1"
)

type JSONMarshal func(v any) ([]byte, error)

type JSONUnmarshal func(data []byte, v any) error

// Client manages communication with the SwitchBot API.
type Client struct {
	token       string
	secret      string
	jsonEncoder JSONMarshal
	jsonDecoder JSONUnmarshal
	httpClient  *http.Client
	baseURL     *url.URL
	_           struct{}
}

// ClientOption defines a function type for configuring the Client.
type ClientOption func(*Client) error

// WithHTTPClient sets a custom HTTP client for the SwitchBot Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			// Or return an error, depending on desired behavior
			c.httpClient = http.DefaultClient
		} else {
			c.httpClient = httpClient
		}
		return nil
	}
}

// WithBaseURL sets a custom base URL for the SwitchBot Client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("invalid base URL %q: %w", baseURL, err)
		}
		c.baseURL = parsedURL
		return nil
	}
}

// WithJSONEncoder sets a custom JSON handler for marshalling.
func WithJSONEncoder(encoder JSONMarshal) ClientOption {
	return func(c *Client) error {
		if encoder == nil {
			return fmt.Errorf("JSONEncoder cannot be nil")
		}
		c.jsonEncoder = encoder
		return nil
	}
}

// WithJSONDecoder sets a custom JSON handler for un marshalling.
func WithJSONDecoder(decoder JSONUnmarshal) ClientOption {
	return func(c *Client) error {
		if decoder == nil {
			return fmt.Errorf("JSONDecoder cannot be nil")
		}
		c.jsonDecoder = decoder
		return nil
	}
}

// NewClient creates a new SwitchBot API client with optional configurations.
func NewClient(token, secret string, options ...ClientOption) (*Client, error) {
	if token == "" || secret == "" {
		return nil, fmt.Errorf("token and secret must not be empty")
	}

	baseURL, _ := url.Parse(DefaultBaseURL) // Error ignored as DefaultBaseURL is static

	// Initialize client with defaults
	client := &Client{
		httpClient:  http.DefaultClient, // Default HTTP client
		baseURL:     baseURL,
		token:       token,
		secret:      secret,
		jsonEncoder: json.Marshal,   // Default JSON encoder
		jsonDecoder: json.Unmarshal, // Default JSON decoder
	}

	// Apply all provided options
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	return client, nil
}

// --- Generic Request Handling ---

// Response is the generic structure for SwitchBot API responses.
type Response struct {
	StatusCode int             `json:"statusCode"`
	Message    string          `json:"message"`
	Body       json.RawMessage `json:"body"` // Use json.RawMessage to delay parsing specific body structures
}

// doRequest performs the actual HTTP request with authentication and error handling.
func (c *Client) doRequest(ctx context.Context, method, path string, requestBody interface{}) (*Response, error) {
	relURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q: %w", path, err)
	}
	absURL := c.baseURL.ResolveReference(relURL)

	var bodyReader io.Reader
	var reqBodyBytes []byte // Store request body bytes for potential logging or retries
	if requestBody != nil {
		reqBodyBytes, err = c.jsonEncoder(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, absURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthorizationHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to %s: %w", absURL.String(), err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", absURL.String(), err)
	}

	// Attempt to parse into the standard SwitchBot response structure first
	var apiResp Response
	if err := c.jsonDecoder(respBodyBytes, &apiResp); err != nil {
		// If parsing fails, check HTTP status for error indication
		if resp.StatusCode >= 400 {
			return nil, &APIError{
				StatusCode: resp.StatusCode, // Use HTTP status as primary code
				Message:    fmt.Sprintf("Received HTTP %d error with unparsable body", resp.StatusCode),
				Body:       json.RawMessage(respBodyBytes), // Include raw body
				Err:        err,                            // Include parsing error
			}
		}
		// If HTTP status is OK (2xx/3xx) but body is not standard JSON, it's unusual
		return nil, fmt.Errorf("failed to unmarshal successful response (HTTP %d) body: %w, body: %s", resp.StatusCode, err, string(respBodyBytes))
	}

	// Check SwitchBot API specific status code for application-level errors
	// StatusCode 100 is the primary success indicator from SwitchBot.
	// Other codes (even with HTTP 200 OK) usually indicate specific issues.
	if apiResp.StatusCode != 100 {
		// Check if it's a known error code based on documentation
		knownErrorCodes := map[int]bool{
			151: true, // device type error
			152: true, // device not found
			160: true, // command not supported
			161: true, // device offline
			171: true, // hub offline
			190: true, // internal error / invalid command format
			// Add other known non-100 error codes if necessary
		}
		if knownErrorCodes[apiResp.StatusCode] {
			return nil, &APIError{
				StatusCode: apiResp.StatusCode,
				Message:    apiResp.Message,
				Body:       apiResp.Body,
				Err:        fmt.Errorf("received API status code %d", apiResp.StatusCode),
			}
		}
		// If it's not 100 and not a known error code, it might be unexpected or for async ops.
		// Return the response but let caller be aware. Consider logging a warning.
		// fmt.Printf("Warning: Received non-100 API status code %d: %s\n", apiResp.StatusCode, apiResp.Message)
	}
	// Also check HTTP status code for client/server errors (redundant but safe)
	if resp.StatusCode >= 400 {
		errToReturn := &APIError{
			StatusCode: resp.StatusCode, // Prioritize HTTP status code for 4xx/5xx
			Message:    apiResp.Message, // Use message from parsed body if available
			Body:       apiResp.Body,
			Err:        fmt.Errorf("received HTTP status code %d", resp.StatusCode),
		}
		// If message was empty in parsed body, use default HTTP status text
		if errToReturn.Message == "" {
			errToReturn.Message = fmt.Sprintf("Received HTTP %d error", resp.StatusCode)
		}
		return nil, errToReturn
	}

	// If API status code is 100 and HTTP status is OK, return the successful response
	return &apiResp, nil
}
