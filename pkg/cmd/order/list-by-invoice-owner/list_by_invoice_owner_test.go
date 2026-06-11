package listbyinvoiceowner

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdListByInvoiceOwner(f) }

func TestOrderListByInvoiceOwner_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/invoiceOwner/A00000001", map[string]interface{}{
		"success": true,
		"orders": []map[string]interface{}{
			{
				"orderNumber":           "O-00000001",
				"status":                "Completed",
				"orderDate":             "2024-01-01",
				"existingAccountNumber": "A00000001",
				"createdDate":           "2024-01-02",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-invoice-owner", "A00000001")

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stdout, "Completed")
}

func TestOrderListByInvoiceOwner_WithPaging(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders/invoiceOwner/A00000001", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "50", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"orders": []map[string]interface{}{
				{
					"orderNumber":           "O-00000002",
					"status":                "Pending",
					"orderDate":             "2024-02-01",
					"existingAccountNumber": "A00000001",
					"createdDate":           "2024-02-02",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-invoice-owner", "A00000001", "--page", "2", "--page-size", "50")

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000002")
}

func TestOrderListByInvoiceOwner_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "list-by-invoice-owner")

	assert.Error(t, err)
}

func TestOrderListByInvoiceOwner_NextPageHint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"orders":   []map[string]interface{}{{"orderNumber": "O-00000001"}},
			"nextPage": "https://rest.example.com/v1/orders/invoiceOwner/A00000001?page=2",
		})
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-invoice-owner", "A00000001")

	require.NoError(t, err)
	assert.Contains(t, stderr, "More results available")
}
