package changelog

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdChangelog(f) }

func TestChangelog_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscription-change-logs/S-00000001", map[string]interface{}{
		"success": true,
		"changeLogs": []map[string]interface{}{
			{"type": "Create", "date": "2026-01-01"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "changelog", "S-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "changeLogs")
}

// TestChangelog_Version folds in the old `changelog-version` via --version.
func TestChangelog_Version(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscription-change-logs/S-00000001/versions/2", map[string]interface{}{
		"success":    true,
		"changeLogs": []map[string]interface{}{{"type": "Update", "version": 2}},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "changelog", "S-00000001", "--version", "2")
	require.NoError(t, err)
	assert.Contains(t, stdout, "changeLogs")
}

// TestChangelog_ByOrder folds in the old `changelog-by-order` via --order.
func TestChangelog_ByOrder(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscription-change-logs/orders/O-00000001", map[string]interface{}{
		"success":    true,
		"changeLogs": []map[string]interface{}{{"type": "Update", "orderNumber": "O-00000001"}},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "changelog", "--order", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "changeLogs")
}

func TestChangelog_RequiresSubscriptionOrOrder(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "changelog")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "a <subscription-number> argument or --order is required")
}

func TestChangelog_OrderMutuallyExclusive(t *testing.T) {
	// --order with a subscription-number arg must be rejected before any request
	// (nil handler asserts no HTTP call).
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "changelog", "S-00000001", "--order", "O-00000001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--order cannot be combined")

	_, _, err = cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "changelog", "--order", "O-1", "--version", "2")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--order cannot be combined")
}
