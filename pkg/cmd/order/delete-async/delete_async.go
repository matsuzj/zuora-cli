// Package deleteasync implements the "zr order delete-async" command.
package deleteasync

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDeleteAsync creates the order delete-async command.
func NewCmdDeleteAsync(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete-async <order-number>",
		Short: "Delete an order asynchronously",
		Long: `Delete a Zuora order asynchronously. Returns a job ID.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr order delete-async O-00000001 --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runDeleteAsync(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "deletion")
	return cmd
}

func runDeleteAsync(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/async/orders/%s", url.PathEscape(orderNumber)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Job ID", Value: cmdutil.GetString(raw, "jobId")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if jobID := cmdutil.GetString(raw, "jobId"); jobID != "" {
				return fmt.Sprintf("Async order deletion started. Job ID: %s\nCheck status: zr order job-status %s\n", jobID, jobID)
			}
			return ""
		},
	})
}
