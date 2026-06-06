package list

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
	root.PersistentFlags().Bool("csv", false, "")
	payment := &cobra.Command{Use: "payment"}
	payment.AddCommand(NewCmdList(f))
	root.AddCommand(payment)
	return root
}

func TestPaymentList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/transactions/payments/accounts/A001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"payments": []map[string]interface{}{
				{
					"id":            "pay-001",
					"paymentNumber": "P-00001",
					"effectiveDate": "2026-01-15",
					"amount":        200.00,
					"status":        "Processed",
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
	root.SetArgs([]string{"payment", "list", "--account", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "P-00001")
	assert.Contains(t, out.String(), "Processed")
}

func TestPaymentList_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"payments": []map[string]interface{}{
				{"id": "pay-001", "paymentNumber": "P-00001", "amount": 200.00, "status": "Processed"},
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "list", "--account", "A001", "--csv"})
	require.NoError(t, root.Execute())

	output := out.String()
	// CSV: a header row plus the data row, comma-separated.
	assert.Contains(t, output, "P-00001")
	assert.Contains(t, output, ",")
	assert.Contains(t, output, "Processed")
}

func TestPaymentList_RequiresAccountFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "list"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
