// Package previewchange implements the "zr subscription preview-change" command.
package previewchange

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPreviewChange creates the subscription preview-change command.
func NewCmdPreviewChange(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "preview-change <subscription-key>",
		Short: "Preview changes to a subscription",
		Long: `Preview changes to an existing Zuora subscription without applying them.

Examples:
  zr subscription preview-change SUB-001 --body @changes.json
  zr sub preview-change SUB-001 --body '{"update":[...]}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runPreviewChange(cmd, f, args[0], body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runPreviewChange(cmd *cobra.Command, f *factory.Factory, key, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/preview", url.PathEscape(key))
	resp, err := client.Post(path, bodyReader, api.WithCheckSuccess())
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
