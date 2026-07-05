package metrics

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdMetrics(f) }

func TestRampMetrics_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/ramps/R-00000001/ramp-metrics", map[string]interface{}{
		"success":    true,
		"rampNumber": "R-00000001",
		"rampMetrics": []map[string]interface{}{
			{
				"name": "Total Contract Value",
				"tcb":  12000,
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics", "R-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "R-00000001")
	assert.Contains(t, stdout, "Total Contract Value")
}

// TestRampMetrics_ByOrder folds in the old `metrics-by-order` via --order.
func TestRampMetrics_ByOrder(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001/ramp-metrics", map[string]interface{}{
		"success":     true,
		"rampMetrics": []map[string]interface{}{{"name": "Order Fixture Metric", "tcb": 4321.5}},
	})
	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics", "--order", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "rampMetrics")
	// The entry's VALUES must survive the passthrough — the envelope key alone
	// matches any rendering that echoes the key the test itself injected. (#483)
	assert.Contains(t, stdout, `"name": "Order Fixture Metric"`)
	assert.Contains(t, stdout, `"tcb": 4321.5`)
}

// TestRampMetrics_BySubscription folds in the old `metrics-by-subscription`.
func TestRampMetrics_BySubscription(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S00000001/ramp-metrics", map[string]interface{}{
		"success":     true,
		"rampMetrics": []map[string]interface{}{{"name": "Subscription Fixture Metric", "tcb": 8765.25}},
	})
	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics", "--subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "rampMetrics")
	// Entry values, not just the envelope key the test injected. (#483)
	assert.Contains(t, stdout, `"name": "Subscription Fixture Metric"`)
	assert.Contains(t, stdout, `"tcb": 8765.25`)
}

func TestRampMetrics_RequiresSelector(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one of <ramp-number>, --order, or --subscription is required")
}

func TestRampMetrics_SelectorsMutuallyExclusive(t *testing.T) {
	// More than one selector is rejected before any request (nil handler).
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics", "R-1", "--order", "O-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specify only one of")

	_, _, err = cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics", "--order", "O-1", "--subscription", "A-S1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specify only one of")
}

func TestRampMetrics_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Ramp not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics", "R-INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ramp not found")
}
