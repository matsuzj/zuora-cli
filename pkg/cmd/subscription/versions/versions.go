// Package versions implements the "zr subscription versions" command.
package versions

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdVersions creates the subscription versions command.
func NewCmdVersions(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions <subscription-key> <version>",
		Short: "Get a specific subscription version",
		Long:  `Get details for a specific version of a Zuora subscription.`,
		Example: `  zr subscription versions A-S00000001 1
  zr sub versions A-S00000001 2 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersions(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runVersions(cmd *cobra.Command, f *factory.Factory, key, version string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/subscriptions/%s/versions/%s", url.PathEscape(key), url.PathEscape(version)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "Version", Value: cmdutil.GetString(raw, "version")},
				{Key: "Name", Value: cmdutil.GetString(raw, "name")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Term Type", Value: cmdutil.GetString(raw, "termType")},
				{Key: "Term Start Date", Value: cmdutil.GetString(raw, "termStartDate")},
				{Key: "Term End Date", Value: cmdutil.GetString(raw, "termEndDate")},
				{Key: "Contract Effective Date", Value: cmdutil.GetString(raw, "contractEffectiveDate")},
			}
		},
	})
}
