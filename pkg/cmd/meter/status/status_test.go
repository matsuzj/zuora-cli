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
	handler := cmdtest.OK(t, "GET", "/meters/meter123/1/runStatus", map[string]interface{}{
		"meterId":   "meter123",
		"version":   "1",
		"status":    "COMPLETED",
		"runType":   "FULL",
		"startTime": "2025-01-01T00:00:00Z",
		"endTime":   "2025-01-01T01:00:00Z",
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "status", "meter123", "1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "meter123")
	assert.Contains(t, stdout, "COMPLETED")
}

func TestMeterStatus_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "status", "meter123")
	assert.Error(t, err)
}

func TestMeterStatus_NoArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "status")
	assert.Error(t, err)
}
