// Package list implements the "zr account list" command.
package list

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the account list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List billing accounts",
		Long:  `List Zuora billing accounts via Object Query API.`,
		Example: `  zr account list
  zr account list --page-size 5
  zr account list --filter "status.EQ:Active"
  zr account list --json`,
		Flags: []listcmd.Flag{
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, IntDefault: 20},
			{Name: "cursor", Query: "cursor", Usage: "Pagination cursor"},
			{Name: "filter", Query: "filter[]", Usage: "Filter expressions (repeatable)", Repeatable: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/object-query/accounts"
		},
		ItemsKey: "data",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "NAME", Key: "name"},
			{Header: "NUMBER", Key: "accountNumber"},
			{Header: "STATUS", Key: "status"},
			{Header: "BALANCE", Key: "balance", Money: true},
			{Header: "CREATED", Key: "createdDate"},
		},
		NextPage: listcmd.NextPage{Flag: "cursor"},
	})
}
