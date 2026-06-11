package metricsbyorder

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdMetricsByOrder(f) }

func TestRampMetricsByOrder_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001/ramp-metrics", map[string]interface{}{
		"success":     true,
		"orderNumber": "O-00000001",
		"rampMetrics": []map[string]interface{}{
			{"rampNumber": "R-00000001", "tcb": 1200.0, "tcv": 1000.0},
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics-by-order", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stdout, "R-00000001")
}

func TestRampMetricsByOrder_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics-by-order")
	assert.Error(t, err)
}

func TestRampMetricsByOrder_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Order not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics-by-order", "O-INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
}
