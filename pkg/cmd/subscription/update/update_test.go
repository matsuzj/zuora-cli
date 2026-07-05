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

func TestSubscriptionUpdate_Success(t *testing.T) {
	// JSONBody: the --body payload must reach the server intact. (#484)
	handler := cmdtest.Expect{
		Method:   "PUT",
		Path:     "/v1/subscriptions/SUB-001",
		JSONBody: `{"notes":"updated"}`,
		Respond:  map[string]interface{}{"success": true},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "update", "SUB-001", "--body", `{"notes":"updated"}`)
	require.NoError(t, err)
	// Label-bound (#483): bare Contains "true" matches the success flag anywhere.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Subscription SUB-001 updated.")
}

func TestSubscriptionUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update", "SUB-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestSubscriptionUpdate_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
