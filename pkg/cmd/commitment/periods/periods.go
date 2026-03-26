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
}

// NewCmdPeriods creates the commitment periods command.
func NewCmdPeriods(f *factory.Factory) *cobra.Command {
	opts := &periodsOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "periods",
		Short: "List commitment periods",
		Long: `List periods for a Zuora commitment.

Examples:
  zr commitment periods --commitment CMT-00000001
  zr commitment periods --commitment CMT-00000001 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Commitment == "" {
				return fmt.Errorf("--commitment is required")
			}
			return runPeriods(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Commitment, "commitment", "", "Commitment key (required)")

	return cmd
}

func runPeriods(cmd *cobra.Command, opts *periodsOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get("/v1/commitments/periods",
		api.WithQuery("commitmentKey", opts.Commitment),
		api.WithCheckSuccess(),
	)
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
