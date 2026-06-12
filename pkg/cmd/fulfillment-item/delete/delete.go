// Package delete implements the "zr fulfillment-item delete" command.
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

// NewCmdDelete creates the fulfillment-item delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <fulfillment-item-id>",
		Short: "Delete a fulfillment item",
		Long: `Delete a Zuora fulfillment item.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr fulfillment-item delete 2c92c0f8... --confirm`,
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, itemID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/fulfillment-items/%s", url.PathEscape(itemID)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Fulfillment item %s deleted.\n", itemID),
		func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		})
}
