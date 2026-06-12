// Package delete implements the "zr omnichannel delete" command.
package delete

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	Factory *factory.Factory
	Confirm bool
}

// NewCmdDelete creates the omnichannel delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <subscription-key>",
		Short: "Delete an omni-channel subscription",
		Long: `Delete a Zuora omni-channel subscription.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr omnichannel delete S-001 --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runDelete(cmd, opts, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &opts.Confirm, "deletion")

	return cmd
}

func runDelete(cmd *cobra.Command, opts *deleteOptions, subscriptionKey string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(
		fmt.Sprintf("/v1/omni-channel-subscriptions/%s", url.PathEscape(subscriptionKey)),
	)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Omni-channel subscription %s deleted.\n", subscriptionKey),
		func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		})
}
