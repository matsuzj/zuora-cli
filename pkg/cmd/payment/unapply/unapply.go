// Package unapply implements the "zr payment unapply" command.
package unapply

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type unapplyOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdUnapply creates the payment unapply command.
func NewCmdUnapply(f *factory.Factory) *cobra.Command {
	opts := &unapplyOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "unapply <payment-id>",
		Short: "Unapply a payment from invoices and debit memos",
		Long: `Unapply a Zuora payment, detaching it from the invoices and debit memos it
was applied to so the balance can be reapplied elsewhere.`,
		Example: `  zr payment unapply 2c92c0f8... --body @unapply.json
  zr payment unapply 2c92c0f8... --body '{"invoices":[{"invoiceId":"2c92c0f8...","amount":50}]}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnapply(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runUnapply(cmd *cobra.Command, opts *unapplyOptions, paymentID string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/payments/%s/unapply", url.PathEscape(paymentID)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				// The Payments response field is "number" (matching payment
				// get/apply/refund, all live-verified); "paymentNumber" never
				// existed.
				{Key: "Payment Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Unapplied Amount", Value: cmdutil.GetMoney(raw, "unappliedAmount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetBool(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Payment %s unapplied.\n", paymentID)
		},
	})
}
