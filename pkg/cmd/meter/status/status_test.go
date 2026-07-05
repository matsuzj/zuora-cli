package status

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdStatus(f) }

func TestMeterStatus_Success(t *testing.T) {
	// Doc-verified mediation envelope (#486): {success, data:{runStatus,
	// runStatusDescription}} — the old flat meterId/version/status/runType/
	// startTime/endTime fixture encoded keys the API never returns.
	handler := cmdtest.OK(t, "GET", "/meters/meter123/1/runStatus", map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"runStatus":            13,
			"runStatusDescription": "CONSUME_COMPLETED",
		},
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "status", "meter123", "1")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Run Status:\s+13$`, stdout)
	assert.Regexp(t, `(?m)^Description:\s+CONSUME_COMPLETED$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestMeterStatus_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "status", "meter123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 1")
}

func TestMeterStatus_NoArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 0")
}
