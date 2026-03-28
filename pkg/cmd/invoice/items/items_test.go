package items

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
	invoice := &cobra.Command{Use: "invoice"}
	invoice.AddCommand(NewCmdItems(f))
	root.AddCommand(invoice)
	return root
}

func TestInvoiceItems_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001/items", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"invoiceItems": []map[string]interface{}{
				{
					"id":               "item-001",
					"subscriptionName": "S-00001",
					"chargeAmount":     150.00,
					"chargeDate":       "2026-01-15",
					"chargeName":       "Monthly Fee",
				},
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "items", "inv-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "S-00001")
	assert.Contains(t, out.String(), "Monthly Fee")
}

func TestInvoiceItems_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "items"})
	err := root.Execute()

	assert.Error(t, err)
}
