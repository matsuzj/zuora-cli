// Package listlegacy implements the "zr product list-legacy" command.
package listlegacy

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listLegacyOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdListLegacy creates the product list-legacy command.
func NewCmdListLegacy(f *factory.Factory) *cobra.Command {
	opts := &listLegacyOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list-legacy",
		Short: "List legacy products",
		Long:  `List Zuora legacy products via the commerce API.`,
		Example: `  zr product list-legacy --body @query.json
  zr product list-legacy --body '{"page":0,"page_size":20}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runListLegacy(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runListLegacy(cmd *cobra.Command, opts *listLegacyOptions) error {
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

	resp, err := client.Post("/commerce/legacy/products/list", bodyReader)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
