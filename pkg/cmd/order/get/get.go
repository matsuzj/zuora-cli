// Package get implements the "zr order get" command.
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

// NewCmdGet creates the order get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <order-number>",
		Short: "Get order details",
		Long: `Get detailed information about a Zuora order.

Examples:
  zr order get O-00000001
  zr order get O-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/orders/%s", url.PathEscape(orderNumber)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// Zuora GET /v1/orders/{orderNumber} returns data nested under "order" key
	order, _ := raw["order"].(map[string]interface{})
	if order == nil {
		order = raw
	}

	fields := []output.DetailField{
		{Key: "Order Number", Value: getString(order, "orderNumber")},
		{Key: "Status", Value: getString(order, "status")},
		{Key: "Order Date", Value: getString(order, "orderDate")},
		{Key: "Account Number", Value: getString(order, "existingAccountNumber")},
		{Key: "Description", Value: getString(order, "description")},
		{Key: "Created Date", Value: getString(order, "createdDate")},
		{Key: "Created By", Value: getString(order, "createdBy")},
		{Key: "Updated Date", Value: getString(order, "updatedDate")},
		{Key: "Updated By", Value: getString(order, "updatedBy")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
