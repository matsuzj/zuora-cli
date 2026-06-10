package paymentmethodscascading

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
	acct := &cobra.Command{Use: "account"}
	acct.AddCommand(NewCmdPaymentMethodsCascading(f))
	root.AddCommand(acct)
	return root
}

func TestPaymentMethodsCascading_Detail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/accounts/A001/payment-methods/cascading", r.URL.Path)
		w.WriteHeader(200)
		// Cascading config response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":                       true,
			"paymentMethodId":               "pm-parent",
			"paymentMethodCascadingConsent": true,
			"paymentMethodType":             "CreditCard",
			"creditCardMaskNumber":          "****9999",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "payment-methods-cascading", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "pm-parent")
	assert.Contains(t, output, "CreditCard")
}

func TestPaymentMethodsCascading_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 50000040, "message": "Account not found"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "payment-methods-cascading", "bad-key"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
