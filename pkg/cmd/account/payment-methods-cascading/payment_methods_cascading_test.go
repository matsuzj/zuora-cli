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
	// Doc-verified shape (#416): {consent, priorities:[{paymentMethodId,
	// order}], success} — the old six flat keys do not exist in the response.
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001/payment-methods/cascading", map[string]interface{}{
		"success": true,
		"consent": true,
		"priorities": []interface{}{
			map[string]interface{}{"paymentMethodId": "pm-primary-42", "order": 1},
			map[string]interface{}{"paymentMethodId": "pm-backup-77", "order": 2},
		},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-cascading", "A001")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Consent:\s+true$`, stdout)
	assert.Regexp(t, `(?m)^Priority 1:\s+pm-primary-42$`, stdout)
	assert.Regexp(t, `(?m)^Priority 2:\s+pm-backup-77$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestPaymentMethodsCascading_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "payment-methods-cascading", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
