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

func TestInvoicePost_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/post", map[string]interface{}{
		"id":            "inv-001",
		"invoiceNumber": "INV00001",
		"status":        "Posted",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "post", "inv-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Posted")
}

func TestInvoicePost_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "post")
	assert.Error(t, err)
}

func TestInvoicePost_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730020, "Invoice is not in draft status")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "post", "inv-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in draft status")
}
