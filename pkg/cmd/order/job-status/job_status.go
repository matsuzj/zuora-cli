// Package jobstatus implements the "zr order job-status" command.
package jobstatus

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdJobStatus creates the order job-status command.
func NewCmdJobStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job-status <job-id>",
		Short: "Get async job status",
		Long: `Get the status of an asynchronous order job.

Examples:
  zr order job-status 2c92c0f9876...
  zr order job-status 2c92c0f9876... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobStatus(cmd, f, args[0])
		},
	}
	return cmd
}

func runJobStatus(cmd *cobra.Command, f *factory.Factory, jobID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/async-jobs/%s", url.PathEscape(jobID)))
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
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Result", Value: getString(raw, "result")},
		{Key: "Order Number", Value: getString(raw, "orderNumber")},
		{Key: "Account Number", Value: getString(raw, "accountNumber")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
