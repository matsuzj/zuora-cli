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
}

// NewCmdRefund creates the payment refund command.
func NewCmdRefund(f *factory.Factory) *cobra.Command {
	opts := &refundOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "refund <payment-id>",
		Short: "Refund a payment",
		Long:  `Create a refund for a Zuora payment.`,
		Example: `  zr payment refund 2c92c0f8... --body @refund.json
  zr payment refund 2c92c0f8... --body '{"amount":50,"type":"External"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRefund(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

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
				{Key: "ID", Value: cmdutil.GetDecimal(raw, "id")},
				{Key: "Refund Number", Value: cmdutil.GetDecimal(raw, "refundNumber")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
				{Key: "Success", Value: cmdutil.GetDecimal(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := cmdutil.GetDecimal(raw, "id"); id != "" {
				return fmt.Sprintf("Refund %s created for payment %s.\n", id, paymentID)
			}
			return ""
		},
	})
}
