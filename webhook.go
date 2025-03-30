package switchbot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhookSetupRequest is the request body for setting up a webhook.
type WebhookSetupRequest struct {
	Action     string `json:"action"` // Should be "setupWebhook"
	URL        string `json:"url"`
	DeviceList string `json:"deviceList"` // Currently only "ALL" is supported
	_          struct{}
}

// SetupWebhook configures the URL to receive webhook events.
func (c *Client) SetupWebhook(ctx context.Context, webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhookURL cannot be empty")
	}
	reqBody := WebhookSetupRequest{
		Action:     "setupWebhook",
		URL:        webhookURL,
		DeviceList: "ALL", // Per documentation
	}
	path := fmt.Sprintf("/%s/webhook/setupWebhook", apiVersion)
	_, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	return err
}

// WebhookQueryRequest is the request body for querying webhook configurations.
type WebhookQueryRequest struct {
	Action string   `json:"action"`         // "queryUrl" or "queryDetails"
	URLs   []string `json:"urls,omitempty"` // Required only for "queryDetails"
	_      struct{}
}

// WebhookQueryURLResponse represents the response containing configured webhook URLs.
type WebhookQueryURLResponse struct {
	URLs []string `json:"urls"`
	_    struct{}
}

// WebhookDetails represents the detailed configuration of a specific webhook URL.
type WebhookDetails struct {
	URL            string `json:"url"`
	DeviceList     string `json:"deviceList"`     // e.g., "ALL"
	CreateTime     int64  `json:"createTime"`     // Unix timestamp (likely milliseconds)
	LastUpdateTime int64  `json:"lastUpdateTime"` // Unix timestamp (likely milliseconds)
	Enable         bool   `json:"enable"`         // Whether the webhook is active
	_              struct{}
}

// QueryWebhookURL retrieves the list of configured webhook URLs.
func (c *Client) QueryWebhookURL(ctx context.Context) ([]string, error) {
	reqBody := WebhookQueryRequest{Action: "queryUrl"}
	path := fmt.Sprintf("/%s/webhook/queryWebhook", apiVersion)
	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}

	var queryResp WebhookQueryURLResponse
	if err := json.Unmarshal(resp.Body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal QueryWebhookURL response body: %w, body: %s", err, string(resp.Body))
	}
	return queryResp.URLs, nil
}

// QueryWebhookDetails retrieves the detailed configuration for the specified webhook URLs.
func (c *Client) QueryWebhookDetails(ctx context.Context, urls []string) ([]WebhookDetails, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("at least one URL must be provided for queryDetails")
	}
	reqBody := WebhookQueryRequest{Action: "queryDetails", URLs: urls}
	path := fmt.Sprintf("/%s/webhook/queryWebhook", apiVersion)
	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}

	var details []WebhookDetails
	if err := json.Unmarshal(resp.Body, &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal QueryWebhookDetails response body: %w, body: %s", err, string(resp.Body))
	}
	return details, nil
}

// WebhookUpdateRequest is the request body for updating webhook configurations.
type WebhookUpdateRequest struct {
	Action string        `json:"action"` // Should be "updateWebhook"
	Config WebhookConfig `json:"config"`
	_      struct{}
}

// WebhookConfig represents the configuration fields that can be updated for a webhook.
type WebhookConfig struct {
	URL    string `json:"url"`    // The URL whose config is being updated
	Enable bool   `json:"enable"` // The new state (true=enabled, false=disabled)
	_      struct{}
}

// UpdateWebhook enables or disables updates for a specific configured webhook URL.
func (c *Client) UpdateWebhook(ctx context.Context, webhookURL string, enable bool) error {
	if webhookURL == "" {
		return fmt.Errorf("webhookURL cannot be empty")
	}
	reqBody := WebhookUpdateRequest{
		Action: "updateWebhook",
		Config: WebhookConfig{
			URL:    webhookURL,
			Enable: enable,
		},
	}
	path := fmt.Sprintf("/%s/webhook/updateWebhook", apiVersion)
	_, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	return err
}

// WebhookDeleteRequest is the request body for deleting a webhook configuration.
type WebhookDeleteRequest struct {
	Action string `json:"action"` // Should be "deleteWebhook"
	URL    string `json:"url"`    // The URL to delete
	_      struct{}
}

// DeleteWebhook removes the configuration for a specific webhook URL.
func (c *Client) DeleteWebhook(ctx context.Context, webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhookURL cannot be empty")
	}
	reqBody := WebhookDeleteRequest{
		Action: "deleteWebhook",
		URL:    webhookURL,
	}
	path := fmt.Sprintf("/%s/webhook/deleteWebhook", apiVersion)
	_, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	return err
}
