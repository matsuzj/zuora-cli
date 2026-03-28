package versions

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
	sub := &cobra.Command{Use: "subscription"}
	sub.AddCommand(NewCmdVersions(f))
	root.AddCommand(sub)
	return root
}

func TestSubscriptionVersions_Detail(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/A-S001/versions/1", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "sub-1", "subscriptionNumber": "A-S001", "version": 1,
			"name": "Gold Plan", "status": "Active", "termType": "TERMED",
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "versions", "A-S001", "1"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Gold Plan")
	assert.Contains(t, output, "A-S001")
}

func TestSubscriptionVersions_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "versions", "A-S001"})
	err := root.Execute()

	assert.Error(t, err)
}
