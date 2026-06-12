// Package create implements the "zr omnichannel create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the omnichannel create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an omni-channel subscription",
		Long:  `Create a new Zuora omni-channel subscription.`,
		Example: `  zr omnichannel create --body @omnichannel.json
  zr omnichannel create --body '{"subscriptionKey":"S-001",...}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/omni-channel-subscriptions",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Subscription Key", Value: cmdutil.GetString(raw, "subscriptionKey")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if key := cmdutil.GetString(raw, "subscriptionKey"); key != "" {
				return fmt.Sprintf("Omni-channel subscription %s created.\n", key)
			}
			return ""
		},
	})
}
