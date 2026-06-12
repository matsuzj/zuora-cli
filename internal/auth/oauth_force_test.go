package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ForceRefresh must fetch a brand-new token from the OAuth endpoint even when a
// still-valid token is cached. This is the distinguishing behavior from Token,
// which would return the cached value without contacting the server. A regression
// that made ForceRefresh consult the cache (like Token does) would silently keep
// returning a stale/revoked token after a 401, which is exactly what ForceRefresh
// exists to avoid.
func TestForceRefresh_IgnoresValidCacheAndFetchesNew(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		require.Equal(t, "/oauth/token", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"forced-token","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: srv.URL}
	// Seed a still-valid cached token. Token() would return this untouched.
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "cached-token",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}))

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", "test-secret"))
	ts := &TokenSource{Config: cfg, Creds: creds}

	// Sanity guard: with the same valid cache, Token returns the cached value
	// and does NOT hit the server.
	tok, err := ts.Token("sandbox")
	require.NoError(t, err)
	require.Equal(t, "cached-token", tok)
	require.Equal(t, int32(0), atomic.LoadInt32(&hits), "Token must serve a valid cache without a network call")

	// ForceRefresh must bypass that valid cache and fetch anew.
	forced, err := ts.ForceRefreshContext(context.Background(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, "forced-token", forced, "ForceRefresh must return the freshly fetched token, not the cache")
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits), "ForceRefresh must contact the OAuth endpoint exactly once")

	// The new token must replace the cached one and be persisted.
	cached, err := cfg.Token("sandbox")
	require.NoError(t, err)
	assert.Equal(t, "forced-token", cached.AccessToken, "ForceRefresh must overwrite the cached token")
	assert.Equal(t, 1, cfg.SaveCallCount, "the refreshed token must be saved")
}

// ForceRefreshContext is the cancellable variant; it must thread the context to
// the OAuth request. A context cancelled before the call must abort without
// contacting the server, proving the forced path honours cancellation rather
// than blocking on the HTTP client timeout.
func TestForceRefreshContext_CancelledBeforeCall_Aborts(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write([]byte(`{"access_token":"x","expires_in":3600}`))
	}))
	defer srv.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: srv.URL}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "id", "secret"))
	ts := &TokenSource{Config: cfg, Creds: creds}

	ctx, cancel := newCancelledContext()
	defer cancel()

	start := time.Now()
	_, err := ts.ForceRefreshContext(ctx, "sandbox")
	elapsed := time.Since(start)

	require.Error(t, err, "a cancelled context must abort the forced refresh")
	assert.Less(t, elapsed, 2*time.Second, "must fail fast, not wait for the client timeout")
	assert.Equal(t, int32(0), atomic.LoadInt32(&hits), "no OAuth request should be sent once the context is cancelled")
}
