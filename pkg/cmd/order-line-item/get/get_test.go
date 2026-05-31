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
	oli := &cobra.Command{Use: "order-line-item"}
	oli.AddCommand(NewCmdGet(f))
	root.AddCommand(oli)
	return root
}

func TestOrderLineItemGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/order-line-items/OLI-001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"id":          "OLI-001",
			"itemName":    "Widget",
			"itemNumber":  "OLI-N-001",
			"orderNumber": "O-00000001",
			"quantity":    2,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order-line-item", "get", "OLI-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "OLI-001")
	assert.Contains(t, out.String(), "Widget")
}

func TestOrderLineItemGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order-line-item", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
