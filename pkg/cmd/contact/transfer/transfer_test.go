package transfer

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdTransfer(f) }

func TestContactTransfer_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/contacts/c-123/transfer", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "transfer", "c-123", "--body", `{"destinationAccountId":"a-2"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Contact c-123 transferred.")
}

func TestContactTransfer_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "transfer", "c-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
