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

func TestFulfillmentItemGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/fulfillment-items/item-001", map[string]interface{}{
		"success":        true,
		"id":             "item-001",
		"fulfillmentKey": "F-00000001",
		"quantity":       5,
		"description":    "Test Item",
	})

	stdout, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "get", "item-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "item-001")
	assert.Contains(t, stdout, "F-00000001")
}

func TestFulfillmentItemGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "get")
	assert.Error(t, err)
}
