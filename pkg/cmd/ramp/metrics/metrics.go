// Package metrics implements the "zr ramp metrics" command.
package metrics

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdMetrics creates the ramp metrics command.
func NewCmdMetrics(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics <ramp-number>",
		Short: "Get ramp metrics",
		Long:  `Get metrics for a Zuora ramp.`,
		Example: `  zr ramp metrics R-00000001
  zr ramp metrics R-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetrics(cmd, f, args[0])
		},
	}
	return cmd
}

func runMetrics(cmd *cobra.Command, f *factory.Factory, rampNumber string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/ramps/%s/ramp-metrics", url.PathEscape(rampNumber)))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
