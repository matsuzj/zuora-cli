// Package deleteasync implements the "zr order delete-async" command.
package deleteasync

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDeleteAsync creates the order delete-async command.
func NewCmdDeleteAsync(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-async <order-number>",
		Short: "Delete an order asynchronously",
		Long: `Delete a Zuora order asynchronously. Returns a job ID.

Examples:
  zr order delete-async O-00000001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteAsync(cmd, f, args[0])
		},
	}
	return cmd
}

func runDeleteAsync(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/async/orders/%s", url.PathEscape(orderNumber)), api.WithCheckSuccess())
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
		fmt.Fprintf(f.IOStreams.ErrOut, "Async order deletion started. Job ID: %s\n", jobID)
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
