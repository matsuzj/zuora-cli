// Package files implements the "zr invoice files" command.
package files

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdFiles creates the invoice files command.
func NewCmdFiles(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files <invoice-id>",
		Short: "List invoice files",
		Long:  `List all files associated with a Zuora invoice.`,
		Example: `  zr invoice files 2c92c0f8...
  zr invoice files 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFiles(cmd, f, args[0])
		},
	}
	return cmd
}

func runFiles(cmd *cobra.Command, f *factory.Factory, invoiceID string) error {
	fmtOpts := output.FromCmd(cmd)

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/invoices/%s/files", url.PathEscape(invoiceID)))
	if err != nil {
		return err
	}

	// Live-verified shape (2026-07-02): {"invoiceFiles":[{id, versionNumber,
	// pdfFileUrl}], "success":true}. versionNumber is a large integer (epoch
	// millis), so json.Number keeps it out of scientific notation without a
	// float round-trip.
	var body struct {
		InvoiceFiles []struct {
			ID            string      `json:"id"`
			VersionNumber json.Number `json:"versionNumber"`
			PDFFileURL    string      `json:"pdfFileUrl"`
		} `json:"invoiceFiles"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "ID"},
		{Header: "VERSION"},
		{Header: "PDF_URL"},
	}
	rows := make([][]string, len(body.InvoiceFiles))
	for i, file := range body.InvoiceFiles {
		rows[i] = []string{file.ID, file.VersionNumber.String(), file.PDFFileURL}
	}

	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}
