package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ——— moved verbatim from auth_test.go (P4-3 test consolidation) ———

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

func TestToken_RefusesCleartextOAuthToRemoteHost(t *testing.T) {
	// An http:// base URL to a non-loopback host must be refused BEFORE the
	// client_secret is POSTed, so credentials never travel in cleartext. The
	// loopback case (http://127.0.0.1) is exercised by the httptest-backed
	// refresh tests above, which still pass. (#439)
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: "http://proxy.example.com"}
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "old-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}))
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", "test-secret"))

	ts := &TokenSource{Config: cfg, Creds: creds}

	_, err := ts.Token("sandbox")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plaintext HTTP")
	assert.Contains(t, err.Error(), "proxy.example.com")
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

// TestToken_ServerError_ExitCode covers that a 5xx from the OAuth server is
// classified as a server error (exit 4, matching APIError) rather than a
// credential error (exit 2), and that the status is recorded on the AuthError.
func TestToken_ServerError_ExitCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable) // 503
		w.Write([]byte(`{"error":"temporarily_unavailable"}`))
	}))
	defer server.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "id", "secret"))

	ts := &TokenSource{
		Config: cfg,
		Creds:  creds,
	}

	_, err := ts.Token("sandbox")
	require.Error(t, err)

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, 503, authErr.StatusCode)
	assert.Equal(t, 4, authErr.ExitCode(), "OAuth 5xx must map to the server exit code (4)")
}

// ——— moved verbatim from oauth_force_test.go (P4-3 test consolidation) ———

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

// ——— moved verbatim from oauth_edge_test.go (P4-3 test consolidation) ———

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
	_, err := ts.refresh(context.Background(), "sandbox")
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
	_, err := ts.refresh(context.Background(), "sandbox")
	require.Error(t, err)
}

func TestRefresh_HTTPError_TruncatesBody(t *testing.T) {
	// The body must be LONGER than the 200-character cut so this test
	// actually exercises the truncation branch — the original fixture sent
	// 26 bytes and passed with the branch deleted (hollow-test audit).
	longBody := `{"error":"invalid_client","detail":"` + strings.Repeat("x", 300) + `"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(longBody))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	_, err := ts.refresh(context.Background(), "sandbox")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Contains(t, err.Error(), "...", "long bodies must be truncated with an ellipsis")
	assert.NotContains(t, err.Error(), strings.Repeat("x", 201), "no more than 200 chars of the body may leak")
}

func TestRefresh_StoresExpiry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"access_token":"tok","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	ts := newTokenSource(t, srv.URL)
	tok, err := ts.refresh(context.Background(), "sandbox")
	require.NoError(t, err)
	assert.Equal(t, "tok", tok)

	cached, err := ts.Config.Token("sandbox")
	require.NoError(t, err)
	require.NotNil(t, cached)
	assert.True(t, cached.IsValid(), "a freshly refreshed token must be valid")
}

// ——— moved verbatim from oauth_redirect_test.go (P4-3 test consolidation) ———

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

// Same protection must apply when the caller INJECTS an http.Client without its
// own redirect policy — otherwise the client_secret would still leak on a
// redirect. (Gap found by a second-opinion review of the default-only fix.)
func TestOAuth_InjectedClient_DoesNotFollowRedirect(t *testing.T) {
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

	// Injected client with NO CheckRedirect of its own.
	injected := &http.Client{Timeout: 10 * time.Second}
	ts := &TokenSource{Config: cfg, Creds: creds, HTTPClient: injected}

	_, err := ts.Token("sandbox")
	require.Error(t, err)
	assert.False(t, attackerGotSecret, "client_secret must not leak even with an injected client")
	assert.Nil(t, injected.CheckRedirect, "the caller's injected client must not be mutated")
}

// ——— moved verbatim from oauth_ctx_test.go (P4-3 test consolidation) ———

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

// ─── P6-2: auth observability ───

// TestTokenSource_LogfObservabilityAndLeakDenial captures the Logf stream
// across cache-hit, refresh, and force-refresh paths, asserting the expected
// event lines AND — the mandatory leak guard — that neither the client secret
// nor any token value ever reaches the log.
func TestTokenSource_LogfObservabilityAndLeakDenial(t *testing.T) {
	const secret = "SUPER-SECRET-VALUE"
	const issuedToken = "ISSUED-TOKEN-VALUE"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": issuedToken,
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}

	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sandbox", "test-id", secret))

	var buf strings.Builder
	ts := &TokenSource{
		Config: cfg,
		Creds:  creds,
		Logf:   func(format string, args ...any) { fmt.Fprintf(&buf, format, args...) },
	}

	// Cache miss → refresh. This phase builds the client_secret request body,
	// so its log output is the highest-risk for an accidental credential leak.
	_, err := ts.TokenContext(context.Background(), "sandbox")
	require.NoError(t, err)
	missOut := buf.String()
	// MockCredentialStore IS a StaticCredentialStore — the label must say so
	// (Codex: auth login passes one built from flags/prompt, not the keyring).
	assert.Contains(t, missOut, `* auth: fetching token for environment "sandbox" (credentials from explicitly provided values)`)
	assert.Contains(t, missOut, "* auth: token acquired, expires in 3600s")

	// Cache hit.
	buf.Reset()
	_, err = ts.TokenContext(context.Background(), "sandbox")
	require.NoError(t, err)
	hitOut := buf.String()
	assert.Contains(t, hitOut, `* auth: cache hit for environment "sandbox"`)

	// Force refresh.
	buf.Reset()
	_, err = ts.ForceRefreshContext(context.Background(), "sandbox")
	require.NoError(t, err)
	forceOut := buf.String()
	assert.Contains(t, forceOut, `* auth: force-refreshing token for environment "sandbox"`)

	// Leak denial across EVERY phase logged in this test — most importantly the
	// cache-miss refresh above. Earlier this only checked the force-refresh
	// buffer, because the two buf.Reset() calls discarded the miss/hit output
	// before the assertions ran, leaving the highest-risk phase unguarded.
	full := missOut + hitOut + forceOut
	assert.NotContains(t, full, secret, "client secret must never be logged")
	assert.NotContains(t, full, issuedToken, "token values must never be logged")
	assert.NotContains(t, full, "test-id", "client id is deliberately not logged")
}

// TestTokenSource_EnvVarCredentialSourceLine pins the env-pair source label.
func TestTokenSource_EnvVarCredentialSourceLine(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok", "token_type": "bearer", "expires_in": 60,
		})
	}))
	defer server.Close()

	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}

	var buf strings.Builder
	ts := &TokenSource{
		Config: cfg,
		Creds:  &envVarStore{clientID: "id", clientSecret: "sec"},
		Logf:   func(format string, args ...any) { fmt.Fprintf(&buf, format, args...) },
	}
	_, err := ts.ForceRefreshContext(context.Background(), "sandbox")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "credentials from the ZR_CLIENT_ID/ZR_CLIENT_SECRET env vars")
}

// TestTokenSource_NilLogfIsSilent: the zero value stays a no-op.
func TestTokenSource_NilLogfIsSilent(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: "https://rest.test.zuora.com"}
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "cached-token",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}))
	ts := &TokenSource{Config: cfg, Creds: NewMockCredentialStore()}
	_, err := ts.TokenContext(context.Background(), "sandbox")
	require.NoError(t, err) // reaching here without panic proves the nil guard
}

// raceSafeStore is a minimal thread-safe ConfigStore for concurrency tests.
// (config.MockConfig re-points its delegate's maps on every call via sync(),
// which is a data race under concurrent callers — fine for the serial tests
// it serves, unusable here.)
type raceSafeStore struct {
	mu    sync.Mutex
	env   *config.Environment
	tok   *config.TokenEntry
	saves int
}

func (s *raceSafeStore) Token(string) (*config.TokenEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tok, nil
}

func (s *raceSafeStore) SetToken(_ string, tok *config.TokenEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tok = tok
	return nil
}

func (s *raceSafeStore) Environment(string) (*config.Environment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.env, nil
}

func (s *raceSafeStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves++
	return nil
}

// TestTokenContext_ConcurrentSingleFlight pins the design's core guarantee:
// N concurrent callers with an expired cache produce exactly ONE OAuth POST
// (the per-env lock serializes; the post-lock cache re-check makes waiters
// adopt the winner's token instead of refreshing again). Deleting the
// re-check in TokenContext makes this fail with N POSTs.
func TestTokenContext_ConcurrentSingleFlight(t *testing.T) {
	var posts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&posts, 1)
		time.Sleep(20 * time.Millisecond) // widen the stampede window
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "single-flight-token",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	store := &raceSafeStore{
		env: &config.Environment{BaseURL: srv.URL},
		tok: &config.TokenEntry{AccessToken: "stale", ExpiresAt: time.Now().Add(-time.Hour)},
	}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("sf-env", "test-id", "test-secret"))
	ts := &TokenSource{Config: store, Creds: creds}

	const n = 10
	var wg sync.WaitGroup
	tokens := make([]string, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tokens[i], errs[i] = ts.TokenContext(context.Background(), "sf-env")
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		require.NoError(t, errs[i], "caller %d", i)
		assert.Equal(t, "single-flight-token", tokens[i], "caller %d must observe the winner's token", i)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(&posts),
		"concurrent callers must be deduplicated into a single OAuth POST")
	store.mu.Lock()
	defer store.mu.Unlock()
	assert.Equal(t, 1, store.saves, "exactly one refresh persists")
}

// TestForceRefreshContext_SerializesConcurrentRefreshes pins the OTHER side
// of the contract: forced refreshes are deliberately NOT deduplicated (each
// caller demanded a fresh token) but must still serialize on the per-env
// lock so they cannot stampede the OAuth endpoint in parallel.
func TestForceRefreshContext_SerializesConcurrentRefreshes(t *testing.T) {
	var posts, inFlight, maxInFlight int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			seen := atomic.LoadInt32(&maxInFlight)
			if cur <= seen || atomic.CompareAndSwapInt32(&maxInFlight, seen, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&posts, 1)
		atomic.AddInt32(&inFlight, -1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "forced-token",
			"expires_in":   3600,
		})
	}))
	defer srv.Close()

	store := &raceSafeStore{env: &config.Environment{BaseURL: srv.URL}}
	creds := NewMockCredentialStore()
	require.NoError(t, creds.Set("ff-env", "test-id", "test-secret"))
	ts := &TokenSource{Config: store, Creds: creds}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tok, err := ts.ForceRefreshContext(context.Background(), "ff-env")
			assert.NoError(t, err)
			assert.Equal(t, "forced-token", tok)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(2), atomic.LoadInt32(&posts), "forced refreshes are not deduplicated")
	assert.Equal(t, int32(1), atomic.LoadInt32(&maxInFlight), "but they must never hit the OAuth endpoint concurrently")
}
