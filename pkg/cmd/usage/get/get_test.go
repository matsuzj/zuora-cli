package get

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
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
	usage := &cobra.Command{Use: "usage"}
	usage.AddCommand(NewCmdGet(f))
	root.AddCommand(usage)
	return root
}

func TestUsageGet_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/object/usage/2c92a0f96bd", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Id":            "2c92a0f96bd",
			"AccountId":     "acc123",
			"Quantity":      10,
			"StartDateTime": "2026-01-01",
			"UOM":           "Each",
			"Status":        "Processed",
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "get", "2c92a0f96bd"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "2c92a0f96bd")
	assert.Contains(t, out.String(), "Each")
}

func TestUsageGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
