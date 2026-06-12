// Package list implements the "zr debitmemo list" command.
package list

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the debitmemo list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List debit memos",
		Long:  `List Zuora debit memos, optionally filtered by account or status.`,
		Example: `  zr debitmemo list
  zr debitmemo list --account-number A00000001
  zr debitmemo list --account-id 8aca... --status Posted
  zr debitmemo list --page-size 10 --json`,
		Flags: []listcmd.Flag{
			{Name: "account-id", Query: "accountId", Usage: "Filter by Zuora account ID"},
			{Name: "account-number", Query: "accountNumber", Usage: "Filter by account number"},
			{Name: "status", Query: "status", Usage: "Filter by status (e.g. Draft, Posted, Cancelled)", Enum: []string{"Draft", "Posted", "Cancelled"}},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/v1/debitmemos"
		},
		ItemsKey: "debitmemos",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "NUMBER", Key: "number"},
			{Header: "DATE", Key: "debitMemoDate"},
			{Header: "AMOUNT", Key: "amount", Money: true},
			{Header: "BALANCE", Key: "balance", Money: true},
			{Header: "STATUS", Key: "status"},
			{Header: "ACCOUNT", Key: "accountNumber"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
