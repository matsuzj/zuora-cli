package list

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestPaymentList_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/transactions/payments/accounts/A001", map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"id":            "pay-001",
				"paymentNumber": "P-00001",
				"effectiveDate": "2026-01-15",
				"amount":        200.00,
				"status":        "Processed",
			},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "list", "--account", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "P-00001")
	assert.Contains(t, stdout, "Processed")
}

func TestPaymentList_CSV(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"payments": []map[string]interface{}{
			{"id": "pay-001", "paymentNumber": "P-00001", "amount": 200.00, "status": "Processed"},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "list", "--account", "A001", "--csv")
	require.NoError(t, err)

	// CSV: a header row plus the data row, comma-separated.
	assert.Contains(t, stdout, "P-00001")
	assert.Contains(t, stdout, ",")
	assert.Contains(t, stdout, "Processed")
}

func TestPaymentList_RequiresAccountFlag(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
