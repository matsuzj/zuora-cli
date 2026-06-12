// Package list implements the "zr order list" command.
package list

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the order list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List orders",
		Long:  `List Zuora orders.`,
		Example: `  zr order list
  zr order list --status Completed
  zr order list --page 2 --page-size 10 --json`,
		Flags: []listcmd.Flag{
			{Name: "status", Query: "status", Usage: "Filter by order status"},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/v1/orders"
		},
		ItemsKey: "orders",
		Columns: []listcmd.ColumnSpec{
			{Header: "ORDER_NUMBER", Key: "orderNumber"},
			{Header: "STATUS", Key: "status"},
			{Header: "ORDER_DATE", Key: "orderDate"},
			{Header: "ACCOUNT", Key: "existingAccountNumber"},
			{Header: "DESCRIPTION", Key: "description"},
			{Header: "CREATED", Key: "createdDate"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
