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
	handler := cmdtest.OK(t, "GET", "/v1/order-line-items/OLI-001", map[string]interface{}{
		"success":     true,
		"id":          "OLI-001",
		"itemName":    "Widget",
		"itemNumber":  "OLI-N-001",
		"orderNumber": "O-00000001",
		"quantity":    2,
	})

	stdout, _, err := cmdtest.Run(t, "order-line-item", newCmd, handler, "order-line-item", "get", "OLI-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "OLI-001")
	assert.Contains(t, stdout, "Widget")
}

func TestOrderLineItemGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, nil, "order-line-item", "get")
	assert.Error(t, err)
}
