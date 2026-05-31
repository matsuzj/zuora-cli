package update

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
	oli := &cobra.Command{Use: "order-line-item"}
	oli.AddCommand(NewCmdUpdate(f))
	root.AddCommand(oli)
	return root
}

func TestOrderLineItemUpdate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/order-line-items/OLI-001", r.URL.Path)

		raw, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, "Updated description", got["description"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      "OLI-001",
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order-line-item", "update", "OLI-001", "--body", `{"description":"Updated description"}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "Order line item OLI-001 updated.")
}

func TestOrderLineItemUpdate_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order-line-item", "update", "OLI-001"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestOrderLineItemUpdate_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order-line-item", "update", "--body", `{"description":"x"}`})
	err := root.Execute()

	assert.Error(t, err)
}
