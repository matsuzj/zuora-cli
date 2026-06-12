// Package preview implements the "zr order preview" command.
package preview

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPreview creates the order preview command.
func NewCmdPreview(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview an order",
		Long:  `Preview a Zuora order without creating it.`,
		Example: `  zr order preview --body @preview.json
  zr order preview --body '{"existingAccountNumber":"A001",...}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreview(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runPreview(cmd *cobra.Command, f *factory.Factory, body string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/orders/preview", bodyReader)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
