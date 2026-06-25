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

func TestFulfillmentGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/fulfillments/F-00000001", map[string]interface{}{
		"success": true,
		"fulfillment": map[string]interface{}{
			"id":                "8aca-ful-id",
			"fulfillmentNumber": "F-00000001",
			"state":             "Executed",
			"orderLineItemId":   "OLI-001",
			"quantity":          5,
			// Real shape: nested under a "fulfillment" object; the number is
			// "fulfillmentNumber" (no top-level "key") and the date is "fulfillmentDate".
			"fulfillmentDate": "2026-05-30",
		},
	})

	stdout, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "get", "F-00000001")

	require.NoError(t, err)
	// Label-bound (F-08): each value under its own label.
	assert.Regexp(t, `(?m)^Fulfillment Number:\s+F-00000001$`, stdout) // not the absent "key"
	assert.Regexp(t, `(?m)^State:\s+Executed$`, stdout)                // nested state
	assert.Regexp(t, `(?m)^Order Line Item ID:\s+OLI-001$`, stdout)    // nested orderLineItemId
	assert.Regexp(t, `(?m)^Date:\s+2026-05-30$`, stdout)               // fulfillmentDate, not absent "date"
}

func TestFulfillmentGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, nil, "fulfillment", "get")
	assert.Error(t, err)
}

func TestFulfillmentGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Fulfillment not found")

	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "get", "F-99999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Fulfillment not found")
}
