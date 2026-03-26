// Package periods implements the "zr commitment periods" command.
package periods

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type periodsOptions struct {
	Factory    *factory.Factory
	Commitment string
	Account    string
	StartDate  string
	EndDate    string
}

// NewCmdPeriods creates the commitment periods command.
func NewCmdPeriods(f *factory.Factory) *cobra.Command {
	opts := &periodsOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "periods",
		Short: "List commitment periods",
		Long: `List periods for a Zuora commitment.

Use --commitment for a specific commitment, or --account + --start-date + --end-date
to query by account and date range.

Examples:
  zr commitment periods --commitment CMT-00000001
  zr commitment periods --account A00000001 --start-date 2026-01-01 --end-date 2026-12-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Commitment == "" && opts.Account == "" {
				return fmt.Errorf("--commitment or --account (with --start-date and --end-date) is required")
			}
			if opts.Account != "" && (opts.StartDate == "" || opts.EndDate == "") {
				return fmt.Errorf("--start-date and --end-date are required when using --account")
			}
			return runPeriods(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Commitment, "commitment", "", "Commitment key")
	cmd.Flags().StringVar(&opts.Account, "account", "", "Account number")
	cmd.Flags().StringVar(&opts.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.EndDate, "end-date", "", "End date (YYYY-MM-DD)")

	return cmd
}

func runPeriods(cmd *cobra.Command, opts *periodsOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var reqOpts []api.RequestOption
	reqOpts = append(reqOpts, api.WithCheckSuccess())
	if opts.Commitment != "" {
		reqOpts = append(reqOpts, api.WithQuery("commitmentKey", opts.Commitment))
	} else {
		reqOpts = append(reqOpts, api.WithQuery("accountNumber", opts.Account))
		reqOpts = append(reqOpts, api.WithQuery("startDate", opts.StartDate))
		reqOpts = append(reqOpts, api.WithQuery("endDate", opts.EndDate))
	}
	resp, err := client.Get("/v1/commitments/periods", reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" {
		return output.PrintJSON(f.IOStreams, resp.Body, fmtOpts.JQ)
	}
	if fmtOpts.Template != "" {
		return output.PrintTemplate(f.IOStreams, resp.Body, fmtOpts.Template)
	}
	return output.PrintJSON(f.IOStreams, resp.Body, "")
}
