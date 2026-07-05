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
	// Expect.JSONBody pins that the --body payload reaches the server intact
	// (#484): a command that dropped or mangled the body would fail here.
	handler := cmdtest.Expect{
		Method:   "PUT",
		Path:     "/v1/contacts/c-123",
		JSONBody: `{"firstName":"Jane"}`,
		Respond:  map[string]interface{}{"success": true},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "update", "c-123", "--body", `{"firstName":"Jane"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the command's only detail row is Success, read from
	// the response "success" key — a bare Contains "true" would pass on any
	// stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Contact c-123 updated.")
}

func TestContactUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "update", "c-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestContactUpdate_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "update")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
