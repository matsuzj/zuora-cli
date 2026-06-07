package get

import (
	"encoding/json"
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
	ramp := &cobra.Command{Use: "ramp"}
	ramp.AddCommand(NewCmdGet(f))
	root.AddCommand(ramp)
	return root
}

func TestRampGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/ramps/R-00000001", r.URL.Path)
		w.WriteHeader(200)
		// Real shape: the ramp is nested under a "ramp" object and its number
		// field is "number" (not "rampNumber").
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"ramp": map[string]interface{}{
				"number":             "R-00000001",
				"name":               "Test Ramp",
				"description":        "Ramp description",
				"subscriptionNumber": "A-S00000001",
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get", "R-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	outStr := out.String()
	assert.Contains(t, outStr, "R-00000001") // ramp.number (was read from flat "rampNumber")
	assert.Contains(t, outStr, "Test Ramp")
	assert.Contains(t, outStr, "A-S00000001") // proves the nested read works
}

func TestRampGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestRampGet_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000000, "message": "Ramp not found"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get", "R-INVALID"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ramp not found")
}
