package reverse

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdReverse(f) }

func TestInvoiceReverse_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/reverse", map[string]interface{}{
		"id":      "inv-001",
		"status":  "Reversed",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "reverse", "inv-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Reversed")
}

func TestInvoiceReverse_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "reverse", "inv-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoiceReverse_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "reverse")
	assert.Error(t, err)
}
