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
			// GET /v1/order-line-items/{id} nests the item under
			// "orderLineItem" (same wrapper class as fulfillment/ramp get,
			// fixed in #56); fall back to the top level so an unwrapped
			// response still renders.
			oli, _ := raw["orderLineItem"].(map[string]interface{})
			if oli == nil {
				oli = raw
			}
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(oli, "id")},
				{Key: "Item Name", Value: cmdutil.GetString(oli, "itemName")},
				{Key: "Item Number", Value: cmdutil.GetString(oli, "itemNumber")},
				{Key: "Item Type", Value: cmdutil.GetString(oli, "itemType")},
				{Key: "Item State", Value: cmdutil.GetString(oli, "itemState")},
				{Key: "Order Number", Value: cmdutil.GetString(oli, "orderNumber")},
				{Key: "Amount", Value: cmdutil.GetMoney(oli, "amount")},
				{Key: "Quantity", Value: cmdutil.GetDecimal(oli, "quantity")},
			}
		},
	})
}
