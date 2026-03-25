// Package list implements the "zr order list" command.
package list

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory  *factory.Factory
	Status   string
	Page     string
	PageSize string
}

// NewCmdList creates the order list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List orders",
		Long: `List Zuora orders.

Examples:
  zr order list
  zr order list --status Completed
  zr order list --page 2 --page-size 10 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by order status")
	cmd.Flags().StringVar(&opts.Page, "page", "", "Page number")
	cmd.Flags().StringVar(&opts.PageSize, "page-size", "", "Number of results per page")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var reqOpts []api.RequestOption
	if opts.Status != "" {
		reqOpts = append(reqOpts, api.WithQuery("status", opts.Status))
	}
	if opts.Page != "" {
		reqOpts = append(reqOpts, api.WithQuery("page", opts.Page))
	}
	if opts.PageSize != "" {
		reqOpts = append(reqOpts, api.WithQuery("pageSize", opts.PageSize))
	}

	reqOpts = append(reqOpts, api.WithCheckSuccess())
	resp, err := client.Get("/v1/orders", reqOpts...)
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
			Description   string `json:"description"`
		} `json:"orders"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ORDER_NUMBER", Field: "orderNumber"},
		{Header: "STATUS", Field: "status"},
		{Header: "ORDER_DATE", Field: "orderDate"},
		{Header: "ACCOUNT", Field: "existingAccountNumber"},
		{Header: "DESCRIPTION", Field: "description"},
		{Header: "CREATED", Field: "createdDate"},
	}

	rows := make([][]string, len(body.Orders))
	for i, o := range body.Orders {
		rows[i] = []string{
			o.OrderNumber,
			o.Status,
			o.OrderDate,
			o.AccountNumber,
			o.Description,
			o.CreatedDate,
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
