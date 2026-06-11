// Package list implements the "zr plan list" command.
package list

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdList creates the plan list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commerce plans",
		Long: `List Zuora commerce plans.

Examples:
  zr plan list --body @query.json
  zr plan list --body '{"page":0,"page_size":20}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runList(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/commerce/plans/list", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
