// Package post implements the "zr invoice post" command.
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

// NewCmdPost creates the invoice post command.
func NewCmdPost(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post <invoice-id>",
		Short: "Post a draft invoice",
		Long: `Post a draft Zuora invoice, transitioning it to Posted status.

Examples:
  zr invoice post 2c92c0f8...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPost(cmd, f, args[0])
		},
	}
	return cmd
}

func runPost(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/invoices/%s/post", url.PathEscape(invoiceID)),
		// Zuora's invoice post endpoint binds a Map body parameter and
		// returns HTTP 415 when the request carries no Content-Type. The
		// client sets Content-Type only when a body is present, so send an
		// explicit empty JSON object (live-verified 2026-06-12).
		Body: strings.NewReader("{}"),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Invoice Number", Value: cmdutil.GetString(raw, "invoiceNumber")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Invoice %s posted.\n", invoiceID)
		},
	})
}
