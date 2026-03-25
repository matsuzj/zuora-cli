package summary

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	meter := &cobra.Command{Use: "meter"}
	meter.AddCommand(NewCmdSummary(f))
	root.AddCommand(meter)
	return root
}

func TestMeterSummary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/meters/meter123/summary", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "FULL", reqBody["runType"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"meterId": "meter123",
			"runType": "FULL",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "summary", "meter123", "--run-type", "FULL"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "meter123")
	assert.Contains(t, out.String(), "FULL")
}

func TestMeterSummary_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "FULL", reqBody["runType"])
		assert.Equal(t, "2026-01-01", reqBody["startDate"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "summary", "meter123", "--run-type", "FULL", "--body", `{"startDate":"2026-01-01"}`})
	err := root.Execute()

	require.NoError(t, err)
}

func TestMeterSummary_RequiresRunType(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "summary", "meter123"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--run-type is required")
}

func TestMeterSummary_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "summary", "--run-type", "FULL"})
	err := root.Execute()

	assert.Error(t, err)
}
