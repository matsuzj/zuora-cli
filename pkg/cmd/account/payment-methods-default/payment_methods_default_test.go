package paymentmethodsdefault

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPaymentMethodsDefault(f) }

func TestPaymentMethodsDefault_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001/payment-methods/default", map[string]interface{}{
		"id": "pm-1", "type": "CreditCard", "creditCardMaskNumber": "****1234",
		"expirationMonth": "12", "expirationYear": "2027", "status": "Active",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-default", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CreditCard")
	assert.Contains(t, stdout, "Active")
}

func TestPaymentMethodsDefault_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "No default payment method found for account")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-default", "A001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No default payment method found for account")
}
