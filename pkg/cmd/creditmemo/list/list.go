// Package list implements the "zr creditmemo list" command.
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
	Factory       *factory.Factory
	AccountID     string
	AccountNumber string
	Status        string
	Page          string
	PageSize      string
}

// NewCmdList creates the creditmemo list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List credit memos",
		Long: `List Zuora credit memos, optionally filtered by account or status.

Examples:
  zr creditmemo list
  zr creditmemo list --account-number A00000001
  zr creditmemo list --account-id 8aca... --status Posted
  zr creditmemo list --page-size 10 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.AccountID, "account-id", "", "Filter by Zuora account ID")
	cmd.Flags().StringVar(&opts.AccountNumber, "account-number", "", "Filter by account number")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (e.g. Draft, Posted, Cancelled)")
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
	if opts.AccountID != "" {
		reqOpts = append(reqOpts, api.WithQuery("accountId", opts.AccountID))
	}
	if opts.AccountNumber != "" {
		reqOpts = append(reqOpts, api.WithQuery("accountNumber", opts.AccountNumber))
	}
	if opts.Status != "" {
		reqOpts = append(reqOpts, api.WithQuery("status", opts.Status))
	}
	if opts.Page != "" {
		reqOpts = append(reqOpts, api.WithQuery("page", opts.Page))
	}
	if opts.PageSize != "" {
		reqOpts = append(reqOpts, api.WithQuery("pageSize", opts.PageSize))
	}

	resp, err := client.Get("/v1/creditmemos", reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		CreditMemos []struct {
			ID             string  `json:"id"`
			Number         string  `json:"number"`
			CreditMemoDate string  `json:"creditMemoDate"`
			Amount         float64 `json:"amount"`
			Balance        float64 `json:"balance"`
			Status         string  `json:"status"`
			AccountNumber  string  `json:"accountNumber"`
		} `json:"creditmemos"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID"},
		{Header: "NUMBER"},
		{Header: "DATE"},
		{Header: "AMOUNT"},
		{Header: "BALANCE"},
		{Header: "STATUS"},
		{Header: "ACCOUNT"},
	}

	rows := make([][]string, len(body.CreditMemos))
	for i, cm := range body.CreditMemos {
		rows[i] = []string{
			cm.ID,
			cm.Number,
			cm.CreditMemoDate,
			fmt.Sprintf("%.2f", cm.Amount),
			fmt.Sprintf("%.2f", cm.Balance),
			cm.Status,
			cm.AccountNumber,
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
