package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTokenSource builds a TokenSource whose "sandbox" environment points at the
// given test server, with credentials already present.
func newTokenSource(t *testing.T, baseURL string) *TokenSource {
	t.Helper()
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: baseURL}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", "test-secret"))
	return &TokenSource{Config: cfg, Creds: creds}
}

func TestRefresh_EmptyAccessToken_ErrorsAndDoesNotCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"access_token":"","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	_, err := ts.Refresh("sandbox")
	require.Error(t, err, "an empty access_token must be rejected, not cached")

	cached, _ := ts.Config.Token("sandbox")
	if cached != nil {
		assert.Empty(t, cached.AccessToken, "an empty token must not be cached as if valid")
	}
}

func TestRefresh_NonJSON200_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html>not json</html>`))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	_, err := ts.Refresh("sandbox")
	require.Error(t, err)
}

func TestRefresh_HTTPError_TruncatesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	_, err := ts.Refresh("sandbox")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestRefresh_StoresExpiry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"access_token":"tok","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	tok, err := ts.Refresh("sandbox")
	require.NoError(t, err)
	assert.Equal(t, "tok", tok)

	cached, err := ts.Config.Token("sandbox")
	require.NoError(t, err)
	require.NotNil(t, cached)
	assert.True(t, cached.IsValid(), "a freshly refreshed token must be valid")
}
