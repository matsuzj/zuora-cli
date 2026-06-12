package list

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestInvoiceList_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/transactions/invoices/accounts/A00000001", map[string]interface{}{
		"invoices": []map[string]interface{}{
			{
				"id":            "inv-001",
				"invoiceNumber": "INV00001",
				"invoiceDate":   "2026-01-15",
				"amount":        100.50,
				"balance":       50.25,
				"status":        "Posted",
			},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "list", "--account-key", "A00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "INV00001")
	assert.Contains(t, stdout, "Posted")
}

func TestInvoiceList_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"invoices": []map[string]interface{}{
			{
				"id":            "inv-001",
				"invoiceNumber": "INV00001",
				"invoiceDate":   "2026-01-15",
				"amount":        100.50,
				"balance":       50.25,
				"status":        "Posted",
			},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "list", "--account-key", "A00000001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"invoiceNumber"`)
}

func TestInvoiceList_RequiresAccount(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestInvoiceList_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "list", "--account-key", "INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}

// TestInvoiceList_DeprecatedAccountAliasStillWorks pins the P5-1 deprecation
// contract: the old --account spelling keeps feeding the account-key path
// through v0.5.x (removed in v0.6.0) and satisfies the required check.
func TestInvoiceList_DeprecatedAccountAliasStillWorks(t *testing.T) {
	var gotPath string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		fmt.Fprint(w, `{"invoices": []}`)
	}

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "list", "--account", "A00000001")
	require.NoError(t, err)
	assert.Equal(t, "/v1/transactions/invoices/accounts/A00000001", gotPath)
}
