package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A redirecting OAuth token endpoint must NOT have its redirect followed, so the
// client_secret in the POST body can never be forwarded to the redirect target.
func TestOAuth_DoesNotFollowRedirect_SecretNotLeaked(t *testing.T) {
	var attackerGotSecret bool
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("client_secret") != "" {
			attackerGotSecret = true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+"/oauth/token", http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: origin.URL}
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "old-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}))

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", "SUPERSECRET"))

	// HTTPClient nil => the default client, which refuses to follow redirects.
	ts := &TokenSource{Config: cfg, Creds: creds}

	_, err := ts.Token("sandbox")
	require.Error(t, err, "a redirecting token endpoint must not yield a token")
	assert.False(t, attackerGotSecret, "client_secret must never be forwarded to the redirect target")
}
