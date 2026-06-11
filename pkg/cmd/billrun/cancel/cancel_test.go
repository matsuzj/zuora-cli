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

func TestBillRunCancel_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/cancel", map[string]interface{}{
		"id":      "br-001",
		"status":  "Cancelled",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "cancel", "br-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Cancelled")
}

func TestBillRunCancel_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "cancel", "br-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestBillRunCancel_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "cancel")
	assert.Error(t, err)
}
