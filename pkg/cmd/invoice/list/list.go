// Package list implements the "zr invoice list" command.
package list

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the invoice list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List invoices for an account",
		Long:  `List all invoices associated with a Zuora billing account.`,
		Example: `  zr invoice list --account-key A00000001
  zr invoice list --account-key A00000001 --json
  zr invoice list --account-key A00000001 --page-size 10`,
		Flags: []listcmd.Flag{
			{Name: "account-key", Usage: "Account key (ID or account number)", Required: true},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/transactions/invoices/accounts/%s", url.PathEscape(flags["account-key"]))
		},
		ItemsKey: "invoices",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "INVOICE_NUMBER", Key: "invoiceNumber"},
			{Header: "INVOICE_DATE", Key: "invoiceDate"},
			{Header: "AMOUNT", Key: "amount", Money: true},
			{Header: "BALANCE", Key: "balance", Money: true},
			{Header: "STATUS", Key: "status"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
