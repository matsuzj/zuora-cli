package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdAuth(f))
	return root
}

func TestAuthLogin_WithFlags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "new-token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}
	f := factory.NewTestFactory(ios, cfg, server.URL, "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "login", "--client-id", "test-id", "--client-secret", "test-secret"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Logged in to sandbox")

	token, err := cfg.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "new-token", token.AccessToken)
}

// TestAuthLogin_VerboseDoesNotLeakSecret pins the command-layer guarantee that
// `auth login --verbose` never echoes the client secret or the issued token to
// stderr (login.go wires ts.Logf → ErrOut). The unit-level oauth leak guard
// covers TokenSource.Logf in isolation; this covers the wiring end-to-end and
// would catch a future direct ErrOut write in runLogin that bypasses Logf. The
// root registers the global flags so --verbose resolves, as in production.
func TestAuthLogin_VerboseDoesNotLeakSecret(t *testing.T) {
	const secret = "super-secret-value"
	const issued = "issued-token-value"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": issued, "token_type": "bearer", "expires_in": 3600,
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}
	f := factory.NewTestFactory(ios, cfg, server.URL, "")

	root := &cobra.Command{Use: "zr"}
	globalflags.Register(root) // defines --verbose so login.go's GetCount resolves
	root.AddCommand(NewCmdAuth(f))
	root.SetArgs([]string{"auth", "login", "--verbose", "--client-id", "test-id", "--client-secret", secret})
	require.NoError(t, root.Execute())

	out := errOut.String()
	assert.Contains(t, out, "* auth:", "verbose output must actually be produced (else NotContains is vacuous)")
	assert.NotContains(t, out, secret, "client secret must never appear in verbose stderr")
	assert.NotContains(t, out, issued, "the issued token must never appear in verbose stderr")
}

func TestAuthLogout(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "old-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}))
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "logout"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Logged out of sandbox")

	token, _ := cfg.Token("sandbox")
	assert.Nil(t, token)
}

func TestAuthStatus_Valid(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}))
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "status"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Environment: sandbox")
	assert.Contains(t, output, "valid")
}

func TestAuthStatus_NoToken(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "status"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "not authenticated")
}

// TestAuthStatus_JSON pins the output-consistency fix (#453): `auth status
// --json` must emit structured JSON, not silently fall back to the plain-text
// key/value form. Removing the format-flag branch makes json.Unmarshal fail.
func TestAuthStatus_JSON(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}))
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	globalflags.Register(root)
	root.SetArgs([]string{"auth", "status", "--json"})
	require.NoError(t, root.Execute())

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got),
		"auth status --json must emit valid JSON, not plain text")
	assert.Equal(t, "sandbox", got["environment"])
	assert.Equal(t, "keyring", got["credentials"])
	tok, ok := got["token"].(map[string]interface{})
	require.True(t, ok, "token must be a nested object")
	assert.Equal(t, "valid", tok["status"])
}

func TestAuthToken(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	// Set a valid cached token so TokenSource returns it without needing credentials
	require.NoError(t, cfg.SetToken("sandbox", &config.TokenEntry{
		AccessToken: "cached-token-123",
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}))
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "token"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "cached-token-123\n", out.String())
}

// A hung OAuth endpoint must be interruptible with Ctrl-C. Before the
// context wiring, login/token went through context.Background() wrappers and
// could not be cancelled until the 30s HTTP timeout fired.
func TestAuthLogin_HungOAuthEndpointCancelsPromptly(t *testing.T) {
	// The handler blocks on a test-owned channel (NOT r.Context().Done():
	// with an unread POST body, net/http does not reliably cancel the
	// handler context on client disconnect, which would deadlock
	// server.Close()). LIFO defers: close(unblock) releases the handler
	// before server.Close() waits on it.
	unblock := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-unblock // hang until the test tears down
	}))
	defer server.Close()
	defer close(unblock)

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}
	f := factory.NewTestFactory(ios, cfg, server.URL, "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "login", "--client-id", "id", "--client-secret", "secret"})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := root.ExecuteContext(ctx)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Less(t, elapsed, 500*time.Millisecond, "Ctrl-C must interrupt a hung OAuth request promptly")
}

func TestAuthToken_HungOAuthEndpointCancelsPromptly(t *testing.T) {
	// See the comment in TestAuthLogin_HungOAuthEndpointCancelsPromptly for
	// why the handler blocks on a test-owned channel.
	unblock := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-unblock
	}))
	defer server.Close()
	defer close(unblock)

	t.Setenv("ZR_CLIENT_ID", "id")
	t.Setenv("ZR_CLIENT_SECRET", "secret")

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.Envs["sandbox"] = &config.Environment{BaseURL: server.URL}
	f := factory.NewTestFactory(ios, cfg, server.URL, "")

	root := newTestRoot(f)
	root.SetArgs([]string{"auth", "token"})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := root.ExecuteContext(ctx)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Less(t, elapsed, 500*time.Millisecond, "Ctrl-C must interrupt a hung OAuth request promptly")
}
