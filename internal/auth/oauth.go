package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
func (ts *TokenSource) refresh(ctx context.Context, envName string) (string, error) {
	clientID, clientSecret, err := ts.Creds.Get(envName)
	if err != nil {
		return "", err
	}
	// Log the credential SOURCE only — never any credential value.
	credSource := "keyring"
	if _, ok := ts.Creds.(*envVarStore); ok {
		credSource = "env vars (ZR_CLIENT_ID/ZR_CLIENT_SECRET)"
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
