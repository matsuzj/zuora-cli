// Package get implements the "zr debitmemo get" command.
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

// NewCmdGet creates the debitmemo get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <debit-memo-id>",
		Short: "Get debit memo details",
		Long: `Get detailed information about a Zuora debit memo.

Examples:
  zr debitmemo get 2c92c0f8...
  zr debitmemo get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, debitMemoID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/debitmemos/%s", url.PathEscape(debitMemoID)))
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
		{Key: "Number", Value: cmdutil.GetDecimal(raw, "number")},
		{Key: "Debit Memo Date", Value: cmdutil.GetDecimal(raw, "debitMemoDate")},
		{Key: "Due Date", Value: cmdutil.GetDecimal(raw, "dueDate")},
		{Key: "Amount", Value: cmdutil.GetDecimal(raw, "amount")},
		{Key: "Balance", Value: cmdutil.GetDecimal(raw, "balance")},
		{Key: "Tax Amount", Value: cmdutil.GetDecimal(raw, "taxAmount")},
		{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
		{Key: "Currency", Value: cmdutil.GetDecimal(raw, "currency")},
		{Key: "Reason Code", Value: cmdutil.GetDecimal(raw, "reasonCode")},
		{Key: "Account ID", Value: cmdutil.GetDecimal(raw, "accountId")},
		{Key: "Account Number", Value: cmdutil.GetDecimal(raw, "accountNumber")},
		{Key: "Created Date", Value: cmdutil.GetDecimal(raw, "createdDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
