// Package changelogversion implements the "zr subscription changelog-version" command.
package changelogversion

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdChangelogVersion creates the subscription changelog-version command.
func NewCmdChangelogVersion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog-version <subscription-number> <version>",
		Short: "Get subscription change log for a specific version",
		Long: `Get change log for a specific subscription version.

Examples:
  zr subscription changelog-version S-00000001 1
  zr subscription changelog-version S-00000001 2 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelogVersion(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runChangelogVersion(cmd *cobra.Command, f *factory.Factory, subscriptionNumber, version string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(
		fmt.Sprintf("/v1/subscription-change-logs/%s/versions/%s",
			url.PathEscape(subscriptionNumber),
			url.PathEscape(version),
		),
	)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
