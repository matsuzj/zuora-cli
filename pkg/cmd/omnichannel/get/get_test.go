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
	handler := cmdtest.OK(t, "GET", "/v1/omni-channel-subscriptions/S-001", map[string]interface{}{
		"success":         true,
		"subscriptionKey": "S-001",
		"status":          "Active",
		"channel":         "Web",
	})

	stdout, _, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "get", "S-001")
	require.NoError(t, err)
	// Label-bound (F-08): value under its own label.
	assert.Regexp(t, `(?m)^Subscription Key:\s+S-001$`, stdout)
}

func TestOmnichannelGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, nil, "omnichannel", "get")
	assert.Error(t, err)
}
