// Package items implements the "zr invoice items" command.
package items

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdItems creates the invoice items command.
func NewCmdItems(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items <invoice-id>",
		Short: "List invoice items",
		Long: `List all items for a Zuora invoice.

Examples:
  zr invoice items 2c92c0f8...
  zr invoice items 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runItems(cmd, f, args[0])
		},
	}
	return cmd
}

func runItems(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/invoices/%s/items", url.PathEscape(invoiceID)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		InvoiceItems []struct {
			ID           string  `json:"id"`
			Subscription string  `json:"subscriptionName"`
			ChargeAmount float64 `json:"chargeAmount"`
			ChargeDate   string  `json:"chargeDate"`
			ChargeName   string  `json:"chargeName"`
		} `json:"invoiceItems"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "SUBSCRIPTION", Field: "subscriptionName"},
		{Header: "CHARGE_AMOUNT", Field: "chargeAmount"},
		{Header: "CHARGE_DATE", Field: "chargeDate"},
		{Header: "CHARGE_NAME", Field: "chargeName"},
	}

	rows := make([][]string, len(body.InvoiceItems))
	for i, item := range body.InvoiceItems {
		rows[i] = []string{
			item.ID,
			item.Subscription,
			fmt.Sprintf("%.2f", item.ChargeAmount),
			item.ChargeDate,
			item.ChargeName,
		}
	}

	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}
