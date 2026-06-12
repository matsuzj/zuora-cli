// Package apply implements the "zr payment apply" command.
package apply

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type applyOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdApply creates the payment apply command.
func NewCmdApply(f *factory.Factory) *cobra.Command {
	opts := &applyOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "apply <payment-id>",
		Short: "Apply a payment to invoices",
		Long:  `Apply a Zuora payment to one or more invoices.`,
		Example: `  zr payment apply 2c92c0f8... --body @apply.json
  zr payment apply 2c92c0f8... --body '{"invoices":[{"invoiceId":"2c92c0f8...","amount":50}]}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runApply(cmd *cobra.Command, opts *applyOptions, paymentID string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/payments/%s/apply", url.PathEscape(paymentID)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetDecimal(raw, "id")},
				{Key: "Payment Number", Value: cmdutil.GetDecimal(raw, "paymentNumber")},
				{Key: "Amount", Value: cmdutil.GetDecimal(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
				{Key: "Success", Value: cmdutil.GetDecimal(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Payment %s applied.\n", paymentID)
		},
	})
}
