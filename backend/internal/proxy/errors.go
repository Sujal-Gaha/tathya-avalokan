package proxy

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"tathya-avalokan/backend/internal/response"
)

// ClassifyExecutionError analyzes an error from QueryContext/ExecContext and returns
// the corresponding HTTP status, error code, and error details map.
func ClassifyExecutionError(err error) (int, string, string, map[string]any) {
	if err == nil {
		return http.StatusOK, "", "", nil
	}

	details := make(map[string]any)

	// 1. Timeout / Deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "deadline exceeded") {
		return http.StatusRequestTimeout, response.ErrCodeQueryTimeout, "Query execution exceeded timeout limit", nil
	}

	errMsg := err.Error()

	// 2. Syntax / Database Execution errors
	return http.StatusBadRequest, response.ErrCodeQueryExecution, errMsg, details
}

// ClassifyConnectionError analyzes an error from PingContext or database connection handshake.
func ClassifyConnectionError(err error, host string, port int, driverType string) (int, string, string, map[string]any) {
	details := map[string]any{
		"host":        host,
		"port":        port,
		"driver_type": driverType,
	}

	msg := "Unable to connect to target database instance"
	if err != nil {
		msg = err.Error()
	}

	return http.StatusBadGateway, response.ErrCodeConnectionFailed, msg, details
}
