package items

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdItems(f) }

func TestInvoiceItems_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/inv-001/items", map[string]interface{}{
		"invoiceItems": []map[string]interface{}{
			{
				"id":               "item-001",
				"subscriptionName": "S-00001",
				"chargeAmount":     150.00,
				"chargeDate":       "2026-01-15",
				"chargeName":       "Monthly Fee",
			},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "items", "inv-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "S-00001")
	assert.Contains(t, stdout, "Monthly Fee")
	// Pin every declared column's cell (#483): ID/CHARGE_AMOUNT/CHARGE_DATE
	// were fixtured but unasserted — a struct-tag typo would render an empty
	// cell while the test stayed green.
	assert.Contains(t, stdout, "item-001")   // ID
	assert.Contains(t, stdout, "150.00")     // CHARGE_AMOUNT (%.2f)
	assert.Contains(t, stdout, "2026-01-15") // CHARGE_DATE
}

func TestInvoiceItems_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "items")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
