// Package paymentmethodscascading implements the "zr account payment-methods-cascading" command.
package paymentmethodscascading

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPaymentMethodsCascading creates the account payment-methods-cascading command.
func NewCmdPaymentMethodsCascading(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment-methods-cascading <account-key>",
		Short: "Get cascading payment method for an account",
		Long:  `Get the cascading (inherited) payment method for a Zuora billing account.`,
		Example: `  zr account payment-methods-cascading A00000001
  zr account payment-methods-cascading A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCascading(cmd, f, args[0])
		},
	}
	return cmd
}

func runCascading(cmd *cobra.Command, f *factory.Factory, key string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/accounts/%s/payment-methods/cascading", url.PathEscape(key)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real shape per the official API reference (doc-verified 2026-07-05,
			// #416): {consent: bool, priorities: [{paymentMethodId, order}],
			// success} — every previously read flat key (paymentMethodId at top
			// level, paymentMethodCascadingConsent, paymentMethodType/Number,
			// creditCardType/MaskNumber) is absent, so the whole view was blank.
			// LIVE-UNVERIFIED(cascading payment-methods shape {consent,priorities[]}; since 2026-07-05; trigger: tenant with cascading payments enabled)
			fields := []output.DetailField{
				{Key: "Consent", Value: cmdutil.GetString(raw, "consent")},
			}
			if priorities, ok := raw["priorities"].([]interface{}); ok {
				for i, p := range priorities {
					entry, _ := p.(map[string]interface{})
					label := fmt.Sprintf("Priority %s", cmdutil.GetString(entry, "order"))
					if cmdutil.GetString(entry, "order") == "" {
						label = fmt.Sprintf("Priority #%d", i+1)
					}
					fields = append(fields, output.DetailField{
						Key: label, Value: cmdutil.GetString(entry, "paymentMethodId"),
					})
				}
			}
			fields = append(fields, output.DetailField{
				Key: "Success", Value: cmdutil.GetString(raw, "success"),
			})
			return fields
		},
	})
}
