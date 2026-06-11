package post

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPost(f) }

func TestBillRunPost_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/post", map[string]interface{}{
		"id":            "br-001",
		"billRunNumber": "BR-00000001",
		"status":        "Posted",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "post", "br-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Posted")
}

func TestBillRunPost_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "post")
	assert.Error(t, err)
}
