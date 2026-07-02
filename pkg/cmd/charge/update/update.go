// Package update implements the "zr charge update" command.
package update

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdUpdate creates the charge update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &updateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a commerce charge",
		Long: `Update a Zuora commerce charge.

The charge ID is specified in the request body.`,
		Example: `  zr charge update --body @charge.json
  zr charge update --body '{"id":"...","name":"Updated Charge"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

	resp, err := client.Put("/commerce/charges", bodyReader)
	if err != nil {
		return err
	}

	return output.RenderJSONWithMessage(f.IOStreams, resp.Body, fmtOpts, "Charge updated.\n")
}
