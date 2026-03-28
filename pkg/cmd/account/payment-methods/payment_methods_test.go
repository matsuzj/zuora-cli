package paymentmethods

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
	acct := &cobra.Command{Use: "account"}
	acct.AddCommand(NewCmdPaymentMethods(f))
	root.AddCommand(acct)
	return root
}

func TestPaymentMethods_Table(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/accounts/A001/payment-methods", r.URL.Path)
		w.WriteHeader(200)
		// Zuora API uses "returnedPaymentMethodType" as envelope key
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"returnedPaymentMethodType": []map[string]interface{}{
				{"id": "pm-1", "type": "CreditCard", "creditCardMaskNumber": "************1234", "isDefault": true, "status": "Active"},
				{"id": "pm-2", "type": "ACH", "accountNumber": "****5678", "isDefault": false, "status": "Active"},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "payment-methods", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "CreditCard")
	assert.Contains(t, output, "1234")
	assert.Contains(t, output, "true")
}
