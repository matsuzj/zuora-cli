// Package paymentmethods implements the "zr account payment-methods" command.
package paymentmethods

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPaymentMethods creates the account payment-methods command.
func NewCmdPaymentMethods(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment-methods <account-key>",
		Short: "List payment methods for an account",
		Long: `List all payment methods associated with a Zuora billing account.

Examples:
  zr account payment-methods A00000001
  zr account payment-methods A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPaymentMethods(cmd, f, args[0])
		},
	}
	return cmd
}

func runPaymentMethods(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/accounts/%s/payment-methods", url.PathEscape(key)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// Zuora API returns payment methods under a dynamic key named after the
	// payment method type (e.g. "creditcard", "creditcardreferencetransaction").
	// We scan the response map for any array value to find the payment methods.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	type paymentMethod struct {
		ID                string `json:"id"`
		Type              string `json:"type"`
		CreditCardMaskNum string `json:"creditCardMaskNumber"`
		AccountNumber     string `json:"accountNumber"`
		IsDefault         bool   `json:"isDefault"`
		Status            string `json:"status"`
	}

	var methods []paymentMethod
	skipKeys := map[string]bool{"success": true, "defaultPaymentMethodId": true, "paymentGateway": true}
	for k, v := range raw {
		if skipKeys[k] {
			continue
		}
		var arr []paymentMethod
		if err := json.Unmarshal(v, &arr); err == nil && len(arr) > 0 {
			methods = append(methods, arr...)
		}
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "TYPE", Field: "type"},
		{Header: "LAST4", Field: "last4"},
		{Header: "DEFAULT", Field: "default"},
		{Header: "STATUS", Field: "status"},
	}

	rows := make([][]string, len(methods))
	for i, pm := range methods {
		last4 := lastN(pm.CreditCardMaskNum, 4)
		if last4 == "" {
			last4 = lastN(pm.AccountNumber, 4)
		}
		def := "false"
		if pm.IsDefault {
			def = "true"
		}
		rows[i] = []string{pm.ID, pm.Type, last4, def, pm.Status}
	}

	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}

func lastN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
