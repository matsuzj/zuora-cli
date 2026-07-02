// Package transfer implements the "zr payment transfer" command.
package transfer

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type transferOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdTransfer creates the payment transfer command.
func NewCmdTransfer(f *factory.Factory) *cobra.Command {
	opts := &transferOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "transfer <payment-id>",
		Short: "Transfer an unapplied payment to another account",
		Long: `Transfer the unapplied balance of a Zuora payment to another customer account.

The payment must have an unapplied balance; the target account is given in the
request body.`,
		Example: `  zr payment transfer 2c92c0f8... --body @transfer.json
  zr payment transfer 2c92c0f8... --body '{"accountId":"2c92c0f8..."}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransfer(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runTransfer(cmd *cobra.Command, opts *transferOptions, paymentID string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/payments/%s/transfer", url.PathEscape(paymentID)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				// The Payments response field is "number" (matching payment
				// get/apply/refund, all live-verified); "paymentNumber" never
				// existed.
				{Key: "Payment Number", Value: cmdutil.GetString(raw, "number")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Amount", Value: cmdutil.GetMoney(raw, "amount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetBool(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Payment %s transferred.\n", paymentID)
		},
	})
}
