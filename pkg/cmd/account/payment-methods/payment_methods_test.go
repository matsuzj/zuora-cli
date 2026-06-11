package paymentmethods

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPaymentMethods(f) }

func TestPaymentMethods_Table(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001/payment-methods", map[string]interface{}{
		"success": true,
		// Zuora API uses "returnedPaymentMethodType" as envelope key
		"returnedPaymentMethodType": []map[string]interface{}{
			{"id": "pm-1", "type": "CreditCard", "creditCardMaskNumber": "************1234", "isDefault": true, "status": "Active"},
			{"id": "pm-2", "type": "ACH", "accountNumber": "****5678", "isDefault": false, "status": "Active"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CreditCard")
	assert.Contains(t, stdout, "1234")
	assert.Contains(t, stdout, "true")
}

func TestPaymentMethods_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
