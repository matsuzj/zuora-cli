package paymentmethodsdefault

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
	acct := &cobra.Command{Use: "account"}
	acct.AddCommand(NewCmdPaymentMethodsDefault(f))
	root.AddCommand(acct)
	return root
}

func TestPaymentMethodsDefault_Detail(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/accounts/A001/payment-methods/default", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "pm-1", "type": "CreditCard", "creditCardMaskNumber": "****1234",
			"expirationMonth": "12", "expirationYear": "2027", "status": "Active",
			"success": true,
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "payment-methods-default", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "CreditCard")
	assert.Contains(t, output, "Active")
}
