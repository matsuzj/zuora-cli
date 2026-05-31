package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A context cancelled before the refresh starts must abort immediately with the
// context error, rather than contacting the OAuth endpoint or blocking until the
// client timeout.
func TestTokenContext_AlreadyCancelled_Aborts(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(200)
		w.Write([]byte(`{"access_token":"x","expires_in":3600}`))
	}))
	defer srv.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: srv.URL}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "id", "secret"))
	ts := &TokenSource{Config: cfg, Creds: creds}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before the call

	start := time.Now()
	_, err := ts.TokenContext(ctx, "sandbox")
	elapsed := time.Since(start)

	require.Error(t, err, "a cancelled context must abort the refresh")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 2*time.Second, "must fail fast, not wait for the client timeout")
	assert.Equal(t, 0, hits, "no OAuth request should be sent once the context is cancelled")
}

// A context cancelled mid-flight (while the OAuth endpoint is slow) must abort
// well before the 30s client timeout. The server responds after a short delay
// so the test never hangs even if cancellation propagation regresses.
func TestTokenContext_CancelMidFlight_AbortsFast(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond slowly; a working cancellation should beat this.
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"access_token":"x","expires_in":3600}`))
	}))
	defer srv.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: srv.URL}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "id", "secret"))
	ts := &TokenSource{Config: cfg, Creds: creds}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()

	start := time.Now()
	_, err := ts.TokenContext(ctx, "sandbox")
	elapsed := time.Since(start)

	require.Error(t, err, "a cancelled mid-flight refresh must return an error")
	assert.Less(t, elapsed, 2*time.Second, "cancellation must abort well before the 30s client timeout")
}
