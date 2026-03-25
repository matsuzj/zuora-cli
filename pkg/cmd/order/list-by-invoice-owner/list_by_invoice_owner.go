// Package listbyinvoiceowner implements the "zr order list-by-invoice-owner" command.
package listbyinvoiceowner

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
	Page     string
	PageSize string
}

// NewCmdListByInvoiceOwner creates the order list-by-invoice-owner command.
func NewCmdListByInvoiceOwner(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list-by-invoice-owner <account-number>",
		Short: "List orders by invoice owner account",
		Long: `List Zuora orders for an invoice owner account.

Examples:
  zr order list-by-invoice-owner A00000001
  zr order list-by-invoice-owner A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.Page, "page", "", "Page number")
	cmd.Flags().StringVar(&opts.PageSize, "page-size", "", "Number of results per page")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions, accountNumber string) error {
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
	resp, err := client.Get(fmt.Sprintf("/v1/orders/invoiceOwner/%s", url.PathEscape(accountNumber)), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Orders []struct {
			OrderNumber   string `json:"orderNumber"`
			Status        string `json:"status"`
			CreatedDate   string `json:"createdDate"`
			AccountNumber string `json:"existingAccountNumber"`
			OrderDate     string `json:"orderDate"`
		} `json:"orders"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ORDER_NUMBER", Field: "orderNumber"},
		{Header: "STATUS", Field: "status"},
		{Header: "ORDER_DATE", Field: "orderDate"},
		{Header: "ACCOUNT", Field: "existingAccountNumber"},
		{Header: "CREATED", Field: "createdDate"},
	}

	rows := make([][]string, len(body.Orders))
	for i, o := range body.Orders {
		rows[i] = []string{
			o.OrderNumber,
			o.Status,
			o.OrderDate,
			o.AccountNumber,
			o.CreatedDate,
		}
	}

	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}
