package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelopeServer returns an HTTP-200 handler that responds with a Zuora
// success=false envelope carrying `code` for the first `failures` calls, then
// success=true. It records the call count, every request body, and every
// Idempotency-Key so tests can prove resend behavior.
type envelopeServer struct {
	mu       sync.Mutex
	failures int
	code     string // numeric Zuora code, emitted unquoted (e.g. "53100050")
	multi    bool   // emit two reasons instead of one
	calls    int
	bodies   []string
	idemKeys []string
}

func (s *envelopeServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.calls++
		n := s.calls
		s.bodies = append(s.bodies, string(body))
		s.idemKeys = append(s.idemKeys, r.Header.Get("Idempotency-Key"))
		s.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		if n <= s.failures {
			if s.multi {
				fmt.Fprintf(w, `{"success":false,"reasons":[{"code":%s,"message":"a"},{"code":%s,"message":"b"}]}`, s.code, s.code)
			} else {
				fmt.Fprintf(w, `{"success":false,"reasons":[{"code":%s,"message":"transient"}]}`, s.code)
			}
			return
		}
		w.Write([]byte(`{"success":true}`))
	}
}

func (s *envelopeServer) count() int { s.mu.Lock(); defer s.mu.Unlock(); return s.calls }

// A transient (code ending 50) success=false on an idempotent GET is retried
// until it succeeds.
func TestDo_TransientSuccessFalse_GET_Retried(t *testing.T) {
	es := &envelopeServer{failures: 2, code: "53100050"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	var buf strings.Builder
	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	c.SetVerbose(&buf)

	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.Equal(t, 3, es.count(), "2 transient failures then success = 3 calls")
	assert.Contains(t, buf.String(), "success=false transient error", "each retry should log a backoff line")
}

// POST is retried (it carries an Idempotency-Key) AND the request body is
// resent intact on every attempt — the regression this guards: doWithRetry
// consumes req.Body, so a naive resend would transmit an empty body.
func TestDo_TransientSuccessFalse_POST_ResendsBodyAndKey(t *testing.T) {
	es := &envelopeServer{failures: 2, code: "53100050"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Post("/v1/orders", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)

	es.mu.Lock()
	defer es.mu.Unlock()
	require.Equal(t, 3, es.calls)
	for i, b := range es.bodies {
		assert.Equal(t, `{"a":1}`, b, "attempt %d must resend the full body, not an empty one", i+1)
	}
	require.NotEmpty(t, es.idemKeys[0])
	for i, k := range es.idemKeys {
		assert.Equal(t, es.idemKeys[0], k, "attempt %d must reuse the same Idempotency-Key so the resend dedupes", i+1)
	}
}

// PUT is never retried on a transient success=false: it carries no
// Idempotency-Key (Zuora rejects PUT+key), so a resend could double-apply.
func TestDo_TransientSuccessFalse_PUT_NotRetried(t *testing.T) {
	es := &envelopeServer{failures: 99, code: "53100050"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Put("/v1/accounts/A-1", strings.NewReader(`{}`))
	require.Error(t, err)
	assert.Equal(t, 1, es.count(), "PUT must not be retried")
}

// A non-transient code (suffix 40 = Not Found) is not retried even on a GET.
func TestDo_NonTransientSuccessFalse_GET_NotRetried(t *testing.T) {
	es := &envelopeServer{failures: 99, code: "53100040"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, 1, es.count(), "non-transient code must not be retried")
}

// A multi-reason envelope is conservatively NOT retried (parseAPIError leaves
// Code empty for >1 reason, so a batch with any non-transient reason is safe).
func TestDo_MultiReasonSuccessFalse_GET_NotRetried(t *testing.T) {
	es := &envelopeServer{failures: 99, code: "53100050", multi: true}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, 1, es.count(), "multi-reason envelope must not be retried")
}

// A persistently transient success=false exhausts the retries and surfaces the
// envelope error.
func TestDo_TransientSuccessFalse_Exhausts(t *testing.T) {
	es := &envelopeServer{failures: 99, code: "53100050"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, maxRetries+1, es.count(), "all attempts made before giving up")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "53100050", apiErr.Code)
}

// WithoutCheckSuccess delivers a success=false body uninterpreted (raw zr api
// GET passthrough) — no error, no retry.
func TestDo_WithoutCheckSuccess_PassesThrough(t *testing.T) {
	es := &envelopeServer{failures: 99, code: "53100050"}
	srv := httptest.NewServer(es.handler())
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	resp, err := c.Get("/v1/test", WithoutCheckSuccess())
	require.NoError(t, err)
	assert.Equal(t, 1, es.count(), "passthrough must not retry")
	assert.Contains(t, string(resp.Body), `"success":false`)
}

func TestIsTransientBodyCode(t *testing.T) {
	for _, code := range []string{"53100050", "53100061", "53100070", "53100099", "50", "99"} {
		assert.True(t, isTransientBodyCode(code), "%s should be transient", code)
	}
	for _, code := range []string{"53100040", "53100030", "53100020", "53100000", "INVALID", ""} {
		assert.False(t, isTransientBodyCode(code), "%s should NOT be transient", code)
	}
}

// A request that hits BOTH transient HTTP failures (5xx) AND a transient
// success=false must share ONE retry budget across doWithRetry's loop and the
// outer success-envelope loop — the two must not multiply. Regression guard for
// the nested-budget defect: with a per-loop budget the pattern below (three
// 500s then a 200 success=false, repeating) could send (maxRetries+1)^2 = 16
// times; sharing the budget caps it at maxRetries+1.
func TestDo_SharedBudget_MixedTransientDoesNotMultiply(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if (n-1)%4 < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":false,"reasons":[{"code":53100050,"message":"locked"}]}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	mu.Lock()
	got := calls
	mu.Unlock()
	assert.LessOrEqual(t, got, maxRetries+1,
		"transport and success-envelope retries must share one budget, not multiply (got %d sends)", got)
}
