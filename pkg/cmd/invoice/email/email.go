// Package email implements the "zr invoice email" command.
package email

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type emailOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdEmail creates the invoice email command.
func NewCmdEmail(f *factory.Factory) *cobra.Command {
	opts := &emailOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "email <invoice-id>",
		Short: "Email an invoice",
		Long: `Email a Zuora invoice to specified recipients.

Examples:
  zr invoice email 2c92c0f8... --body '{"emailAddresses":"user@example.com","useEmailTemplateSetting":true}'
  zr invoice email 2c92c0f8... --body @email.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runEmail(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runEmail(cmd *cobra.Command, opts *emailOptions, invoiceID string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/invoices/%s/emails", url.PathEscape(invoiceID)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Invoice %s email sent.\n", invoiceID)
		},
	})
}
