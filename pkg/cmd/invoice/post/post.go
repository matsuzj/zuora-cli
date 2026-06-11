// Package post implements the "zr invoice post" command.
package post

import (
	"encoding/json"
	"fmt"
	"net/url"

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
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/invoices/%s/post", url.PathEscape(invoiceID)), nil)
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Invoice %s posted.\n", invoiceID)
	return nil
}
