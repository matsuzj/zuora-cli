// Package get implements the "zr subscription get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the subscription get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription-key>",
		Short: "Get subscription details",
		Long: `Get detailed information about a Zuora subscription.

The subscription-key can be a subscription ID or subscription number.`,
		Example: `  zr subscription get A-S00000001
  zr sub get 402880ec12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, key string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/subscriptions/%s", url.PathEscape(key)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "Name", Value: cmdutil.GetString(raw, "name")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Account Name", Value: cmdutil.GetString(raw, "accountName")},
				{Key: "Term Type", Value: cmdutil.GetString(raw, "termType")},
				{Key: "Term Start Date", Value: cmdutil.GetString(raw, "termStartDate")},
				{Key: "Term End Date", Value: cmdutil.GetString(raw, "termEndDate")},
				{Key: "Current Term", Value: cmdutil.GetString(raw, "currentTerm")},
				{Key: "Current Term Period", Value: cmdutil.GetString(raw, "currentTermPeriodType")},
				{Key: "Auto Renew", Value: cmdutil.GetString(raw, "autoRenew")},
				{Key: "Contract Effective Date", Value: cmdutil.GetString(raw, "contractEffectiveDate")},
				{Key: "Service Activation Date", Value: cmdutil.GetString(raw, "serviceActivationDate")},
			}
		},
	})
}
