// Package post implements the "zr billrun post" command.
package post

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdPost creates the billrun post command.
func NewCmdPost(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post <bill-run-id>",
		Short: "Post a bill run",
		Long: `Post a Zuora bill run, finalizing its generated invoices and credit memos.

Examples:
  zr billrun post 2c92c0f8...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPost(cmd, f, args[0])
		},
	}
	return cmd
}

func runPost(cmd *cobra.Command, f *factory.Factory, billRunID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/bill-runs/%s/post", url.PathEscape(billRunID)),
		// Zuora's bill-run post endpoint binds a Map body parameter and
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
			return fmt.Sprintf("Bill run %s posted.\n", billRunID)
		},
	})
}
