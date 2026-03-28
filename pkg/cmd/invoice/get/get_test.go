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
	invoice := &cobra.Command{Use: "invoice"}
	invoice.AddCommand(NewCmdGet(f))
	root.AddCommand(invoice)
	return root
}

func TestInvoiceGet_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "inv-001",
			"invoiceNumber": "INV00001",
			"invoiceDate":   "2026-01-15",
			"dueDate":       "2026-02-15",
			"amount":        100.50,
			"balance":       50.25,
			"status":        "Posted",
			"accountId":     "acc-001",
			"createdDate":   "2026-01-10T10:00:00Z",
			"success":       true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "get", "inv-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "INV00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestInvoiceGet_JSON(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "inv-001",
			"invoiceNumber": "INV00001",
			"success":       true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "get", "inv-001", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"invoiceNumber"`)
}

func TestInvoiceGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "get"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestInvoiceGet_SuccessFalse(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000040, "message": "Invoice not found"},
			},
		})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "get", "bad-id"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invoice not found")
}
