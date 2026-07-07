// Package create implements the "zr omnichannel create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the omnichannel create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an omni-channel subscription",
		Long:  `Create a new Zuora omni-channel subscription.`,
		Example: `  zr omnichannel create --body @omnichannel.json
  zr omnichannel create --body '{"externalSubscriptionId":"ext-sub-1","externalSourceSystem":"AppleAppStore"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/omni-channel-subscriptions",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real POST response per the official API reference (doc-verified
			// 2026-07-05, #414): {subscriptionId, subscriptionNumber, accountId,
			// accountNumber, success, …} — there is no subscriptionKey key, so
			// the success message never fired and the detail row was blank.
			// LIVE-UNVERIFIED(omni-channel create POST response {subscriptionId,subscriptionNumber,...}; since 2026-07-05; trigger: tenant with Omni-Channel provisioned)
			return []output.DetailField{
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if num := cmdutil.GetString(raw, "subscriptionNumber"); num != "" {
				return fmt.Sprintf("Omni-channel subscription %s created.\n", num)
			}
			return ""
		},
	})
}
