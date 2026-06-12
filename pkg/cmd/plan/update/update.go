// Package update implements the "zr plan update" command.
package update

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdUpdate creates the plan update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &updateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a commerce plan",
		Long: `Update a Zuora commerce plan.

The plan ID is specified in the request body.`,
		Example: `  zr plan update --body @plan.json
  zr plan update --body '{"id":"...","name":"Updated Plan"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runUpdate(cmd *cobra.Command, opts *updateOptions) error {
	f := opts.Factory
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put("/commerce/plans", bodyReader)
	if err != nil {
		return err
	}

	if fmtOpts.JQ != "" || fmtOpts.Template != "" {
		_, err := output.RenderJSON(f.IOStreams, resp.Body, fmtOpts)
		return err
	}

	if err := output.PrintJSON(f.IOStreams, resp.Body, ""); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Plan updated.\n")
	return nil
}
