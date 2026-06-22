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
	// REAL response shape (Zuora API reference): the item nests under a
	// "fulfillmentItem" object and the documented fields are
	// id/fulfillmentId/itemIdentifier/description. The old flat fixture
	// (top-level id/fulfillmentKey/quantity) masked the bug — every field
	// rendered empty live, and "fulfillmentKey"/"quantity" are not real fields.
	handler := cmdtest.OK(t, "GET", "/v1/fulfillment-items/item-001", map[string]interface{}{
		"success": true,
		"fulfillmentItem": map[string]interface{}{
			"id":             "item-001",
			"fulfillmentId":  "8aca-ful-id",
			"itemIdentifier": "EXT-12345",
			"description":    "Test Item",
		},
	})

	stdout, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "get", "item-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "item-001")    // id
	assert.Contains(t, stdout, "8aca-ful-id") // fulfillmentId (was the absent flat "fulfillmentKey")
	assert.Contains(t, stdout, "EXT-12345")   // itemIdentifier (previously not rendered)
	assert.Contains(t, stdout, "Test Item")   // nested description
}

func TestFulfillmentItemGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "get")
	assert.Error(t, err)
}

func TestFulfillmentItemGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Fulfillment item not found")

	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "get", "item-999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Fulfillment item not found")
}
