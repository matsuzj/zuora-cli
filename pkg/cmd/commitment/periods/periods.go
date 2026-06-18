// Package periods implements the "zr commitment periods" command.
package periods

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
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

Use --commitment for a specific commitment, or --account-number + --start-date + --end-date
to query by account and date range.`,
		Example: `  zr commitment periods --commitment CMT-00000001
  zr commitment periods --account-number A00000001 --start-date 2026-01-01 --end-date 2026-12-31`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Commitment != "" && opts.Account != "" {
				return fmt.Errorf("--commitment and --account-number are mutually exclusive")
			}
			if opts.Commitment == "" && opts.Account == "" {
				return fmt.Errorf("--commitment or --account-number (with --start-date and --end-date) is required")
			}
			if opts.Account != "" && (opts.StartDate == "" || opts.EndDate == "") {
				return fmt.Errorf("--start-date and --end-date are required when using --account-number")
			}
			return runPeriods(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Commitment, "commitment", "", "Commitment key")
	cmdutil.AddAccountNumberFlag(cmd, &opts.Account)
	cmd.Flags().StringVar(&opts.StartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.EndDate, "end-date", "", "End date (YYYY-MM-DD)")

	return cmd
}

func runPeriods(cmd *cobra.Command, opts *periodsOptions) error {
	f := opts.Factory
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var reqOpts []api.RequestOption
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

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
