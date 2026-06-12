// Package get implements the "zr order-line-item get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the order-line-item get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <item-id>",
		Short: "Get order line item details",
		Long:  `Get detailed information about a Zuora order line item.`,
		Example: `  zr order-line-item get 2c92c0f9876...
  zr order-line-item get 2c92c0f9876... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, itemID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/order-line-items/%s", url.PathEscape(itemID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Item Name", Value: cmdutil.GetString(raw, "itemName")},
				{Key: "Item Number", Value: cmdutil.GetString(raw, "itemNumber")},
				{Key: "Item Type", Value: cmdutil.GetString(raw, "itemType")},
				{Key: "Item State", Value: cmdutil.GetString(raw, "itemState")},
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Amount", Value: cmdutil.GetString(raw, "amount")},
				{Key: "Quantity", Value: cmdutil.GetString(raw, "quantity")},
			}
		},
	})
}
