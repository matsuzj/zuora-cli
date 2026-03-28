package list

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
	invoice.AddCommand(NewCmdList(f))
	root.AddCommand(invoice)
	return root
}

func TestInvoiceList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/transactions/invoices/accounts/A00000001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"invoices": []map[string]interface{}{
				{
					"id":            "inv-001",
					"invoiceNumber": "INV00001",
					"invoiceDate":   "2026-01-15",
					"amount":        100.50,
					"balance":       50.25,
					"status":        "Posted",
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
	root.SetArgs([]string{"invoice", "list", "--account", "A00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "INV00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestInvoiceList_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"invoices": []map[string]interface{}{
				{
					"id":            "inv-001",
					"invoiceNumber": "INV00001",
					"invoiceDate":   "2026-01-15",
					"amount":        100.50,
					"balance":       50.25,
					"status":        "Posted",
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
	root.SetArgs([]string{"invoice", "list", "--account", "A00000001", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"invoiceNumber"`)
}

func TestInvoiceList_RequiresAccount(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "list"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestInvoiceList_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000040, "message": "Account not found"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "list", "--account", "INVALID"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
