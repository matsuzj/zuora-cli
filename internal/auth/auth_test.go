package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToken_CachedValid(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: "http://unused"}
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "cached-token",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}))

	ts := &TokenSource{
		Config: cfg,
		Creds:  NewMockCredentialStore(),
	}

	token, err := ts.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "cached-token", token)
}

func TestToken_CachedExpired_Refresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))
		assert.Equal(t, "test-id", r.PostForm.Get("client_id"))
		assert.Equal(t, "test-secret", r.PostForm.Get("client_secret"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "new-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "old-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}))

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", "test-secret"))

	ts := &TokenSource{
		Config: cfg,
		Creds:  creds,
	}

	token, err := ts.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "new-token", token)

	// Verify token was persisted
	cached, err := cfg.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "new-token", cached.AccessToken)
	assert.True(t, cached.IsValid())
	assert.Equal(t, 1, cfg.SaveCallCount)
}

func TestToken_NoCredentials(t *testing.T) {
	cfg := config.NewMockConfig()
	creds := NewMockCredentialStore()

	ts := &TokenSource{
		Config: cfg,
		Creds:  creds,
	}

	_, err := ts.Token("sandbox")
	assert.Error(t, err)

	var authErr *AuthError
	assert.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Hint, "zr auth login")
}

func TestToken_AuthFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer server.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "bad-id", "bad-secret"))

	ts := &TokenSource{
		Config: cfg,
		Creds:  creds,
	}

	_, err := ts.Token("sandbox")
	assert.Error(t, err)

	var authErr *AuthError
	assert.ErrorAs(t, err, &authErr)
	assert.Equal(t, 2, authErr.ExitCode())
}
