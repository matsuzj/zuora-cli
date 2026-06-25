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

func TestInvoiceGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/inv-001", map[string]interface{}{
		"id":            "inv-001",
		"invoiceNumber": "INV00001",
		"invoiceDate":   "2026-01-15",
		"dueDate":       "2026-02-15",
		"amount":        100.50,
		"balance":       50.25,
		"status":        "Posted",
		"accountId":     "acc-001",
		"createdDate":   "2026-01-10T10:00:00Z",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "get", "inv-001")
	require.NoError(t, err)
	// Label-bound (F-08): each value under its own label.
	assert.Regexp(t, `(?m)^Invoice Number:\s+INV00001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Posted$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.50$`, stdout) // money
	assert.Regexp(t, `(?m)^Balance:\s+50\.25$`, stdout) // money
	// String fields are read flat with GetString (the real invoice GET is a flat
	// object with these as strings — #340/F-17). Pin each so a wrong key renders
	// empty and fails here.
	assert.Regexp(t, `(?m)^ID:\s+inv-001$`, stdout)
	assert.Regexp(t, `(?m)^Invoice Date:\s+2026-01-15$`, stdout)
	assert.Regexp(t, `(?m)^Due Date:\s+2026-02-15$`, stdout)
	assert.Regexp(t, `(?m)^Account ID:\s+acc-001$`, stdout)
	assert.Regexp(t, `(?m)^Created Date:\s+2026-01-10T10:00:00Z$`, stdout)
}

func TestInvoiceGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":            "inv-001",
		"invoiceNumber": "INV00001",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "get", "inv-001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"invoiceNumber"`)
}

func TestInvoiceGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestInvoiceGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Invoice not found")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "get", "bad-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invoice not found")
}
