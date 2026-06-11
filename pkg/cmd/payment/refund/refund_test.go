package refund

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRefund(f) }

func TestPaymentRefund_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/payments/pay-001/refunds", map[string]interface{}{
		"id":           "ref-001",
		"refundNumber": "R-00001",
		"amount":       50.00,
		"status":       "Processed",
		"success":      true,
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "refund", "pay-001", "--body", `{"amount":50,"type":"External"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "R-00001")
	assert.Contains(t, stdout, "Processed")
}

func TestPaymentRefund_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "refund", "pay-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
