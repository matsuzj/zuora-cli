// Package previewasync implements the "zr order preview-async" command.
package previewasync

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPreviewAsync creates the order preview-async command.
func NewCmdPreviewAsync(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:     "preview-async",
		Short:   "Preview an order asynchronously",
		Long:    `Preview a Zuora order asynchronously. Returns a job ID.`,
		Example: `  zr order preview-async --body @preview.json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runPreviewAsync(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runPreviewAsync(cmd *cobra.Command, f *factory.Factory, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/async/orders/preview",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Job ID", Value: cmdutil.GetString(raw, "jobId")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if jobID := cmdutil.GetString(raw, "jobId"); jobID != "" {
				return fmt.Sprintf("Async order preview started. Job ID: %s\nCheck status: zr order job-status %s\n", jobID, jobID)
			}
			return ""
		},
	})
}
