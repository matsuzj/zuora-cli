// Package reverse implements the "zr invoice reverse" command.
package reverse

import (
	"encoding/json"
	"fmt"
	"net/url"

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

This action is irreversible. Use --confirm to proceed.

Examples:
  zr invoice reverse 2c92c0f8... --confirm`,
		Args: cobra.ExactArgs(1),
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
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/invoices/%s/reverse", url.PathEscape(invoiceID)), nil)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "Invoice Number", Value: cmdutil.GetString(raw, "invoiceNumber")},
		{Key: "Status", Value: cmdutil.GetString(raw, "status")},
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Invoice %s reversed.\n", invoiceID)
	return nil
}
