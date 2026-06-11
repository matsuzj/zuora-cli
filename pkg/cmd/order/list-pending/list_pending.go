// Package listpending implements the "zr order list-pending" command.
package listpending

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdListPending creates the order list-pending command.
func NewCmdListPending(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list-pending <subscription-key>",
		Short: "List pending orders for a subscription",
		Long: `List pending Zuora orders for a subscription number or key.

Examples:
  zr order list-pending A-S00000001
  zr order list-pending A-S00000001 --json`,
		Args: cobra.ExactArgs(1),
		Flags: []listcmd.Flag{
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page"},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/orders/subscription/%s/pending", url.PathEscape(args[0]))
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
