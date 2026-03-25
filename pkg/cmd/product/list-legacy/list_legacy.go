// Package listlegacy implements the "zr product list-legacy" command.
package listlegacy

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
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
		Long: `List Zuora legacy products via the commerce API.

Examples:
  zr product list-legacy --body @query.json
  zr product list-legacy --body '{"page":0,"page_size":20}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runListLegacy(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runListLegacy(cmd *cobra.Command, opts *listLegacyOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/commerce/legacy/products/list", bodyReader, api.WithCheckSuccess())
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
	return output.PrintJSON(f.IOStreams, resp.Body, "")
}
