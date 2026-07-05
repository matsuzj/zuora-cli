package create

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestContactCreate_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/contacts", map[string]interface{}{
		"success": true,
		"id":      "c-new",
	})

	stdout, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "create", "--body", `{"accountId":"a-1","firstName":"John","lastName":"Doe","country":"US"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "c-new")
	assert.Contains(t, stderr, "Contact c-new created.")
}

func TestContactCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "create")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
