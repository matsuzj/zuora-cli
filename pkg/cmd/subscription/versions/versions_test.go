package versions

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdVersions(f) }

func TestSubscriptionVersions_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S001/versions/1", map[string]interface{}{
		"id": "sub-1", "subscriptionNumber": "A-S001", "version": 1,
		"name": "Gold Plan", "status": "Active", "termType": "TERMED",
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "versions", "A-S001", "1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Gold Plan")
	assert.Contains(t, stdout, "A-S001")
}

func TestSubscriptionVersions_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "versions", "A-S001")
	assert.Error(t, err)
}

func TestSubscriptionVersions_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Subscription version not found")

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "versions", "bad-key", "99")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription version not found")
}
