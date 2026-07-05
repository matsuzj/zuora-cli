package getbysubscription

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGetBySubscription(f) }

func TestRampGetBySubscription_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S00000001/ramps", map[string]interface{}{
		"success": true,
		"ramps": []map[string]interface{}{
			{
				"rampNumber":         "R-00000001",
				"name":               "Test Ramp",
				"subscriptionNumber": "A-S00000001",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "get-by-subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "R-00000001")
	assert.Contains(t, stdout, "Test Ramp")
}

func TestRampGetBySubscription_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "get-by-subscription")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestRampGetBySubscription_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Subscription not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "get-by-subscription", "A-INVALID")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}
