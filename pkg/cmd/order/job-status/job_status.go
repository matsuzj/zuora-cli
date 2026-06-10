// Package jobstatus implements the "zr order job-status" command.
package jobstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type jobStatusOptions struct {
	Watch    bool
	Interval time.Duration
	Timeout  time.Duration
}

// NewCmdJobStatus creates the order job-status command.
func NewCmdJobStatus(f *factory.Factory) *cobra.Command {
	opts := &jobStatusOptions{}

	cmd := &cobra.Command{
		Use:   "job-status <job-id>",
		Short: "Get async job status",
		Long: `Get the status of an asynchronous order job.

Use --watch to poll until the job completes; --interval controls the
polling cadence and --timeout gives up after a duration (0 = no limit).
Ctrl-C cancels immediately, including mid-interval.

Examples:
  zr order job-status 2c92c0f9876...
  zr order job-status 2c92c0f9876... --watch
  zr order job-status 2c92c0f9876... --watch --interval 10s --timeout 5m
  zr order job-status 2c92c0f9876... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runJobStatus(cmd, f, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Watch, "watch", false, "Poll until job completes")
	cmd.Flags().DurationVar(&opts.Interval, "interval", 5*time.Second, "Polling interval for --watch")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 0, "Give up watching after this duration (0 = no limit)")
	return cmd
}

func runJobStatus(cmd *cobra.Command, f *factory.Factory, opts *jobStatusOptions, jobID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	if opts.Watch && opts.Interval <= 0 {
		return fmt.Errorf("--interval must be positive (got %s)", opts.Interval)
	}

	// The command context is cancelled by Ctrl-C (signal.NotifyContext in
	// main); --timeout layers a deadline on top of it. Re-point the client at
	// the derived context so an in-flight status request observes the
	// deadline too — not just the sleep between polls.
	ctx := cmd.Context()
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
		client.SetContext(ctx)
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

		status := cmdutil.GetString(raw, "status")

		fields := []output.DetailField{
			{Key: "Job ID", Value: cmdutil.GetString(raw, "jobId")},
			{Key: "Status", Value: status},
			{Key: "Result", Value: cmdutil.GetString(raw, "result")},
			{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
			{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
			{Key: "Success", Value: cmdutil.GetString(raw, "success")},
		}

		if !opts.Watch || isTerminalStatus(status) {
			return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
		}

		// Show progress and poll again. SleepContext (not time.Sleep!) so
		// Ctrl-C and --timeout interrupt mid-interval instead of being held
		// hostage for up to a full interval.
		fmt.Fprintf(f.IOStreams.ErrOut, "Job %s: %s (polling in %s...)\n", jobID, status, opts.Interval)
		if err := cmdutil.SleepContext(ctx, opts.Interval); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("gave up waiting for job %s after %s (last status: %s)", jobID, opts.Timeout, status)
			}
			return err
		}
	}
}

func isTerminalStatus(status string) bool {
	switch status {
	case "Completed", "Failed", "Error", "Cancelled":
		return true
	}
	return false
}
