// Package create implements the "zr billrun create" command.
package create

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the billrun create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a bill run",
		Long: `Create a Zuora bill run.

Examples:
  zr billrun create --body @billrun.json
  zr billrun create --body '{"batches":["AllBatches"],"targetDate":"2026-06-30"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)
	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/bill-runs", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "Bill Run Number", Value: cmdutil.GetString(raw, "billRunNumber")},
		{Key: "Status", Value: cmdutil.GetString(raw, "status")},
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if id := cmdutil.GetString(raw, "id"); id != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Bill run %s created.\n", id)
	}
	return nil
}
