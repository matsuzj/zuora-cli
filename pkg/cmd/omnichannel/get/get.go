// Package get implements the "zr omnichannel get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the omnichannel get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription-key>",
		Short: "Get an omni-channel subscription",
		Long: `Get detailed information about a Zuora omni-channel subscription.

Examples:
  zr omnichannel get S-001
  zr omnichannel get S-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, subscriptionKey string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(
		fmt.Sprintf("/v1/omni-channel-subscriptions/%s", url.PathEscape(subscriptionKey)),
		api.WithCheckSuccess(),
	)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Subscription Key", Value: getString(raw, "subscriptionKey")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Channel", Value: getString(raw, "channel")},
		{Key: "Created Date", Value: getString(raw, "createdDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
