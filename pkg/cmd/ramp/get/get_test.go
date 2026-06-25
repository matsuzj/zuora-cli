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

func TestRampGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/ramps/R-00000001", map[string]interface{}{
		"success": true,
		"ramp": map[string]interface{}{
			// Real shape: the ramp is nested under a "ramp" object and its number
			// field is "number" (not "rampNumber").
			"number":             "R-00000001",
			"name":               "Test Ramp",
			"description":        "Ramp description",
			"subscriptionNumber": "A-S00000001",
		},
	})

	stdout, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "get", "R-00000001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Ramp Number:\s+R-00000001$`, stdout) // ramp.number (nested), not flat "rampNumber"
	assert.Regexp(t, `(?m)^Name:\s+Test Ramp$`, stdout)
	assert.Regexp(t, `(?m)^Subscription Number:\s+A-S00000001$`, stdout) // proves the nested read works
}

func TestRampGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "ramp", newCmd, nil, "ramp", "get")
	assert.Error(t, err)
}

func TestRampGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Ramp not found")

	_, _, err := cmdtest.Run(t, "ramp", newCmd, handler, "ramp", "get", "R-INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Ramp not found")
}
