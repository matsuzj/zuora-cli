package list

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestRampList_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S00000001/ramps", map[string]interface{}{
		"success": true,
		"ramps": []map[string]interface{}{
			{"rampNumber": "R-00000001", "name": "Test Ramp"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "list", "--subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "R-00000001")
	assert.Contains(t, stdout, "Test Ramp")
}

func TestRampList_RequiresSubscription(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "subscription" not set`)
}
