// Package get implements the "zr creditmemo get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the creditmemo get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <credit-memo-id>",
		Short: "Get credit memo details",
		Long: `Get detailed information about a Zuora credit memo.

Examples:
  zr creditmemo get 2c92c0f8...
  zr creditmemo get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, creditMemoID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/creditmemos/%s", url.PathEscape(creditMemoID)), api.WithCheckSuccess())
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
		{Key: "Number", Value: getString(raw, "number")},
		{Key: "Credit Memo Date", Value: getString(raw, "creditMemoDate")},
		{Key: "Amount", Value: getString(raw, "amount")},
		{Key: "Applied Amount", Value: getString(raw, "appliedAmount")},
		{Key: "Refund Amount", Value: getString(raw, "refundAmount")},
		{Key: "Balance", Value: getString(raw, "balance")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Currency", Value: getString(raw, "currency")},
		{Key: "Reason Code", Value: getString(raw, "reasonCode")},
		{Key: "Referred Invoice ID", Value: getString(raw, "referredInvoiceId")},
		{Key: "Account ID", Value: getString(raw, "accountId")},
		{Key: "Account Number", Value: getString(raw, "accountNumber")},
		{Key: "Created Date", Value: getString(raw, "createdDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	// JSON numbers decode to float64; format without scientific notation so
	// monetary amounts (e.g. 1000000) render as "1000000", not "1e+06".
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v)
}
