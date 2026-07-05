package balance

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdBalance(f) }

func TestCommitmentBalance_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/commitments/CMT-00000001/balancepreview", map[string]interface{}{
		"success":         true,
		"commitmentKey":   "CMT-00000001",
		"totalAmount":     1000.0,
		"consumedAmount":  250.0,
		"remainingAmount": 750.0,
		"currency":        "USD",
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "balance", "CMT-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CMT-00000001")
	assert.Contains(t, stdout, "remainingAmount")
}

func TestCommitmentBalance_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "balance")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
