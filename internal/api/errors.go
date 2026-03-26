// Package api provides an HTTP client for Zuora APIs.
package api

import (
	"encoding/json"
	"fmt"
)

// APIError represents a Zuora API error response.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Raw        string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Zuora API error (HTTP %d)\n  Code: %s\n  Message: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("Zuora API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// ExitCode returns 3 for client errors (4xx) and 4 for server errors (5xx).
func (e *APIError) ExitCode() int {
	if e.StatusCode >= 500 {
		return 4
	}
	return 3
}

// parseAPIError attempts to parse a Zuora error response body.
func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Raw:        string(body),
	}

	// Try Zuora v1 format: {"success": false, "reasons": [{"code": ..., "message": ...}]}
	var v1 struct {
		Success bool `json:"success"`
		Reasons []struct {
			Code    json.RawMessage `json:"code"`
			Message string          `json:"message"`
		} `json:"reasons"`
	}
	if err := json.Unmarshal(body, &v1); err == nil && len(v1.Reasons) > 0 {
		// Unquote the code (may be a string like "INVALID" or a number like 53100020)
		code := string(v1.Reasons[0].Code)
		var codeStr string
		if err := json.Unmarshal(v1.Reasons[0].Code, &codeStr); err == nil {
			code = codeStr
		}
		apiErr.Code = code
		apiErr.Message = v1.Reasons[0].Message
		return apiErr
	}

	// Try alternative format: {"error": {"code": "...", "message": "..."}}
	var v2 struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &v2); err == nil && v2.Error.Code != "" {
		apiErr.Code = v2.Error.Code
		apiErr.Message = v2.Error.Message
		return apiErr
	}

	// Try simple message format
	var v3 struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &v3); err == nil && v3.Message != "" {
		apiErr.Message = v3.Message
		return apiErr
	}

	apiErr.Message = string(body)
	return apiErr
}

// ReadOnlyError is returned when a write operation is blocked in read-only mode.
type ReadOnlyError struct {
	Method string
	Path   string
}

func (e *ReadOnlyError) Error() string {
	if e.Method != "" && e.Path != "" {
		return fmt.Sprintf("blocked: %s %s not allowed in read-only mode. Remove --read-only flag or unset ZR_READ_ONLY to enable write operations", e.Method, e.Path)
	}
	return "blocked: write operation not allowed in read-only mode. Remove --read-only flag or unset ZR_READ_ONLY to enable write operations"
}

// ExitCode returns 5 for read-only violations (1=general, 2=auth, 3=4xx, 4=5xx, 5=read-only).
func (e *ReadOnlyError) ExitCode() int { return 5 }
