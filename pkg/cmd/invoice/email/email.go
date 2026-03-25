// Package email implements the "zr invoice email" command.
package email

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
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

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runEmail(cmd *cobra.Command, opts *emailOptions, invoiceID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("/v1/invoices/%s/emails", url.PathEscape(invoiceID)), bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Invoice %s email sent.\n", invoiceID)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
