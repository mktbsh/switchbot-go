package switchbot

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	testCases := []struct {
		name          string
		apiError      APIError
		expectedError string // Expected substring or full string
	}{
		{
			name: "Simple error with status and message",
			apiError: APIError{
				StatusCode: 161,
				Message:    "device offline",
			},
			expectedError: "SwitchBot API error: statusCode=161, message='device offline'",
		},
		{
			name: "Error with status, message, and simple body",
			apiError: APIError{
				StatusCode: 152,
				Message:    "device not found",
				Body:       json.RawMessage(`{"details":"invalid id"}`),
			},
			// Body should be included, formatting might vary slightly
			expectedError: `SwitchBot API error: statusCode=152, message='device not found', body={"details":"invalid id"}`,
		},
		{
			name: "Error with status, message, and nested body",
			apiError: APIError{
				StatusCode: 190,
				Message:    "command format error",
				Body:       json.RawMessage(`{"error":{"code":10,"reason":"bad parameter"}}`),
			},
			expectedError: `SwitchBot API error: statusCode=190, message='command format error', body={"error":{"code":10,"reason":"bad parameter"}}`,
		},
		{
			name: "Error with status, message, and underlying Go error",
			apiError: APIError{
				StatusCode: 500,
				Message:    "internal server error",
				Err:        errors.New("underlying network issue"),
			},
			expectedError: "SwitchBot API error: statusCode=500, message='internal server error' (caused by: underlying network issue)",
		},
		{
			name: "Error with all fields",
			apiError: APIError{
				StatusCode: 401,
				Message:    "Unauthorized",
				Body:       json.RawMessage(`{"reason":"invalid token"}`),
				Err:        errors.New("authentication failed"),
			},
			expectedError: `SwitchBot API error: statusCode=401, message='Unauthorized', body={"reason":"invalid token"} (caused by: authentication failed)`,
		},
		{
			name: "Error with empty body",
			apiError: APIError{
				StatusCode: 160,
				Message:    "command not supported",
				Body:       json.RawMessage(""),
			},
			// Body should NOT be included
			expectedError: "SwitchBot API error: statusCode=160, message='command not supported'",
		},
		{
			name: "Error with nil body",
			apiError: APIError{
				StatusCode: 160,
				Message:    "command not supported",
				Body:       nil,
			},
			// Body should NOT be included
			expectedError: "SwitchBot API error: statusCode=160, message='command not supported'",
		},
		{
			name: "Error with empty object body {}",
			apiError: APIError{
				StatusCode: 171,
				Message:    "hub offline",
				Body:       json.RawMessage("{}"),
			},
			// Body should NOT be included by isEmptyJSONBody logic
			expectedError: "SwitchBot API error: statusCode=171, message='hub offline'",
		},
		{
			name: "Error with null body",
			apiError: APIError{
				StatusCode: 171,
				Message:    "hub offline",
				Body:       json.RawMessage("null"),
			},
			// Body should NOT be included by isEmptyJSONBody logic
			expectedError: "SwitchBot API error: statusCode=171, message='hub offline'",
		},
		{
			name: "Error with status only (message might be auto-generated)",
			apiError: APIError{
				StatusCode: 404,
				// Message: "", // Intentionally empty
				Err: errors.New("resource not found"),
			},
			// Expecting the status code and the underlying error
			expectedError: "SwitchBot API error: statusCode=404, message='' (caused by: resource not found)",
		},
		{
			name: "Error with non-JSON body", // Simulate case where body parsing failed
			apiError: APIError{
				StatusCode: 502,
				Message:    "Bad Gateway",
				Body:       json.RawMessage("This is not JSON"),
				Err:        errors.New("upstream error"),
			},
			// Body should be included as raw string
			expectedError: "SwitchBot API error: statusCode=502, message='Bad Gateway', body=This is not JSON (caused by: upstream error)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.apiError.Error()

			// Using Contains because exact formatting of the body might vary slightly
			// depending on how it was originally marshaled if we were to re-marshal it.
			// For stricter tests, compare field by field or use a JSON comparison library.
			if !strings.Contains(actualError, tc.expectedError) {
				// For simple cases without body/err, direct comparison is fine.
				if tc.apiError.Body == nil && tc.apiError.Err == nil {
					if actualError != tc.expectedError {
						t.Errorf("Error() = %q; want %q", actualError, tc.expectedError)
					}
				} else {
					// Log both for easier debugging if Contains fails
					t.Errorf("Error() = %q; expected to contain %q", actualError, tc.expectedError)
				}
			}

			// Additional check: ensure the basic prefix is always there
			if !strings.HasPrefix(actualError, "SwitchBot API error:") {
				t.Errorf("Error() message %q missing expected prefix", actualError)
			}
		})
	}
}
