// Package schedules implements the "zr commitment schedules" command.
package schedules

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSchedules creates the commitment schedules command.
func NewCmdSchedules(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedules <commitment-key>",
		Short: "Get commitment schedules",
		Long: `Get schedules for a Zuora commitment.

Examples:
  zr commitment schedules CMT-00000001
  zr commitment schedules CMT-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSchedules(cmd, f, args[0])
		},
	}
	return cmd
}

func runSchedules(cmd *cobra.Command, f *factory.Factory, commitmentKey string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/commitments/%s/schedules", url.PathEscape(commitmentKey)), api.WithCheckSuccess())
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
