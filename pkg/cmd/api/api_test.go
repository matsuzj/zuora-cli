package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	root.PersistentFlags().Bool("csv", false, "Output as CSV")
	root.AddCommand(NewCmdAPI(f))
	return root
}

func TestAPI_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/accounts", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accounts": []map[string]string{{"name": "Test"}},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Test")
}

// A non-JSON 2xx body (e.g. a text/CSV download or a proxy's HTML page) must be
// passed through to stdout and exit 0 — not dropped/errored — on the raw `api`
// escape hatch's default (no --jq/--template) path.
func TestAPI_NonJSONBody_PassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("plain text body, not json"))
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/export"})
	err := root.Execute()

	require.NoError(t, err, "non-JSON 2xx must not error on the default api path")
	assert.Contains(t, out.String(), "plain text body, not json")
}

func TestAPI_POST_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/test", "-H", "X-Custom:custom-value"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestAPI_JQFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"name": "Account1"},
				{"name": "Account2"},
			},
		})
	}))
	defer server.Close()

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": "INVALID", "message": "bad request"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/bad"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zuora API error")
}

// Raw GET passthrough must deliver success:false bodies uninterpreted —
// scripts read raw envelopes; only mutating methods get the default check.
func TestAPIGet_SuccessFalseBodyPassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":false,"reasons":[{"code":1,"message":"logical failure"}]}`))
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/raw-read"})
	err := root.Execute()

	require.NoError(t, err, "GET passthrough must not interpret the envelope")
	assert.Contains(t, out.String(), `"success": false`)
}

func TestAPIPost_SuccessFalseBodyErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":false,"reasons":[{"code":1,"message":"write failed"}]}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/raw-write", "-X", "POST", "--body", "{}"})
	err := root.Execute()

	require.Error(t, err, "mutating methods keep the success check")
	assert.Contains(t, err.Error(), "write failed")
}
