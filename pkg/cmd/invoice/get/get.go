// Package get implements the "zr invoice get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the invoice get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <invoice-id>",
		Short: "Get invoice details",
		Long:  `Get detailed information about a Zuora invoice.`,
		Example: `  zr invoice get 2c92c0f8...
  zr invoice get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/invoices/%s", url.PathEscape(invoiceID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// GET /v1/invoices/{id} returns a FLAT object (no "invoice" wrapper —
			// verified against a live invoice). id/invoiceNumber/dates/status/
			// accountId are STRINGS → GetString; only amount/balance are numeric →
			// GetMoney. (Previously these string fields used GetDecimal, which only
			// happened to pass strings through via %v — the wrong helper, #340/F-17.)
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Invoice Number", Value: cmdutil.GetString(raw, "invoiceNumber")},
				{Key: "Invoice Date", Value: cmdutil.GetString(raw, "invoiceDate")},
				{Key: "Due Date", Value: cmdutil.GetString(raw, "dueDate")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Balance", Value: cmdutil.GetMoney(raw, "balance")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Created Date", Value: cmdutil.GetString(raw, "createdDate")},
			}
		},
	})
}
