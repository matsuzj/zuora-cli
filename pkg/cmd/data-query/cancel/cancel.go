// Package cancel implements the "zr data-query cancel" command.
package cancel

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCancel creates the data-query cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "cancel <job-id>",
		Short: "Cancel a Data Query job",
		Long: `Cancel a Data Query job.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr data-query cancel 2c92c0f8... --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runCancel(cmd, f, args[0])
		},
	}
	cmdutil.AddConfirmFlag(cmd, &confirm, "cancellation")
	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, jobID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}
	resp, err := client.Delete(dqutil.JobPath(jobID))
	if err != nil {
		return err
	}
	return cmdutil.RenderDeleteResult(f.IOStreams, resp, output.FromCmd(cmd),
		fmt.Sprintf("Data Query job %s cancelled.\n", jobID),
		func(raw map[string]interface{}) []output.DetailField {
			return dqutil.DetailFields(dqutil.UnwrapData(raw))
		})
}
