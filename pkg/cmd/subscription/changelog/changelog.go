// Package changelog implements the "zr subscription changelog" command.
package changelog

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdChangelog creates the subscription changelog command.
func NewCmdChangelog(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog <subscription-number>",
		Short: "Get subscription change logs",
		Long: `Get change logs for a Zuora subscription.

Examples:
  zr subscription changelog S-00000001
  zr subscription changelog S-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelog(cmd, f, args[0])
		},
	}
	return cmd
}

func runChangelog(cmd *cobra.Command, f *factory.Factory, subscriptionNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(
		fmt.Sprintf("/v1/subscription-change-logs/%s", url.PathEscape(subscriptionNumber)),
		api.WithCheckSuccess(),
	)
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
