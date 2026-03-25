// Package get implements the "zr order get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

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

	resp, err := client.Get(fmt.Sprintf("/v1/orders/%s", url.PathEscape(orderNumber)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Order Number", Value: getString(raw, "orderNumber")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Order Date", Value: getString(raw, "orderDate")},
		{Key: "Account Number", Value: getString(raw, "existingAccountNumber")},
		{Key: "Description", Value: getString(raw, "description")},
		{Key: "Created Date", Value: getString(raw, "createdDate")},
		{Key: "Created By", Value: getString(raw, "createdBy")},
		{Key: "Updated Date", Value: getString(raw, "updatedDate")},
		{Key: "Updated By", Value: getString(raw, "updatedBy")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
