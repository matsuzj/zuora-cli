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
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/SUB-001", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "update", "SUB-001", "--body", `{"notes":"updated"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Subscription SUB-001 updated.")
}

func TestSubscriptionUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update", "SUB-001")
	assert.Error(t, err)
}

func TestSubscriptionUpdate_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update")
	assert.Error(t, err)
}
