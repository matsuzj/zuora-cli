package run

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRun(f) }

func TestMeterRun_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/meters/run/meter123/1", map[string]interface{}{
		"success": true,
		"message": "Meter run started",
	})

	stdout, stderr, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "run", "meter123", "1")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Meter run started")
	assert.Contains(t, stderr, "Meter run started.")
}

func TestMeterRun_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "run", "meter123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 1")
}

func TestMeterRun_NoArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "run")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 0")
}
