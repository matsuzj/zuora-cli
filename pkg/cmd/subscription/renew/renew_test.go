package renew

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRenew(f) }

func TestSubscriptionRenew_Success(t *testing.T) {
	// JSONBody: the --body payload must reach the server intact. (#484)
	handler := cmdtest.Expect{
		Method:   "PUT",
		Path:     "/v1/subscriptions/SUB-001/renew",
		JSONBody: `{"collect":true}`,
		Respond:  map[string]interface{}{"success": true},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "renew", "SUB-001", "--body", `{"collect":true}`)
	require.NoError(t, err)
	// Label-bound: bare Contains "true" matches the success flag anywhere. (#483)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Subscription SUB-001 renewed.")
}

func TestSubscriptionRenew_WithoutBody(t *testing.T) {
	// Without --body the command must still send an explicit "{}" payload.
	handler := cmdtest.Expect{
		Method:   "PUT",
		Path:     "/v1/subscriptions/SUB-001/renew",
		JSONBody: `{}`,
		Respond:  map[string]interface{}{"success": true},
	}.Handler(t)

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "renew", "SUB-001")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Subscription SUB-001 renewed.")
}

func TestSubscriptionRenew_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "renew")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
