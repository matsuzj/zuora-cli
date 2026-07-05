package revert

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRevert(f) }

func TestOrderRevert_Success(t *testing.T) {
	// Expect.JSONBody pins that the --body payload reaches the server intact
	// (#484): a command that dropped or mangled the body would fail here.
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/v1/orders/O-00000001/revert",
		JSONBody: `{"orderDate":"2026-01-01"}`,
		Respond: map[string]interface{}{
			"success":     true,
			"orderNumber": "O-00000001",
		},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "revert", "O-00000001", "--body", `{"orderDate":"2026-01-01"}`, "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	// Label-bound (F-08, #483): values under their own labels — a response-key
	// typo would render "" while a bare Contains stayed green.
	assert.Regexp(t, `(?m)^Order Number:\s+O-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Order O-00000001 reverted.")
}

func TestOrderRevert_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "revert", "O-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
