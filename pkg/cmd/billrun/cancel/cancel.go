// Package cancel implements the "zr billrun cancel" command.
package cancel

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCancel creates the billrun cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "cancel <bill-run-id>",
		Short: "Cancel a bill run",
		Long: `Cancel a Zuora bill run.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr billrun cancel 2c92c0f8... --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runCancel(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "cancellation")
	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, billRunID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/bill-runs/%s/cancel", url.PathEscape(billRunID)),
		// Zuora's bill-run cancel endpoint binds a Map body parameter and
		// returns HTTP 415 when the request carries no Content-Type. The
		// client sets Content-Type only when a body is present, so send an
		// explicit empty JSON object (live-verified 2026-06-12).
		Body: strings.NewReader("{}"),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Bill Run Number", Value: cmdutil.GetString(raw, "billRunNumber")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Bill run %s cancelled.\n", billRunID)
		},
	})
}
