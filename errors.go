package switchbot

import (
	"encoding/json"
	"fmt"
	"strings"
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
	var sb strings.Builder // Use strings.Builder for efficient string concatenation
	sb.WriteString(fmt.Sprintf("SwitchBot API error: statusCode=%d, message='%s'", e.StatusCode, e.Message))

	// Check if the body is non-empty AND not just "{}", "null" etc. before adding it
	// (Re-using the logic from utils.go/isEmptyJSONBody conceptually)
	bodyStr := string(e.Body)
	isBodyEmptyOrNull := isEmptyJSONBody(e.Body)

	if !isBodyEmptyOrNull {
		sb.WriteString(fmt.Sprintf(", body=%s", bodyStr))
	}

	// Add the underlying error if it exists
	if e.Err != nil {
		sb.WriteString(fmt.Sprintf(" (caused by: %v)", e.Err))
	}

	return sb.String()
}
