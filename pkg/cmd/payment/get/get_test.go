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
	payment := &cobra.Command{Use: "payment"}
	payment.AddCommand(NewCmdGet(f))
	root.AddCommand(payment)
	return root
}

func TestPaymentGet_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/payments/pay-001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "pay-001",
			"paymentNumber": "P-00000001",
			"effectiveDate": "2026-01-15",
			"amount":        100.00,
			"status":        "Processed",
			"type":          "External",
			"accountId":     "acc-001",
			"gatewayState":  "Settled",
			"createdDate":   "2026-01-10T10:00:00Z",
			"success":       true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "get", "pay-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "P-00000001")
	assert.Contains(t, out.String(), "Processed")
}

func TestPaymentGet_JSON(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "pay-001",
			"paymentNumber": "P-00000001",
			"success":       true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "get", "pay-001", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"paymentNumber"`)
}

func TestPaymentGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "get"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestPaymentGet_SuccessFalse(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000040, "message": "Payment not found"},
			},
		})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "get", "bad-id"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")
}
