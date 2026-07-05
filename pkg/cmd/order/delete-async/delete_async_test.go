package deleteasync

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDeleteAsync(f) }

func TestOrderDeleteAsync_Success(t *testing.T) {
	handler := cmdtest.OK(t, "DELETE", "/v1/async/orders/O-00000001", map[string]interface{}{
		"success": true,
		"jobId":   "job-12345",
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "delete-async", "O-00000001", "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stderr, "job-12345")
	assert.Contains(t, stderr, "Async order deletion started.")
}

func TestOrderDeleteAsync_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Order not found")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "delete-async", "O-00000002", "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
}

func TestOrderDeleteAsync_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "delete-async")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestOrderDeleteAsync_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "order", newCmd, "order", "delete-async", "O-00000001")
}
