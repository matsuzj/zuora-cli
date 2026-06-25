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
			return []output.DetailField{
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
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
