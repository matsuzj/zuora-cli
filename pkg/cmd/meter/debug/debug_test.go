package debug

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDebug(f) }

func TestMeterDebug_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/meters/debug/meter123/1", map[string]interface{}{
		"success": true,
		"message": "Meter debug started",
	})

	stdout, stderr, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "debug", "meter123", "1")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Meter debug started")
	assert.Contains(t, stderr, "Meter debug started.")
}

func TestMeterDebug_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "debug", "meter123")
	assert.Error(t, err)
}

func TestMeterDebug_NoArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "debug")
	assert.Error(t, err)
}
