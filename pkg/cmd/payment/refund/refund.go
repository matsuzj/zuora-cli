// Package refund implements the "zr payment refund" command.
package refund

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
		Long: `Create a refund for a Zuora payment.

Examples:
  zr payment refund 2c92c0f8... --body @refund.json
  zr payment refund 2c92c0f8... --body '{"amount":50,"type":"External"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runRefund(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runRefund(cmd *cobra.Command, opts *refundOptions, paymentID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("/v1/payments/%s/refunds", url.PathEscape(paymentID)), bodyReader, api.WithCheckSuccess())
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
		{Key: "Refund Number", Value: getString(raw, "refundNumber")},
		{Key: "Amount", Value: getString(raw, "amount")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if id := getString(raw, "id"); id != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Refund %s created for payment %s.\n", id, paymentID)
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
