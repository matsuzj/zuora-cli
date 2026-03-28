package refund

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
	payment.AddCommand(NewCmdRefund(f))
	root.AddCommand(payment)
	return root
}

func TestPaymentRefund_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/payments/pay-001/refunds", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           "ref-001",
			"refundNumber": "R-00001",
			"amount":       50.00,
			"status":       "Processed",
			"success":      true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "refund", "pay-001", "--body", `{"amount":50,"type":"External"}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "R-00001")
	assert.Contains(t, out.String(), "Processed")
}

func TestPaymentRefund_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"payment", "refund", "pay-001"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
