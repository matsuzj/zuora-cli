package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestOmnichannelGet_Success(t *testing.T) {
	// Doc-verified flat shape (#414): the old fixture's subscriptionKey/status/
	// channel keys do not exist in the real response.
	handler := cmdtest.OK(t, "GET", "/v1/omni-channel-subscriptions/S-001", map[string]interface{}{
		"success":                true,
		"id":                     "omni-9001",
		"subscriptionNumber":     "OCS-00042",
		"state":                  "Active",
		"externalState":          "ACTIVE",
		"externalSourceSystem":   "AppleAppStore",
		"externalSubscriptionId": "ext-sub-777",
		"autoRenew":              true,
		"currency":               "JPY",
	})

	stdout, _, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "get", "S-001")
	require.NoError(t, err)
	// Label-bound (F-08): each value under its own label.
	assert.Regexp(t, `(?m)^ID:\s+omni-9001$`, stdout)
	assert.Regexp(t, `(?m)^Subscription Number:\s+OCS-00042$`, stdout)
	assert.Regexp(t, `(?m)^State:\s+Active$`, stdout)
	assert.Regexp(t, `(?m)^External State:\s+ACTIVE$`, stdout)
	assert.Regexp(t, `(?m)^Source System:\s+AppleAppStore$`, stdout)
	assert.Regexp(t, `(?m)^External Subscription ID:\s+ext-sub-777$`, stdout)
	assert.Regexp(t, `(?m)^Auto Renew:\s+true$`, stdout)
	assert.Regexp(t, `(?m)^Currency:\s+JPY$`, stdout)
}

func TestOmnichannelGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, nil, "omnichannel", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
