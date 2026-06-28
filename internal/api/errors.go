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
	// IdemKey is the Idempotency-Key the failed POST/PATCH carried, surfaced in
	// the SafeToRetry hint so the user can quote it to Zuora support without
	// having to re-run under -v. Empty when the request carried no key.
	IdemKey string
	// Err is the underlying transport error, when this APIError wraps one
	// (StatusCode 0). Response-derived errors leave it nil.
	Err error
}

// Unwrap exposes the underlying transport error to errors.Is/As chains.
func (e *APIError) Unwrap() error { return e.Err }

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
		if e.IdemKey != "" {
			msg += fmt.Sprintf("\n  Idempotency-Key: %s", e.IdemKey)
		}
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

	// Zuora reports a logical failure as a list of reasons in one of two shapes:
	// v1 REST uses lowercase {"reasons":[{"code","message"}]}, Object CRUD uses
	// uppercase {"Errors":[{"Code","Message"}]}. Both collapse onto applyReasons.
	var v1 struct {
		Reasons []struct {
			Code    json.RawMessage `json:"code"`
			Message string          `json:"message"`
		} `json:"reasons"`
	}
	if err := json.Unmarshal(body, &v1); err == nil && len(v1.Reasons) > 0 {
		reasons := make([]failureReason, len(v1.Reasons))
		for i, r := range v1.Reasons {
			reasons[i] = failureReason{code: decodeCode(r.Code), message: r.Message}
		}
		apiErr.applyReasons(reasons)
		return apiErr
	}

	// Object-CRUD failure: {"Success":false,"Errors":[{"Code","Message"}]}.
	var objCRUD struct {
		Errors []struct {
			Code    json.RawMessage `json:"Code"`
			Message string          `json:"Message"`
		} `json:"Errors"`
	}
	if err := json.Unmarshal(body, &objCRUD); err == nil && len(objCRUD.Errors) > 0 {
		reasons := make([]failureReason, len(objCRUD.Errors))
		for i, e := range objCRUD.Errors {
			reasons[i] = failureReason{code: decodeCode(e.Code), message: e.Message}
		}
		apiErr.applyReasons(reasons)
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

// failureReason is one entry of a Zuora logical-failure list — the lowercase v1
// "reasons" array or the uppercase Object-CRUD "Errors" array.
type failureReason struct {
	code    string
	message string
}

// decodeCode unquotes a Zuora reason/error code, which may be a JSON string like
// "INVALID_VALUE" or a number like 53100020.
func decodeCode(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// applyReasons fills Code/Message from the parsed failure reasons: a single
// reason becomes Code+Message; multiple are listed (Zuora frequently returns
// several validation failures at once, and surfacing only the first hides the
// rest).
func (e *APIError) applyReasons(reasons []failureReason) {
	if len(reasons) == 1 {
		e.Code = reasons[0].code
		e.Message = reasons[0].message
		return
	}
	parts := make([]string, len(reasons))
	for i, r := range reasons {
		if r.code != "" {
			parts[i] = fmt.Sprintf("[%s] %s", r.code, r.message)
		} else {
			parts[i] = r.message
		}
	}
	e.Message = fmt.Sprintf("%d errors:\n  - %s", len(parts), strings.Join(parts, "\n  - "))
}

// ReadOnlyError is returned when a write operation is blocked in read-only mode.
type ReadOnlyError struct {
	Method string
	Path   string
	// Hint, when non-empty, appends extra guidance to the message (e.g. the
	// Data Query opt-in toggle). It is empty for ordinary write blocks, so the
	// existing message — and the tests asserting it — stay unchanged.
	Hint string
}

func (e *ReadOnlyError) Error() string {
	msg := "blocked: write operation not allowed in read-only mode. Remove --read-only flag or unset ZR_READ_ONLY to enable write operations"
	if e.Method != "" && e.Path != "" {
		msg = fmt.Sprintf("blocked: %s %s not allowed in read-only mode. Remove --read-only flag or unset ZR_READ_ONLY to enable write operations", e.Method, e.Path)
	}
	if e.Hint != "" {
		msg += ". " + e.Hint
	}
	return msg
}

// ExitCode returns 5 for read-only violations (1=general, 2=auth, 3=4xx, 4=5xx, 5=read-only).
func (e *ReadOnlyError) ExitCode() int { return 5 }

// successEnvelopeError reports the logical failure carried by an HTTP-2xx
// body whose Zuora success flag is false (v1 REST uses lowercase "success",
// Object CRUD uses uppercase "Success"). Returns nil for non-JSON bodies and
// for bodies without a success flag, so non-envelope responses pass through.
func successEnvelopeError(statusCode int, body []byte) error {
	var envelope struct {
		Success      *bool `json:"success"`
		SuccessUpper *bool `json:"Success"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return nil
	}
	if envelope.Success != nil && !*envelope.Success {
		return parseAPIError(statusCode, body)
	}
	if envelope.SuccessUpper != nil && !*envelope.SuccessUpper {
		return parseAPIError(statusCode, body)
	}
	return nil
}
