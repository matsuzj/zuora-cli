package paymentmethodscascading

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPaymentMethodsCascading(f) }

func TestPaymentMethodsCascading_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001/payment-methods/cascading", map[string]interface{}{
		"success":                       true,
		"paymentMethodId":               "pm-parent",
		"paymentMethodCascadingConsent": true,
		"paymentMethodType":             "CreditCard",
		"creditCardMaskNumber":          "****9999",
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-cascading", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "pm-parent")
	assert.Contains(t, stdout, "CreditCard")
}

func TestPaymentMethodsCascading_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-cascading", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
