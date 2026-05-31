package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// reauthHintSub and retryHintSub are distinctive substrings of the two
// independent hints appended by APIError.Error.
const (
	reauthHintSub = "zr auth login"
	retryHintSub  = "Idempotency-Key"
)

// TestAPIError_Error_UnauthorizedHint pins that a 401 (and only a 401) appends
// the re-auth hint, while a non-retryable 401 does not also add the retry hint.
func TestAPIError_Error_UnauthorizedHint(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusUnauthorized, Message: "denied"}).Error()
	assert.Contains(t, got, "HTTP 401")
	assert.Contains(t, got, "denied")
	assert.Contains(t, got, reauthHintSub)
	assert.NotContains(t, got, retryHintSub)
}

// TestAPIError_Error_SafeToRetryHint pins that SafeToRetry appends the
// idempotency hint and that a non-401 status does not add the re-auth hint.
func TestAPIError_Error_SafeToRetryHint(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusServiceUnavailable, SafeToRetry: true}).Error()
	assert.Contains(t, got, "HTTP 503")
	assert.Contains(t, got, retryHintSub)
	assert.NotContains(t, got, reauthHintSub)
}

// TestAPIError_Error_UnauthorizedAndSafeToRetry pins that a 401 that is also
// safe to retry shows BOTH hints (the two branches are independent).
func TestAPIError_Error_UnauthorizedAndSafeToRetry(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusUnauthorized, SafeToRetry: true}).Error()
	assert.Contains(t, got, "HTTP 401")
	assert.Contains(t, got, reauthHintSub)
	assert.Contains(t, got, retryHintSub)
}

// TestAPIError_Error_NoHints pins the plain form: not a 401 and not retryable
// means neither hint is appended.
func TestAPIError_Error_NoHints(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusInternalServerError, Message: "boom"}).Error()
	assert.Contains(t, got, "HTTP 500")
	assert.Contains(t, got, "boom")
	assert.NotContains(t, got, reauthHintSub)
	assert.NotContains(t, got, retryHintSub)
}

// TestAPIError_Error_CodeBranch pins that a populated Code switches to the
// multi-line "Code:/Message:" form rather than the inline form.
func TestAPIError_Error_CodeBranch(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusBadRequest, Code: "INVALID_VALUE", Message: "bad field"}).Error()
	assert.Contains(t, got, "Code: INVALID_VALUE")
	assert.Contains(t, got, "Message: bad field")
	assert.True(t, strings.Contains(got, "\n"), "code form should be multi-line")
}

// TestReadOnlyError_Error_BothForms pins both ReadOnlyError messages: the
// detailed form when Method+Path are set, and the generic fallback otherwise.
func TestReadOnlyError_Error_BothForms(t *testing.T) {
	withPath := (&ReadOnlyError{Method: http.MethodDelete, Path: "/v1/accounts/A-1"}).Error()
	assert.Contains(t, withPath, "read-only mode")
	assert.Contains(t, withPath, http.MethodDelete)
	assert.Contains(t, withPath, "/v1/accounts/A-1")

	generic := (&ReadOnlyError{}).Error()
	assert.Contains(t, generic, "read-only mode")
	assert.Contains(t, generic, "write operation")
	assert.NotContains(t, generic, http.MethodDelete)
}
