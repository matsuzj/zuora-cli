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
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/SUB-001/renew", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "renew", "SUB-001", "--body", `{"collect":true}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Subscription SUB-001 renewed.")
}

func TestSubscriptionRenew_WithoutBody(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/SUB-001/renew", map[string]interface{}{
		"success": true,
	})

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "renew", "SUB-001")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Subscription SUB-001 renewed.")
}

func TestSubscriptionRenew_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "renew")
	assert.Error(t, err)
}
