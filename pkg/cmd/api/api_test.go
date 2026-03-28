package api

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	// Register global persistent flags that api command reads
	root.PersistentFlags().String("jq", "", "Filter JSON output")
	root.PersistentFlags().Bool("json", false, "Output as JSON")
	root.PersistentFlags().String("template", "", "Format output")
	root.AddCommand(NewCmdAPI(f))
	return root
}

func TestAPI_GET(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/accounts", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accounts": []map[string]string{{"name": "Test"}},
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Test")
}

func TestAPI_POST_WithBody(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"123"}`))
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "-X", "POST", "/v1/orders", "--body", `{"name":"test"}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "123")
}

func TestAPI_POST_WithFileBody(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))

	tmpFile := filepath.Join(t.TempDir(), "body.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"name":"from-file"}`), 0600))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "-X", "POST", "/v1/orders", "--body", "@" + tmpFile})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "ok")
}

func TestAPI_CustomHeaders(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/test", "-H", "X-Custom:custom-value"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestAPI_JQFilter(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"name": "Account1"},
				{"name": "Account2"},
			},
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/test", "--jq", ".data[].name"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Account1")
	assert.Contains(t, output, "Account2")
}

func TestAPI_ErrorResponse(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": "INVALID", "message": "bad request"},
			},
		})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/bad"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zuora API error")
}
