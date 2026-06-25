// Package create implements the "zr subscription create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCreate creates the subscription create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a subscription",
		Long:  `Create a new Zuora subscription.`,
		Example: `  zr subscription create --body @subscription.json
  zr sub create --body '{"accountKey":"A001","termType":"TERMED",...}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runCreate(cmd *cobra.Command, f *factory.Factory, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/subscriptions",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
				{Key: "Subscription Number", Value: cmdutil.GetString(raw, "subscriptionNumber")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if num := cmdutil.GetString(raw, "subscriptionNumber"); num != "" {
				return fmt.Sprintf("Subscription %s created.\n", num)
			}
			return ""
		},
	})
}
