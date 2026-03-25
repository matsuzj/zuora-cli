// Package files implements the "zr invoice files" command.
package files

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdFiles creates the invoice files command.
func NewCmdFiles(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files <invoice-id>",
		Short: "List invoice files",
		Long: `List all files associated with a Zuora invoice.

Output is always JSON due to the complex structure of file URLs.

Examples:
  zr invoice files 2c92c0f8...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFiles(cmd, f, args[0])
		},
	}
	return cmd
}

func runFiles(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/invoices/%s/files", url.PathEscape(invoiceID)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return output.PrintJSON(f.IOStreams, resp.Body, fmtOpts.JQ)
}
