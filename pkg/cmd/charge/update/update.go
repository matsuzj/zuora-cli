// Package update implements the "zr charge update" command.
package update

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
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

The charge ID is specified in the request body.

Examples:
  zr charge update --body @charge.json
  zr charge update --body '{"id":"...","name":"Updated Charge"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdate(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runUpdate(cmd *cobra.Command, opts *updateOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put("/commerce/charges", bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" {
		return output.PrintJSON(f.IOStreams, resp.Body, fmtOpts.JQ)
	}
	if fmtOpts.Template != "" {
		return output.PrintTemplate(f.IOStreams, resp.Body, fmtOpts.Template)
	}

	if err := output.PrintJSON(f.IOStreams, resp.Body, ""); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Charge updated.\n")
	return nil
}
