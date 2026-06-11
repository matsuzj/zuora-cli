// Package listbysubscription implements the "zr order list-by-subscription" command.
package listbysubscription

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdListBySubscription creates the order list-by-subscription command.
func NewCmdListBySubscription(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list-by-subscription <subscription-key>",
		Short: "List orders by subscription",
		Long: `List Zuora orders for a subscription number or key.

Examples:
  zr order list-by-subscription A-S00000001
  zr order list-by-subscription A-S00000001 --json`,
		Args: cobra.ExactArgs(1),
		Flags: []listcmd.Flag{
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page"},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/orders/subscription/%s", url.PathEscape(args[0]))
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
