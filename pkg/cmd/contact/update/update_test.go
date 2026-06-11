package update

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdate(f) }

func TestContactUpdate_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/contacts/c-123", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "update", "c-123", "--body", `{"firstName":"Jane"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Contact c-123 updated.")
}

func TestContactUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "update", "c-123")
	assert.Error(t, err)
}

func TestContactUpdate_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "update")
	assert.Error(t, err)
}
