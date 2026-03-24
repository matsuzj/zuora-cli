package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
)

// TokenSource manages OAuth 2.0 token acquisition and caching.
type TokenSource struct {
	Config     config.Config
	Creds      CredentialStore
	HTTPClient *http.Client
}

// Token returns a valid access token for the given environment.
// If a cached token is still valid, it is returned. Otherwise, a new token is fetched.
func (ts *TokenSource) Token(envName string) (string, error) {
	cached, err := ts.Config.Token(envName)
	if err != nil {
		return "", err
	}
	if cached.IsValid() {
		return cached.AccessToken, nil
	}
	return ts.Refresh(envName)
}

// Refresh fetches a new token from the OAuth endpoint.
func (ts *TokenSource) Refresh(envName string) (string, error) {
	clientID, clientSecret, err := ts.Creds.Get(envName)
	if err != nil {
		return "", err
	}

	env, err := ts.Config.Environment(envName)
	if err != nil {
		return "", err
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

	resp, err := httpClient.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
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
		return "", &AuthError{
			Message: fmt.Sprintf("authentication failed (HTTP %d): %s", resp.StatusCode, errBody),
			Hint:    "Check your Client ID and Client Secret.",
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

	return tokenResp.AccessToken, nil
}
