// Package usageratedetail implements the "zr invoice usage-rate-detail" command.
package usageratedetail

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUsageRateDetail creates the invoice usage-rate-detail command.
func NewCmdUsageRateDetail(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage-rate-detail <invoice-item-id>",
		Short: "Get usage rate detail for an invoice item",
		Long: `Get detailed usage rate information for a Zuora invoice item.

Output is always JSON due to the complex nested structure.

Examples:
  zr invoice usage-rate-detail 2c92c0f8...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUsageRateDetail(cmd, f, args[0])
		},
	}
	return cmd
}

func runUsageRateDetail(cmd *cobra.Command, f *factory.Factory, itemID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/invoices/invoice-item/%s/usage-rate-detail", url.PathEscape(itemID)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" {
		return output.PrintJSON(f.IOStreams, resp.Body, fmtOpts.JQ)
	}
	if fmtOpts.Template != "" {
		return output.PrintTemplate(f.IOStreams, resp.Body, fmtOpts.Template)
	}
	return output.PrintJSON(f.IOStreams, resp.Body, "")
}
