package switchbot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Scene represents a manual scene defined in the SwitchBot app.
type Scene struct {
	SceneID   string `json:"sceneId"`
	SceneName string `json:"sceneName"`
	_         struct{}
}

// GetScenes retrieves the list of manual scenes configured by the user.
func (c *Client) GetScenes(ctx context.Context) ([]Scene, error) {
	path := fmt.Sprintf("/%s/scenes", apiVersion)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var scenes []Scene
	if err := json.Unmarshal(resp.Body, &scenes); err != nil {
		// Handle case where body might be empty array "[]" which is valid JSON
		if string(resp.Body) == "[]" {
			return []Scene{}, nil // Return empty slice
		}
		return nil, fmt.Errorf("failed to unmarshal GetScenes response body: %w, body: %s", err, string(resp.Body))
	}

	return scenes, nil
}

// ExecuteScene triggers the execution of a specific manual scene.
// The response body is typically empty ({}) on success.
func (c *Client) ExecuteScene(ctx context.Context, sceneID string) error {
	if sceneID == "" {
		return fmt.Errorf("sceneID cannot be empty")
	}
	path := fmt.Sprintf("/%s/scenes/%s/execute", apiVersion, sceneID)
	_, err := c.doRequest(ctx, http.MethodPost, path, nil)
	return err
}
