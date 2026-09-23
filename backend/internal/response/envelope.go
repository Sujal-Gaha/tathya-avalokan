package response

import (
	"encoding/json"
	"net/http"
)

// Standard Error Codes
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeReadOnly        = "READ_ONLY_VIOLATION"
	ErrCodeQueryExecution  = "QUERY_EXECUTION_ERROR"
	ErrCodeQueryTimeout    = "QUERY_TIMEOUT"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
)

// ErrorPayload represents the standardized API error structure.
type ErrorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// Envelope represents the unified API response format.
type Envelope struct {
	Data     any            `json:"data"`
	Error    *ErrorPayload  `json:"error"`
	Metadata map[string]any `json:"metadata"`
}

// SendJSON writes a success response enveloped in the standard JSON structure.
func SendJSON(w http.ResponseWriter, status int, data any, metadata map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if metadata == nil {
		metadata = make(map[string]any)
	}

	env := Envelope{
		Data:     data,
		Error:    nil,
		Metadata: metadata,
	}

	_ = json.NewEncoder(w).Encode(env)
}

// SendError writes an error response enveloped in the standard JSON structure.
func SendError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	env := Envelope{
		Data: nil,
		Error: &ErrorPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
		Metadata: make(map[string]any),
	}

	_ = json.NewEncoder(w).Encode(env)
}
