// Package list implements the "zr subscription list" command.
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
	Factory      *factory.Factory
	Account      string
	PageSize     string
	Page         string
	ChargeDetail string
}

// NewCmdList creates the subscription list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subscriptions for an account",
		Long: `List all subscriptions associated with a Zuora billing account.

Examples:
  zr subscription list --account A00000001
  zr subscription list --account A00000001 --json
  zr sub list --account A00000001 --page-size 5 --page 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Account, "account", "", "Account key (required)")
	cmd.Flags().StringVar(&opts.PageSize, "page-size", "", "Number of results per page")
	cmd.Flags().StringVar(&opts.Page, "page", "", "Page number (1-based)")
	cmd.Flags().StringVar(&opts.ChargeDetail, "charge-detail", "", "Charge detail level")
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
	if opts.PageSize != "" {
		reqOpts = append(reqOpts, api.WithQuery("pageSize", opts.PageSize))
	}
	if opts.Page != "" {
		reqOpts = append(reqOpts, api.WithQuery("page", opts.Page))
	}
	if opts.ChargeDetail != "" {
		reqOpts = append(reqOpts, api.WithQuery("charge-detail", opts.ChargeDetail))
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/accounts/%s", url.PathEscape(opts.Account)), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Subscriptions []struct {
			ID                 string `json:"id"`
			SubscriptionNumber string `json:"subscriptionNumber"`
			Name               string `json:"name"`
			Status             string `json:"status"`
			TermType           string `json:"termType"`
			TermStartDate      string `json:"termStartDate"`
			TermEndDate        string `json:"termEndDate"`
		} `json:"subscriptions"`
		NextPage string `json:"nextPage"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "NUMBER", Field: "subscriptionNumber"},
		{Header: "NAME", Field: "name"},
		{Header: "STATUS", Field: "status"},
		{Header: "TERM_TYPE", Field: "termType"},
		{Header: "START", Field: "termStartDate"},
		{Header: "END", Field: "termEndDate"},
	}

	rows := make([][]string, len(body.Subscriptions))
	for i, s := range body.Subscriptions {
		rows[i] = []string{
			s.ID,
			s.SubscriptionNumber,
			s.Name,
			s.Status,
			s.TermType,
			s.TermStartDate,
			s.TermEndDate,
		}
	}

	if err := output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols); err != nil {
		return err
	}

	// Show pagination hint in table mode, preserving original flags
	if body.NextPage != "" && !fmtOpts.JSON && fmtOpts.JQ == "" && fmtOpts.Template == "" {
		if u, err := url.Parse(body.NextPage); err == nil {
			if p := u.Query().Get("page"); p != "" {
				hint := fmt.Sprintf("zr subscription list --account %s --page %s", opts.Account, p)
				if opts.PageSize != "" {
					hint += " --page-size " + opts.PageSize
				}
				if opts.ChargeDetail != "" {
					hint += " --charge-detail " + opts.ChargeDetail
				}
				fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Next page:\n  %s\n", hint)
			} else {
				fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Use --json to see nextPage URL.\n")
			}
		} else {
			fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Use --json to see nextPage URL.\n")
		}
	}

	return nil
}
