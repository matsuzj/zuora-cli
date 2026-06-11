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

func TestUsageGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/object/usage/2c92a0f96bd", map[string]interface{}{
		"Id":            "2c92a0f96bd",
		"AccountId":     "acc123",
		"Quantity":      10,
		"StartDateTime": "2026-01-01",
		"UOM":           "Each",
		"Status":        "Processed",
	})

	stdout, _, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "get", "2c92a0f96bd")
	require.NoError(t, err)
	assert.Contains(t, stdout, "2c92a0f96bd")
	assert.Contains(t, stdout, "Each")
}

func TestUsageGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "get")
	assert.Error(t, err)
}
