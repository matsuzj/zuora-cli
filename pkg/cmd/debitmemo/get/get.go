// Package get implements the "zr debitmemo get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the debitmemo get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <debit-memo-id>",
		Short: "Get debit memo details",
		Long:  `Get detailed information about a Zuora debit memo.`,
		Example: `  zr debitmemo get 2c92c0f8...
  zr debitmemo get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, debitMemoID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/debitmemos/%s", url.PathEscape(debitMemoID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Debit Memo Date", Value: cmdutil.GetString(raw, "debitMemoDate")},
				{Key: "Due Date", Value: cmdutil.GetString(raw, "dueDate")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Balance", Value: cmdutil.GetMoney(raw, "balance")},
				// taxAmount is monetary — use GetMoney for consistent 2-decimal
				// formatting with the sibling Amount/Balance fields. See #423.
				{Key: "Tax Amount", Value: cmdutil.GetMoney(raw, "taxAmount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Currency", Value: cmdutil.GetString(raw, "currency")},
				{Key: "Reason Code", Value: cmdutil.GetString(raw, "reasonCode")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Created Date", Value: cmdutil.GetString(raw, "createdDate")},
			}
		},
	})
}
