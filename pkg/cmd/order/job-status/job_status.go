// Package jobstatus implements the "zr order job-status" command.
package jobstatus

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type jobStatusOptions struct {
	Watch bool
}

// NewCmdJobStatus creates the order job-status command.
func NewCmdJobStatus(f *factory.Factory) *cobra.Command {
	opts := &jobStatusOptions{}

	cmd := &cobra.Command{
		Use:   "job-status <job-id>",
		Short: "Get async job status",
		Long: `Get the status of an asynchronous order job.

Use --watch to poll until the job completes.

Examples:
  zr order job-status 2c92c0f9876...
  zr order job-status 2c92c0f9876... --watch
  zr order job-status 2c92c0f9876... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobStatus(cmd, f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Watch, "watch", false, "Poll until job completes")
	return cmd
}

func runJobStatus(cmd *cobra.Command, f *factory.Factory, opts *jobStatusOptions, jobID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/async-jobs/%s", url.PathEscape(jobID))

	for {
		resp, err := client.Get(path, api.WithCheckSuccess())
		if err != nil {
			return err
		}

		fmtOpts := output.FromCmd(cmd)

		var raw map[string]interface{}
		if err := json.Unmarshal(resp.Body, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		status := getString(raw, "status")

		fields := []output.DetailField{
			{Key: "Job ID", Value: getString(raw, "jobId")},
			{Key: "Status", Value: status},
			{Key: "Result", Value: getString(raw, "result")},
			{Key: "Order Number", Value: getString(raw, "orderNumber")},
			{Key: "Account Number", Value: getString(raw, "accountNumber")},
			{Key: "Success", Value: getString(raw, "success")},
		}

		if !opts.Watch || isTerminalStatus(status) {
			return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
		}

		// Show progress and poll again
		fmt.Fprintf(f.IOStreams.ErrOut, "Job %s: %s (polling in 5s...)\n", jobID, status)
		time.Sleep(5 * time.Second)
	}
}

func isTerminalStatus(status string) bool {
	switch status {
	case "Completed", "Failed", "Error", "Cancelled":
		return true
	}
	return false
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
