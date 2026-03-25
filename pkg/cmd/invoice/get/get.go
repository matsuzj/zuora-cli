// Package get implements the "zr invoice get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the invoice get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <invoice-id>",
		Short: "Get invoice details",
		Long: `Get detailed information about a Zuora invoice.

Examples:
  zr invoice get 2c92c0f8...
  zr invoice get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/invoices/%s", url.PathEscape(invoiceID)), api.WithCheckSuccess())
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
		{Key: "Invoice Number", Value: getString(raw, "invoiceNumber")},
		{Key: "Invoice Date", Value: getString(raw, "invoiceDate")},
		{Key: "Due Date", Value: getString(raw, "dueDate")},
		{Key: "Amount", Value: getString(raw, "amount")},
		{Key: "Balance", Value: getString(raw, "balance")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Account ID", Value: getString(raw, "accountId")},
		{Key: "Created Date", Value: getString(raw, "createdDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
