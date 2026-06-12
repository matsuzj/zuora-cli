// Package listbysubscriptionowner implements the "zr order list-by-subscription-owner" command.
package listbysubscriptionowner

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdListBySubscriptionOwner creates the order list-by-subscription-owner command.
func NewCmdListBySubscriptionOwner(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list-by-subscription-owner <account-number>",
		Short: "List orders by subscription owner account",
		Long:  `List Zuora orders for a subscription owner account.`,
		Example: `  zr order list-by-subscription-owner A00000001
  zr order list-by-subscription-owner A00000001 --json`,
		Args: cobra.ExactArgs(1),
		Flags: []listcmd.Flag{
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page"},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/orders/subscriptionOwner/%s", url.PathEscape(args[0]))
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
