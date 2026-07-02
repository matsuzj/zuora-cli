// Package signup implements the "zr signup" command.
package signup

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSignup creates the signup command.
func NewCmdSignup(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "signup",
		Short: "Create account, payment method, and subscription in one call",
		Long: `Sign up a new customer by creating an account, payment method, and subscription together.

This uses the Zuora Sign Up API (POST /v1/sign-up) which combines
account creation, payment method setup, and subscription creation in a single call.`,
		Example: `  zr signup --body @signup.json
  zr signup --body '{"accountInfo":{...},"subscriptionInfo":{...}}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSignup(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runSignup(cmd *cobra.Command, f *factory.Factory, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/sign-up",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// The Sign-Up API returns the immediate follow-on references a
			// completing onboarding flow needs next (order/invoice/payment/
			// credit-memo numbers) alongside the account+subscription. These are
			// flat top-level string ids per Zuora's Sign-Up response; paidAmount
			// is monetary (GetMoney → fixed two decimals). Absent fields render
			// blank, so this is additive and safe. NOTE: the response shape could
			// not be live-verified — this tenant returns HTTP 500 (69000060) on
			// sign-up — so the field names are sourced from Zuora's API docs.
			return []output.DetailField{
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Invoice ID", Value: cmdutil.GetString(raw, "invoiceId")},
				{Key: "Invoice Number", Value: cmdutil.GetString(raw, "invoiceNumber")},
				{Key: "Payment ID", Value: cmdutil.GetString(raw, "paymentId")},
				{Key: "Payment Number", Value: cmdutil.GetString(raw, "paymentNumber")},
				{Key: "Credit Memo ID", Value: cmdutil.GetString(raw, "creditMemoId")},
				{Key: "Credit Memo Number", Value: cmdutil.GetString(raw, "creditMemoNumber")},
				{Key: "Paid Amount", Value: cmdutil.GetMoney(raw, "paidAmount")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if num := cmdutil.GetString(raw, "accountNumber"); num != "" {
				return fmt.Sprintf("Sign-up complete. Account %s created.\n", num)
			}
			return ""
		},
	})
}
