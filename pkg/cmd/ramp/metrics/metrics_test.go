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

func TestRampMetrics_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "metrics")
	assert.Error(t, err)
}

func TestRampMetrics_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Ramp not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "metrics", "R-INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ramp not found")
}
