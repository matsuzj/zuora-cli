package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestOrderLineItemGet_Success(t *testing.T) {
	// REAL response shape: the item nests under "orderLineItem" (the old flat
	// fixture masked the bug — all 8 fields rendered empty live).
	handler := cmdtest.OK(t, "GET", "/v1/order-line-items/OLI-001", map[string]interface{}{
		"success": true,
		"orderLineItem": map[string]interface{}{
			"id":          "OLI-001",
			"itemName":    "Widget",
			"itemNumber":  "OLI-N-001",
			"itemType":    "Product",
			"itemState":   "Executing",
			"orderNumber": "O-00000001",
			"amount":      2000000.0,
			"quantity":    2.0,
		},
	})

	stdout, _, err := cmdtest.Run(t, "order-line-item", newCmd, handler, "order-line-item", "get", "OLI-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^ID:\s+OLI-001$`, stdout)
	assert.Regexp(t, `(?m)^Item Name:\s+Widget$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+2000000\.00$`, stdout) // money, not 2e+06
	assert.Regexp(t, `(?m)^Item State:\s+Executing$`, stdout)
}

func TestOrderLineItemGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, nil, "order-line-item", "get")
	assert.Error(t, err)
}
