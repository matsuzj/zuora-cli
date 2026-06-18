// Package list implements the "zr subscription list" command.
package list

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the subscription list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List subscriptions for an account",
		Long:  `List all subscriptions associated with a Zuora billing account.`,
		Example: `  zr subscription list --account-key A00000001
  zr subscription list --account-key A00000001 --json
  zr sub list --account-key A00000001 --page-size 5 --page 2`,
		Flags: []listcmd.Flag{
			{Name: "account-key", Usage: "Account key (ID or account number)", Required: true},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
			{Name: "page", Query: "page", Usage: "Page number (1-based)"},
			{Name: "charge-detail", Query: "charge-detail", Usage: "Charge detail level"},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/subscriptions/accounts/%s", url.PathEscape(flags["account-key"]))
		},
		ItemsKey: "subscriptions",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "NUMBER", Key: "subscriptionNumber"},
			{Header: "NAME", Key: "name"},
			{Header: "STATUS", Key: "status"},
			{Header: "TERM_TYPE", Key: "termType"},
			{Header: "START", Key: "termStartDate"},
			{Header: "END", Key: "termEndDate"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})
}
