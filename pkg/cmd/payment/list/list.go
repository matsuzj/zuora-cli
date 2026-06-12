// Package list implements the "zr payment list" command.
package list

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the payment list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List payments for an account",
		Long:  `List all payments associated with a Zuora billing account.`,
		Example: `  zr payment list --account-key A00000001
  zr payment list --account-key A00000001 --json
  zr payment list --account-key A00000001 --page-size 10`,
		Flags: []listcmd.Flag{
			{Name: "account-key", Usage: "Account key (ID or account number)", Required: true, DeprecatedName: "account"},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/transactions/payments/accounts/%s", url.PathEscape(flags["account-key"]))
		},
		ItemsKey: "payments",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "PAYMENT_NUMBER", Key: "paymentNumber"},
			{Header: "EFFECTIVE_DATE", Key: "effectiveDate"},
			{Header: "AMOUNT", Key: "amount", Money: true},
			{Header: "STATUS", Key: "status"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
