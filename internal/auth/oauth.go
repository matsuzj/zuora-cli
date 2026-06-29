package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
)

// refreshLocks serializes token refreshes per environment so concurrent callers
// do not each POST to the OAuth endpoint; the second caller observes the token
// the first one cached.
var refreshLocks sync.Map // envName -> *sync.Mutex

// ConfigStore is the slice of the config surface the token source actually
// needs — cached token reads/writes, environment lookup, persistence —
// declared consumer-side so auth does not depend on the full config.Config
// interface. config.Config satisfies it.
type ConfigStore interface {
	Token(envName string) (*config.TokenEntry, error)
	SetToken(envName string, token *config.TokenEntry) error
	Environment(name string) (*config.Environment, error)
	Save() error
}

// TokenSource manages OAuth 2.0 token acquisition and caching.
type TokenSource struct {
	Config     ConfigStore
	Creds      CredentialStore
	HTTPClient *http.Client
	// Logf, when non-nil, receives verbose diagnostic lines (P6-2). It must
	// never be called with secret material — only event names, environment
	// names, and non-sensitive metadata.
	Logf func(format string, args ...any)
}

// logf is the nil-guarded Logf entry point used at the observability sites.
func (ts *TokenSource) logf(format string, args ...any) {
	if ts.Logf != nil {
		ts.Logf(format, args...)
	}
}

// Token returns a valid access token for the given environment.
// If a cached token is still valid, it is returned. Otherwise, a new token is fetched.
func (ts *TokenSource) Token(envName string) (string, error) {
	return ts.TokenContext(context.Background(), envName)
}

// TokenContext is Token with a context so a token fetch can be cancelled
// (e.g. Ctrl-C) before the actual API request begins.
func (ts *TokenSource) TokenContext(ctx context.Context, envName string) (string, error) {
	cached, err := ts.Config.Token(envName)
	if err != nil {
		return "", err
	}
	if cached.IsValid() {
		ts.logf("* auth: cache hit for environment %q\n", envName)
		return cached.AccessToken, nil
	}

	// Serialize refreshes per environment to avoid duplicate token requests.
	defer lockEnv(envName)()

	// Re-check: another goroutine may have refreshed while we waited.
	if cached, err := ts.Config.Token(envName); err == nil && cached.IsValid() {
		ts.logf("* auth: cache hit (post-lock) for environment %q\n", envName)
		return cached.AccessToken, nil
	}
	return ts.refresh(ctx, envName)
}

// ForceRefreshContext fetches a new token unconditionally (bypassing the
// cache) while still serializing per environment via the same single-flight
// lock as TokenContext, so a forced refresh (e.g. after a 401) cannot
// stampede the OAuth endpoint alongside concurrent callers.
func (ts *TokenSource) ForceRefreshContext(ctx context.Context, envName string) (string, error) {
	ts.logf("* auth: force-refreshing token for environment %q\n", envName)
	defer lockEnv(envName)()
	return ts.refresh(ctx, envName)
}

// lockEnv takes the per-environment single-flight lock and returns the
// unlock; callers defer the returned func immediately.
func lockEnv(envName string) func() {
	muAny, _ := refreshLocks.LoadOrStore(envName, &sync.Mutex{})
	mu := muAny.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// refresh fetches a new token from the OAuth endpoint using the given context.
// insecureCleartextHost reports whether rawURL would transmit OAuth credentials
// in cleartext: an http:// scheme to a non-loopback host. http:// to localhost /
// 127.0.0.1 / ::1 is treated as safe (local dev/proxy, not network-exposed); any
// https:// URL or unparseable input is treated as safe here (ValidateBaseURL has
// already rejected non-http/https schemes upstream). Returns the host for the
// error message.
func insecureCleartextHost(rawURL string) (host string, insecure bool) {
	u, err := url.Parse(rawURL)
	if err != nil || !strings.EqualFold(u.Scheme, "http") {
		return "", false
	}
	h := u.Hostname()
	if strings.EqualFold(h, "localhost") {
		return h, false
	}
	if ip := net.ParseIP(h); ip != nil && ip.IsLoopback() {
		return h, false
	}
	return h, true
}

func (ts *TokenSource) refresh(ctx context.Context, envName string) (string, error) {
	clientID, clientSecret, err := ts.Creds.Get(envName)
	if err != nil {
		return "", err
	}
	// Log the credential SOURCE only — never any credential value. The
	// three-way split keeps `auth login --verbose` honest: its credentials
	// come from flags/prompt (a StaticCredentialStore), not the keyring.
	credSource := "the OS keyring"
	switch ts.Creds.(type) {
	case *envVarStore:
		credSource = "the ZR_CLIENT_ID/ZR_CLIENT_SECRET env vars"
	case *StaticCredentialStore:
		credSource = "explicitly provided values"
	}
	ts.logf("* auth: fetching token for environment %q (credentials from %s)\n", envName, credSource)

	env, err := ts.Config.Environment(envName)
	if err != nil {
		return "", err
	}
	if err := config.ValidateBaseURL(env.BaseURL); err != nil {
		return "", &AuthError{
			Message: fmt.Sprintf("environment %q has an invalid base URL: %v", envName, err),
			Hint:    "Check your environment configuration.",
		}
	}
	// Refuse to POST the client_secret over plaintext HTTP to a non-loopback
	// host — it would travel in cleartext on the wire. http:// to a loopback
	// host (localhost / 127.0.0.1 / ::1) is permitted for local development and
	// proxies, where there is no network exposure. (#439)
	if host, insecure := insecureCleartextHost(env.BaseURL); insecure {
		return "", &AuthError{
			Message: fmt.Sprintf("refusing to send OAuth credentials to %q over plaintext HTTP (client_secret would be exposed)", host),
			Hint:    "Use an https:// base URL. Plain http:// is allowed only for loopback hosts (localhost).",
		}
	}

	tokenURL := env.BaseURL + "/oauth/token"
	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	httpClient := ts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	// A correct OAuth token endpoint never redirects. Refuse to follow any
	// redirect (return the 3xx as-is) so the client_secret in the request body
	// can never be forwarded to a different (attacker) host. This applies to an
	// injected ts.HTTPClient too (unless it set its own policy); copy it first so
	// the caller's shared client is not mutated.
	if httpClient.CheckRedirect == nil {
		cp := *httpClient
		cp.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
		httpClient = &cp
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("connecting to %s: %w", tokenURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Truncate response body to avoid leaking credentials echoed by OAuth servers
		errBody := string(respBody)
		if len(errBody) > 200 {
			errBody = errBody[:200] + "..."
		}
		hint := "Check your Client ID and Client Secret."
		if resp.StatusCode >= 500 {
			hint = "The OAuth server returned a server error; this is likely transient. Try again shortly."
		}
		return "", &AuthError{
			Message:    fmt.Sprintf("authentication failed (HTTP %d): %s", resp.StatusCode, errBody),
			Hint:       hint,
			StatusCode: resp.StatusCode,
		}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", &AuthError{
			Message: "token response missing access_token",
			Hint:    "The OAuth endpoint returned an unexpected response.",
		}
	}

	entry := &config.TokenEntry{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}
	if err := ts.Config.SetToken(envName, entry); err != nil {
		return "", err
	}
	if err := ts.Config.Save(); err != nil {
		return "", err
	}
	ts.logf("* auth: token acquired, expires in %ds\n", tokenResp.ExpiresIn)

	return tokenResp.AccessToken, nil
}
