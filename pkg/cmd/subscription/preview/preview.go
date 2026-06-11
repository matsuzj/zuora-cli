// Package preview implements the "zr subscription preview" command.
package preview

import (
	"fmt"

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
		Long: `Preview a new Zuora subscription without creating it.

Examples:
  zr subscription preview --body @preview.json
  zr sub preview --body '{"accountKey":"A001","termType":"TERMED",...}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runPreview(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runPreview(cmd *cobra.Command, f *factory.Factory, body string) error {
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

	fmtOpts := output.FromCmd(cmd)

	if handled, err := output.RenderJSON(f.IOStreams, resp.Body, fmtOpts); handled || err != nil {
		return err
	}
	return output.PrintJSON(f.IOStreams, resp.Body, "")
}
