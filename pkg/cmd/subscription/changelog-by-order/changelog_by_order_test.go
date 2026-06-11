package changelogbyorder

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdChangelogByOrder(f) }

func TestChangelogByOrder_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscription-change-logs/orders/O-00000001", map[string]interface{}{
		"success": true,
		"changeLogs": []map[string]interface{}{
			{"type": "Update", "orderNumber": "O-00000001"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "changelog-by-order", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "changeLogs")
}
