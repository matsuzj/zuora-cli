package cancel

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCancel(f) }

func TestOrderCancel_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/orders/O-00000001/cancel", map[string]interface{}{
		"success":     true,
		"orderNumber": "O-00000001",
		"status":      "Cancelled",
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "cancel", "O-00000001", "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	// Label-bound (F-08, #483): each value under its own label — a response-key
	// typo would render "" while a bare Contains stayed green.
	assert.Regexp(t, `(?m)^Order Number:\s+O-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Cancelled$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Order O-00000001 cancelled.")
}

func TestOrderCancel_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "order", newCmd, "order", "cancel", "O-00000001")
}
