// Package update implements the "zr product update" command.
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

// NewCmdUpdate creates the product update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &updateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a commerce product",
		Long: `Update a Zuora commerce product.

The product ID is specified in the request body.

Examples:
  zr product update --body @product.json
  zr product update --body '{"id":"...","name":"Updated Product"}'`,
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
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put("/commerce/products", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" || fmtOpts.Template != "" {
		_, err := output.RenderJSON(f.IOStreams, resp.Body, fmtOpts)
		return err
	}
	if fmtOpts.CSV && !fmtOpts.JSON {
		// Bare --csv is an explicit error on JSON-only commands; the
		// JSON-family flags keep their documented precedence over --csv.
		return output.ErrCSVUnsupportedJSONOnly
	}

	if err := output.PrintJSON(f.IOStreams, resp.Body, ""); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Product updated.\n")
	return nil
}
