package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRoot is retained for api_extra_test.go and api_coverage_test.go which
// still reference it. Migrate those files to retire this helper.
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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdAPI(f) }

func TestAPI_GET(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts", map[string]interface{}{
		"accounts": []map[string]string{{"name": "Test"}},
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/accounts")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Test")
}

// A non-JSON 2xx body (e.g. a text/CSV download or a proxy's HTML page) must be
// passed through to stdout and exit 0 — not dropped/errored — on the raw `api`
// escape hatch's default (no --jq/--template) path.
func TestAPI_NonJSONBody_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("plain text body, not json"))
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/export")
	require.NoError(t, err, "non-JSON 2xx must not error on the default api path")
	assert.Contains(t, stdout, "plain text body, not json")
}

func TestAPI_POST_WithBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"123"}`))
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "-X", "POST", "/v1/orders", "--body", `{"name":"test"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "123")
}

func TestAPI_POST_WithFileBody(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "body.json")
	require.NoError(t, os.WriteFile(tmpFile, []byte(`{"name":"from-file"}`), 0600))

	handler := cmdtest.OK(t, "", "", map[string]interface{}{"ok": true})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "-X", "POST", "/v1/orders", "--body", "@"+tmpFile)
	require.NoError(t, err)
	assert.Contains(t, stdout, "ok")
}

func TestAPI_CustomHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	})

	_, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/test", "-H", "X-Custom:custom-value")
	require.NoError(t, err)
}

func TestAPI_JQFilter(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"data": []map[string]string{
			{"name": "Account1"},
			{"name": "Account2"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/test", "--jq", ".data[].name")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Account1")
	assert.Contains(t, stdout, "Account2")
}

func TestAPI_ErrorResponse(t *testing.T) {
	handler := cmdtest.Status(t, "", "/v1/bad", 400, map[string]interface{}{
		"success": false,
		"reasons": []map[string]interface{}{
			{"code": "INVALID", "message": "bad request"},
		},
	})

	_, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/bad")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Zuora API error")
}

// Raw GET passthrough must deliver success:false bodies uninterpreted —
// scripts read raw envelopes; only mutating methods get the default check.
func TestAPIGet_SuccessFalseBodyPassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":false,"reasons":[{"code":1,"message":"logical failure"}]}`))
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/raw-read")
	require.NoError(t, err, "GET passthrough must not interpret the envelope")
	assert.Contains(t, stdout, `"success": false`)
}

func TestAPIPost_SuccessFalseBodyErrors(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"success":false,"reasons":[{"code":1,"message":"write failed"}]}`))
	})

	_, _, err := cmdtest.Run(t, "", newCmd, handler, "api", "/v1/raw-write", "-X", "POST", "--body", "{}")
	require.Error(t, err, "mutating methods keep the success check")
	assert.Contains(t, err.Error(), "write failed")
}
