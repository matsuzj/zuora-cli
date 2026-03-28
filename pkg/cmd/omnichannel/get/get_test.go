package get

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
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
	omni := &cobra.Command{Use: "omnichannel"}
	omni.AddCommand(NewCmdGet(f))
	root.AddCommand(omni)
	return root
}

func TestOmnichannelGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/omni-channel-subscriptions/S-001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"subscriptionKey": "S-001",
			"status":          "Active",
			"channel":         "Web",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"omnichannel", "get", "S-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "S-001")
}

func TestOmnichannelGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"omnichannel", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
