// Package paymentmethodsdefault implements the "zr account payment-methods-default" command.
package paymentmethodsdefault

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPaymentMethodsDefault creates the account payment-methods-default command.
func NewCmdPaymentMethodsDefault(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment-methods-default <account-key>",
		Short: "Get default payment method for an account",
		Long: `Get the default payment method for a Zuora billing account.

Examples:
  zr account payment-methods-default A00000001
  zr account payment-methods-default A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDefault(cmd, f, args[0])
		},
	}
	return cmd
}

func runDefault(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/accounts/%s/payment-methods/default", url.PathEscape(key)))
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
		{Key: "Type", Value: getString(raw, "type")},
		{Key: "Card Number", Value: getString(raw, "creditCardMaskNumber")},
		{Key: "Expiration Month", Value: getString(raw, "expirationMonth")},
		{Key: "Expiration Year", Value: getString(raw, "expirationYear")},
		{Key: "Status", Value: getString(raw, "status")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
