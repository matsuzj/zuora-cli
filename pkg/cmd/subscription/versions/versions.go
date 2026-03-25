// Package versions implements the "zr subscription versions" command.
package versions

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdVersions creates the subscription versions command.
func NewCmdVersions(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions <subscription-key> <version>",
		Short: "Get a specific subscription version",
		Long: `Get details for a specific version of a Zuora subscription.

Examples:
  zr subscription versions A-S00000001 1
  zr sub versions A-S00000001 2 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersions(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runVersions(cmd *cobra.Command, f *factory.Factory, key, version string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/%s/versions/%s", url.PathEscape(key), url.PathEscape(version)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "id")},
		{Key: "Subscription Number", Value: getString(raw, "subscriptionNumber")},
		{Key: "Version", Value: getString(raw, "version")},
		{Key: "Name", Value: getString(raw, "name")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Term Type", Value: getString(raw, "termType")},
		{Key: "Term Start Date", Value: getString(raw, "termStartDate")},
		{Key: "Term End Date", Value: getString(raw, "termEndDate")},
		{Key: "Contract Effective Date", Value: getString(raw, "contractEffectiveDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
