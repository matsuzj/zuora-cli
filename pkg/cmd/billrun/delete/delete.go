// Package delete implements the "zr billrun delete" command.
package delete

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDelete creates the billrun delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <bill-run-id>",
		Short: "Delete a bill run",
		Long: `Delete a Zuora bill run.

This action is irreversible. Use --confirm to proceed. Only bill runs that
have not been posted can be deleted.

Examples:
  zr billrun delete 2c92c0f8... --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runDelete(cmd, f, args[0])
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm the deletion")
	return cmd
}

func runDelete(cmd *cobra.Command, f *factory.Factory, billRunID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/bill-runs/%s", url.PathEscape(billRunID)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Bill run %s deleted.\n", billRunID)
	return nil
}
