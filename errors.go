package switchbot

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error response from the SwitchBot API.
type APIError struct {
	Body    json.RawMessage `json:"body"`
	Message string          `json:"message"`
	// Body might contain more details for specific errors.
	// Underlying HTTP error or context, if any
	Err error

	StatusCode int `json:"statusCode"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("SwitchBot API error: statusCode=%d, message='%s'", e.StatusCode, e.Message)
	// Avoid printing empty or null body
	if isEmptyJSONBody(e.Body) {
		msg += fmt.Sprintf(", body=%s", string(e.Body))
	}
	if e.Err != nil {
		msg += fmt.Sprintf(" (caused by: %v)", e.Err)
	}
	return msg
}
