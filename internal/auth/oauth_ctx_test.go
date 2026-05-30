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

// A cancelled context must abort a token refresh before/while contacting the
// OAuth endpoint, rather than blocking until the client timeout.
func TestTokenContext_CancelAbortsRefresh(t *testing.T) {
	// Server hangs longer than the test would tolerate; cancellation must win.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the client cancels
	}))
	defer srv.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: srv.URL}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "id", "secret"))
	ts := &TokenSource{Config: cfg, Creds: creds}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(50 * time.Millisecond); cancel() }()

	start := time.Now()
	_, err := ts.TokenContext(ctx, "sandbox")
	elapsed := time.Since(start)

	require.Error(t, err, "a cancelled refresh must return an error")
	assert.Less(t, elapsed, 5*time.Second, "cancellation must abort well before the 30s client timeout")
}
