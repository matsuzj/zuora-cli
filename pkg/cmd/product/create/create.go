// Package create implements the "zr product create" command.
package create

import (
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

// NewCmdCreate creates the product create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a commerce product",
		Long:  `Create a new Zuora commerce product.`,
		Example: `  zr product create --body @product.json
  zr product create --body '{"name":"My Product","description":"A product"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
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

	resp, err := client.Post("/commerce/products", bodyReader)
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Product created.\n")
	return nil
}
