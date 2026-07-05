package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errors.go behaviors: Error() hint rendering and ExitCode mapping
// (consolidated verbatim from errors_hints_test.go + exitcode_test.go).

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

// TestAPIError_Error_SafeToRetryHintIncludesIdemKey pins that the SafeToRetry
// hint surfaces the Idempotency-Key VALUE when present, and omits the
// "Idempotency-Key: <value>" line when the key is empty (the generic sentence,
// which uses "Idempotency-Key," with a comma, still appears).
func TestAPIError_Error_SafeToRetryHintIncludesIdemKey(t *testing.T) {
	withKey := (&APIError{StatusCode: http.StatusInternalServerError, SafeToRetry: true, IdemKey: "zr-abc123"}).Error()
	assert.Contains(t, withKey, "Idempotency-Key: zr-abc123")

	noKey := (&APIError{StatusCode: http.StatusInternalServerError, SafeToRetry: true}).Error()
	assert.Contains(t, noKey, retryHintSub, "the generic retry hint still appears")
	assert.NotContains(t, noKey, "Idempotency-Key: ", "no key line when IdemKey is empty")
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

// TestParseAPIError_ObjectCRUDSingleError pins that the uppercase Object-CRUD
// failure envelope ({"Success":false,"Errors":[{"Code","Message"}]}) is parsed
// into a clean Code/Message — before this branch existed it fell through to the
// raw-body fallback, leaking the whole JSON blob as the message.
func TestParseAPIError_ObjectCRUDSingleError(t *testing.T) {
	e := parseAPIError(http.StatusBadRequest, []byte(`{"Success":false,"Errors":[{"Code":"INVALID_VALUE","Message":"bad quantity"}]}`))
	assert.Equal(t, "INVALID_VALUE", e.Code)
	assert.Equal(t, "bad quantity", e.Message)
}

// TestParseAPIError_ObjectCRUDMultipleErrors pins that every uppercase error is
// surfaced, mirroring the v1 multi-reason behavior.
func TestParseAPIError_ObjectCRUDMultipleErrors(t *testing.T) {
	body := []byte(`{"Success":false,"Errors":[{"Code":"C1","Message":"first"},{"Code":"REQUIRED_VALUE_MISSING","Message":"second"}]}`)
	got := parseAPIError(http.StatusBadRequest, body).Error()
	assert.Contains(t, got, "2 errors")
	assert.Contains(t, got, "first")
	assert.Contains(t, got, "second")
	assert.Contains(t, got, "REQUIRED_VALUE_MISSING")
}

// The error path is not just HTTP 200 + {"success":false}: real failures arrive
// as 4xx/5xx with assorted bodies. The next tests pin the remaining parse shapes
// and that the HTTP status is preserved regardless of body.

func TestParseAPIError_V2ErrorObjectShape(t *testing.T) {
	// Newer REST surfaces a single error object: {"error":{"code","message"}}.
	e := parseAPIError(http.StatusBadRequest, []byte(`{"error":{"code":"INVALID_TOKEN","message":"token expired"}}`))
	assert.Equal(t, "INVALID_TOKEN", e.Code)
	assert.Equal(t, "token expired", e.Message)
	assert.Equal(t, http.StatusBadRequest, e.StatusCode)
}

func TestParseAPIError_V3TopLevelMessageShape(t *testing.T) {
	// Some gateways return a bare {"message":"..."} with no reasons/error object.
	e := parseAPIError(http.StatusNotFound, []byte(`{"message":"resource not found"}`))
	assert.Equal(t, "resource not found", e.Message)
	assert.Equal(t, http.StatusNotFound, e.StatusCode)
}

func TestParseAPIError_NonJSONBodyFallsBackToRaw(t *testing.T) {
	// A proxy/CDN 502 often returns HTML or plain text, not JSON; the raw body
	// becomes the message instead of being silently dropped.
	e := parseAPIError(http.StatusBadGateway, []byte("<html>502 Bad Gateway</html>"))
	assert.Contains(t, e.Message, "502 Bad Gateway")
	assert.Equal(t, http.StatusBadGateway, e.StatusCode)
}

func TestParseAPIError_EmptyBodyKeepsStatus(t *testing.T) {
	// A bodyless 5xx must still produce a non-nil error carrying the status.
	e := parseAPIError(http.StatusServiceUnavailable, nil)
	require.NotNil(t, e)
	assert.Equal(t, http.StatusServiceUnavailable, e.StatusCode)
	assert.Empty(t, e.Message)
}

func TestParseAPIError_OversizedBodyTruncated(t *testing.T) {
	// A huge non-JSON body is truncated with an ellipsis so the error stays
	// readable rather than dumping kilobytes.
	big := strings.Repeat("x", maxRawErrorBody+500)
	e := parseAPIError(http.StatusInternalServerError, []byte(big))
	assert.Equal(t, maxRawErrorBody+len("..."), len(e.Message))
	assert.True(t, strings.HasSuffix(e.Message, "..."), "oversized body must be truncated with ...")
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

// exitCodeOf mirrors cmd/zr/main.go's exitCode(): an error implementing
// ExitCode() returns its code, otherwise a general error is exit 1.
func exitCodeOf(err error) int {
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

// TestAPIError_Unwrap_ContextCanceled pins the Unwrap contract the CLI exit
// code depends on: an APIError wrapping a cancellation must satisfy
// errors.Is(err, context.Canceled) so cmd/zr maps Ctrl-C to exit 130 instead
// of misreporting it as an API failure.
func TestAPIError_Unwrap_ContextCanceled(t *testing.T) {
	err := &APIError{StatusCode: 0, Err: context.Canceled}
	assert.True(t, errors.Is(err, context.Canceled),
		"APIError.Unwrap must expose the wrapped transport error to errors.Is")
	assert.NoError(t, (&APIError{StatusCode: 400}).Unwrap(),
		"a response-derived APIError wraps nothing")
}

func TestAPIError_TransportFailure_ExitCode(t *testing.T) {
	assert.Equal(t, 1, (&APIError{StatusCode: 0}).ExitCode(),
		"a transport failure (no HTTP response) is not a 4xx client error; it must exit 1")
}

// TestClient_TransportError_ExitCodeConsistentAcrossMethods covers the audit
// finding that the same transport failure exited 1 on idempotent GET/DELETE but
// 3 (mislabeled "4xx client error") on POST/PUT/PATCH. Both must now yield 1.
func TestClient_TransportError_ExitCodeConsistentAcrossMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL
	srv.Close() // closing guarantees connection-refused at addr

	client := newNoSleepClient(WithBaseURL(addr))

	_, postErr := client.Post("/v1/x", strings.NewReader(`{}`))
	require.Error(t, postErr)
	assert.Equal(t, 1, exitCodeOf(postErr), "POST transport failure must exit 1")

	_, getErr := client.Get("/v1/x")
	require.Error(t, getErr)
	assert.Equal(t, 1, exitCodeOf(getErr), "GET transport failure must exit 1 (consistent with POST)")
}
