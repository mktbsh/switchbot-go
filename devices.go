package switchbot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Device represents a generic physical device structure from the device list.
// Use map[string]interface{} for flexibility due to varying fields per deviceType.
type Device map[string]interface{}

// InfraredRemoteDevice represents a virtual infrared remote device from the device list.
type InfraredRemoteDevice struct {
	DeviceID    string `json:"deviceId"`
	DeviceName  string `json:"deviceName"`
	RemoteType  string `json:"remoteType"`
	HubDeviceID string `json:"hubDeviceId"`
	_           struct{}
}

// GetDevicesResponse holds the structured response for the GetDevices endpoint.
type GetDevicesResponse struct {
	DeviceList         []Device               `json:"deviceList"`
	InfraredRemoteList []InfraredRemoteDevice `json:"infraredRemoteList"`
	_                  struct{}
}

// GetDevices retrieves the list of all physical and virtual infrared devices associated with the account.
func (c *Client) GetDevices(ctx context.Context) (*GetDevicesResponse, error) {
	path := fmt.Sprintf("/%s/devices", apiVersion)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err // Error already wrapped in doRequest
	}

	var devicesResp GetDevicesResponse
	if err := json.Unmarshal(resp.Body, &devicesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetDevices response body: %w, body: %s", err, string(resp.Body))
	}

	return &devicesResp, nil
}

// DeviceStatus represents the status of a device.
// Use map[string]interface{} for flexibility as the structure is highly dependent on deviceType.
type DeviceStatus map[string]interface{}

// GetDeviceStatus retrieves the current status of a specific physical device.
func (c *Client) GetDeviceStatus(ctx context.Context, deviceID string) (DeviceStatus, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("deviceID cannot be empty")
	}
	path := fmt.Sprintf("/%s/devices/%s/status", apiVersion, deviceID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var status DeviceStatus
	// Handle potentially empty body for devices without status (though unlikely based on docs)
	if isEmptyJSONBody(resp.Body) {
		if err := json.Unmarshal(resp.Body, &status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal GetDeviceStatus response body for %s: %w, body: %s", deviceID, err, string(resp.Body))
		}
	} else {
		// Return an empty map if the body is empty, though the API usually returns structured data or an error.
		status = make(DeviceStatus)
	}

	return status, nil
}

// CommandRequest represents the JSON body for sending a command to a device.
type CommandRequest struct {
	Command     string      `json:"command"`
	CommandType string      `json:"commandType"`
	Parameter   interface{} `json:"parameter"` // Use "default" or specific structure (map/struct)
	_           struct{}
}

// CommandResponse represents the response body after sending a command.
// Often empty ({}), but can contain fields like "commandId" for asynchronous operations (e.g., Keypad).
// Use map[string]interface{} for flexibility.
type CommandResponse map[string]interface{}

// SendDeviceCommand sends a control command to a specific device (physical or virtual IR).
// parameter: Use "default" for simple commands, or a map/struct for complex ones (e.g., setAll, setMode).
// commandType: Use "command" (default) for standard commands, "customize" for IR custom buttons.
func (c *Client) SendDeviceCommand(ctx context.Context, deviceID string, command string, parameter interface{}, commandType string) (CommandResponse, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("deviceID cannot be empty")
	}
	if command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	// Set defaults if not provided
	effectiveParameter := parameter
	if effectiveParameter == nil {
		effectiveParameter = "default"
	}
	effectiveCommandType := commandType
	if effectiveCommandType == "" {
		effectiveCommandType = "command" // Default for most API commands
	}

	reqBody := CommandRequest{
		Command:     command,
		Parameter:   effectiveParameter,
		CommandType: effectiveCommandType,
	}

	path := fmt.Sprintf("/%s/devices/%s/commands", apiVersion, deviceID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}

	var cmdResp CommandResponse
	// Handle potentially empty body for successful commands
	if isEmptyJSONBody(resp.Body) {
		if err := json.Unmarshal(resp.Body, &cmdResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SendDeviceCommand response body for %s: %w, body: %s", deviceID, err, string(resp.Body))
		}
	} else {
		cmdResp = make(CommandResponse) // Return empty map for empty body
	}

	return cmdResp, nil
}
