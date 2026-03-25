// Package list implements the "zr account list" command.
package list

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory  *factory.Factory
	PageSize int
	Cursor   string
	Filters  []string
}

// NewCmdList creates the account list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List billing accounts",
		Long: `List Zuora billing accounts via Object Query API.

Examples:
  zr account list
  zr account list --page-size 5
  zr account list --filter "status.EQ:Active"
  zr account list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Number of results per page")
	cmd.Flags().StringVar(&opts.Cursor, "cursor", "", "Pagination cursor")
	cmd.Flags().StringArrayVar(&opts.Filters, "filter", nil, "Filter expressions (repeatable)")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var reqOpts []api.RequestOption
	reqOpts = append(reqOpts, api.WithQuery("pageSize", strconv.Itoa(opts.PageSize)))
	if opts.Cursor != "" {
		reqOpts = append(reqOpts, api.WithQuery("cursor", opts.Cursor))
	}
	if len(opts.Filters) > 0 {
		reqOpts = append(reqOpts, api.WithQuerySlice("filter[]", opts.Filters))
	}

	resp, err := client.Get("/object-query/accounts", reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Data []struct {
			ID            string  `json:"id"`
			Name          string  `json:"name"`
			AccountNumber string  `json:"accountNumber"`
			Status        string  `json:"status"`
			Balance       float64 `json:"balance"`
			CreatedDate   string  `json:"createdDate"`
		} `json:"data"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "NAME", Field: "name"},
		{Header: "NUMBER", Field: "accountNumber"},
		{Header: "STATUS", Field: "status"},
		{Header: "BALANCE", Field: "balance"},
		{Header: "CREATED", Field: "createdDate"},
	}

	rows := make([][]string, len(body.Data))
	for i, a := range body.Data {
		rows[i] = []string{
			a.ID,
			a.Name,
			a.AccountNumber,
			a.Status,
			fmt.Sprintf("%.2f", a.Balance),
			a.CreatedDate,
		}
	}

	if err := output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols); err != nil {
		return err
	}

	// Show pagination hint with cursor value in table mode
	if body.NextPage != "" && !fmtOpts.JSON && fmtOpts.JQ == "" && fmtOpts.Template == "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Next page:\n  zr account list --cursor %q\n", body.NextPage)
	}

	return nil
}
