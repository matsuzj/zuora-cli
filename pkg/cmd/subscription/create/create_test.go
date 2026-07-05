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

func TestSubscriptionCreate_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/subscriptions", map[string]interface{}{
		"success":            true,
		"subscriptionId":     "sub-123",
		"subscriptionNumber": "A-S001",
	})

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "create", "--body", `{"accountKey":"A001"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "A-S001")
	assert.Contains(t, stderr, "Subscription A-S001 created.")
}

func TestSubscriptionCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "create")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
