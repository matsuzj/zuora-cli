// Package create implements the "zr payment create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the payment create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment",
		Long: `Create a new Zuora payment.

Examples:
  zr payment create --body @payment.json
  zr payment create --body '{"amount":100,"accountId":"2c92c0f8...","effectiveDate":"2026-01-01"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/payments",
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
			if id := cmdutil.GetDecimal(raw, "id"); id != "" {
				return fmt.Sprintf("Payment %s created.\n", id)
			}
			return ""
		},
	})
}
