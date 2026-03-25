// Package get implements the "zr subscription get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the subscription get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subscription-key>",
		Short: "Get subscription details",
		Long: `Get detailed information about a Zuora subscription.

The subscription-key can be a subscription ID or subscription number.

Examples:
  zr subscription get A-S00000001
  zr sub get 402880ec12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/%s", url.PathEscape(key)))
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
		{Key: "Name", Value: getString(raw, "name")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Account ID", Value: getString(raw, "accountId")},
		{Key: "Account Number", Value: getString(raw, "accountNumber")},
		{Key: "Account Name", Value: getString(raw, "accountName")},
		{Key: "Term Type", Value: getString(raw, "termType")},
		{Key: "Term Start Date", Value: getString(raw, "termStartDate")},
		{Key: "Term End Date", Value: getString(raw, "termEndDate")},
		{Key: "Current Term", Value: getString(raw, "currentTerm")},
		{Key: "Current Term Period", Value: getString(raw, "currentTermPeriodType")},
		{Key: "Auto Renew", Value: getString(raw, "autoRenew")},
		{Key: "Contract Effective Date", Value: getString(raw, "contractEffectiveDate")},
		{Key: "Service Activation Date", Value: getString(raw, "serviceActivationDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
