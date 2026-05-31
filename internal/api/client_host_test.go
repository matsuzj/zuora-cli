package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckHost_RelativeAllowed pins that a relative path (no host) is allowed:
// it is already rooted at the configured base URL.
func TestCheckHost_RelativeAllowed(t *testing.T) {
	c := NewClient(WithBaseURL("https://rest.zuora.com"))
	require.NoError(t, c.checkHost("/v1/accounts"))
}

// TestCheckHost_SameHostAbsoluteAllowed pins that an absolute URL on the same
// host as baseURL is allowed (e.g. a pagination nextPage link).
func TestCheckHost_SameHostAbsoluteAllowed(t *testing.T) {
	c := NewClient(WithBaseURL("https://rest.zuora.com"))
	require.NoError(t, c.checkHost("https://rest.zuora.com/v1/accounts?page=2"))
}

// TestCheckHost_EmptyBaseHostAllowsCrossHost pins that when baseURL has no host
// (base.Host == ""), the cross-host guard is skipped and any absolute URL is
// allowed through.
func TestCheckHost_EmptyBaseHostAllowsCrossHost(t *testing.T) {
	c := NewClient(WithBaseURL(""))
	require.NoError(t, c.checkHost("https://anything.example.com/v1/accounts"))
}

// TestCheckHost_MalformedAbsoluteURL pins the parse-error branch: an absolute
// URL that url.Parse rejects is wrapped as an "invalid request URL" error.
func TestCheckHost_MalformedAbsoluteURL(t *testing.T) {
	c := NewClient(WithBaseURL("https://rest.zuora.com"))
	err := c.checkHost("https://rest.zuora.com:notaport/v1/accounts")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request URL")
}

// TestCheckHost_CrossHostBlocked pins the security guard: a parseable absolute
// URL on a different host is refused, and the error names both the offending
// host and the configured host.
func TestCheckHost_CrossHostBlocked(t *testing.T) {
	c := NewClient(WithBaseURL("https://rest.zuora.com"))
	err := c.checkHost("https://evil.example.com/v1/accounts")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to send credentials")
	assert.Contains(t, err.Error(), "evil.example.com")
	assert.Contains(t, err.Error(), "rest.zuora.com")
}

// TestDo_CrossHostBlocked drives the guard through the public Do path against an
// httptest server: the cross-host request is refused and never reaches the
// server (so the bearer token is not leaked off-host).
func TestDo_CrossHostBlocked(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(
		WithBaseURL(srv.URL),
		WithTokenSource(func(context.Context) (string, error) { return "secret", nil }),
	)
	_, err := c.Do(http.MethodGet, "https://evil.example.com/v1/accounts")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to send credentials")
	assert.False(t, hit, "cross-host request must not reach the server")
}

// TestDo_SameHostAllowed pins that a same-host request succeeds end-to-end
// through the guard (the allow path that returns a Response).
func TestDo_SameHostAllowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.Do(http.MethodGet, "/v1/accounts")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestDo_IdempotencyKeyPerPostRequestAndPrefix pins newIdempotencyKey behaviour
// observed through real POSTs: each POST carries an Idempotency-Key header, the
// key has the zr- prefix, and two POSTs produce two distinct keys.
func TestDo_IdempotencyKeyPerPostRequestAndPrefix(t *testing.T) {
	var (
		mu   sync.Mutex
		keys []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	for i := 0; i < 2; i++ {
		_, err := c.Do(http.MethodPost, "/v1/accounts", WithBody(strings.NewReader("{}")))
		require.NoError(t, err)
	}

	require.Len(t, keys, 2)
	for _, k := range keys {
		assert.True(t, strings.HasPrefix(k, "zr-"), "key %q should have zr- prefix", k)
	}
	assert.NotEqual(t, keys[0], keys[1], "two POSTs must use distinct idempotency keys")
}

// TestDo_GetHasNoIdempotencyKey pins that GET requests do not carry an
// Idempotency-Key header (only POST/PATCH do).
func TestDo_GetHasNoIdempotencyKey(t *testing.T) {
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key = r.Header.Get("Idempotency-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Do(http.MethodGet, "/v1/accounts")
	require.NoError(t, err)
	assert.Empty(t, key)
}
