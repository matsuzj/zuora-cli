package setcascading

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSetCascading(f) }

func TestSetCascading_Success(t *testing.T) {
	// JSONBody: the --body payload must reach the server intact. (#484)
	handler := cmdtest.Expect{
		Method:   "PUT",
		Path:     "/v1/accounts/A001/payment-methods/cascading",
		JSONBody: `{"paymentMethodId":"pm-1"}`,
		Respond:  map[string]interface{}{"success": true},
	}.Handler(t)

	_, stderr, err := cmdtest.Run(t, "account", newCmd, handler, "account", "set-cascading", "A001", "--body", `{"paymentMethodId":"pm-1"}`)

	require.NoError(t, err)
	assert.Contains(t, stderr, "Cascading payment methods updated")
}

func TestSetCascading_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "account", newCmd, nil, "account", "set-cascading", "A001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
