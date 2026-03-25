// Package get implements the "zr order-line-item get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the order-line-item get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <item-id>",
		Short: "Get order line item details",
		Long: `Get detailed information about a Zuora order line item.

Examples:
  zr order-line-item get 2c92c0f9876...
  zr order-line-item get 2c92c0f9876... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, itemID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/order-line-items/%s", url.PathEscape(itemID)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "id")},
		{Key: "Item Name", Value: getString(raw, "itemName")},
		{Key: "Item Number", Value: getString(raw, "itemNumber")},
		{Key: "Item Type", Value: getString(raw, "itemType")},
		{Key: "Item State", Value: getString(raw, "itemState")},
		{Key: "Order Number", Value: getString(raw, "orderNumber")},
		{Key: "Amount", Value: getString(raw, "amount")},
		{Key: "Quantity", Value: getString(raw, "quantity")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
