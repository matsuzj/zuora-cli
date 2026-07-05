package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noSleep replaces the backoff seam so retry tests run instantly while still
// honoring context cancellation.
func noSleep(ctx context.Context, _ time.Duration) error {
	return ctx.Err()
}

func newNoSleepClient(opts ...ClientOption) *Client {
	base := []ClientOption{func(c *Client) { c.sleep = noSleep }}
	return NewClient(append(base, opts...)...)
}

func TestRetry_GET_5xx_Retries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, int32(maxRetries+1), atomic.LoadInt32(&calls), "GET 5xx should retry up to maxRetries")
}

func TestRetry_POST_5xx_NotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Post("/v1/accounts", strings.NewReader(`{}`))
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "POST 5xx must NOT be retried (no duplicate mutation)")
}

func TestRetry_PATCH_5xx_NotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(503)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Do(http.MethodPatch, "/v1/accounts/1", WithBody(strings.NewReader(`{}`)))
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "PATCH 5xx must NOT be retried")
}

func TestRetry_PUT_5xx_NotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Put("/v1/accounts/1", strings.NewReader(`{}`))
	require.Error(t, err)
	// PUT carries no Idempotency-Key (Zuora rejects it) and Zuora's PUTs are
	// often non-idempotent actions, so a 5xx must NOT auto-retry (double-apply).
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "PUT 5xx must NOT be retried")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.False(t, apiErr.SafeToRetry, "a keyless PUT must not be advertised as safe-to-retry")
}

func TestRetry_5xx_PreservesZuoraError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte(`{"reasons":[{"code":"SERVICE_DOWN","message":"maintenance window"}]}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "SERVICE_DOWN", apiErr.Code, "exhausted retry must surface Zuora's real error code")
	assert.Contains(t, apiErr.Message, "maintenance window")
}

func TestRetry_POST_TransportError_NotRetried(t *testing.T) {
	// Closed server -> transport error on connect.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := newNoSleepClient(WithBaseURL(url))
	_, err := c.Post("/v1/accounts", strings.NewReader(`{}`))
	require.Error(t, err, "POST transport error must surface immediately, not be retried")
}

func TestRetry_429_RetriesAndSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			w.Write([]byte(`{"message":"slow down"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	resp, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls), "one 429 then one success = exactly 2 calls")
}

// TestRetry_429_Exhausted pins the give-up path: when every attempt is rate-
// limited, doWithRetry exhausts maxRetries+1 attempts and surfaces the final
// 429 as a non-nil *APIError (not a nil error, not a hang). The existing 429
// tests only cover "one 429 then success", never the loop-exhaustion exit.
func TestRetry_429_Exhausted(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"rate limited"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, int32(maxRetries+1), atomic.LoadInt32(&calls),
		"all attempts must be made before giving up")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr,
		"exhausted 429 retries must surface as *APIError, not nil or a generic error")
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
}

// TestRetry_POST_5xx_PopulatesIdemKey pins that a non-retried POST 5xx surfaces
// the request's Idempotency-Key on the APIError, so the SafeToRetry hint can
// show the key the user should quote to Zuora support.
func TestRetry_POST_5xx_PopulatesIdemKey(t *testing.T) {
	var sawKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey = r.Header.Get("Idempotency-Key")
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Post("/v1/orders", strings.NewReader(`{}`))
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.NotEmpty(t, sawKey)
	assert.Equal(t, sawKey, apiErr.IdemKey, "the hint key must match the key actually sent")
	assert.Contains(t, apiErr.Error(), "Idempotency-Key: "+apiErr.IdemKey)
}

func TestRetry_POST_429_SendsIdempotencyKeyStableAcrossRetries(t *testing.T) {
	var keys []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Post("/v1/accounts", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.NotEmpty(t, keys[0], "POST must carry an Idempotency-Key so a 429 replay is deduplicated server-side")
	assert.Equal(t, keys[0], keys[1], "Idempotency-Key must be stable across retries")
}

func TestParseRetryAfter(t *testing.T) {
	// Numeric seconds.
	d, ok := parseRetryAfter("120")
	assert.True(t, ok)
	assert.Equal(t, 120*time.Second, d)

	// Negative clamps to zero.
	d, ok = parseRetryAfter("-5")
	assert.True(t, ok)
	assert.Equal(t, time.Duration(0), d)

	// HTTP-date form must be honored (the regression that was silently lost).
	future := time.Now().Add(90 * time.Second).UTC().Format(http.TimeFormat)
	d, ok = parseRetryAfter(future)
	assert.True(t, ok, "an HTTP-date Retry-After must be parsed, not ignored")
	assert.Greater(t, d, 30*time.Second)

	// Past date clamps to zero.
	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	d, ok = parseRetryAfter(past)
	assert.True(t, ok)
	assert.Equal(t, time.Duration(0), d)

	// Empty / garbage -> not usable.
	_, ok = parseRetryAfter("")
	assert.False(t, ok)
	_, ok = parseRetryAfter("soon")
	assert.False(t, ok)
}

func TestRetry_POST_401Refresh_KeepsIdempotencyKeyStable(t *testing.T) {
	var keys []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithTokenSource(func(context.Context) (string, error) { return "t1", nil }),
		WithRefreshToken(func(context.Context) (string, error) { return "t2", nil }),
	)
	_, err := c.Post("/v1/accounts", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.NotEmpty(t, keys[0])
	assert.Equal(t, keys[0], keys[1], "Idempotency-Key must stay stable across a 401-refresh resend, not be regenerated")
}

func TestRetry_BodyReplayedOnRetry(t *testing.T) {
	var bodies []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests) // 429 is retried for any method
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	// A 429 is retried for any method (it was rate-limited, not processed); the
	// body must be replayed intact on the retried request.
	_, err := c.Put("/v1/accounts/1", strings.NewReader(`{"name":"x"}`))
	require.NoError(t, err)
	require.Len(t, bodies, 2)
	assert.Equal(t, `{"name":"x"}`, bodies[0])
	assert.Equal(t, `{"name":"x"}`, bodies[1], "retried request must resend the full body, not an empty one")
}

func TestRetry_401_RefreshError_Aborts(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(401)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithTokenSource(func(context.Context) (string, error) { return "tok", nil }),
		WithRefreshToken(func(context.Context) (string, error) { return "", assertErr{} }),
	)
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "a failed refresh must abort, not loop")
}

func TestRetry_ContextCancelled_StopsImmediately(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(500)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	c.SetContext(ctx)
	_, err := c.Get("/v1/test")
	require.Error(t, err)
}

type assertErr struct{}

func (assertErr) Error() string { return "refresh failed" }

// ——— moved verbatim from saferetry_test.go (P4-2 test consolidation) ———

// A POST that fails with 5xx is not retried, but the returned error must be
// marked SafeToRetry so the CLI can tell the user it is safe to re-run (the
// request carries an Idempotency-Key).
func TestRetry_POST_5xx_MarkedSafeToRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"reasons":[{"code":"X","message":"boom"}]}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Post("/v1/orders", strings.NewReader(`{}`))
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.SafeToRetry, "POST 5xx error must be marked safe to re-run")
	assert.Contains(t, apiErr.Error(), "Idempotency-Key")
	assert.Contains(t, apiErr.Error(), "HTTP 409")
}

// A GET 5xx (retried and exhausted) must NOT be marked SafeToRetry (it's a read).
func TestRetry_GET_5xx_NotMarkedSafeToRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte(`{"reasons":[{"code":"Y","message":"down"}]}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.False(t, apiErr.SafeToRetry, "a GET is a read; no safe-to-retry mutation hint")
}

// A POST transport error (server unreachable) is surfaced immediately, marked
// SafeToRetry, and never auto-retried.
func TestRetry_POST_TransportError_MarkedSafeToRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // now unreachable -> transport error

	c := newNoSleepClient(WithBaseURL(url))
	_, err := c.Post("/v1/orders", strings.NewReader(`{}`))
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.SafeToRetry)
}

// ─── P6-1: verbose retry visibility ───

// TestVerbose_BackoffAndIdempotent5xxLines pins the retry-loop diagnostics:
// a 5xx on an idempotent GET logs the 5xx decision and the backoff before
// each re-attempt.
func TestVerbose_BackoffAndIdempotent5xxLines(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(500)
			w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)

	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "* HTTP 500 on idempotent GET, will retry")
	assert.Contains(t, out, "s backoff (attempt 2/4)")
}

// TestVerbose_RetryAfterLine pins the 429 Retry-After diagnostic.
func TestVerbose_RetryAfterLine(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
			w.Write([]byte(`{"message":"slow down"}`))
			return
		}
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)

	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "* HTTP 429, honoring Retry-After: 2s (cap 60s)")
}

// TestVerbose_TokenRefreshMasksBearer is the mandatory P6-1 leak guard: the
// 401-refresh diagnostics must show the hardcoded "Bearer ***" and the real
// refreshed token value must NEVER reach the verbose stream.
func TestVerbose_TokenRefreshMasksBearer(t *testing.T) {
	const secretToken = "SECRET-REFRESHED-TOKEN-VALUE"
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{"message":"expired"}`))
			return
		}
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithRefreshToken(func(context.Context) (string, error) { return secretToken, nil }),
	)
	var buf strings.Builder
	c.SetVerbose(&buf)

	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "* HTTP 401, refreshing token and resending")
	assert.Contains(t, out, "(Bearer ***)")
	assert.NotContains(t, out, secretToken, "refreshed token value must never be logged")
}

// TestVerbose_OffProducesNoStarLines: without SetVerbose the retry loop stays
// silent (vlogf no-op).
func TestVerbose_OffProducesNoStarLines(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(500)
			w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	// nothing to assert on a writer (none set); reaching here without a
	// panic proves the nil-writer no-op path.
}

// ─── P6-3: gated body logging ───

func TestVerboseBody_RequestAndResponseAtLevel2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"marker":"RESP-BODY"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	_, err := c.Post("/v1/test", strings.NewReader(`{"marker":"REQ-BODY"}`))
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, `> {"marker":"REQ-BODY"}`)
	assert.Contains(t, out, `RESP-BODY`)
}

// TestVerboseBody_RedactsSecrets pins the end-to-end masking (#325): sensitive
// field VALUES in both the request and the response body must be redacted before
// they reach the verbose log, even at level 2.
func TestVerboseBody_RedactsSecrets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"token":"resp-secret-xyz"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	_, err := c.Post("/v1/payment-methods",
		strings.NewReader(`{"cardNumber":"4111111111111111","cardSecurityCode":"999888777"}`))
	require.NoError(t, err)

	out := buf.String()
	assert.NotContains(t, out, "4111111111111111", "request card number must be redacted")
	assert.NotContains(t, out, "999888777", "request security code must be redacted")
	assert.NotContains(t, out, "resp-secret-xyz", "response token must be redacted")
	assert.Contains(t, out, "***REDACTED***", "the redaction marker must appear")
}

func TestVerboseBody_Level1OmitsBodies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"marker":"RESP-BODY"}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf) // level 1 only

	_, err := c.Post("/v1/test", strings.NewReader(`{"marker":"REQ-BODY"}`))
	require.NoError(t, err)
	out := buf.String()
	assert.NotContains(t, out, "REQ-BODY", "level 1 must not log request bodies (PII)")
	assert.NotContains(t, out, "RESP-BODY", "level 1 must not log response bodies (PII)")
}

func TestVerboseBody_TruncatesAtCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	big := `{"pad":"` + strings.Repeat("x", maxBodyLog) + `"}`
	_, err := c.Post("/v1/test", strings.NewReader(big))
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "[body truncated at 4096 bytes]")
}

func TestVerboseBody_MultipartSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	_, err := c.Post("/v1/test", strings.NewReader("SECRET-FILE-CONTENT"),
		WithHeader("Content-Type", "multipart/form-data; boundary=xyz"))
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "[multipart body omitted]")
	assert.NotContains(t, out, "SECRET-FILE-CONTENT")
}

// TestVerboseBody_RetryLayerErrorBodies pins the Codex finding: bodies of
// responses consumed INSIDE the retry loop (429/5xx) must still surface
// under level-2 verbose — they are the main diagnostic payload.
func TestVerboseBody_RetryLayerErrorBodies(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(500)
			w.Write([]byte(`{"message":"ERR-BODY-MARKER"}`))
			return
		}
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ERR-BODY-MARKER",
		"retry-layer error bodies must surface under -vv")
}

// TestVerboseBody_MultipartSkipCaseInsensitive pins the MIME case rule: a
// "Multipart/Form-Data" spelling must also be skipped.
func TestVerboseBody_MultipartSkipCaseInsensitive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var buf strings.Builder
	c.SetVerbose(&buf)
	c.SetVerboseBody()

	_, err := c.Post("/v1/test", strings.NewReader("SECRET-UPLOAD"),
		WithHeader("Content-Type", "Multipart/Form-Data; boundary=xyz"))
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "[multipart body omitted]")
	assert.NotContains(t, buf.String(), "SECRET-UPLOAD")
}

// TestParseRetryAfter_OverflowSecondsClamped pins the int64-overflow guard:
// pre-fix, "9223372037" seconds wrapped NEGATIVE when multiplied by
// time.Second, slipped past the caller's `d > maxRetryAfter` cap (negative is
// not greater), and turned the rate-limit wait into zero-delay hot retries —
// the exact hostile-header scenario the cap exists to prevent.
func TestParseRetryAfter_OverflowSecondsClamped(t *testing.T) {
	for _, v := range []string{"9223372037", "9223372036854775807"} {
		d, ok := parseRetryAfter(v)
		require.True(t, ok, v)
		assert.GreaterOrEqual(t, d, time.Duration(0), "parsed duration must never be negative (%s)", v)
		assert.Equal(t, maxRetryAfter, d, v)
	}
	// Beyond int64 entirely: Atoi fails and it is not an HTTP date -> unusable.
	_, ok := parseRetryAfter("92233720368547758070")
	assert.False(t, ok)
}

// TestRetry_429_RetryAfterCapped pins the 60s cap end-to-end through the 429
// path: both a large-but-sane server value (3600) and an overflowing hostile
// value must produce exactly one recorded sleep of maxRetryAfter.
func TestRetry_429_RetryAfterCapped(t *testing.T) {
	for _, header := range []string{"3600", "9223372037"} {
		t.Run(header, func(t *testing.T) {
			var calls int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&calls, 1) == 1 {
					w.Header().Set("Retry-After", header)
					w.WriteHeader(429)
					w.Write([]byte(`{"message":"rate limited"}`))
					return
				}
				w.WriteHeader(200)
				w.Write([]byte(`{"ok":true}`))
			}))
			defer srv.Close()

			var slept []time.Duration
			c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
			c.sleep = func(ctx context.Context, d time.Duration) error {
				slept = append(slept, d)
				return ctx.Err()
			}

			resp, err := c.Get("/v1/test")
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			require.Len(t, slept, 1, "Retry-After present: exactly one honored wait, no extra backoff")
			assert.Equal(t, maxRetryAfter, slept[0], "wait must be capped at maxRetryAfter, never negative/zero")
		})
	}
}

// TestClient_401_RefreshThen401_SurfacesAuthErrorWithoutLoop pins the
// !tokenRefreshed guard: a second 401 after the one allowed refresh must
// surface the auth error (with its login hint) instead of refreshing again —
// exactly 2 HTTP calls and 2 token-source calls (initial + one refresh).
func TestClient_401_RefreshThen401_SurfacesAuthErrorWithoutLoop(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer srv.Close()

	var tokenCalls int32
	c := newNoSleepClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithTokenSource(func(context.Context) (string, error) {
			n := atomic.AddInt32(&tokenCalls, 1)
			return fmt.Sprintf("token-%d", n), nil
		}),
	)

	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls), "second 401 must not trigger another refresh/resend cycle")
	assert.Equal(t, int32(2), atomic.LoadInt32(&tokenCalls), "exactly one initial token + one refresh")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.StatusCode)
	assert.Contains(t, err.Error(), "zr auth login", "the second 401 must surface the auth hint")
}
