// Package paymentmethodsdefault implements the "zr account payment-methods-default" command.
package paymentmethodsdefault

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPaymentMethodsDefault creates the account payment-methods-default command.
func NewCmdPaymentMethodsDefault(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment-methods-default <account-key>",
		Short: "Get default payment method for an account",
		Long:  `Get the default payment method for a Zuora billing account.`,
		Example: `  zr account payment-methods-default A00000001
  zr account payment-methods-default A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefault(cmd, f, args[0])
		},
	}
	return cmd
}

func runDefault(cmd *cobra.Command, f *factory.Factory, key string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/accounts/%s/payment-methods/default", url.PathEscape(key)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Type", Value: cmdutil.GetString(raw, "type")},
				// Masked card number is "cardNumber" on the REST payment-method
				// object (live-verified), not the ZOQL "creditCardMaskNumber". (#421)
				{Key: "Card Number", Value: cmdutil.GetString(raw, "cardNumber")},
				{Key: "Expiration Month", Value: cmdutil.GetString(raw, "expirationMonth")},
				{Key: "Expiration Year", Value: cmdutil.GetString(raw, "expirationYear")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
			}
		},
	})
}
