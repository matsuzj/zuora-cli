// Package get implements the "zr payment get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the payment get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <payment-id>",
		Short: "Get payment details",
		Long:  `Get detailed information about a Zuora payment.`,
		Example: `  zr payment get 2c92c0f8...
  zr payment get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, paymentID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/payments/%s", url.PathEscape(paymentID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				// The response field is "number" (matching creditmemo/
				// debitmemo); "paymentNumber" never existed — verified live.
				{Key: "Payment Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Effective Date", Value: cmdutil.GetString(raw, "effectiveDate")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Type", Value: cmdutil.GetString(raw, "type")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Gateway State", Value: cmdutil.GetString(raw, "gatewayState")},
				{Key: "Created Date", Value: cmdutil.GetString(raw, "createdDate")},
			}
		},
	})
}
