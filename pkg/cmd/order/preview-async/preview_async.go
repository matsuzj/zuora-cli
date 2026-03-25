// Package previewasync implements the "zr order preview-async" command.
package previewasync

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPreviewAsync creates the order preview-async command.
func NewCmdPreviewAsync(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "preview-async",
		Short: "Preview an order asynchronously",
		Long: `Preview a Zuora order asynchronously. Returns a job ID.

Examples:
  zr order preview-async --body @preview.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runPreviewAsync(cmd, f, body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runPreviewAsync(cmd *cobra.Command, f *factory.Factory, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/async/orders/preview", bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Job ID", Value: getString(raw, "jobId")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if jobID := getString(raw, "jobId"); jobID != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Async order preview started. Job ID: %s\n", jobID)
		fmt.Fprintf(f.IOStreams.ErrOut, "Check status: zr order job-status %s\n", jobID)
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
