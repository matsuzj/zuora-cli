package changelogversion

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdChangelogVersion(f) }

func TestChangelogVersion_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscription-change-logs/S-00000001/versions/2", map[string]interface{}{
		"success": true,
		"changeLog": map[string]interface{}{
			"type":    "Update",
			"version": 2,
		},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "changelog-version", "S-00000001", "2")
	require.NoError(t, err)
	assert.Contains(t, stdout, "changeLog")
	// Entry values, not just the envelope key the test injected. (#483)
	assert.Contains(t, stdout, `"type": "Update"`)
	assert.Contains(t, stdout, `"version": 2`)
}

func TestChangelogVersion_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "changelog-version", "S-00000001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 1")
}
