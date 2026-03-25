// Package apply implements the "zr payment apply" command.
package apply

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
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
		Long: `Apply a Zuora payment to one or more invoices.

Examples:
  zr payment apply 2c92c0f8... --body @apply.json
  zr payment apply 2c92c0f8... --body '{"invoices":[{"invoiceId":"2c92c0f8...","amount":50}]}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runApply(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runApply(cmd *cobra.Command, opts *applyOptions, paymentID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/payments/%s/apply", url.PathEscape(paymentID)), bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "id")},
		{Key: "Payment Number", Value: getString(raw, "paymentNumber")},
		{Key: "Amount", Value: getString(raw, "amount")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Payment %s applied.\n", paymentID)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
