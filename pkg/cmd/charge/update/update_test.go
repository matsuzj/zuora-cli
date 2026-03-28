package update

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
	charge := &cobra.Command{Use: "charge"}
	charge.AddCommand(NewCmdUpdate(f))
	root.AddCommand(charge)
	return root
}

func TestChargeUpdate_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/commerce/charges", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "chg-001",
			"name": "Updated Charge",
		})
	}))

	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"charge", "update", "--body", `{"id":"chg-001","name":"Updated Charge"}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "chg-001")
	assert.Contains(t, out.String(), "Updated Charge")
	assert.Contains(t, errOut.String(), "Charge updated.")
}

func TestChargeUpdate_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"charge", "update"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
