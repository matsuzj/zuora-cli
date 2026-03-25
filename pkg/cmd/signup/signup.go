// Package signup implements the "zr signup" command.
package signup

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
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
account creation, payment method setup, and subscription creation in a single call.

Examples:
  zr signup --body @signup.json
  zr signup --body '{"accountInfo":{...},"subscriptionInfo":{...}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runSignup(cmd, f, body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runSignup(cmd *cobra.Command, f *factory.Factory, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/sign-up", bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Account ID", Value: getString(raw, "accountId")},
		{Key: "Account Number", Value: getString(raw, "accountNumber")},
		{Key: "Subscription ID", Value: getString(raw, "subscriptionId")},
		{Key: "Subscription Number", Value: getString(raw, "subscriptionNumber")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if num := getString(raw, "accountNumber"); num != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Sign-up complete. Account %s created.\n", num)
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
