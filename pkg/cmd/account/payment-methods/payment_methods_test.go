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
		// Zuora groups payment methods under a type-named key (e.g. "creditcard");
		// the card object's masked PAN is "cardNumber" (live-verified). (#421)
		"creditcard": []map[string]interface{}{
			{"id": "pm-1", "type": "CreditCard", "cardNumber": "************1234", "isDefault": true, "status": "Active"},
			{"id": "pm-2", "type": "ACH", "accountNumber": "****5678", "isDefault": false, "status": "Active"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CreditCard")
	// LAST4 is sourced from "cardNumber"; bites if production reverts to the
	// (absent) "creditCardMaskNumber" key — the card row's last4 goes blank.
	assert.Contains(t, stdout, "1234")
	// Row-bound (#483): every declared column's cell on its own row — a bare
	// Contains "true" matches a boolean anywhere in any rendering.
	assert.Regexp(t, `(?m)^│\s*pm-1\s*│\s*CreditCard\s*│\s*1234\s*│\s*true\s*│\s*Active\s*│`, stdout)
	assert.Regexp(t, `(?m)^│\s*pm-2\s*│\s*ACH\s*│\s*5678\s*│\s*false\s*│\s*Active\s*│`, stdout)
}

func TestPaymentMethods_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
