// Package api provides an HTTP client for Zuora APIs.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// maxRawErrorBody caps how much of an unparseable error body is echoed to the
// user, mirroring the OAuth path, so a large HTML gateway page is not dumped.
const maxRawErrorBody = 500

// APIError represents a Zuora API error response.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Raw        string
	// SafeToRetry is set on a non-idempotent (POST/PATCH) failure that was not
	// retried automatically, to tell the user the command can be safely re-run
	// because it carries an Idempotency-Key.
	SafeToRetry bool
}

func (e *APIError) Error() string {
	// A 2xx response carrying success=false means the HTTP call succeeded but
	// Zuora reported an application-level failure. Calling that "error (HTTP
	// 200)" is self-contradictory, so frame it as a request failure instead.
	header := fmt.Sprintf("Zuora API error (HTTP %d)", e.StatusCode)
	if e.StatusCode >= 200 && e.StatusCode < 300 {
		header = fmt.Sprintf("Zuora API request failed (HTTP %d, success=false)", e.StatusCode)
	}

	var msg string
	if e.Code != "" {
		msg = fmt.Sprintf("%s\n  Code: %s\n  Message: %s", header, e.Code, e.Message)
	} else {
		msg = fmt.Sprintf("%s: %s", header, e.Message)
	}
	if e.StatusCode == http.StatusUnauthorized {
		msg += "\n  Hint: credentials may be expired. Run: zr auth login"
	}
	if e.SafeToRetry {
		msg += "\n  Hint: this write was not retried automatically. It is safe to run the" +
			" command again — it carries an Idempotency-Key, so if the original" +
			" request did go through, the retry returns HTTP 409 instead of" +
			" creating a duplicate."
	}
	return msg
}

// ExitCode maps the HTTP status to a CLI exit code:
// 2 for 401 (auth — matches AuthError), 4 for 5xx (server), 1 for a transport
// failure (no HTTP response), 3 for other 4xx.
func (e *APIError) ExitCode() int {
	if e.StatusCode == http.StatusUnauthorized {
		return 2
	}
	if e.StatusCode == 0 {
		// Transport failure (no HTTP response: connection refused, DNS, timeout).
		// It is NOT a 4xx client error. The idempotent-method retry path returns
		// the bare transport error (exit 1); non-idempotent methods wrap it in an
		// APIError{StatusCode:0}, so map that to exit 1 too for a consistent,
		// method-independent exit code.
		return 1
	}
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
		// decodeCode unquotes a reason code, which may be a JSON string like
		// "INVALID" or a number like 53100020.
		decodeCode := func(raw json.RawMessage) string {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil {
				return s
			}
			return string(raw)
		}
		if len(v1.Reasons) == 1 {
			apiErr.Code = decodeCode(v1.Reasons[0].Code)
			apiErr.Message = v1.Reasons[0].Message
			return apiErr
		}
		// Zuora frequently returns several validation failures at once. Showing
		// only the first hides the rest, so list every reason.
		parts := make([]string, len(v1.Reasons))
		for i, r := range v1.Reasons {
			if code := decodeCode(r.Code); code != "" {
				parts[i] = fmt.Sprintf("[%s] %s", code, r.Message)
			} else {
				parts[i] = r.Message
			}
		}
		apiErr.Message = fmt.Sprintf("%d errors:\n  - %s", len(parts), strings.Join(parts, "\n  - "))
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

	msg := string(body)
	if len(msg) > maxRawErrorBody {
		msg = msg[:maxRawErrorBody] + "..."
	}
	apiErr.Message = msg
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
