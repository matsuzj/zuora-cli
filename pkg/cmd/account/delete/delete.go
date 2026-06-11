// Package delete implements the "zr account delete" command.
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

// NewCmdDelete creates the account delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <account-key>",
		Short: "Delete a billing account",
		Long: `Delete a Zuora billing account. This is an async operation.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr account delete A00000001 --confirm`,
		Args: cobra.ExactArgs(1),
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, key string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/accounts/%s", url.PathEscape(key)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Account %s deleted.\n", key),
		func(raw map[string]interface{}) []output.DetailField {
			fields := []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
			if jobID, ok := raw["jobId"].(string); ok {
				fields = append(fields, output.DetailField{Key: "Job ID", Value: jobID})
			}
			if jobStatus, ok := raw["jobStatus"].(string); ok {
				fields = append(fields, output.DetailField{Key: "Job Status", Value: jobStatus})
			}
			return fields
		})
}
