// Package cancel implements the "zr payment cancel" command.
package cancel

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCancel creates the payment cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "cancel <payment-id>",
		Short: "Cancel a payment",
		Long: `Cancel a Zuora payment.

Cancelling a payment reverses it and cannot be undone. Use --confirm to proceed.`,
		Example: `  zr payment cancel 2c92c0f8... --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runCancel(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "cancellation")

	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, paymentID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/payments/%s/cancel", url.PathEscape(paymentID)),
		// Zuora's payment lifecycle PUTs bind a Map body parameter and return
		// HTTP 415 when the request carries no Content-Type. The client sets
		// Content-Type only when a body is present, so send an explicit empty
		// JSON object — the same contract as invoice/billrun post.
		Body: strings.NewReader("{}"),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				// The Payments response field is "number" (matching payment
				// get/apply/refund, all live-verified); "paymentNumber" never
				// existed.
				{Key: "Payment Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetBool(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Payment %s cancelled.\n", paymentID)
		},
	})
}
