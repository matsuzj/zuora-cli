package scrub

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdScrub(f) }

func TestContactScrub_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/contacts/c-123/scrub", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "scrub", "c-123", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Contact c-123 scrubbed.")
}

func TestContactScrub_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "scrub")
	assert.Error(t, err)
}
