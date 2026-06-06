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

// TestAPIError_Error_2xxLogicalFailure pins that a 2xx carrying success=false
// is framed as a request failure, not the self-contradictory "error (HTTP 200)".
func TestAPIError_Error_2xxLogicalFailure(t *testing.T) {
	got := (&APIError{StatusCode: http.StatusOK, Code: "X", Message: "nope"}).Error()
	assert.Contains(t, got, "request failed", "2xx logical failure should read as a request failure")
	assert.Contains(t, got, "success=false")
	assert.NotContains(t, got, "error (HTTP 200)", "must not call a 200 an HTTP error")
	assert.Contains(t, got, "Message: nope")
}

// TestParseAPIError_SingleReason pins back-compat: one reason keeps the
// structured Code/Message split unchanged.
func TestParseAPIError_SingleReason(t *testing.T) {
	e := parseAPIError(http.StatusBadRequest, []byte(`{"success":false,"reasons":[{"code":"INVALID","message":"bad"}]}`))
	assert.Equal(t, "INVALID", e.Code)
	assert.Equal(t, "bad", e.Message)
}

// TestParseAPIError_MultipleReasons pins that EVERY reason is surfaced, not just
// the first, including a numeric code unquoted to its digits.
func TestParseAPIError_MultipleReasons(t *testing.T) {
	body := []byte(`{"success":false,"reasons":[{"code":"C1","message":"first"},{"code":53100020,"message":"second"}]}`)
	got := parseAPIError(http.StatusBadRequest, body).Error()
	assert.Contains(t, got, "2 errors")
	assert.Contains(t, got, "first")
	assert.Contains(t, got, "second")
	assert.Contains(t, got, "C1")
	assert.Contains(t, got, "53100020", "numeric reason codes should appear as digits, not quoted")
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
