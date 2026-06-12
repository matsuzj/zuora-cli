// Package listbyinvoiceowner implements the "zr order list-by-invoice-owner" command.
package listbyinvoiceowner

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdListByInvoiceOwner creates the order list-by-invoice-owner command.
func NewCmdListByInvoiceOwner(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list-by-invoice-owner <account-number>",
		Short: "List orders by invoice owner account",
		Long:  `List Zuora orders for an invoice owner account.`,
		Example: `  zr order list-by-invoice-owner A00000001
  zr order list-by-invoice-owner A00000001 --json`,
		Args: cobra.ExactArgs(1),
		Flags: []listcmd.Flag{
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/orders/invoiceOwner/%s", url.PathEscape(args[0]))
		},
		ItemsKey: "orders",
		Columns: []listcmd.ColumnSpec{
			{Header: "ORDER_NUMBER", Key: "orderNumber"},
			{Header: "STATUS", Key: "status"},
			{Header: "ORDER_DATE", Key: "orderDate"},
			{Header: "ACCOUNT", Key: "existingAccountNumber"},
			{Header: "CREATED", Key: "createdDate"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
