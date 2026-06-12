// Package delete implements the "zr usage delete" command.
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

// NewCmdDelete creates the usage delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a usage record",
		Long: `Delete a usage record by ID via the CRUD API.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr usage delete 2c92a0f96bd... --confirm`,
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, id string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/object/usage/%s", url.PathEscape(id)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Usage record %s deleted.\n", id),
		func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "Id")},
				{Key: "Success", Value: cmdutil.GetString(raw, "Success")},
			}
		})
}
