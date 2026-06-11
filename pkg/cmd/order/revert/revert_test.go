package revert

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRevert(f) }

func TestOrderRevert_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/orders/O-00000001/revert", map[string]interface{}{
		"success":     true,
		"orderNumber": "O-00000001",
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "revert", "O-00000001", "--body", `{"orderDate":"2026-01-01"}`, "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stderr, "Order O-00000001 reverted.")
}

func TestOrderRevert_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "revert", "O-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
