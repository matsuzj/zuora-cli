package metricsbysubscription

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdMetricsBySubscription(f) }

func TestRampMetricsBySubscription_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S00000001/ramp-metrics", map[string]interface{}{
		"success": true,
		"rampMetrics": []map[string]interface{}{
			{
				"rampNumber":         "R-00000001",
				"subscriptionNumber": "A-S00000001",
				"totalGrossTcb":      1200.0,
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics-by-subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "R-00000001")
	assert.Contains(t, stdout, "A-S00000001")
}

func TestRampMetricsBySubscription_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics-by-subscription")
	assert.Error(t, err)
}

func TestRampMetricsBySubscription_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Subscription not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics-by-subscription", "A-INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}
