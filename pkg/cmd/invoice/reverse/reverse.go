// Package reverse implements the "zr invoice reverse" command.
package reverse

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdReverse creates the invoice reverse command.
func NewCmdReverse(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "reverse <invoice-id>",
		Short: "Reverse a posted invoice",
		Long: `Reverse a posted Zuora invoice.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr invoice reverse 2c92c0f8... --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runReverse(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "reversal")
	return cmd
}

func runReverse(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/invoices/%s/reverse", url.PathEscape(invoiceID)),
		// Zuora's invoice reverse endpoint binds a Map body parameter and
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
			return fmt.Sprintf("Invoice %s reversed.\n", invoiceID)
		},
	})
}
