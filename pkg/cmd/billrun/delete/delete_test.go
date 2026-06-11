package delete

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDelete(f) }

func TestBillRunDelete_Success(t *testing.T) {
	handler := cmdtest.OK(t, "DELETE", "/v1/bill-runs/br-001", map[string]interface{}{"success": true})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "delete", "br-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
}

func TestBillRunDelete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "delete", "br-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestBillRunDelete_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730030, "Cannot delete a posted bill run")

	_, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "delete", "br-001", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete a posted bill run")
}
