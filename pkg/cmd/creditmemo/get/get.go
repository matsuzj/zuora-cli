// Package get implements the "zr creditmemo get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the creditmemo get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <credit-memo-id>",
		Short: "Get credit memo details",
		Long:  `Get detailed information about a Zuora credit memo.`,
		Example: `  zr creditmemo get 2c92c0f8...
  zr creditmemo get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, creditMemoID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/creditmemos/%s", url.PathEscape(creditMemoID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetDecimal(raw, "id")},
				{Key: "Number", Value: cmdutil.GetDecimal(raw, "number")},
				{Key: "Credit Memo Date", Value: cmdutil.GetDecimal(raw, "creditMemoDate")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Applied Amount", Value: cmdutil.GetMoney(raw, "appliedAmount")},
				{Key: "Refund Amount", Value: cmdutil.GetMoney(raw, "refundAmount")},
				// Credit memos expose the outstanding credit as "unappliedAmount"
				// (live-verified); there is no top-level "balance" field (that is a
				// debit-memo field). See #418.
				{Key: "Balance", Value: cmdutil.GetMoney(raw, "unappliedAmount")},
				{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
				{Key: "Currency", Value: cmdutil.GetDecimal(raw, "currency")},
				{Key: "Reason Code", Value: cmdutil.GetDecimal(raw, "reasonCode")},
				{Key: "Referred Invoice ID", Value: cmdutil.GetDecimal(raw, "referredInvoiceId")},
				{Key: "Account ID", Value: cmdutil.GetDecimal(raw, "accountId")},
				{Key: "Account Number", Value: cmdutil.GetDecimal(raw, "accountNumber")},
				{Key: "Created Date", Value: cmdutil.GetDecimal(raw, "createdDate")},
			}
		},
	})
}
