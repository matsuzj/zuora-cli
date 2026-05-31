package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
