// Package refund implements the "zr payment refund" command.
package refund

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type refundOptions struct {
	Factory *factory.Factory
	Body    string
	Confirm bool
}

// NewCmdRefund creates the payment refund command.
func NewCmdRefund(f *factory.Factory) *cobra.Command {
	opts := &refundOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "refund <payment-id>",
		Short: "Refund a payment",
		Long: `Create a refund for a Zuora payment.

Disbursing a refund cannot be undone once processed. Use --confirm to proceed.`,
		Example: `  zr payment refund 2c92c0f8... --body @refund.json --confirm
  zr payment refund 2c92c0f8... --body '{"amount":50,"type":"External"}' --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runRefund(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)
	cmdutil.AddConfirmFlag(cmd, &opts.Confirm, "refund")

	return cmd
}

func runRefund(cmd *cobra.Command, opts *refundOptions, paymentID string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/payments/%s/refunds", url.PathEscape(paymentID)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				// The Refunds response field is "number" (matching payment
				// get/apply/create, all live-verified) — there is no
				// "refundNumber". It is a string ID, so use GetString. See #420.
				{Key: "Refund Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetBool(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := cmdutil.GetString(raw, "id"); id != "" {
				return fmt.Sprintf("Refund %s created for payment %s.\n", id, paymentID)
			}
			return ""
		},
	})
}
