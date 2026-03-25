// Package list implements the "zr payment list" command.
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

// NewCmdList creates the payment list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payments for an account",
		Long: `List all payments associated with a Zuora billing account.

Examples:
  zr payment list --account A00000001
  zr payment list --account A00000001 --json
  zr payment list --account A00000001 --page-size 10`,
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
	resp, err := client.Get(fmt.Sprintf("/v1/transactions/payments/accounts/%s", url.PathEscape(opts.Account)), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Payments []struct {
			ID            string  `json:"id"`
			PaymentNumber string  `json:"paymentNumber"`
			EffectiveDate string  `json:"effectiveDate"`
			Amount        float64 `json:"amount"`
			Status        string  `json:"status"`
		} `json:"payments"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "PAYMENT_NUMBER", Field: "paymentNumber"},
		{Header: "EFFECTIVE_DATE", Field: "effectiveDate"},
		{Header: "AMOUNT", Field: "amount"},
		{Header: "STATUS", Field: "status"},
	}

	rows := make([][]string, len(body.Payments))
	for i, p := range body.Payments {
		rows[i] = []string{
			p.ID,
			p.PaymentNumber,
			p.EffectiveDate,
			fmt.Sprintf("%.2f", p.Amount),
			p.Status,
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
