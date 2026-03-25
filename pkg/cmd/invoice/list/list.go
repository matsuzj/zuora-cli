// Package list implements the "zr invoice list" command.
package list

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory  *factory.Factory
	Account  string
	Page     string
	PageSize string
}

// NewCmdList creates the invoice list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices for an account",
		Long: `List all invoices associated with a Zuora billing account.

Examples:
  zr invoice list --account A00000001
  zr invoice list --account A00000001 --json
  zr invoice list --account A00000001 --page-size 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Account, "account", "", "Account key (required)")
	cmd.Flags().StringVar(&opts.Page, "page", "", "Page number")
	cmd.Flags().StringVar(&opts.PageSize, "page-size", "", "Number of results per page")
	_ = cmd.MarkFlagRequired("account")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var reqOpts []api.RequestOption
	if opts.Page != "" {
		reqOpts = append(reqOpts, api.WithQuery("page", opts.Page))
	}
	if opts.PageSize != "" {
		reqOpts = append(reqOpts, api.WithQuery("pageSize", opts.PageSize))
	}

	reqOpts = append(reqOpts, api.WithCheckSuccess())
	resp, err := client.Get(fmt.Sprintf("/v1/transactions/invoices/accounts/%s", url.PathEscape(opts.Account)), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Invoices []struct {
			ID            string  `json:"id"`
			InvoiceNumber string  `json:"invoiceNumber"`
			InvoiceDate   string  `json:"invoiceDate"`
			Amount        float64 `json:"amount"`
			Balance       float64 `json:"balance"`
			Status        string  `json:"status"`
		} `json:"invoices"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "INVOICE_NUMBER", Field: "invoiceNumber"},
		{Header: "INVOICE_DATE", Field: "invoiceDate"},
		{Header: "AMOUNT", Field: "amount"},
		{Header: "BALANCE", Field: "balance"},
		{Header: "STATUS", Field: "status"},
	}

	rows := make([][]string, len(body.Invoices))
	for i, inv := range body.Invoices {
		rows[i] = []string{
			inv.ID,
			inv.InvoiceNumber,
			inv.InvoiceDate,
			fmt.Sprintf("%.2f", inv.Amount),
			fmt.Sprintf("%.2f", inv.Balance),
			inv.Status,
		}
	}

	if err := output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols); err != nil {
		return err
	}

	if body.NextPage != "" && !fmtOpts.JSON && fmtOpts.JQ == "" && fmtOpts.Template == "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Use --json to see nextPage URL.\n")
	}

	return nil
}
