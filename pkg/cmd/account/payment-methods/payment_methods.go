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

	// Zuora API uses "returnedPaymentMethodType" as the envelope key (not "paymentMethods")
	var body struct {
		PaymentMethods []struct {
			ID                  string  `json:"id"`
			Type                string  `json:"type"`
			CreditCardMaskNum   string  `json:"creditCardMaskNumber"`
			AccountNumber       string  `json:"accountNumber"`
			IsDefault           bool    `json:"isDefault"`
			Status              string  `json:"status"`
		} `json:"returnedPaymentMethodType"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID", Field: "id"},
		{Header: "TYPE", Field: "type"},
		{Header: "LAST4", Field: "last4"},
		{Header: "DEFAULT", Field: "default"},
		{Header: "STATUS", Field: "status"},
	}

	rows := make([][]string, len(body.PaymentMethods))
	for i, pm := range body.PaymentMethods {
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
