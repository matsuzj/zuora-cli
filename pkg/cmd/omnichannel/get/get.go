// Package get implements the "zr omnichannel get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the omnichannel get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription-key>",
		Short: "Get an omni-channel subscription",
		Long:  `Get detailed information about a Zuora omni-channel subscription.`,
		Example: `  zr omnichannel get S-001
  zr omnichannel get S-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, subscriptionKey string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/omni-channel-subscriptions/%s", url.PathEscape(subscriptionKey)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real shape per the official API reference (doc-verified 2026-07-05,
			// #414): the response is flat and carries NONE of the previously read
			// keys — subscriptionKey is only the path-parameter name, the states
			// are `state` (Zuora) and `externalState` (store), the source app
			// store is `externalSourceSystem`, and there is no createdDate.
			// LIVE-UNVERIFIED(omni-channel subscription GET flat shape incl. state/externalState; since 2026-07-05; trigger: tenant with Omni-Channel provisioned)
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "State", Value: cmdutil.GetString(raw, "state")},
				{Key: "External State", Value: cmdutil.GetString(raw, "externalState")},
				{Key: "Source System", Value: cmdutil.GetString(raw, "externalSourceSystem")},
				{Key: "External Subscription ID", Value: cmdutil.GetString(raw, "externalSubscriptionId")},
				{Key: "Auto Renew", Value: cmdutil.GetString(raw, "autoRenew")},
				{Key: "Currency", Value: cmdutil.GetString(raw, "currency")},
			}
		},
	})
}
