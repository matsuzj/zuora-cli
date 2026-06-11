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

func TestOrderCreate_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/orders", map[string]interface{}{
		"success":     true,
		"orderNumber": "O-00000001",
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create", "--body", `{"existingAccountNumber":"A001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stderr, "Order O-00000001 created.")
}

func TestOrderCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
