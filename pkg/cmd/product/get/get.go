// Package get implements the "zr product get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the product get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <product-key>",
		Short: "Get a commerce product",
		Long: `Get detailed information about a Zuora commerce product.

Examples:
  zr product get PROD-001
  zr product get PROD-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, productKey string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/commerce/products/%s", url.PathEscape(productKey)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Name", Value: cmdutil.GetString(raw, "name")},
				{Key: "SKU", Value: cmdutil.GetString(raw, "sku")},
				{Key: "Description", Value: cmdutil.GetString(raw, "description")},
				{Key: "Start Date", Value: cmdutil.GetString(raw, "start_date")},
				{Key: "End Date", Value: cmdutil.GetString(raw, "end_date")},
			}
		},
	})
}
