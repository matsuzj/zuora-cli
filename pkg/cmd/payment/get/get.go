// Package get implements the "zr payment get" command.
package get

import (
	"encoding/json"
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
		Long: `Get detailed information about a Zuora payment.

Examples:
  zr payment get 2c92c0f8...
  zr payment get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, paymentID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/payments/%s", url.PathEscape(paymentID)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetDecimal(raw, "id")},
		{Key: "Payment Number", Value: cmdutil.GetDecimal(raw, "paymentNumber")},
		{Key: "Effective Date", Value: cmdutil.GetDecimal(raw, "effectiveDate")},
		{Key: "Amount", Value: cmdutil.GetDecimal(raw, "amount")},
		{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
		{Key: "Type", Value: cmdutil.GetDecimal(raw, "type")},
		{Key: "Account ID", Value: cmdutil.GetDecimal(raw, "accountId")},
		{Key: "Gateway State", Value: cmdutil.GetDecimal(raw, "gatewayState")},
		{Key: "Created Date", Value: cmdutil.GetDecimal(raw, "createdDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
