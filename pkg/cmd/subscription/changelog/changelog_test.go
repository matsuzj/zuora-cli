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
