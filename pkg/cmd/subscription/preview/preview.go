// Package preview implements the "zr subscription preview" command.
package preview

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPreview creates the subscription preview command.
func NewCmdPreview(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview a subscription",
		Long:  `Preview a new Zuora subscription without creating it.`,
		Example: `  zr subscription preview --body @preview.json
  zr sub preview --body '{"accountKey":"A001","termType":"TERMED",...}'`,
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

	resp, err := client.Post("/v1/subscriptions/preview", bodyReader)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
