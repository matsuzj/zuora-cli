package activate

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
	order := &cobra.Command{Use: "order"}
	order.AddCommand(NewCmdActivate(f))
	root.AddCommand(order)
	return root
}

func TestOrderActivate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orders/O-00000001/activate", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"orderNumber": "O-00000001",
			"status":      "Completed",
		})
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "activate", "O-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "O-00000001")
	assert.Contains(t, errOut.String(), "Order O-00000001 activated.")
}
