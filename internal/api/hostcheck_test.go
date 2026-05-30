package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A cross-host absolute URL must be refused so the bearer token is never sent
// to a host other than the configured environment.
func TestClient_CrossHostAbsoluteURL_Refused(t *testing.T) {
	var gotAuth string
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer attacker.Close()

	c := NewClient(
		WithBaseURL("https://rest.zuora.com"),
		WithTokenSource(func() (string, error) { return "secret-token", nil }),
	)
	_, err := c.Get(attacker.URL + "/v1/accounts")
	require.Error(t, err, "an off-host absolute URL must be refused")
	assert.Contains(t, err.Error(), "refusing to send credentials")
	assert.Empty(t, gotAuth, "the bearer token must never reach the off-host server")
}

// A same-host absolute URL (e.g. a pagination nextPage) is allowed.
func TestClient_SameHostAbsoluteURL_Allowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	resp, err := c.Get(srv.URL + "/v1/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// Relative paths resolve against the base URL and are always allowed.
func TestClient_RelativePath_Allowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/accounts")
	require.NoError(t, err)
}

// Context cancellation during retry backoff surfaces the cancellation error,
// not a stale prior API error from an earlier attempt.
func TestRetry_ContextCancel_ReturnsContextError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"reasons":[{"code":"X","message":"boom"}]}`))
	}))
	defer srv.Close()

	// The backoff seam returns a context error on the first wait, simulating
	// Ctrl-C after the first 500 response.
	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	c.sleep = func(_ context.Context, _ time.Duration) error { return context.Canceled }
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled, "cancellation must surface the context error, not the stale 500 APIError")
	var apiErr *APIError
	assert.False(t, errors.As(err, &apiErr), "must not surface the stale APIError on cancel")
}
