package auth

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
	"net/http"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
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
