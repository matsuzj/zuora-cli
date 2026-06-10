// Package paymentmethodscascading implements the "zr account payment-methods-cascading" command.
package paymentmethodscascading

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
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
		Long: `Get the cascading (inherited) payment method for a Zuora billing account.

Examples:
  zr account payment-methods-cascading A00000001
  zr account payment-methods-cascading A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCascading(cmd, f, args[0])
		},
	}
	return cmd
}

func runCascading(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/accounts/%s/payment-methods/cascading", url.PathEscape(key)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// This endpoint returns the cascading payment method configuration
	fields := []output.DetailField{
		{Key: "Payment Method ID", Value: cmdutil.GetString(raw, "paymentMethodId")},
		{Key: "Cascading Consent", Value: cmdutil.GetString(raw, "paymentMethodCascadingConsent")},
		{Key: "Payment Method Type", Value: cmdutil.GetString(raw, "paymentMethodType")},
		{Key: "Payment Method Number", Value: cmdutil.GetString(raw, "paymentMethodNumber")},
		{Key: "Credit Card Type", Value: cmdutil.GetString(raw, "creditCardType")},
		{Key: "Credit Card Number", Value: cmdutil.GetString(raw, "creditCardMaskNumber")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
