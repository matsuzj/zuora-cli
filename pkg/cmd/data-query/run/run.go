// Package run implements the "zr data-query run" command (submit + poll +
// optional download).
package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type runOptions struct {
	submit          dqutil.SubmitFlags
	Output          string
	Interval        time.Duration
	Timeout         time.Duration
	DownloadTimeout time.Duration
}

// NewCmdRun creates the data-query run command.
func NewCmdRun(f *factory.Factory) *cobra.Command {
	opts := &runOptions{}
	cmd := &cobra.Command{
		Use:   `run ["<SQL>"]`,
		Short: "Submit a Data Query, wait for it, and optionally download the result",
		Long: `Submit a Data Query job, poll until it completes, and (with --output) download
the result file.

Provide the SQL as an argument or via --file (exactly one). Progress and the
summary go to stderr; with "--output -" the raw result bytes stream to stdout
(job metadata is never written to stdout in that mode, so --json/--csv have no
stdout effect there).

Note: the global 'zr --timeout' (see 'zr --help') bounds the WHOLE command
run; --wait-timeout bounds only the submit+poll wait and --download-timeout
the download. (While the deprecated --timeout alias exists, the global flag
is hidden from this help.)`,
		Example: `  zr data-query run "SELECT accountnumber FROM account" --output result.json
  zr data-query run --file q.sql --output - > result.jsonl
  zr data-query run "SELECT 1" --interval 3s --wait-timeout 5m`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(cmd, f, opts, args)
		},
	}
	dqutil.AddSubmitFlags(cmd.Flags(), &opts.submit)
	dqutil.RegisterSubmitCompletions(cmd)
	cmd.Flags().StringVar(&opts.Output, "output", "", "Write the result file here (- for stdout)")
	cmd.Flags().DurationVar(&opts.Interval, "interval", 5*time.Second, "Polling interval")
	// --wait-timeout is the primary name (#456): the old local --timeout
	// shadowed the global persistent `zr --timeout` in help output. The old
	// name stays registered (hidden, deprecated) for back-compat; both bind
	// the same variable.
	cmd.Flags().DurationVar(&opts.Timeout, "wait-timeout", 0, "Give up the submit+poll wait after this duration (0 = no limit); distinct from the global 'zr --timeout'")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 0, "Deprecated alias of --wait-timeout")
	_ = cmd.Flags().MarkDeprecated("timeout", "use --wait-timeout instead")
	cmd.Flags().DurationVar(&opts.DownloadTimeout, "download-timeout", 10*time.Minute, "Maximum time for the result download")
	return cmd
}

func runRun(cmd *cobra.Command, f *factory.Factory, opts *runOptions, args []string) error {
	sql, err := dqutil.ResolveSQL(args, opts.submit.File)
	if err != nil {
		return err
	}
	if opts.Interval <= 0 {
		return fmt.Errorf("--interval must be positive (got %s)", opts.Interval)
	}
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// baseCtx carries Ctrl-C and the global `zr --timeout`. The submit+poll
	// phase is bounded by the local --timeout (pollCtx); the download phase
	// derives its OWN budget from baseCtx (not pollCtx) so a long poll near the
	// --timeout deadline does not starve --download-timeout.
	baseCtx := cmd.Context()
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	pollCtx := baseCtx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		pollCtx, cancel = context.WithTimeout(baseCtx, opts.Timeout)
		defer cancel()
	}
	// Re-point the client at the poll context so submit and each poll request
	// observe Ctrl-C and the poll --timeout.
	client.SetContext(pollCtx)

	// Submit.
	body, err := dqutil.BuildSubmitBody(sql, &opts.submit)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	var reqOpts []api.RequestOption
	if opts.submit.IdempotencyKey != "" {
		reqOpts = append(reqOpts, api.WithHeader("Idempotency-Key", opts.submit.IdempotencyKey))
	}
	resp, err := client.Post(dqutil.SubmitPath, bytes.NewReader(body), reqOpts...)
	if err != nil {
		return err
	}
	d, err := dqutil.DecodeData(resp.Body)
	if err != nil {
		return err
	}
	jobID := cmdutil.GetString(d, "id")
	if jobID == "" {
		return fmt.Errorf("submit returned no job id: %s", strings.TrimSpace(string(resp.Body)))
	}
	status := cmdutil.GetString(d, "queryStatus")
	lastBody := resp.Body

	// Poll until the job reaches a terminal state.
	for !dqutil.IsTerminalStatus(status) {
		// jobID and status are response-derived: sanitize so hostile values
		// cannot write escape codes to the terminal via the progress line.
		fmt.Fprintf(f.IOStreams.ErrOut, "Data Query job %s: %s (polling in %s...)\n",
			output.SanitizeInline(jobID), output.SanitizeInline(status), opts.Interval)
		if err := cmdutil.SleepContext(pollCtx, opts.Interval); err != nil {
			// Only frame a deadline as "gave up after <timeout>" when the LOCAL
			// --timeout set it. With opts.Timeout==0 the deadline came from the
			// global `zr --timeout`, so waitErr would print "after 0s" and a
			// "raise --timeout" hint that does not apply — return the raw error,
			// mirroring the mid-GET branch below. (#428)
			if opts.Timeout > 0 {
				return waitErr(err, jobID, opts.Timeout, status)
			}
			return err
		}
		gresp, err := client.Get(dqutil.JobPath(jobID))
		if err != nil {
			// A deadline that fires mid-request should read like one that fires
			// mid-sleep, not as a raw "context deadline exceeded".
			if opts.Timeout > 0 && errors.Is(pollCtx.Err(), context.DeadlineExceeded) {
				return waitErr(context.DeadlineExceeded, jobID, opts.Timeout, status)
			}
			return err
		}
		if d, err = dqutil.DecodeData(gresp.Body); err != nil {
			return err
		}
		status = cmdutil.GetString(d, "queryStatus")
		lastBody = gresp.Body
	}

	switch strings.ToLower(status) {
	case "failed":
		return fmt.Errorf("data-query job %s failed: %s", jobID,
			dqutil.FirstNonEmpty(cmdutil.GetString(d, "errorMessage"), cmdutil.GetString(d, "message"), cmdutil.GetString(d, "error"), "(no error message)"))
	case "cancelled", "canceled":
		return fmt.Errorf("data-query job %s was cancelled", jobID)
	}

	// Completed.
	if opts.Output != "" {
		dataFile := cmdutil.GetString(d, "dataFile")
		if dataFile == "" {
			return fmt.Errorf("completed data-query job %s has no dataFile URL", jobID)
		}
		// Derive from baseCtx, not pollCtx: the download gets its own
		// --download-timeout budget and is not bounded by the poll --timeout.
		dctx := baseCtx
		if opts.DownloadTimeout > 0 {
			var dcancel context.CancelFunc
			dctx, dcancel = context.WithTimeout(baseCtx, opts.DownloadTimeout)
			defer dcancel()
		}
		// Production passes nil so DownloadStream builds the hardened client; the
		// test seam (DownloadClientForTest) is consulted inside DownloadStream,
		// not wired through production code. (#440)
		if err := dqutil.DownloadToFile(dctx, dataFile, opts.Output, f.IOStreams.Out, nil); err != nil {
			if errors.Is(dctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("downloading data-query result for job %s timed out (after --download-timeout %s)", jobID, opts.DownloadTimeout)
			}
			return err
		}
		dest := opts.Output
		if dest == "-" {
			dest = "stdout"
		}
		fmt.Fprintf(f.IOStreams.ErrOut, "Data Query job %s completed (%s rows); wrote result to %s\n",
			output.SanitizeInline(jobID), output.SanitizeInline(cmdutil.GetDecimal(d, "outputRows")), dest)
		return nil
	}

	// No --output: render the job metadata to stdout (honors --json/--csv).
	return output.RenderDetail(f.IOStreams, lastBody, output.FromCmd(cmd), dqutil.DetailFields(d))
}

// waitErr maps a poll-loop interruption: a deadline becomes a friendly give-up
// message (with a queue hint); other errors (e.g. a Ctrl-C cancellation) pass
// through unchanged.
func waitErr(err error, jobID string, timeout time.Duration, lastStatus string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("gave up waiting for data-query job %s after %s (last status: %s; the job may still be queued by tenant concurrency limits — use `data-query get %s` or raise --wait-timeout)",
			jobID, timeout, lastStatus, jobID)
	}
	return err
}
