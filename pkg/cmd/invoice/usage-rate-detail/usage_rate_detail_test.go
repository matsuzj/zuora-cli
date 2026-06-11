package usageratedetail

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUsageRateDetail(f) }

func TestInvoiceUsageRateDetail_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/invoice-item/item-001/usage-rate-detail", map[string]interface{}{
		"success": true,
		"usageData": []map[string]interface{}{
			{"unitOfMeasure": "GB", "quantity": 100},
		},
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "usage-rate-detail", "item-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "usageData")
}

func TestInvoiceUsageRateDetail_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "usage-rate-detail")
	assert.Error(t, err)
}
